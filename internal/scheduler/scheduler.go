package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"reflect"
	"sync"
	"time"

	"github.com/rs/xid"

	"github.com/TTK4145-students-2019/project-thefuturezebras/pkg/utilities"

	"github.com/davecgh/go-spew/spew"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/common"
	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/elevatordriver"
	"github.com/TTK4145/driver-go/elevio"
)

//SchedulableOrder is an order with a priority and cost
type SchedulableOrder struct {
	common.Order `json:"order"`
	Worker       int       `json:"assignee"`
	Timestamp    time.Time `json:"timestamp"`
	OrderID      string    `json:"order_id"`
	completed    *time.Time
}

//Config contains scheduler configuration variables
type Config struct {
	ElevatorID         int
	NumFloors          int
	FilePath           string
	ElevButtonPressed  <-chan elevio.ButtonEvent
	ElevCompletedOrder <-chan common.Order
	ElevExecuteOrder   chan<- common.Order
	ElevStatus         <-chan common.ElevatorStatus
	//Sets light state - assumed non-blocking
	Lights             chan<- elevatordriver.LightState
	NewOrderSend       chan<- SchedulableOrder
	NewOrderRecv       <-chan SchedulableOrder
	OrderCompletedSend chan<- SchedulableOrder
	OrderCompletedRecv <-chan SchedulableOrder
	CostsSend          chan<- common.OrderCosts
	CostsRecv          <-chan common.OrderCosts
	WorkerLost         <-chan int
}

//Struct containing orders in the different directions
type schedOrders struct {
	OrdersUp   []*SchedulableOrder `json:"orders_up"`
	OrdersDown []*SchedulableOrder `json:"orders_down"`
	OrdersCab  []*SchedulableOrder `json:"orders_cab"`
}

//If for some reason the scheduler generates orders faster than the elevatorcontroller
//we want to only send the latest one when the channel is ready.
//Sending the message using multiple goroutines wouldn't help since the order of the messages is important
func runSendLatestOrder(ctx context.Context, sendChan chan<- common.Order, orderToSend <-chan common.Order) {
	for {
		select {
		case <-ctx.Done():
			return
		//Wait until there is something to send
		case order := <-orderToSend:
			//Try send the order - or overwrite the order with a new one if received
			select {
			//Abort if context finished
			case <-ctx.Done():
				return
			//Send order if channel is ready
			case sendChan <- order:
			//Update order to send if a new one is received before the sendChan is ready
			case order = <-orderToSend:
			}
		}
	}

}

