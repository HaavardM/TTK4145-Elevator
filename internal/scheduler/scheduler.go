package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/common"
	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/elevatordriver"
	"github.com/TTK4145/driver-go/elevio"
)

//PriorityOrder is an order with a priority and cost
type PriorityOrder struct {
	common.Order `json:"order"`
	Assignee     int `json:"assignee"`
}

type Config struct {
	ID                 int
	ButtonPressed      <-chan elevio.ButtonEvent
	OrderCompleted     <-chan common.Order
	CurrentOrder       chan<- common.Order
	Lights             chan<- elevatordriver.LightState
	NewOrderSend       chan<- PriorityOrder
	NewOrderRecv       <-chan PriorityOrder
	OrderCompletedSend chan<- PriorityOrder
	OrderCompletedRecv <-chan PriorityOrder
	NumFloors          int
}

type schedOrders struct {
	ordersUp   []*PriorityOrder
	ordersDown []*PriorityOrder
	ordersCab  []*PriorityOrder
}

func Run(ctx context.Context, conf Config) {
	localOrders := schedOrders{
		ordersUp:   make([]*PriorityOrder, conf.NumFloors-1),
		ordersDown: make([]*PriorityOrder, conf.NumFloors-1),
	}

	globalOrders := schedOrders{
		ordersUp:   make([]*PriorityOrder, conf.NumFloors-1),
		ordersDown: make([]*PriorityOrder, conf.NumFloors-1),
	}

	for {
		select {
		case <-ctx.Done():
			return
		case order := <-conf.NewOrderRecv:
			handleNewOrder(&globalOrders, order)
		case order := <-conf.OrderCompleted:
			switch order.Dir {
			case common.UpDir:
				if err := tryRemoveOrderFromSlice(localOrders.ordersUp, order.Floor); err != nil {
					log.Println("Error removing order: ", err)
				}
			case common.DownDir:
				if err := tryRemoveOrderFromSlice(localOrders.ordersDown, order.Floor-1); err != nil {
					log.Println("Error removing order: ", err)
				}
			case common.NoDir:
				if err := tryRemoveOrderFromSlice(localOrders.ordersCab, order.Floor); err != nil {
					log.Println("Error removing order: ", err)
				}
			}
		}
	}

}

func handleNewOrder(orders *schedOrders, order PriorityOrder) error {
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

func tryAddOrderToSlice(slice []*PriorityOrder, pos int, order PriorityOrder) error {
	if pos >= 0 && pos < len(slice) {
		//Identical orders might occur - OK if ID is the same.
		if slice[pos] == nil || order.ID == slice[pos].ID {
			slice[pos] = &order
			return nil
		}
	}
	return errors.New("Invalid index")
}

func tryRemoveOrderFromSlice(slice []*PriorityOrder, pos int) error {
	if pos >= 0 && pos < len(slice) {
		slice[pos] = nil
		return nil
	}
	return errors.New("Invalid index")
}
