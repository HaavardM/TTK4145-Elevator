package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/utilities"

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
}

//Config contains scheduler configuration variables
type Config struct {
	ElevatorID         int
	NumFloors          int
	FolderPath         string
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

	//Stores last sent order to avoid sending duplicates
	var lastOrder *SchedulableOrder

	orderTimeout := time.NewTicker(20 * time.Second)

	//Channel used to avoid select blocking when neccessary
	skipSelect := make(chan struct{}, 1)

	//To avoid deadlocking, we do not want to block the main scheduler thread.
	//The runSkipOldOrders acts as a "middleman", storing the latest order and sends it when
	//the elevatorcontroller is ready.
	orderToElevator := make(chan common.Order)
	go runSendLatestOrder(ctx, conf.ElevExecuteOrder, orderToElevator)

	//Load orders if file exists
	if fileExists(conf.FolderPath) {
		fileOrders, err := readFromOrdersFile(conf.FolderPath)
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

	go sendOrderCosts(conf.CostsSend, workers[conf.ElevatorID])

	for {
		//All blocking operations handled in select!
		select {
		case <-ctx.Done():
			//Delete orders file on clean exit
			deleteOrdersFile(conf.FolderPath)
			return
		case <-skipSelect:
			//Continue after select
		case id := <-conf.WorkerLost:
			delete(workers, id)
			reassignInvalidOrders(ctx, &orders, time.Minute, workers, conf.NewOrderSend)
		case <-orderTimeout.C:
			reassignInvalidOrders(ctx, &orders, time.Minute, workers, conf.NewOrderSend)
		case elevatorStatus := <-conf.ElevStatus:
			if cost, ok := workers[conf.ElevatorID]; ok {
				updateElevatorCost(conf, cost, elevatorStatus)
				//Send cost using deep copy
				//Not critical if multiple of these are sent in wrong order
				go sendOrderCosts(conf.CostsSend, cost)
			} else {
				log.Panicf("Missing elevator cost in costmap")
			}
		case costs := <-conf.CostsRecv:
			workers[costs.ID] = &costs
		case order := <-conf.NewOrderRecv:
			handleNewOrder(&orders, order)
		case order := <-conf.OrderCompletedRecv:
			handleOrderCompleted(&orders, order)
		case order := <-conf.ElevCompletedOrder:
			log.Printf("Elev completed order %v\n", order)
			orders.OrdersCab[order.Floor] = nil
			switch order.Dir {
			case common.DownDir:
				schedOrder := orders.OrdersDown[order.Floor]
				if schedOrder != nil {
					go utilities.SendMessage(ctx, conf.OrderCompletedSend, *schedOrder)
				} else {
					log.Println("Unexpected order completed")
				}
			case common.UpDir, common.NoDir:
				schedOrder := orders.OrdersUp[order.Floor]
				if schedOrder != nil {
					go utilities.SendMessage(ctx, conf.OrderCompletedSend, *schedOrder)
				} else {
					log.Println("Unexpected order completed")
				}
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

		//Save orders to file
		err := saveToOrdersFile(conf.FolderPath, &orders)
		if err != nil {
			log.Panic(err)
		}

		//Set order lights
		for floor, order := range orders.OrdersUp {
			//No up light in top floor
			if floor >= conf.NumFloors {
				continue
			}
			//Receiver never blocks
			conf.Lights <- elevatordriver.LightState{Floor: floor, Type: elevatordriver.UpButtonLight, State: (order != nil)}
		}
		for floor, order := range orders.OrdersDown {
			//No down light in base floor
			if floor <= 0 {
				continue
			}
			//Receiver never blocks
			conf.Lights <- elevatordriver.LightState{Floor: floor, Type: elevatordriver.DownButtonLight, State: (order != nil)}
		}
		for floor, order := range orders.OrdersCab {
			//Receiver never blocks
			conf.Lights <- elevatordriver.LightState{Floor: floor, Type: elevatordriver.InternalButtonLight, State: (order != nil)}
		}

		//Find next order and send to elevatorcontroller
		order := findHighestPriority(&orders, workers[conf.ElevatorID], conf.ElevatorID)
		if order != nil && order != lastOrder {
			lastOrder = order
			//Guranteed to not block by the receiver runSkipOldOrders
			orderToElevator <- order.Order
		}
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

		//If assignee (elevator id) does not exist
		if _, ok := workers[order.Worker]; !ok {
			renewOrder = true
		}

		if renewOrder {
			worker := selectWorker(workers, order.Floor, order.Dir)
			newOrder := createOrder(order.Floor, order.Dir, worker)
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
			go utilities.SendMessage(ctx, send, *order)
		}
	}
}

func handleElevHallBtnPressed(ctx context.Context, btn elevio.ButtonEvent, costMap map[int]*common.OrderCosts, sendOrder chan<- SchedulableOrder) {
	switch btn.Button {
	case elevio.BT_HallDown:
		worker := selectWorker(costMap, btn.Floor, common.DownDir)
		order := createOrder(btn.Floor, common.DownDir, worker)
		go utilities.SendMessage(ctx, sendOrder, *order)
	case elevio.BT_HallUp:
		worker := selectWorker(costMap, btn.Floor, common.UpDir)
		order := createOrder(btn.Floor, common.UpDir, worker)
		go utilities.SendMessage(ctx, sendOrder, *order)
	default:
		log.Panic("Invalid button type")
	}
}

func sendOrderCosts(c chan<- common.OrderCosts, costs *common.OrderCosts) {
	//DeepCopy slices
	msg := common.OrderCosts{
		ID:        costs.ID,
		CostsDown: append(make([]float64, 0, len(costs.CostsDown)), costs.CostsDown...),
		CostsUp:   append(make([]float64, 0, len(costs.CostsUp)), costs.CostsUp...),
	}
	c <- msg
}
func selectWorker(workers map[int]*common.OrderCosts, floor int, dir common.Direction) int {
	minCost := math.Inf(1)
	worker := -1
	for k, v := range workers {
		switch dir {
		case common.UpDir:
			if cost := v.CostsUp[floor]; cost < minCost {
				minCost = cost
				worker = k
			}
		case common.DownDir:
			if cost := v.CostsDown[floor]; cost < minCost {
				minCost = cost
				worker = k
			}
		default:
			log.Panicf("Invalid direction %s\n", dir)
		}
	}
	return worker
}

func findHighestPriority(orders *schedOrders, cost *common.OrderCosts, id int) *SchedulableOrder {
	currMinCost := math.Inf(1)
	var currOrder *SchedulableOrder
	//Check cab calls
	for _, order := range orders.OrdersCab {
		if order == nil || order.Worker != id {
			continue
		}
		orderCost := cost.CostsCab[order.Floor]
		if orderCost < currMinCost {
			currMinCost = orderCost
			currOrder = order
		}
	}

	//Check down hall orders
	for _, order := range orders.OrdersDown {
		if order == nil || order.Worker != id {
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
		if order == nil || order.Worker != id {
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

func createOrder(floor int, dir common.Direction, assignee int) *SchedulableOrder {
	return &SchedulableOrder{
		Order: common.Order{
			Floor: floor,
			Dir:   dir,
		},
		Worker:    assignee,
		Timestamp: time.Now(),
	}
}

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

func handleOrderCompleted(orders *schedOrders, order SchedulableOrder) {
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

func tryAddOrderToSlice(slice []*SchedulableOrder, pos int, order SchedulableOrder) error {
	if pos >= 0 && pos < len(slice) {
		slice[pos] = &order
		return nil
	}
	return errors.New("Invalid index")
}

func tryRemoveOrderFromSlice(slice []*SchedulableOrder, pos int) error {
	if pos >= 0 && pos < len(slice) {
		slice[pos] = nil
		return nil
	}
	return errors.New("Invalid index")
}