//Run is the startingpoint for the scheduler module
//The ctx context is used to stop the gorotine if the context expires.
func Run(ctx context.Context, waitGroup *sync.WaitGroup, conf Config) {
	//Used to make sure main routine waits for this goroutine to finish
	defer waitGroup.Done()

	//Contains orders for all floors and directions
	orders := schedOrders{
		OrdersUp:   make([]*SchedulableOrder, conf.NumFloors),
		OrdersDown: make([]*SchedulableOrder, conf.NumFloors),
		OrdersCab:  make([]*SchedulableOrder, conf.NumFloors),
	}
	//Contains the cost for orders to all floor by all elevators
	workers := map[int]*common.OrderCosts{
		conf.ElevatorID: &common.OrderCosts{
			ID:        conf.ElevatorID,
			CostsUp:   make([]float64, conf.NumFloors),
			CostsDown: make([]float64, conf.NumFloors),
			CostsCab:  make([]float64, conf.NumFloors),
		},
	}

	orderTimeout := 20 * time.Second
	orderTimeoutTicker := time.NewTicker(orderTimeout)
	//Channel used to avoid select blocking when neccessary
	skipSelect := make(chan struct{}, 1)

	//To avoid deadlocking, we do not want to block the main scheduler thread.
	//The runSkipOldOrders acts as a "middleman", storing the latest order and sends it when
	//the elevatorcontroller is ready.
	orderToElevator := make(chan common.Order)
	go runSendLatestOrder(ctx, conf.ElevExecuteOrder, orderToElevator)

	//Load orders if file exists
	if fileExists(conf.FilePath) {
		fileOrders, err := readFromOrdersFile(conf.FilePath)
		if err != nil {
			log.Panicf("Error reading from file: %s\n", err)
		}
		log.Printf("Loading orders from file: %+v\n", fileOrders)
		spew.Dump(fileOrders)
		publishAllHallOrders(ctx, fileOrders, conf.NewOrderSend)
		//Replace orders with orders from file
		orders = *fileOrders

		//Skip select to reload orders
		skipSelect <- struct{}{}
	}

	var prevOrder SchedulableOrder
	var elevatorStatus common.ElevatorStatus

	for {
		//All blocking operations handled in select!
		select {
		case <-ctx.Done():
			//Delete orders file on clean exit
			deleteOrdersFile(conf.FilePath)
			return
		case <-skipSelect:
			//Continue after select
		case id := <-conf.WorkerLost:
			delete(workers, id)
			reassignInvalidOrders(ctx, &orders, orderTimeout, workers, conf.NewOrderSend)
		case <-orderTimeoutTicker.C:
			reassignInvalidOrders(ctx, &orders, orderTimeout, workers, conf.NewOrderSend)
		case elevatorStatus = <-conf.ElevStatus:
			//Updates elevator stauts
		case costs := <-conf.CostsRecv:
			//If a new elevator connects - share all orders
			if _, ok := workers[costs.ID]; !ok {
				publishAllHallOrders(ctx, &orders, conf.NewOrderSend)
			}
			workers[costs.ID] = &costs
		case order := <-conf.NewOrderRecv:
			handleNewOrder(&orders, order)
		case order := <-conf.OrderCompletedRecv:
			handleOrderCompleted(&orders, order, conf)
		case order := <-conf.ElevCompletedOrder:
			log.Printf("Elev completed order %v\n", order)
			//Save previous order
			orders.OrdersCab[order.Floor] = nil
			completedTime := time.Now()
			switch order.Dir {
			case common.DownDir:
				schedOrder := orders.OrdersDown[order.Floor]
				if schedOrder != nil {
					//Send order completed event to network when available
					go utilities.SendMessage(ctx, conf.OrderCompletedSend, *schedOrder)
					//Completed but not (yet) acked by network. Do not send it to the elevator again
					schedOrder.completed = &completedTime
				} else {
					log.Println("Unexpected order completed")
				}
			case common.UpDir:
				schedOrder := orders.OrdersUp[order.Floor]
				if schedOrder != nil {
					//Send order completed event to network when available
					go utilities.SendMessage(ctx, conf.OrderCompletedSend, *schedOrder)
					//Completed but not (yet) acked by network. Do not send it to the elevator again
					schedOrder.completed = &completedTime
				} else {
					log.Println("Unexpected order completed")
				}
			case common.NoDir:
				//Already handled before switch
			default:
				log.Panicln("Unexpected direction")
			}

		case btn := <-conf.ElevButtonPressed:
			log.Printf("Button pressed %v\n", btn)
			if btn.Button == elevio.BT_Cab {
				if orders.OrdersCab[btn.Floor] == nil {
					orders.OrdersCab[btn.Floor] = createOrder(btn.Floor, common.NoDir, conf.ElevatorID)
				}
			} else {
				handleElevHallBtnPressed(ctx, btn, workers, conf.NewOrderSend)
			}
		}

		if cost, ok := workers[conf.ElevatorID]; ok {
			updateElevatorCost(cost, elevatorStatus, &orders, conf.ElevatorID)
			//Send cost using deep copy
			//Not critical if multiple of these are sent in wrong order
			go sendOrderCosts(conf.CostsSend, cost)
		} else {
			log.Panicf("Missing elevator cost in costmap")
		}
		//Save orders to file
		err := saveToOrdersFile(conf.FilePath, &orders)
		if err != nil {
			log.Panic(err)
		}

		//Update status lights based on updated orders
		setLightsFromOrders(orders, conf.Lights, conf.NumFloors)

		//Find next order and send to elevatorcontroller
		order := findHighestPriority(&orders, workers[conf.ElevatorID], conf.ElevatorID)
		//Only send new order if not deeply equal to the last one and not nil
		if order != nil && !reflect.DeepEqual(*order, prevOrder) {
			//Guranteed to not block by the receiver runSkipOldOrders
			orderToElevator <- order.Order
			prevOrder = *order
		}
	}
}

