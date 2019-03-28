package elevatorcontroller

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/common"
	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/elevatordriver"
)

const (
	stateMovingUp state = iota + 1
	stateMovingDown
	stateDoorOpen
	stateDoorClosed
	stateEmergency
)

type state int

//Returns a string representation of the elevator state
func (s state) String() string {
	switch s {
	case stateMovingUp:
		return "MOVING UP"
	case stateMovingDown:
		return "MOVING DOWN"
	case stateDoorOpen:
		return "DOOR OPEN"
	case stateDoorClosed:
		return "DOOR CLOSED"
	case stateEmergency:
		return "EMERGENCY"
	default:
		return "Unrecognized state"
	}
}

const doorOpenDuration = 3 * time.Second

//Sends information about the elevators position and direction so it is available for other modules
func sendElevatorStatus(c chan<- common.ElevatorStatus, status common.ElevatorStatus) {
	c <- status
}

//Config used to configure the fsm
type Config struct {
	ElevatorCommand chan<- elevatordriver.Command
	Order           chan common.Order
	ArrivedAtFloor  <-chan int
	NumberOfFloors  int
	OrderCompleted  chan common.Order
	ElevatorStatus  chan<- common.ElevatorStatus
}

//Struct containing variables and channels used by the statemachine
type fsm struct {
	state           state
	timer           *time.Timer
	elevatorCommand chan<- elevatordriver.Command
	currentOrder    *common.Order
	orderCompleted  chan<- common.Order
	status          common.ElevatorStatus
}

//Function initializing the fsm variables
func newFSM(elevatorCommand chan<- elevatordriver.Command, orderCompleted chan<- common.Order) *fsm {
	temp := &fsm{
		state:           stateDoorClosed,
		timer:           time.NewTimer(doorOpenDuration),
		elevatorCommand: elevatorCommand,
		orderCompleted:  orderCompleted,
	}
	if !(temp.timer.Stop()) {
		<-temp.timer.C
	}
	return temp
}

//Run starts the elevatorcontroller fsm
func Run(ctx context.Context, conf Config) {
	fsm := newFSM(conf.ElevatorCommand, conf.OrderCompleted)
	fsm.init(conf)
	for {
		select {
		case order := <-conf.Order:
			fsm.currentOrder = &order
			fsm.handleNewOrders(conf)
		case fsm.status.Floor = <-conf.ArrivedAtFloor:
			fsm.handleAtFloor(conf)
		case <-fsm.timer.C:
			fsm.handleTimerElapsed()
		case <-ctx.Done():
			break
		}
	}
}


//Initializes elevator when starting up so that it knows where it is
func (f *fsm) init(conf Config) {
	f.elevatorCommand <- elevatordriver.MoveUp
	f.status.Floor = <-conf.ArrivedAtFloor
	f.elevatorCommand <- elevatordriver.Stop
}


//Handles incomming orders from the scheduler module
func (f *fsm) handleNewOrders(conf Config) {

	if f.currentOrder == nil {
		return
	}
	//Initializes variables for the statemachine
	targetFloor := f.currentOrder.Floor
	currentFloor := f.status.Floor
	targetDir := f.currentOrder.Dir

	//Elevator out of range
	if (targetDir == common.UpDir && targetFloor >= conf.NumberOfFloors) || (targetDir == common.DownDir && targetFloor <= 0) {
		log.Panic()
	}

	switch f.state {
	case stateMovingDown:
		if currentFloor == targetFloor {
			f.transitionToDoorOpen()
		} else if f.orderAbove(currentFloor) {
			f.transitionToMovingUp(conf)
		}
	case stateMovingUp:
		if currentFloor == targetFloor {
			f.transitionToDoorOpen()
		} else if !f.orderAbove(currentFloor) {
			f.transitionToMovingDown(conf)
		}

	case stateDoorOpen, stateDoorClosed:
		if currentFloor == targetFloor {
			f.transitionToDoorOpen()
		} else if f.orderAbove(currentFloor) {
			f.transitionToMovingUp(conf)
		} else {
			f.transitionToMovingDown(conf)
		}
	}
}


//Handles events that occur when reaching a new floow
func (f *fsm) handleAtFloor(conf Config) {
	sendElevatorStatus(conf.ElevatorStatus, f.status)
	switch f.state {
	case stateMovingUp, stateMovingDown:
		if f.shouldStop(f.status.Floor) {
			f.transitionToDoorOpen()
		}
	}
}

//Handles transition from one state to the open door
func (f *fsm) transitionToDoorOpen() {
	log.Println("Transition to door open")
	f.elevatorCommand <- elevatordriver.Stop
	f.elevatorCommand <- elevatordriver.OpenDoor
	f.timer.Reset(doorOpenDuration)
	f.state = stateDoorOpen

}

//Handles transition from one state to door closed
func (f *fsm) transitionToDoorClosed() {
	log.Println("Transition to door closed")
	f.elevatorCommand <- elevatordriver.CloseDoor
	f.state = stateDoorClosed
	if f.currentOrder != nil {
		f.orderCompleted <- *f.currentOrder
		f.currentOrder = nil
	}
}

//handles transition from one state to moving down
func (f *fsm) transitionToMovingDown(conf Config) {
	log.Println("Transition to moving down")
	f.elevatorCommand <- elevatordriver.MoveDown
	f.elevatorCommand <- elevatordriver.CloseDoor
	f.status.Dir = common.DownDir
	sendElevatorStatus(conf.ElevatorStatus, f.status)
	fmt.Println(f.status)
	f.state = stateMovingDown
}

//Handles transition from one state to moving up
func (f *fsm) transitionToMovingUp(conf Config) {
	log.Println("Transition to moving up")
	f.elevatorCommand <- elevatordriver.MoveUp
	f.elevatorCommand <- elevatordriver.CloseDoor
	f.status.Dir = common.UpDir
	log.Println("Start")
	sendElevatorStatus(conf.ElevatorStatus, f.status)
	log.Println("End")
	f.state = stateMovingUp
}

//Checking if new order is above or below current floor of the elevator
func (f *fsm) orderAbove(floor int) bool {
	targetFloor := f.currentOrder.Floor
	if targetFloor > floor {
		return true
	}
	return false
}


//Checks if we have reached target floor or not
func (f *fsm) shouldStop(floor int) bool {
	if f.currentOrder == nil || floor == f.currentOrder.Floor {
		return true
	}
	return false
}


//Handles switching of state when door-open-timer has elapsed
func (f *fsm) handleTimerElapsed() {
	switch f.state {
	case stateDoorOpen:
		f.transitionToDoorClosed()
	}
}