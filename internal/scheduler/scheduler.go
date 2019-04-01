package scheduler

import (
	"errors"
	"fmt"
	"log"
	"math"
	"reflect"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/rs/xid"

	"github.com/davecgh/go-spew/spew"

	"github.com/TTK4145-students-2019/project-thefuturezebras/pkg/utilities"

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
	HallUp   []*SchedulableOrder `json:"orders_up"`
	HallDown []*SchedulableOrder `json:"orders_down"`
	Cab      []*SchedulableOrder `json:"orders_cab"`
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
		HallUp:   make([]*SchedulableOrder, conf.NumFloors),
		HallDown: make([]*SchedulableOrder, conf.NumFloors),
		Cab:      make([]*SchedulableOrder, conf.NumFloors),
	}
	//Contains the cost for orders to all floor by all elevators
	workers := map[int]*common.OrderCosts{
		conf.ElevatorID: &common.OrderCosts{
			ID:       conf.ElevatorID,
			HallUp:   make([]float64, len(orders.HallUp)),
			HallDown: make([]float64, len(orders.HallDown)),
			Cab:      make([]float64, len(orders.Cab)),
		},
	}

	orderTimeout := 20 * time.Second
	orderTimeoutTicker := time.NewTicker(time.Second)
	//Channel used to avoid select blocking when neccessary
	skipSelect := make(chan struct{}, 1)

	//To avoid deadlocking, we do not want to block the main scheduler thread.
	//The runSkipOldOrders acts as a "middleman", storing the latest order and sends it when
	//the elevatorcontroller is ready
	orderToElevator := make(chan common.Order)
	go runSendLatestOrder(ctx, conf.ElevExecuteOrder, orderToElevator)

	//Load orders if file exists
	if fileExists(conf.FilePath) {
		fileOrders, err := readFromOrdersFile(conf.FilePath)
		if err != nil {
			log.Panicf("Error reading from file: %s\n", err)
		}
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
			//Save previous order
			orders.Cab[order.Floor] = nil
			completedTime := time.Now()
			switch order.Dir {
			case common.DownDir:
				schedOrder := orders.HallDown[order.Floor]
				if schedOrder != nil {
					//Send order completed event to network when available
					go utilities.SendMessage(ctx, conf.OrderCompletedSend, *schedOrder)
					//Completed but not (yet) acked by network. Do not send it to the elevator again
					schedOrder.completed = &completedTime
				} else {
					log.Println("Unexpected order completed")
				}
			case common.UpDir:
				schedOrder := orders.HallUp[order.Floor]
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
			if btn.Button == elevio.BT_Cab {
				if orders.Cab[btn.Floor] == nil {
					orders.Cab[btn.Floor] = createOrder(btn.Floor, common.NoDir, conf.ElevatorID)
				}
			} else {
				handleElevHallBtnPressed(ctx, btn, workers, conf.NewOrderSend)
			}
		}

		//Update elevators cost
		if cost, ok := workers[conf.ElevatorID]; ok {
			newCost := createElevatorCost(elevatorStatus, &orders, conf.ElevatorID)
			if !reflect.DeepEqual(*cost, newCost) {
				*cost = newCost
				//Send cost using deep copy
				//Not critical if multiple of these are sent in wrong order
				go sendOrderCosts(conf.CostsSend, cost)
			}
		} else {
			log.Panicf("Missing elevator cost in costmap")
		}

		//Save orders to file
		err := saveToOrdersFile(conf.FilePath, &orders)
		if err != nil {
			log.Panic(err)
		} else {
			//Lights is only set if the order is saved to file without issues.
			//Update status lights based on updated orders
			setLightsFromOrders(orders, conf.Lights, conf.NumFloors)

		}

		//Find next order and send to elevatorcontroller
		order := getCheapestActiveOrder(&orders, workers[conf.ElevatorID], conf.ElevatorID)
		//Only send new order if not deeply equal to the last one and not nil
		if order != nil && !reflect.DeepEqual(*order, prevOrder) {
			//Guranteed to not block by the receiver runSkipOldOrders
			orderToElevator <- order.Order
			prevOrder = *order
		}
	}
}

//setLightsFromOrders sets the order lights when the order is confirmed and saved to file
func setLightsFromOrders(orders schedOrders, lights chan<- elevatordriver.LightState, numFloors int) {
	//Set order lights
	for floor, order := range orders.HallUp {
		//No up light in top floor
		if floor >= numFloors {
			continue
		}
		//Receiver never blocks
		lights <- elevatordriver.LightState{Floor: floor, Type: elevatordriver.UpButtonLight, State: (order != nil)}
	}
	for floor, order := range orders.HallDown {
		//No down light in base floor
		if floor <= 0 {
			continue
		}
		//Receiver never blocks
		lights <- elevatordriver.LightState{Floor: floor, Type: elevatordriver.DownButtonLight, State: (order != nil)}
	}
	for floor, order := range orders.Cab {
		//Receiver never blocks
		lights <- elevatordriver.LightState{Floor: floor, Type: elevatordriver.InternalButtonLight, State: (order != nil)}
	}

}

