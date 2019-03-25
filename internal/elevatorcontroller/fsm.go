package elevatorcontroller

import (
	"context"
	"log"
	"time"
	"fmt"

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

type Elevatorstatus struct {
	ElevatorDir		Direction
	ElevatorFloor	int
}

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
	Order         chan Order
	ArrivedAtFloor  <-chan int
	NumberOfFloors int 
	OrderCompleted chan Order
	ElevatorInfo chan<- Elevatorstatus		//Blir dette riktig?? Er det en scheduler eller en elevatorcontroller?
}

type fsm struct {
	state           state
	timer           *time.Timer
	elevatorCommand chan<- elevatordriver.Command
	currentOrder	Order
	orderCompleted 	chan<- Order
	currentFloor int
}

const doorOpenDuration = 3*time.Second

func newFSM(elevatorCommand chan<- elevatordriver.Command, orderCompleted chan<-Order) *fsm {
	temp := &fsm{
		state:           stateDoorClosed,
		timer:           time.NewTimer(doorOpenDuration),
		elevatorCommand: elevatorCommand,
		orderCompleted: orderCompleted,
	}
	if !(temp.timer.Stop()) {
		<-temp.timer.C
	}
	return temp
}

func Test(ctx context.Context, conf Config) {
	firstOrder := Order{UP, 2}
	conf.Order <- firstOrder
	for {
		select {
		case  <-conf.OrderCompleted:
			secondOrder := Order{DOWN,1}
			conf.Order <- secondOrder
		}
	}
}

//Run starts the elevatorcontroller fsm
func Run(ctx context.Context, conf Config, elevstat Elevatorstatus) {
	fsm := newFSM(conf.ElevatorCommand, conf.OrderCompleted)
	fsm.transitionToDoorOpen()
	fsm.init(conf)
	log.Println("done")
	for {
		select {
		case fsm.currentOrder = <-conf.Order:
			log.Printf("New orders %v\n", fsm.currentOrder)
			fsm.handleNewOrders(conf, elevstat)
		case fsm.currentFloor = <-conf.ArrivedAtFloor:
			fsm.handleAtFloor(conf, elevstat)
		case <-fsm.timer.C:
			fsm.handleTimerElapsed()
		case <-ctx.Done():
			break
		}
	}
}


//Handles incomming orders from the scheduler module
func (f *fsm) handleNewOrders(conf Config, elevstat Elevatorstatus) {
	targetFloor := f.currentOrder.Floor
	currentFloor := f.currentFloor
	targetDir := f.currentOrder.Dir

	if (targetDir == UP && targetFloor >= conf.NumberOfFloors) || (targetDir == DOWN && targetFloor <= 0) {
		log.Panic()
	}

	switch f.state {
	case stateMovingDown:
		if f.orderAbove(currentFloor) {
			f.transitionToMovingUp(conf, elevstat)
		}
	case stateMovingUp:
		if !f.orderAbove(currentFloor) {
			f.transitionToMovingDown(conf, elevstat)
		}

	case stateDoorOpen, stateDoorClosed:
		if currentFloor == targetFloor {
			f.transitionToDoorOpen()
		} else if f.orderAbove(currentFloor) {
			f.transitionToMovingUp(conf, elevstat)
		} else {
			f.transitionToMovingDown(conf, elevstat)
		}
	}
}

//Checking if new order is above or below current floor of the elevator
func (f *fsm) orderAbove(floor int) bool {
	targetFloor := f.currentOrder.Floor
	if targetFloor > floor {
		return true
	}
	return false
}


//Handles events that occur when reaching a new floow
func (f *fsm) handleAtFloor(conf Config, elevstat Elevatorstatus) {				//julie
	elevstat.ElevatorFloor = f.currentFloor
	conf.ElevatorInfo <- elevstat
	switch f.state {
	case stateMovingUp, stateMovingDown:
		if f.shouldStop(f.currentFloor) {
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
	f.orderCompleted <- f.currentOrder
}


//Handles events when door-open-timer has elapsed
func (f *fsm) handleTimerElapsed() {
	switch f.state {
	case stateDoorOpen:
		f.transitionToDoorClosed()
	}
}

//handles transition from one state to moving down
func (f *fsm) transitionToMovingDown(conf Config, elevstat Elevatorstatus) {							//julie
	log.Println("Transition to moving down")
	f.elevatorCommand <- elevatordriver.MoveDown
	f.elevatorCommand <- elevatordriver.CloseDoor
	elevstat.ElevatorDir = DOWN
	conf.ElevatorInfo <- elevstat
	fmt.Println(elevstat)
	f.state = stateMovingDown
}

//Handles transition from one state to moving up
func (f *fsm) transitionToMovingUp(conf Config, elevstat Elevatorstatus) {								//julie
	log.Println("Transition to moving up")
	f.elevatorCommand <- elevatordriver.MoveUp
	f.elevatorCommand <- elevatordriver.CloseDoor
	elevstat.ElevatorDir = UP
	conf.ElevatorInfo <- elevstat
	f.state = stateMovingUp
}

//Initializes elevator when starting up
func (f *fsm) init(conf Config) {
	f.elevatorCommand <- elevatordriver.MoveUp
	f.currentFloor = <- conf.ArrivedAtFloor
	f.elevatorCommand <- elevatordriver.Stop
}


//Checks if we have reached target floor or not
func (f *fsm) shouldStop(floor int) bool {
	if floor == f.currentOrder.Floor {
		return true
	}
	return false
}
