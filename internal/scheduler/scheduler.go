package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/configuration"

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
	ButtonPressed      <-chan elevio.ButtonEvent
	OrderCompleted     <-chan common.Order
	ExecuteOrder       chan<- common.Order
	Lights             chan<- elevatordriver.LightState
	NewOrderSend       chan<- SchedulableOrder
	NewOrderRecv       <-chan SchedulableOrder
	OrderCompletedSend chan<- SchedulableOrder
	OrderCompletedRecv <-chan SchedulableOrder
	CostsSend          chan<- common.OrderCosts
	CostsRecv          <-chan common.OrderCosts

	NumFloors int
}

type schedOrders struct {
	OrdersUp   []*SchedulableOrder `json:"orders_up"`
	OrdersDown []*SchedulableOrder `json:"orders_down"`
	OrdersCab  []*SchedulableOrder `json:"orders_cab"`
}

//Run is the startingpoint for the scheduler module
//The ctx context is used to stop the gorotine if the context expires.
func Run(ctx context.Context, conf Config) {
	//Contains orders for all floors and directions
	orders := schedOrders{
		OrdersUp:   make([]*SchedulableOrder, conf.NumFloors-1),
		OrdersDown: make([]*SchedulableOrder, conf.NumFloors-1),
		OrdersCab:  make([]*SchedulableOrder, conf.NumFloors),
	}
	elevatorCosts := map[int]common.OrderCosts{
		conf.ElevatorID: common.OrderCosts{
			CostsUp:   make([]float64, conf.NumFloors-1),
			CostsDown: make([]float64, conf.NumFloors-1),
		},
	}
	timer := time.NewTicker(time.Second)
	sysConf := configuration.GetConfig()

	for {
		select {
		case <-ctx.Done():
			//Delete orders file on clean exit
			deleteOrdersFile(sysConf.FolderPath)
			return

		case <-timer.C:
			if costs, ok := elevatorCosts[conf.ElevatorID]; ok {
				sendOrderCosts(conf.CostsSend, &costs)
			} else {
				log.Panicln("This elevators id not in cost map")
			}
		case costs := <-conf.CostsRecv:
			elevatorCosts[costs.ID] = costs
		case order := <-conf.NewOrderRecv:
			handleNewOrder(&orders, order)
		case order := <-conf.OrderCompletedRecv:
			switch order.Dir {
			case common.UpDir:
				if err := tryRemoveOrderFromSlice(orders.OrdersUp, order.Floor); err != nil {
					log.Println("Error removing order: ", err)
				}
			case common.DownDir:
				if err := tryRemoveOrderFromSlice(orders.OrdersDown, order.Floor-1); err != nil {
					log.Println("Error removing order: ", err)
				}
			case common.NoDir:
				if err := tryRemoveOrderFromSlice(orders.OrdersCab, order.Floor); err != nil {
					log.Println("Error removing order: ", err)
				}
			}
		case order := <-conf.OrderCompleted:
			switch order.Dir {
			case common.NoDir:
				//Clear orders
				orders.OrdersCab[order.Floor] = nil
			case common.DownDir:
				schedOrder := orders.OrdersDown[order.Floor-1]
				if schedOrder != nil {
					conf.OrderCompletedSend <- *schedOrder
				} else {
					log.Println("Unexpected order completed")
				}
			case common.UpDir:
				schedOrder := orders.OrdersUp[order.Floor]
				if schedOrder != nil {
					conf.OrderCompletedSend <- *schedOrder
				} else {
					log.Println("Unexpected order completed")
				}
			default:
				log.Panicln("Unexpected direction")
			}

		case btn := <-conf.ButtonPressed:
			switch btn.Button {
			case elevio.BT_Cab:
				if orders.OrdersCab[btn.Floor] == nil {
					orders.OrdersCab[btn.Floor] = createOrder(btn.Floor, common.NoDir, conf.ElevatorID)
				}
			case elevio.BT_HallDown:
				assignee := selectAssignee(elevatorCosts, btn.Floor, common.DownDir)
				order := createOrder(btn.Floor, common.DownDir, assignee)
				conf.NewOrderSend <- *order
			case elevio.BT_HallUp:
				assignee := selectAssignee(elevatorCosts, btn.Floor, common.UpDir)
				order := createOrder(btn.Floor, common.UpDir, assignee)
				conf.NewOrderSend <- *order
			default:
				log.Panic("Invalid button type")
			}
		}

		//TODO Save orders to file
		err := savetofile(sysConf.FolderPath, &orders)
		if err != nil {
			log.Panic(err)
		}

		//Set order lights
		for _, order := range orders.OrdersUp {
			conf.Lights <- elevatordriver.LightState{Floor: order.Floor, Type: elevatordriver.UpButtonLight, State: (order != nil)}
		}
		for _, order := range orders.OrdersDown {
			conf.Lights <- elevatordriver.LightState{Floor: order.Floor, Type: elevatordriver.DownButtonLight, State: (order != nil)}
		}
		for _, order := range orders.OrdersCab {
			conf.Lights <- elevatordriver.LightState{Floor: order.Floor, Type: elevatordriver.InternalButtonLight, State: (order != nil)}
		}

		//Find next order and send to elevatorcontroller
		order := findHighestPriority(&orders, elevatorCosts[conf.ElevatorID], conf.ElevatorID)
		conf.ExecuteOrder <- order.Order
	}
}

func sendOrderCosts(c chan<- common.OrderCosts, costs *common.OrderCosts) {
	//DeepCopy slices
	msg := common.OrderCosts{
		CostsDown: append(make([]float64, 0, len(costs.CostsDown)), costs.CostsDown...),
		CostsUp:   append(make([]float64, 0, len(costs.CostsUp)), costs.CostsUp...),
	}
	c <- msg
}
func selectAssignee(assignees map[int]common.OrderCosts, floor int, dir common.Direction) int {
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
			if cost := v.CostsDown[floor-1]; cost < minCost {
				minCost = cost
				assignee = k
			}
		default:
			log.Panicf("Invalid direction %s\n", dir)
		}
	}
	return assignee
}

func findHighestPriority(orders *schedOrders, cost common.OrderCosts, id int) *SchedulableOrder {
	currMinCost := math.Inf(1)
	var currOrder *SchedulableOrder
	for _, order := range orders.OrdersCab {
		if order.Assignee != id {
			continue
		}
		orderCost := cost.Cab[order.Floor]
		if orderCost < currMinCost {
			currMinCost = orderCost
			currOrder = order
		}
	}
	for _, order := range orders.OrdersDown {
		if order.Assignee != id {
			continue
		}
		orderCost := cost.Down[order.Floor]
		if orderCost < currMinCost {
			currMinCost = orderCost
			currOrder = order
		}
	}
	for _, order := range orders.OrdersUp {
		if order.Assignee != id {
			continue
		}
		orderCost := cost.Up[order.Floor]
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
		if err := tryAddOrderToSlice(orders.OrdersDown, order.Floor-1, order); err != nil {
			return err
		}
	case common.NoDir:
		if err := tryAddOrderToSlice(orders.OrdersDown, order.Floor, order); err != nil {
			return err
		}
	default:
		return errors.New("Error adding order - unrecognized direction: " + fmt.Sprintf("%v", order))
	}
	return nil
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
