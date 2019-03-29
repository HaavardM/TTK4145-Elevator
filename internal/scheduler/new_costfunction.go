package scheduler

import (
	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/common"
)

func updateElevatorCost(conf Config, costs *common.OrderCosts, status common.ElevatorStatus, orders *schedOrders, id int) {
	dirPenalty := 3.0
	cabPenalty := 0.5

	for i := 0; i < conf.NumFloors; i++ {
		posCost := false
		negCost := false
		cost := float64(i-status.Floor)

		//hvis på vei i feil retning(vekk fra ordre)
		if (cost < 0 && status.Dir == common.UpDir) || (cost > 0 && status.Dir == common.DownDir) || cost == 0 && (status.Dir == common.UpDir || status.Dir == common.DownDir) {
			if cost < 0 {
				cost = -cost
			}
			if status.Dir == common.Updir {
				//må først opp og så ned igjen
				//status.floor null-indeks
				pathLength := (conf.NumFloors-(status.Floor+1)*2
			} else if status.Dir == common.DownDir {
				//først ned og så opp
				pathLength := status.Floor*2
			}
			cost += pathLength  
			//straffe ekstra for å være feil
		}

		if cost < 0 {
			negCost = true
			cost = -cost
		} else if cost > 0 {
			posCost = true
		}

		costs.CostsUp[i] = cost
		costs.CostsDown[i] = cost
		costs.CostsCab[i] = cabPenalty + cost

		//straffe ekstra i motsatt retning. Skal ikke strffe at vi vil opp i 1. da det er eneste alternative retning
		if status.Dir == common.DownDir && i!=0 {
			costs.CostsUp[i] = cost + dirPenalty
		}

		//skal ikke straffe at vi ned i 4. da det er eneste alternative retning
		if status.Dir == common.UpDir && i!=3 {
			costs.CostsDown[i] = cost + dirPenalty
		}

		//NoDir på heis, vil straffe ordre over heis som vil ned igjen, med unntak av 4.etasje
		if status.Dir == NoDir && posCost && i != 3 {
			costs.CostsDown[i] = cost + dirPenalty
		}

		//NoDir på heis, vil straffe ordre under som vil opp igjen, med unntak av 1. etasje
		if status.Dir == NoDir && negCost && i != 0 {
			costs.CostsUp[i] = cost + dirPenalty
		}
				//Check down hall orders
		for _, order := range orders.OrdersDown {
			if order != nil || order.Worker == id {
				for j = 0; j < order.Floor+1 ; j++ {
					costs.CostsDown[j] += 1
				}
			}
		}

		//Check up hall orders
		for _, order := range orders.OrdersUp {
			if order != nil || order.Worker == id {
				for j = 0; j < order.Floor+1 ; j++ {
					costs.CostsUp[j] += 1
				}
			}
		}

		//Check cab call orders
		for _, order := range orders.OrdersCab {
			if order != nil || order.Worker == id {
				for j = 0; j < order.Floor+1 ; j++ {
					costs.CostsUp[j] += 1
				}
			}
		}
	}
}		