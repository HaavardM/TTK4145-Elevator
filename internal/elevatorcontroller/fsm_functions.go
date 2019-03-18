/* ALT ER UBRUKT


package elevatorcontroller
import (
	"context"
	"log"
	"time"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/elevatordriver"
)

ordersUp = [NumberOfFloors]int
ordersDown = [NumberOfFloors]int
ordersNoDir = [NumberOfFloors]int


func shouldStop(floor int) bool {
	switch f.state {
	case stateMovingUp:
		if ordersUp[floor] == 1 || ordersNoDir[floor] == 1 || !ordersAbove(floor){
			return true
		}
		return false
	case stateMovingDown:
		if ordersDown[floor] == 1 || ordersNoDir[floor] == 1 || !ordersBelow(floor){
			return true
		}
		return false
	case stateDoorOpen:
	case stateDoorClosed:
	}
	return false
}

func ordersAbove(floor int) bool {
	for i := floor + 1; i < NumberOfFloors; i++ {
		if ordersUp[i] == 1 || ordersNoDir[i] == 1 {
			return true
		}
	}
	return false
}

func ordersBelow(floor int) bool {
	for i := 0; i < floor; i++ {
		if ordersDown[i] == 1 || ordersNoDir[i] == 1 {
			return true
		}
	}
	return false
}

func addOrders(orders []Order) {
	switch Order.Dir {
	case DOWN:
		ordersDown[Order.Floor] == 1
	case UP:
		ordersUp[Order.Floor] == 1
	case NoDirection:
		ordersNoDir[Order.Floor] == 1
	}	
}

func deleteOrders(floor int) {
	//sletter vi alltid alle ordre i alle retninger samtidig??
	ordersDown[floor] == 0
	ordersUp[floor] == 0
	ordersNoDir[floor] == 0
	
}

func chooseDirection(direction direction, state state, int elevatorFloor) int{
	switch (state) {
		case DoorOpen:				//Velger state avhengig av hvilken retning vi var på vei i da vi stoppa
			switch (direction){
			case up:
				if ordersAbove(elevatorFloor) {
					return up
				}
				else if ordersBelow(elevatorFloor){
					direction := down
					return down
				}
				else {
					direction := stop
					return stop
				}
			}
			case down:
				if ordersBelow(elevatorFloor) {
					return down
				}
				else if ordersAbove(elevatorFloor){
					direction := up
					return up
				}
				else {
					direction := stop
					return stop
				}
			}

		case DoorClosed:			//Velger utfra om vi har ubetjente ordre
			if ordersAbove(elevatorFloor) {
				direction := up
				return up
			}
			else if ordersBelow(elevatorFloor) {
				direction := down
				return down
			}
			else {
				direction := stop
				return stop
			}

		case MovingUp:				//fortsetter opp om vi fortsatt har ubetjente ordre opp, setter retning til ned om ikke
			if (direction == up && ordersAbove(elevatorFloor)) {
				direction := up
				return up
			}
			else if (ordersBelow(elevatorFloor)){
				direction := down
				return down
			}
			else {
				direction := stop
				return stop
			}

		case MovingDown:
			if (ordersBelow(elevatorFloor)){
				direction := down
				return down
			}

			else if (direction == up && ordersAbove(elevatorFloor)) {
				direction := up
				return up
			}

			else {
				direction := stop
				return stop
			}
		case Emergency:
		//hva gjør vi her da?
		break
	}
}
*/