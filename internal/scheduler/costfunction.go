package scheduler

import (
	"math"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/common"
)

func createElevatorCost(status common.ElevatorStatus, orders *schedOrders, id int) common.OrderCosts {

	//Count orders
	orderCount := countOrdersWithID(orders, id)

	//Add a small penalty to cab call so that hall orders are prioritized at the same floor
	//Hall orders can clear cab orders, but not the other way around.
	cabPenalty := 0.5
	//Create new cost table
	newCost := common.OrderCosts{
		ID:         id,
		OrderCount: orderCount,
		Cab:        make([]float64, len(orders.Cab)),      //All zeros
		HallUp:     make([]float64, len(orders.HallUp)),   //All zeros
		HallDown:   make([]float64, len(orders.HallDown)), //All zeros
	}

	//We only care about the distance if the elevator has no other orders
	if orderCount == 0 {
		for i := 0; i < len(orders.Cab); i++ {
			cost := math.Abs(float64(i - status.Floor))
			newCost.Cab[i] = cost + 0.5
			newCost.HallDown[i] = cost
			newCost.HallUp[i] = cost
		}
	} else {

		costCounter := 0
		var searchDirIncrement int           //Current direction, either 1 or -1
		var currentQueue []*SchedulableOrder //Current queue (up/down) we are looking in. Must represent same direction as searchdir
		var currCost []float64               //Current cost slice we are adding cost to. Must represent same direction as searchdir
		startFloor := status.Floor           //Floor to start searching

		//Select search parameters based on prefered order direction
		if status.OrderDir == common.UpDir {
			searchDirIncrement = 1
			currentQueue = orders.HallUp
			currCost = newCost.HallUp
		} else {
			searchDirIncrement = -1
			currentQueue = orders.HallDown
			currCost = newCost.HallDown
		}

		//If moving, start from next floor in search
		if status.Moving {
			startFloor += searchDirIncrement
		}

		//Current floor
		currFloor := startFloor

		//Add incrementing cost in prefered direction
		for ; currFloor < len(currentQueue) && currFloor >= 0; currFloor += searchDirIncrement {
			orderCost := float64(costCounter + orderCount)
			currCost[currFloor] = orderCost
			costCounter++
		}

		//Switch cost and order lists based on direction
		if searchDirIncrement > 0 {
			currentQueue = orders.HallDown
			currCost = newCost.HallDown
		} else {
			currentQueue = orders.HallUp
			currCost = newCost.HallUp
		}
		//Switch direction
		searchDirIncrement *= -1

		//Add incrementing cost in opposite direction
		for currFloor += searchDirIncrement; currFloor < len(currentQueue) && currFloor >= 0; currFloor += searchDirIncrement {
			orderCost := float64(costCounter + orderCount)
			currCost[currFloor] = orderCost
			costCounter++
		}

		//Switch back to first direction
		if searchDirIncrement > 0 {
			currentQueue = orders.HallDown
			currCost = newCost.HallDown
		} else {
			currentQueue = orders.HallUp
			currCost = newCost.HallUp
		}
		//Switch direction
		searchDirIncrement *= -1
		//Add incrementing cost for the prefered direction, but for the order that are located in the opposite direction
		//Multiple direction switches will be neccessary to fulfill these orders
		for currFloor += searchDirIncrement; currFloor != startFloor; currFloor += searchDirIncrement {
			orderCost := float64(costCounter + orderCount)
			currCost[currFloor] = orderCost
			costCounter++
		}

		//Invalid cost for impossible orders
		newCost.HallUp[len(newCost.HallUp)-1] = -1
		newCost.HallDown[0] = -1

		//Select cheapest order and add cab penalty for cab orders
		for i := range newCost.Cab {
			newCost.Cab[i] = newCost.HallUp[i]
			if newCost.Cab[i] < 0 || (newCost.HallDown[i] < newCost.Cab[i] && newCost.HallDown[i] >= 0) {
				newCost.Cab[i] = newCost.HallDown[i]
			}
			newCost.Cab[i] += cabPenalty
		}
	}

	//Set the new costs
	return newCost
}

func countOrdersWithID(orders *schedOrders, id int) int {
	orderCount := 0

	allOrders := make([]*SchedulableOrder, 0, len(orders.Cab)+len(orders.HallDown)+len(orders.HallUp))
	allOrders = append(allOrders, orders.Cab...)
	allOrders = append(allOrders, orders.HallDown...)
	allOrders = append(allOrders, orders.HallUp...)
	//Count valid, existing orders for this elevator
	for _, order := range allOrders {
		if order != nil && order.Worker == id {
			orderCount++
		}
	}
	return orderCount
}
