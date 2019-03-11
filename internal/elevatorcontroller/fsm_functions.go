package elevatorcontroller

//Elevator_controller
/*
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


func floorReached(int ElevatorFloor){
	//skal vi sette lamper her, eller ordne det et annet sted? Spesielt denne som settes ved hver etasje
	switch (state) {
		case DoorOpen:
			break

		case DoorClosed:
			break

		case MovingUp:
		case MovingDown:
			if ShouldStop(state,elevatorFloor,buttonMatrix) == true {
				Event <- Stop //jeg skjønner ikke helt hvordan vi gjør dette, skal vi sende beskjed om at vi vil stoppe??
				doorTOnimer() //tar seg av både åpme og lukke dør
				floorServiced(floor, buttonMatrix)

				state = DoorOpen
			}
		case Emergency:
			break
			//skal vi gjøre noe her?
	}
}



//Assuming that we only have one elevator so that the buttonMatrix are all orders assigned
//this specific elevator
//Checking if orders above or below the current floor of the elevator
func ordersAbove(elevatorFloor) bool{
	for floor := elevatorFloor + 1; floor < NumberofFloors; floor++{
		for button := 0; button < 3; button++{
			if buttonMatrix[floor][button]{
				return true
			}
		}
	}
	return false
}

func ordersBelow(int elevatorFloor) bool {
	for floor := 0; floor < elevatorFloor; floor++{
		for button := 0; button < 3; button++{
			if buttonMatrix[floor][button]{
				return true
			}
		}
	}
	return false
}

//Ikke ferdig
//Her skjønner jeg ikke helt hva du har planlagt? Å legge til nye bestillinger i en kø?
func newOrder(int buttonType, int floor){

	switch state{
	case DoorOpen:
	case DoorClosed:
	case MovingUp:
	case MovingDown:
	case Emergency:

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



//return true if there is an order on the current floor going in the right direction
//or there are noe orders above/below.

//Tror det blir feil å sette states her kanskje
func shouldStop(state state, int elevatorFloor, int buttonMatrix) bool{
	switch (state) {
		case MovingUp:
			if (buttonMatrix[elevatorFloor][BUTTON_CALL_UP] == 1){
				return true
			}
			if (ordersAbove() == false){
				return true
			}
			return false

		case MovingDown:
			if (buttonMatrix[elevatorFloor][BUTTON_CALL_DOWN] == 1){
				return true
			}
			if (ordersBelow() == false){
				return true
			}
			return false

		case Emergency:
			break
			//hva gjør vi i emergency?
	}
	return false
}

func floorServiced(direction direction, int elevatorFloor, int buttonMatrix){
	switch (direction){
	case up:

	case down:
	}

}

*/
