package scheduler

import (
	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/common"
)

func updateElevatorCost(conf Config, costs *common.OrderCosts, status common.ElevatorStatus) {
	penaltyToTarget := 7.0
	penaltyTargetDir := 3.0
	cabPenalty := 4.0

	for i := 0; i < conf.NumFloors; i++ {
		cost := float64(i-status.Floor) * 2.0

		if cost < 0 {
			if status.Dir == common.DownDir || status.Dir == common.NoDir {
				cost = -cost
				costs.CostsUp[i] = penaltyTargetDir * cost
				costs.CostsDown[i] = cost
				costs.CostsCab[i] = cabPenalty * cost
			}
		} else if cost > 0 {
			if status.Dir == common.UpDir || status.Dir == common.NoDir {
				costs.CostsUp[i] = cost
				costs.CostsDown[i] = penaltyTargetDir * cost
				costs.CostsCab[i] = cabPenalty * cost
			}
		} else if cost == 0 {
			cost = 1
			if status.Dir == common.NoDir {
				costs.CostsUp[i] = cost
				costs.CostsDown[i] = cost
				costs.CostsCab[i] = cost
			}
		}
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
