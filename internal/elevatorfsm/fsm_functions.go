package fsm

import (
	"fmt"
	"time"
	)
const {
	DoorOpen
	DoorClosed
	MovingUp 
	MovingDown
	Emergency
}


//Dette var quene vi laget på torsdag
var lightMatrix = [N_FLOORS][3]
var buttonMatrix = [N_FLOORS][3]

//Kan også lage de slik:
type queue struct{
	matrix [def.N_Floors][def.N_Buttons]order status
}
//queue representing the buttons on the lift panel


//Dette er timeren vi begynte på på torsdag
func timer(){
	select{
		case 
	}
	timer := time.After(3*time.Second)

}

//Kan bruke dette som en timer
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
//Har ikke funnet ut hva .C er enda

//Har ikke funnet helt riktig syntaks enda, så skrev config.DirStop for nå
func doorTimeout (ch Channels){
	switch state{
	case DoorOpen:
		dir = queue.ChooseDirectioin(floor,dir)
		ch.MotorDir <- dir
		if dir == config.DirStop{
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

//vet ikke helt hvorden en skal gjøre det med MovingUp og MovingDown
func floorReached(ch Channels, newFloor int){
	floor = newFloor
	ch.FloorLamp <-floors
	switch state{
		case MovingUp:
			if queue.ShouldStop(floor,dir){
				ch.doorTimerReset <- true
				queue.RemoveOrdersAt(floor, ch.OutgoingMsg)
				ch.DoorLamp <- true
				dir = def.DirStop
				ch.MotorDir <- dir
				state = DoorOPen
			}
	}
}



//Making functions for checking orders below and above if we need them
//if we are to use them for multiple elevators we need a type elevator struct, that contains the 
//elevator we are looking at
//Elevator            [N_Elevators]Elev
func ordersAbove(elevator Elev) bool{
	for floor := elevator.Floor + 1; floor < N_Floors; floor++{
		for button := 0; button < N_Buttons; button++{
			if eleavtor.queue[floor][button]{
				return true
			}
		}
	}
	return false
}

func ordersBelow(elevator Elev) bool {
	for floor := 0; floor < elevator.Floor; floor++{
		for button := 0; button < N_Buttons; button++{
			if eleavtor.queue[floor][button]{
				return true
			}
		}
	}
	return false
}

//Ikke ferdig
func newOrder(ch Channels, newFloor in){
	floor = newFloor
	switch state{
	case DoorOpen:
	case DoorClosed:
	case MovingUp:
	case MovingDown:
	case Emergency:
	
	}
}