func setLightsFromOrders(orders schedOrders, lights chan<- elevatordriver.LightState, numFloors int) {
	//Set order lights
	for floor, order := range orders.OrdersUp {
		//No up light in top floor
		if floor >= numFloors {
			continue
		}
		//Receiver never blocks
		lights <- elevatordriver.LightState{Floor: floor, Type: elevatordriver.UpButtonLight, State: (order != nil)}
	}
	for floor, order := range orders.OrdersDown {
		//No down light in base floor
		if floor <= 0 {
			continue
		}
		//Receiver never blocks
		lights <- elevatordriver.LightState{Floor: floor, Type: elevatordriver.DownButtonLight, State: (order != nil)}
	}
	for floor, order := range orders.OrdersCab {
		//Receiver never blocks
		lights <- elevatordriver.LightState{Floor: floor, Type: elevatordriver.InternalButtonLight, State: (order != nil)}
	}

}

func reassignInvalidOrders(ctx context.Context, orders *schedOrders, timeout time.Duration, workers map[int]*common.OrderCosts, sendOrder chan<- SchedulableOrder) {
	hallOrders := make([]*SchedulableOrder, 0, len(orders.OrdersDown)+len(orders.OrdersUp))
	hallOrders = append(hallOrders, orders.OrdersDown...)
	hallOrders = append(hallOrders, orders.OrdersUp...)
	//Check for timeout or invalid assignee
	for _, order := range hallOrders {
		renewOrder := false
		if order == nil {
			continue
		}

		//Check if timeout have passed
		if time.Now().Sub(order.Timestamp) > timeout {
			renewOrder = true
		}

		//If order is completed and has been for some time
		if order.completed != nil && time.Now().Sub(*order.completed) > timeout {
			renewOrder = true
		}

		//If assignee (elevator id) does not exist
		if _, ok := workers[order.Worker]; !ok {
			renewOrder = true
		}

		if renewOrder {
			worker := selectWorker(workers, order.Floor, order.Dir)
			newOrder := createOrder(order.Floor, order.Dir, worker)
			//Send new order event to network when available
			go utilities.SendMessage(ctx, sendOrder, *newOrder)
		}
	}

	for _, order := range orders.OrdersCab {
		if order == nil {
			continue
		}

		if time.Now().Sub(order.Timestamp) > timeout {
			log.Printf("Order %v timeout\n", order)
			order.Timestamp = time.Now()
		}
	}
}

func publishAllHallOrders(ctx context.Context, orders *schedOrders, send chan<- SchedulableOrder) {
	//Get hall orders
	hallOrders := make([]*SchedulableOrder, 0, len(orders.OrdersDown)+len(orders.OrdersUp))
	hallOrders = append(hallOrders, orders.OrdersDown...)
	hallOrders = append(hallOrders, orders.OrdersUp...)

	for _, order := range hallOrders {
		if order != nil {
			//Send order to network when available
			go utilities.SendMessage(ctx, send, *order)
		}
	}
}

func handleElevHallBtnPressed(ctx context.Context, btn elevio.ButtonEvent, costMap map[int]*common.OrderCosts, sendOrder chan<- SchedulableOrder) {
	switch btn.Button {
	case elevio.BT_HallDown:
		worker := selectWorker(costMap, btn.Floor, common.DownDir)
		order := createOrder(btn.Floor, common.DownDir, worker)
		//Send new order to network when available
		go utilities.SendMessage(ctx, sendOrder, *order)
	case elevio.BT_HallUp:
		worker := selectWorker(costMap, btn.Floor, common.UpDir)
		order := createOrder(btn.Floor, common.UpDir, worker)
		//Send new order to network when available
		go utilities.SendMessage(ctx, sendOrder, *order)
	default:
		log.Panic("Invalid button type")
	}
}

//Sends the cost of specific orders for the elevators
func sendOrderCosts(c chan<- common.OrderCosts, costs *common.OrderCosts) {
	//DeepCopy slices
	msg := common.OrderCosts{
		ID:         costs.ID,
		OrderCount: costs.OrderCount,
		CostsDown:  append(make([]float64, 0, len(costs.CostsDown)), costs.CostsDown...),
		CostsUp:    append(make([]float64, 0, len(costs.CostsUp)), costs.CostsUp...),
		CostsCab:   append(make([]float64, 0, len(costs.CostsCab)), costs.CostsCab...),
	}
	c <- msg
}

