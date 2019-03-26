package scehduler

func updateElevatorCost(costs *common.OrderCosts, status common.ElevatorStatus) {
	penaltyToTarget := 7
	penaltyTargetDir := 3
	cabPenalty := 4

	for i := 0; i < NumberOfFloors; i++ {
		cost := (i - ElevatorFloot)*2	


		if cost < 0	{							
			if status.Dir == DownDir	||	status.Dir == NoDir {
				cost = -cost 
				costs.CostsUp[i] = penaltyTargetDir*cost
				costs.CostsDown[i] = cost
				costs.CostsCab[i] = cabPenalty*cost 
			}
		}
		else if cost > 0 {									
			if status.Dir == UpDir	||	status.Dir == NoDir {
				costs.CostsUp[i] = cost
				costs.CostsDown[i] = penaltyTargetDir*cost
				costs.CostsCab[i] = cabPenalty*cost
			}
		}
		else if cost == 0	{
			cost = 1
			if status.Dir == NoDir {						
				costs.CostsUp[i] = cost
				costs.CostsDown[i] = cost
				costs.CostsCab[i] = cost
			}
		}
		if (cost < 0 && status.Dir == UpDir)	||	(cost == 0 && status.Dir == UpDir) {
			if cost < 0 {
				cost = -cost
			} else {cost = 1}
			costs.CostsUp[i] = penaltyToTarget*penaltyTargetDir*cost
			costs.CostsDown[i] = penaltyToTarget*cost
			costs.CostsCab[i] = penaltyToTarget*cabPenalty*cost 		
		}
		if  (cost > 0 && status.Dir == DownDir)	||	(cost == 0 && status.Dir == DownDir) {
			if cost == 0 {
				cost = 1
			}
			costs.CostsUp[i] = penaltyToTarget*cost
			costs.CostsDown[i] = penaltyToTarget*penaltyTargetDir*cost
			costs.CostsCab[i] = penaltyToTarget*cabPenalty*cost
		}		
	}
}