package scheduler

/*
import (
	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/common"
)


//Updates the cost of an elevator when the status of the elevator changes.
//Assigning an elevator to an order is based on the cost of the elevator to perform the specific order.
func updateElevatorCost(conf Config, costs *common.OrderCosts, status common.ElevatorStatus) {
	//Penalty if the elevator has to change direction to get to the floor of the order
	penaltyToTarget := 7.0
	//Penalty if the elevator gets to a floor and then has to change direction to perform the order
	penaltyTargetDir := 3.0
	//Penalty for cabc alls so that direction specific orders are prioritized over cabcalls at the same floor.
	//calls up and down also clear cab calls
	cabPenalty := 4.0

	for i := 0; i < conf.NumFloors; i++ {
		//gerneral cost is the distance between the last floor of the elevator and the floor of the order
		cost := float64(i-status.Floor) * 2.0

		//Negative cost means the order is below last floor of the elevator
		if cost < 0 {
			if status.Dir == common.DownDir || status.Dir == common.NoDir {
				cost = -cost
				costs.CostsUp[i] = penaltyTargetDir * cost
				costs.CostsDown[i] = cost
				costs.CostsCab[i] = cabPenalty * cost
			}
		//Positive cost means the order is above last floor of the elevator
		} else if cost > 0 {
			if status.Dir == common.UpDir || status.Dir == common.NoDir {
				costs.CostsUp[i] = cost
				costs.CostsDown[i] = penaltyTargetDir * cost
				costs.CostsCab[i] = cabPenalty * cost
			}
		//Order is placed at the last floor the elevator was at
		} else if cost == 0 {
			cost = 1
			if status.Dir == common.NoDir {
				costs.CostsUp[i] = cost
				costs.CostsDown[i] = cost
				costs.CostsCab[i] = cost
			}
		}
		//Worst case elevator moving up and away from target floor of the order
		if (cost < 0 && status.Dir == common.UpDir) || (cost == 0 && status.Dir == common.UpDir) {
			if cost < 0 {
				cost = -cost
			} else {
				cost = 1
			}
			costs.CostsUp[i] = penaltyToTarget * penaltyTargetDir * cost
			costs.CostsDown[i] = penaltyToTarget * cost
			costs.CostsCab[i] = penaltyToTarget * cabPenalty * cost
		}
		//Worst case elevator moving down and away from target floor of the order
		if (cost > 0 && status.Dir == common.DownDir) || (cost == 0 && status.Dir == common.DownDir) {
			if cost == 0 {
				cost = 1
			}
			costs.CostsUp[i] = penaltyToTarget * cost
			costs.CostsDown[i] = penaltyToTarget * penaltyTargetDir * cost
			costs.CostsCab[i] = penaltyToTarget * cabPenalty * cost
		}
	}
}

*/