//Selects an elevator based on which elevator is the cheapest for that specific order(direction and floor)
func selectWorker(workers map[int]*common.OrderCosts, floor int, dir common.Direction) int {
	minCost := math.Inf(1)
	worker := -1
	for _, v := range workers {
		switch dir {
		case common.NoDir:
			if cost := v.CostsCab[floor]; cost < minCost {
				worker = v.ID
				minCost = cost
			}
		case common.UpDir:
			if cost := v.CostsUp[floor]; cost < minCost {
				worker = v.ID
				minCost = cost
			}
		case common.DownDir:
			if cost := v.CostsDown[floor]; cost < minCost {
				worker = v.ID
				minCost = cost
			}
		default:
			log.Panicln("Unknown directio")
		}
	}
	return worker
}

//Finds the order with the highest priority
func findHighestPriority(orders *schedOrders, cost *common.OrderCosts, id int) *SchedulableOrder {
	currMinCost := math.Inf(1)
	var currOrder *SchedulableOrder
	//Check cab calls
	for _, order := range orders.OrdersCab {
		if order == nil || order.Worker != id || order.completed != nil {
			continue
		}
		orderCost := cost.CostsCab[order.Floor]
		if orderCost < currMinCost {
			currMinCost = orderCost
			currOrder = order
			log.Printf("Best yet (%.1f): %v\n", currMinCost, currOrder)
		}
	}

	//Check down hall orders
	for _, order := range orders.OrdersDown {
		if order == nil || order.Worker != id || order.completed != nil {
			continue
		}
		orderCost := cost.CostsDown[order.Floor]
		if orderCost < currMinCost {
			currMinCost = orderCost
			currOrder = order
		}
	}

	//Check up hall orders
	for _, order := range orders.OrdersUp {
		if order == nil || order.Worker != id || order.completed != nil {
			continue
		}
		orderCost := cost.CostsUp[order.Floor]
		if orderCost < currMinCost {
			currMinCost = orderCost
			currOrder = order
		}
	}
	return currOrder
}

//Creates and order marked with assigned elevator and a timestamp
func createOrder(floor int, dir common.Direction, assignee int) *SchedulableOrder {
	return &SchedulableOrder{
		Order: common.Order{
			Floor: floor,
			Dir:   dir,
		},
		Worker:    assignee,
		Timestamp: time.Now(),
		OrderID:   xid.New().String(),
	}
}

//
func handleNewOrder(orders *schedOrders, order SchedulableOrder) error {
	switch order.Dir {
	case common.UpDir:
		if err := tryAddOrderToSlice(orders.OrdersUp, order.Floor, order); err != nil {
			return err
		}
	case common.DownDir:
		if err := tryAddOrderToSlice(orders.OrdersDown, order.Floor, order); err != nil {
			return err
		}
	case common.NoDir:
		if err := tryAddOrderToSlice(orders.OrdersCab, order.Floor, order); err != nil {
			return err
		}
	default:
		return errors.New("Error adding order - unrecognized direction: " + fmt.Sprintf("%v", order))
	}
	return nil
}

//Handles upcoming events once notice of an order being finished comes in
func handleOrderCompleted(orders *schedOrders, order SchedulableOrder, conf Config) {
	switch order.Dir {
	case common.UpDir:
		if err := tryRemoveOrderFromSlice(orders.OrdersUp, order.Floor); err != nil {
			log.Println("Error removing order: ", err)
		}
	case common.DownDir:
		if err := tryRemoveOrderFromSlice(orders.OrdersDown, order.Floor); err != nil {
			log.Println("Error removing order: ", err)
		}

	}

}

//Adds an order to a slice of scheduled orders
func tryAddOrderToSlice(slice []*SchedulableOrder, pos int, order SchedulableOrder) error {
	if pos >= 0 && pos < len(slice) {
		slice[pos] = &order
		return nil
	}
	return errors.New("Invalid index")
}

//Removed orders from slice scheduled orders
func tryRemoveOrderFromSlice(slice []*SchedulableOrder, pos int) error {
	if pos >= 0 && pos < len(slice) {
		slice[pos] = nil
		return nil
	}
	return errors.New("Invalid index")
}
