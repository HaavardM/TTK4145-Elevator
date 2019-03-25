package elevatorcontroller

import (
	"context"
	"log"
	"time"
	"fmt"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/elevatordriver"
	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/common"
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

/*//Direction used to define preferred elevator direction
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
}*/

//Config used to configure the fsm
type Config struct {
	ElevatorCommand chan<- elevatordriver.Command
	ElevatorEvents  <-chan elevatordriver.Event
	Order         chan common.Order 			//common.Order
	ArrivedAtFloor  <-chan int
	NumberOfFloors int 
	OrderCompleted chan common.Order 			//common.Order
	ElevatorInfo chan<- common.Elevatorstatus
}

type fsm struct {
	state           state
	timer           *time.Timer
	elevatorCommand chan<- elevatordriver.Command
	currentOrder	common.Order 				//common.Order
	orderCompleted 	chan<- common.Order 		//common.Order
	status common.Elevatorstatus
}

const doorOpenDuration = 3*time.Second

func newFSM(elevatorCommand chan<- elevatordriver.Command, orderCompleted chan<-common.Order) *fsm {
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
	firstOrder := common.Order{common.UpDir, 2}
	conf.Order <- firstOrder
	for {
		select {
		case  <-conf.OrderCompleted:
			secondOrder := common.Order{common.DownDir,1}
			conf.Order <- secondOrder
		}
	}
}

//Run starts the elevatorcontroller fsm
func Run(ctx context.Context, conf Config) {
	fsm := newFSM(conf.ElevatorCommand, conf.OrderCompleted)
	fsm.transitionToDoorOpen()
	fsm.init(conf)
	log.Println("done")
	for {
		select {
		case fsm.currentOrder = <-conf.Order:
			log.Printf("New orders %v\n", fsm.currentOrder)
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


//Handles incomming orders from the scheduler module
func (f *fsm) handleNewOrders(conf Config) {
	targetFloor := f.currentOrder.Floor
	currentFloor := f.status.Floor
	targetDir := f.currentOrder.Dir

	if (targetDir == common.UpDir && targetFloor >= conf.NumberOfFloors) || (targetDir == common.DownDir && targetFloor <= 0) {
		log.Panic()
	}

	switch f.state {
	case stateMovingDown:
		if f.orderAbove(currentFloor) {
			f.transitionToMovingUp(conf)
		}
	case stateMovingUp:
		if !f.orderAbove(currentFloor) {
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

//Checking if new order is above or below current floor of the elevator
func (f *fsm) orderAbove(floor int) bool {
	targetFloor := f.currentOrder.Floor
	if targetFloor > floor {
		return true
	}
	return false
}


//Handles events that occur when reaching a new floow
func (f *fsm) handleAtFloor(conf Config) {				//julie
	conf.ElevatorInfo <- f.status
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
func (f *fsm) transitionToMovingDown(conf Config) {							//julie
	log.Println("Transition to moving down")
	f.elevatorCommand <- elevatordriver.MoveDown
	f.elevatorCommand <- elevatordriver.CloseDoor
	f.status.Dir = common.DownDir
	conf.ElevatorInfo <- f.status
	fmt.Println(f.status)
	f.state = stateMovingDown
}

//Handles transition from one state to moving up
func (f *fsm) transitionToMovingUp(conf Config) {								//julie
	log.Println("Transition to moving up")
	f.elevatorCommand <- elevatordriver.MoveUp
	f.elevatorCommand <- elevatordriver.CloseDoor
	f.status.Dir = common.UpDir
	conf.ElevatorInfo <- f.status
	f.state = stateMovingUp
}

//Initializes elevator when starting up
func (f *fsm) init(conf Config) {
	f.elevatorCommand <- elevatordriver.MoveUp
	f.status.Floor = <- conf.ArrivedAtFloor
	f.elevatorCommand <- elevatordriver.Stop
}


//Checks if we have reached target floor or not
func (f *fsm) shouldStop(floor int) bool {
	if floor == f.currentOrder.Floor {
		return true
	}
	return false
}
