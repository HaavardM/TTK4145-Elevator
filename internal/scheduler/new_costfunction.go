package scheduler

import (
	"math"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/common"
)

func updateElevatorCost(costs *common.OrderCosts, status common.ElevatorStatus, orders *schedOrders, id int) {
	//Count orders
	orderCount := 0

	allOrders := make([]*SchedulableOrder, 0, len(orders.OrdersCab)+len(orders.OrdersDown)+len(orders.OrdersUp))
	allOrders = append(allOrders, orders.OrdersCab...)
	allOrders = append(allOrders, orders.OrdersDown...)
	allOrders = append(allOrders, orders.OrdersUp...)
	//Count valid, existing orders for this elevator
	for _, order := range allOrders {
		if order != nil && order.Worker == id {
			orderCount++
		}
	}
	newCost := common.OrderCosts{
		ID:         id,
		OrderCount: orderCount,
		CostsCab:   make([]float64, len(costs.CostsCab)),
		CostsUp:    make([]float64, len(costs.CostsUp)),
		CostsDown:  make([]float64, len(costs.CostsDown)),
	}

	if orderCount == 0 {
		for i := 0; i < len(orders.OrdersCab); i++ {
			cost := math.Abs(float64(i - status.Floor))
			newCost.CostsCab[i] = cost + 0.5
			newCost.CostsDown[i] = cost
			newCost.CostsUp[i] = cost
		}
	} else {

		costCount := 0
		var searchDir int
		var currentQueue []*SchedulableOrder
		var currCost []float64
		startFloor := status.Floor

		if status.OrderDir == common.UpDir {
			searchDir = 1
			currentQueue = orders.OrdersUp
			currCost = newCost.CostsUp
		} else {
			searchDir = -1
			currentQueue = orders.OrdersDown
			currCost = newCost.CostsDown
		}
		if status.Moving {
			startFloor += searchDir
		}
		currFloor := startFloor

		for ; currFloor < len(currentQueue) && currFloor >= 0; currFloor += searchDir {
			orderCost := float64(costCount + orderCount)
			currCost[currFloor] = orderCost
			newCost.CostsCab[currFloor] = orderCost + 0.5
			costCount++
		}

		if searchDir > 0 {
			currentQueue = orders.OrdersDown
			currCost = newCost.CostsDown
		} else {
			currentQueue = orders.OrdersUp
			currCost = newCost.CostsUp
		}
		//Switch direction
		searchDir *= -1
		//Search opposite direction
		for currFloor += searchDir; currFloor < len(currentQueue) && currFloor >= 0; currFloor += searchDir {
			orderCost := float64(costCount + orderCount)
			currCost[currFloor] = orderCost
			if newCost.CostsCab[currFloor] <= 0.25 {
				newCost.CostsCab[currFloor] = orderCost + 0.5
			}
			costCount++
		}

		if searchDir > 0 {
			currentQueue = orders.OrdersDown
			currCost = newCost.CostsDown
		} else {
			currentQueue = orders.OrdersUp
			currCost = newCost.CostsUp
		}
		//Switch direction
		searchDir *= -1
		//Search opposite direction
		for currFloor += searchDir; currFloor != startFloor; currFloor += searchDir {
			orderCost := float64(costCount + orderCount)
			currCost[currFloor] = orderCost
			costCount++
		}
	}
	*costs = newCost
}