//Reassigns orders that have timed out as if it was a new order
func reassignInvalidOrders(ctx context.Context, orders *schedOrders, timeout time.Duration, workers map[int]*common.OrderCosts, sendOrder chan<- SchedulableOrder) {
	hallOrders := make([]*SchedulableOrder, 0, len(orders.HallDown)+len(orders.HallUp))
	hallOrders = append(hallOrders, orders.HallDown...)
	hallOrders = append(hallOrders, orders.HallUp...)
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
			log.Printf("Renewing order %+v\n ", newOrder)
		}
	}

	for _, order := range orders.Cab {
		if order == nil {
			continue
		}

		if time.Now().Sub(order.Timestamp) > timeout {
			order.Timestamp = time.Now()
		}
	}
}

//Sends all current hall orders on the network
func publishAllHallOrders(ctx context.Context, orders *schedOrders, send chan<- SchedulableOrder) {
	//Get hall orders
	hallOrders := make([]*SchedulableOrder, 0, len(orders.HallDown)+len(orders.HallUp))
	hallOrders = append(hallOrders, orders.HallDown...)
	hallOrders = append(hallOrders, orders.HallUp...)

	for _, order := range hallOrders {
		if order != nil {
			//Send order to network when available
			go utilities.SendMessage(ctx, send, *order)
		}
	}
}

//Handles events related to a hall button pressed such as creating an order and assigning an elevator
func handleElevHallBtnPressed(ctx context.Context, btn elevio.ButtonEvent, costMap map[int]*common.OrderCosts, sendOrder chan<- SchedulableOrder) {
	switch btn.Button {
	case elevio.BT_HallDown:
		worker := selectWorker(costMap, btn.Floor, common.DownDir)
		order := createOrder(btn.Floor, common.DownDir, worker)
		//Send new order to network when available
		go utilities.SendMessage(ctx, sendOrder, *order)
		log.Println("New HallDown order assigned to ", worker)
	case elevio.BT_HallUp:
		worker := selectWorker(costMap, btn.Floor, common.UpDir)
		order := createOrder(btn.Floor, common.UpDir, worker)
		//Send new order to network when available
		go utilities.SendMessage(ctx, sendOrder, *order)
		log.Println("New HallUp order assigned to ", worker)
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
		HallDown:   append(make([]float64, 0, len(costs.HallDown)), costs.HallDown...),
		HallUp:     append(make([]float64, 0, len(costs.HallUp)), costs.HallUp...),
		Cab:        append(make([]float64, 0, len(costs.Cab)), costs.Cab...),
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
			if cost := v.Cab[floor]; cost < minCost {
				worker = v.ID
				minCost = cost
			}
		case common.UpDir:
			if cost := v.HallUp[floor]; cost < minCost {
				worker = v.ID
				minCost = cost
			}
		case common.DownDir:
			if cost := v.HallDown[floor]; cost < minCost {
				worker = v.ID
				minCost = cost
			}
		default:
			log.Panicln("Unknown direction")
		}
	}
	return worker
}

//Chooses the cheapest order as the next order to be executed
func getCheapestActiveOrder(orders *schedOrders, cost *common.OrderCosts, id int) *SchedulableOrder {
	currMinCost := math.Inf(1)
	var currOrder *SchedulableOrder

	//Check down hall orders
	for _, order := range orders.HallDown {
		if order == nil || order.Worker != id || order.completed != nil {
			continue
		}
		orderCost := cost.HallDown[order.Floor]
		if orderCost < currMinCost {
			currMinCost = orderCost
			currOrder = order
		}
	}

	//Check up hall orders
	for _, order := range orders.HallUp {
		if order == nil || order.Worker != id || order.completed != nil {
			continue
		}
		orderCost := cost.HallUp[order.Floor]
		if orderCost < currMinCost {
			currMinCost = orderCost
			currOrder = order
		}
	}

	//Check cab calls
	for _, order := range orders.Cab {
		if order == nil || order.Worker != id || order.completed != nil {
			continue
		}
		orderCost := cost.Cab[order.Floor]
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

//Handles orders that have already been created and tries to add them to a slice of same types of orders
func handleNewOrder(orders *schedOrders, order SchedulableOrder) error {
	switch order.Dir {
	case common.UpDir:
		if err := tryAddOrderToSlice(orders.HallUp, order.Floor, order); err != nil {
			return err
		}
	case common.DownDir:
		if err := tryAddOrderToSlice(orders.HallDown, order.Floor, order); err != nil {
			return err
		}
	case common.NoDir:
		if err := tryAddOrderToSlice(orders.Cab, order.Floor, order); err != nil {
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
		if err := tryRemoveOrderFromSlice(orders.HallUp, order.Floor); err != nil {
			log.Println("Error removing order: ", err)
		}
	case common.DownDir:
		if err := tryRemoveOrderFromSlice(orders.HallDown, order.Floor); err != nil {
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
