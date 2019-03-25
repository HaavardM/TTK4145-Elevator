package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

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
	CurrentOrdersSend  chan<- []SchedulableOrder
	CurrentOrdersRecv  <-chan []SchedulableOrder

	NumFloors int
}

type schedOrders struct {
	ordersUp   []*SchedulableOrder
	ordersDown []*SchedulableOrder
	ordersCab  []*SchedulableOrder
}

type elevatorCost struct {
	Up   []float64
	Down []float64
	Cab  []float64
}

//Run is the startingpoint for the scheduler module
//The ctx context is used to stop the gorotine if the context expires.
func Run(ctx context.Context, conf Config) {
	orders := schedOrders{
		ordersUp:   make([]*SchedulableOrder, conf.NumFloors-1),
		ordersDown: make([]*SchedulableOrder, conf.NumFloors-1),
		ordersCab:  make([]*SchedulableOrder, conf.NumFloors),
	}
	timer := time.NewTicker(time.Second)
	elevatorCosts := make(map[int]elevatorCost)

	for {
		select {
		case <-ctx.Done():
			return

		case <-timer.C:
			toSend := make([]SchedulableOrder, 0, len(orders.ordersDown)+len(orders.ordersUp))
			for _, o := range orders.ordersDown {
				toSend = append(toSend, *o)
			}

			for _, o := range orders.ordersUp {
				toSend = append(toSend, *o)
			}
			conf.CurrentOrdersSend <- toSend
		case order := <-conf.NewOrderRecv:
			handleNewOrder(&orders, order)
		case order := <-conf.OrderCompletedRecv:
			switch order.Dir {
			case common.UpDir:
				if err := tryRemoveOrderFromSlice(orders.ordersUp, order.Floor); err != nil {
					log.Println("Error removing order: ", err)
				}
			case common.DownDir:
				if err := tryRemoveOrderFromSlice(orders.ordersDown, order.Floor-1); err != nil {
					log.Println("Error removing order: ", err)
				}
			case common.NoDir:
				if err := tryRemoveOrderFromSlice(orders.ordersCab, order.Floor); err != nil {
					log.Println("Error removing order: ", err)
				}
			}
		case order := <-conf.OrderCompleted:
			switch order.Dir {
			case common.NoDir:
				//Clear orders
				orders.ordersCab[order.Floor] = nil
			case common.DownDir:
				schedOrder := orders.ordersDown[order.Floor-1]
				if schedOrder != nil {
					conf.OrderCompletedSend <- *schedOrder
				} else {
					log.Println("Unexpected order completed")
				}
			case common.UpDir:
				schedOrder := orders.ordersUp[order.Floor]
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
				if orders.ordersCab[btn.Floor] == nil {
					orders.ordersCab[btn.Floor] = createOrder(btn.Floor, common.NoDir, conf.ElevatorID)
				}
			case elevio.BT_HallDown:
				order := createOrder(btn.Floor, common.DownDir, selectAssignee(nil, btn.Floor, common.DownDir))
				conf.NewOrderSend <- *order
			case elevio.BT_HallUp:
				order := createOrder(btn.Floor, common.UpDir, selectAssignee(nil, btn.Floor, common.UpDir))
				conf.NewOrderSend <- *order
			default:
				log.Panic("Invalid button type")
			}
		}

		//TODO Save orders to file

		//Set order lights
		for _, order := range orders.ordersUp {
			conf.Lights <- elevatordriver.LightState{Floor: order.Floor, Type: elevatordriver.UpButtonLight, State: (order != nil)}
		}
		for _, order := range orders.ordersDown {
			conf.Lights <- elevatordriver.LightState{Floor: order.Floor, Type: elevatordriver.DownButtonLight, State: (order != nil)}
		}
		for _, order := range orders.ordersCab {
			conf.Lights <- elevatordriver.LightState{Floor: order.Floor, Type: elevatordriver.InternalButtonLight, State: (order != nil)}
		}

		//Find next order and send to elevatorcontroller
		order := findHighestPriority(&orders, elevatorCosts[conf.ElevatorID], conf.ElevatorID)
		conf.ExecuteOrder <- order.Order

	}
}

func selectAssignee(assignees map[int][]int, floor int, dir common.Direction) int {
	return 1
}

func findHighestPriority(orders *schedOrders, cost elevatorCost, id int) *SchedulableOrder {
	currMinCost := math.Inf(1)
	var currOrder *SchedulableOrder
	for _, order := range orders.ordersCab {
		if order.Assignee != id {
			continue
		}
		orderCost := cost.Cab[order.Floor]
		if orderCost < currMinCost {
			currMinCost = orderCost
			currOrder = order
		}
	}
	for _, order := range orders.ordersDown {
		if order.Assignee != id {
			continue
		}
		orderCost := cost.Down[order.Floor]
		if orderCost < currMinCost {
			currMinCost = orderCost
			currOrder = order
		}
	}
	for _, order := range orders.ordersUp {
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
		if err := tryAddOrderToSlice(orders.ordersUp, order.Floor, order); err != nil {
			return err
		}
	case common.DownDir:
		if err := tryAddOrderToSlice(orders.ordersDown, order.Floor-1, order); err != nil {
			return err
		}
	case common.NoDir:
		if err := tryAddOrderToSlice(orders.ordersDown, order.Floor, order); err != nil {
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
