package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/common"
	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/elevatordriver"
	"github.com/TTK4145/driver-go/elevio"
)

//SchedulableOrder is an order with a priority and cost
type SchedulableOrder struct {
	common.Order `json:"order"`
	Assignee     int       `json:"assignee"`
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
	Lights             chan<- elevatordriver.LightState
	NewOrderSend       chan<- SchedulableOrder
	NewOrderRecv       <-chan SchedulableOrder
	OrderCompletedSend chan<- SchedulableOrder
	OrderCompletedRecv <-chan SchedulableOrder
	CostsSend          chan<- common.OrderCosts
	CostsRecv          <-chan common.OrderCosts
}

type schedOrders struct {
	OrdersUp   []*SchedulableOrder `json:"orders_up"`
	OrdersDown []*SchedulableOrder `json:"orders_down"`
	OrdersCab  []*SchedulableOrder `json:"orders_cab"`
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
	costMap := map[int]*common.OrderCosts{
		conf.ElevatorID: &common.OrderCosts{
			ID:        conf.ElevatorID,
			CostsUp:   make([]float64, conf.NumFloors),
			CostsDown: make([]float64, conf.NumFloors),
			CostsCab:  make([]float64, conf.NumFloors),
		},
	}

	var lastOrder *SchedulableOrder = nil

	//Channel used to avoid select blocking when neccessary
	skip := make(chan struct{})

	//Load orders if file exists
	if fileExists(conf.FolderPath) {
		fileOrders, err := readFromOrdersFile(conf.FolderPath)
		if err != nil {
			log.Panicf("Error reading from file: %s\n", err)
		}
		log.Printf("Loading orders from file: %+v\n", fileOrders)
		spew.Dump(fileOrders)
		publishAllHallOrders(fileOrders, conf.NewOrderSend)
		orders.OrdersCab = fileOrders.OrdersCab
		go func() {
			skip <- struct{}{}
		}()
	}

	//Load system configuration
	for {
		select {
		case <-ctx.Done():
			//Delete orders file on clean exit
			deleteOrdersFile(conf.FolderPath)
			return
		case <-skip:
			//Continue after select
		case elevatorStatus := <-conf.ElevStatus:
			if cost, ok := costMap[conf.ElevatorID]; ok {
				updateElevatorCost(cost, elevatorStatus)
				//Send cost using deep copy
				go sendOrderCosts(conf.CostsSend, cost)
			} else {
				log.Panicf("Missing elevator cost in costmap")
			}
		case costs := <-conf.CostsRecv:
			costMap[costs.ID] = &costs
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
					conf.OrderCompletedSend <- *schedOrder
				} else {
					log.Println("Unexpected order completed")
				}
			case common.UpDir, common.NoDir:
				schedOrder := orders.OrdersUp[order.Floor]
				if schedOrder != nil {
					conf.OrderCompletedSend <- *schedOrder
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
				handleElevUpDownBtnPressed(btn, costMap, conf.NewOrderSend)
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
			conf.Lights <- elevatordriver.LightState{Floor: floor, Type: elevatordriver.UpButtonLight, State: (order != nil)}
		}
		for floor, order := range orders.OrdersDown {
			//No down light in base floor
			if floor <= 0 {
				continue
			}
			conf.Lights <- elevatordriver.LightState{Floor: floor, Type: elevatordriver.DownButtonLight, State: (order != nil)}
		}
		for floor, order := range orders.OrdersCab {
			conf.Lights <- elevatordriver.LightState{Floor: floor, Type: elevatordriver.InternalButtonLight, State: (order != nil)}
		}

		//Find next order and send to elevatorcontroller
		order := findHighestPriority(&orders, costMap[conf.ElevatorID], conf.ElevatorID)
		if order != nil && order != lastOrder {
			lastOrder = order
			conf.ElevExecuteOrder <- order.Order
		}
	}
}

func sendOrderToElev(elev chan<- common.Order, order common.Order) {
	elev <- order
}

func publishAllHallOrders(orders *schedOrders, send chan<- SchedulableOrder) {
	for _, order := range orders.OrdersDown {
		if order != nil {
			send <- *order
		}
	}
	for _, order := range orders.OrdersUp {
		if order != nil {
			send <- *order
		}
	}
}

func handleElevUpDownBtnPressed(btn elevio.ButtonEvent, costMap map[int]*common.OrderCosts, sendOrder chan<- SchedulableOrder) {
	switch btn.Button {
	case elevio.BT_HallDown:
		assignee := selectAssignee(costMap, btn.Floor, common.DownDir)
		order := createOrder(btn.Floor, common.DownDir, assignee)
		sendOrder <- *order
	case elevio.BT_HallUp:
		assignee := selectAssignee(costMap, btn.Floor, common.UpDir)
		order := createOrder(btn.Floor, common.UpDir, assignee)
		sendOrder <- *order
	default:
		log.Panic("Invalid button type")
	}
}

func updateElevatorCost(costs *common.OrderCosts, status common.ElevatorStatus) {

}

func sendOrderCosts(c chan<- common.OrderCosts, costs *common.OrderCosts) {
	//DeepCopy slices
	msg := common.OrderCosts{
		CostsDown: append(make([]float64, 0, len(costs.CostsDown)), costs.CostsDown...),
		CostsUp:   append(make([]float64, 0, len(costs.CostsUp)), costs.CostsUp...),
	}
	c <- msg
}
func selectAssignee(assignees map[int]*common.OrderCosts, floor int, dir common.Direction) int {
	minCost := math.Inf(1)
	assignee := -1
	for k, v := range assignees {
		switch dir {
		case common.UpDir:
			if cost := v.CostsUp[floor]; cost < minCost {
				minCost = cost
				assignee = k
			}
		case common.DownDir:
			if cost := v.CostsDown[floor]; cost < minCost {
				minCost = cost
				assignee = k
			}
		default:
			log.Panicf("Invalid direction %s\n", dir)
		}
	}
	return assignee
}

func findHighestPriority(orders *schedOrders, cost *common.OrderCosts, id int) *SchedulableOrder {
	currMinCost := math.Inf(1)
	var currOrder *SchedulableOrder
	for _, order := range orders.OrdersCab {
		if order == nil || order.Assignee != id {
			continue
		}
		orderCost := cost.CostsCab[order.Floor]
		if orderCost < currMinCost {
			currMinCost = orderCost
			currOrder = order
		}
	}
	for _, order := range orders.OrdersDown {
		if order == nil || order.Assignee != id {
			continue
		}
		orderCost := cost.CostsDown[order.Floor]
		if orderCost < currMinCost {
			currMinCost = orderCost
			currOrder = order
		}
	}
	for _, order := range orders.OrdersUp {
		if order == nil || order.Assignee != id {
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
		Assignee:  assignee,
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
		if slice[pos] == nil {
			slice[pos] = &order
			return nil
		}
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
