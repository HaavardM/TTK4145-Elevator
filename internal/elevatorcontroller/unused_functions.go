package elevatorcontroller

//Elevator_controller

import (
	"fmt"
	"time"
	)


// doorTimer keeps a timer for the door open duration, It resets
//when told to and notifies the state macine when it times out
func doorTimer(timeout chan<- bool, reset <-chan bool){
	const doorOpenTime = 3 * time.Second
	timer := time.NewTimer(0)
	timer.Stop()
	for{
		select{
		case <- reset:
			timer.Reset(doorOPenTime)
		case <-timer.C:
			timer.Stop()
			timeout <- true
		}
	}
}

//Har ikke funnet helt riktig syntaks enda, så skrev config.DirStop for nå
func doorTimeout (ch Channels){
	switch (state) {
	case DoorOpen:
		dir = ChooseDirection(floor,dir)
		ch.MotorDir <- dir
		if dir == config.DirStop{= MovingUp
			state = DoorClosed
		}
		//klarte ikke bestemme meg for beste måte å velg eom det var MovingUp eller MovingDown
		else if dir = config.DirUp {
			state = MovingUp
		}
		else{
			state = MovingDown
		}
	default:

	}
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


func chooseDir(floor int) { //eller chooseState
	switch f.state {
	case stateMovingUp:
		if !ordersAbove(floor) && ordersBelow(floor){
			f.transistionToMovingDown()
		}
		//hvis verken ordre opp eller ned, må vi via door open for å stoppe?
	case stateMovingDown:
		if !ordersBelow(floor) && ordersAbove(floor){
			f.transistionToMovingUP()
		}
	case stateDoorOpen:
	case stateDoorClosed:
	}
}