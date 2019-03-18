package elevatorcontroller

import (
	"context"
	"log"
	"time"

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

//Direction used to define preferred elevator direction
type Direction int

const (
	//UP direction
	UP Direction = iota + 1
	//DOWN direction
	DOWN
	//NoDirection implies direction not important
	NoDirection
)

func (d Direction) String() string {
	switch d {
	case UP:
		return "UP"
	case DOWN:
		return "DOWN"
	case NoDirection:
		return "NO DIRECTION"
	default:
		return "Unrecognized direction"
	}
}

//Order contains information about an elevator order
type Order struct {
	Dir   Direction
	Floor int
}

//Config used to configure the fsm
type Config struct {
	ElevatorCommand chan<- elevatordriver.Command
	ElevatorEvents  <-chan elevatordriver.Event
	Order         <-chan Order
	ArrivedAtFloor  <-chan int
}

type fsm struct {
	state           state
	timer           *time.Timer
	elevatorCommand chan<- elevatordriver.Command
	currentOrder	Order
}

doorOpenDuration = 3 * time.Second

func newFSM(elevatorCommand chan<- elevatordriver.Command) *fsm {
	temp := &fsm{
		state:           stateDoorClosed,
		timer:           time.NewTimer(3 * time.Second),
		elevatorCommand: elevatorCommand,
	}
	if !(temp.timer.Stop()) {
		<-temp.timer.C
	}
	return temp
}

func (f *fsm) test(ctx context.Context, conf Config) {
	init(conf)
	orderCompleted := make(chan Order)
	firstOrder := Order{UP, 2}
	orderCompleted <- firstOrder
	for {
		select {
		case fsm.currentOrder := <-conf.Order:
			log.Printf("New orders %v\n", orders)
			fsm.handleNewOrders(orders)
		case floor := <-conf.ArrivedAtFloor:
			fsm.handleAtFloor(floor)
		case <-fsm.timer.C:
			fsm.handleTimerElapsed()
		case  <- orderCompleted:
			secondOrder := Order{DOWN,1}
			orderCompleted <- secondOrder
		case <-ctx.Done():
			break
		}
	}
}

//Run starts the elevatorcontroller fsm
func Run(ctx context.Context, conf Config) {
	init(floor)
	fsm := newFSM(conf.ElevatorCommand)
	fsm.transitionToDoorOpen()
	for {
		select {
		case fsm.currentOrder := <-conf.Order:
			log.Printf("New orders %v\n", orders)
			fsm.handleNewOrders(orders)
		case floor := <-conf.ArrivedAtFloor:
			fsm.handleAtFloor(floor)
		case <-fsm.timer.C:
			fsm.handleTimerElapsed()
		case <-ctx.Done():
			break
		}
	}
}


//Handles incomming orders from the scheduler module
func (f *fsm) handleNewOrders(floor int) {
	targetFloor := f.currentOrder.Floor
	targetDir := f.currentOrder.Dir

	if (targetDir == UP && floor == NumberOfFloors-1) || (targetDir == DOWN && floor == 0) {
		log.panic()
	}

	switch f.state {
	case stateMovingDown:
		if orderAbove(floor) {
			f.transitionToMovingUp()
		}
	case stateMovingUp:
		if !orderAbove(floor) {
			f.transitionToMovingDown()
		}
	case stateDoorOpen:
		if floor == targetFloor {
			f.timer.Reset(doorOpenDuration)
		}

	case stateDoorClosed:
		if floor == targetFloor {
			f.transitionToDoorOpen()
		}
		if orderAbove(floor) {
			f.transitionToMovingUp()
		}
		if !orderAbove(floor) {
			f.transitionToMovingDown()
		}
	}
}

//Checking if new order is above or below current floor of the elevator
func (f *fsm) orderAbove(floor int) int {
	targetFloor = f.currentOrder.Floor
	if targetFloor > floor {
		return true
	}
	return false
}


//Handles events that occur when reaching a new floow
func (f *fsm) handleAtFloor(floor int) {
	switch f.state {
	case stateMovingUp:
		fallthrough
	case stateMovingDown:
		if shouldStop(floor) {
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
	orderCompleted <- currentOrder
}



//Handles transition from one state to door closed
func (f *fsm) transitionToDoorClosed() {
	log.Println("Transition to door closed")
	f.elevatorCommand <- elevatordriver.CloseDoor
	f.state = stateDoorClosed
}


//Handles events when door-open-timer has elapsed
func (f *fsm) handleTimerElapsed() {
	switch f.state {
	case stateDoorOpen:
		f.transitionToDoorClosed()
	}
}

//handles transition from one state to moving down
func (f *fsm) transitionToMovingDown() {
	log.Println("Transition to moving down")
	f.elevatorCommand <- elevatordriver.MoveDown
	f.state = stateMovingDown
}

//Handles transition from one state to moving up
func (f *fsm) transitionToMovingUp() {
	log.Println("Transition to moving up")
	f.elevatorCommand <- elevatordriver.MoveUp
	f.state = stateMovingUp
}

//Initializes elevator when starting up
func (f *fsm) init(conf Config) {
	f.elevatorCommand <- elevator.MoveUp
	select {
		case floor <- conf.ArrivedAtFloor:
			f.elevatorCommand <- elevator.Stop
	}
}


//Checks if we have reached target floor or not
func (f *fsm) shouldStop(floor int) bool {
	if floor == f.currentOrder.Floor {
		return true
	}
	return false
}