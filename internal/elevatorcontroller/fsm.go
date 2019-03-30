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
	nextOrder       *common.Order
	orderCompleted  chan<- common.Order
	status          common.ElevatorStatus
	statusSend      chan<- common.ElevatorStatus
}

func runSendLatestElevatorStatus(ctx context.Context, send chan<- common.ElevatorStatus, status <-chan common.ElevatorStatus) {
	for {
		select {
		case <-ctx.Done():
			return
		//Wait until there is something to send
		case msg := <-status:
			//Try send the status - or overwrite the status with a new one if received
			select {
			//Abort if context finished
			case <-ctx.Done():
				return
			//Send order if channel is ready
			case send <- msg:
			//Update order to send if a new one is received before the sendChan is ready
			case msg = <-status:
			}
		}
	}

}

func newFSM(elevatorCommand chan<- elevatordriver.Command, orderCompleted chan<- common.Order, statusSend chan<- common.ElevatorStatus) *fsm {
	temp := &fsm{
		state:           stateDoorClosed,
		timer:           time.NewTimer(doorOpenDuration),
		elevatorCommand: elevatorCommand,
		orderCompleted:  orderCompleted,
		statusSend:      statusSend,
	}
	if !(temp.timer.Stop()) {
		<-temp.timer.C
	}
	return temp
}

//Run starts the elevatorcontroller fsm
func Run(ctx context.Context, conf Config) {
	//Create elevator status publisher
	elevatorStatus := make(chan common.ElevatorStatus)
	go runSendLatestElevatorStatus(ctx, conf.ElevatorStatus, elevatorStatus)

	fsm := newFSM(conf.ElevatorCommand, conf.OrderCompleted, elevatorStatus)
	fsm.init(conf)

	for {
		select {
		case nextOrder := <-conf.Order:
			log.Println("New order")
			fsm.handleNewOrders(conf, nextOrder)
		case fsm.status.Floor = <-conf.ArrivedAtFloor:
			fsm.handleAtFloor(conf)
			log.Println("New floor")
		case <-fsm.timer.C:
			fsm.handleTimerElapsed(conf)
			log.Println("New timer")
		case <-ctx.Done():
			break
		}
		elevatorStatus <- fsm.status
	}
}

//Initializes elevator when starting up so that it knows where it is
func (f *fsm) init(conf Config) {
	f.elevatorCommand <- elevatordriver.MoveUp
	f.status.Floor = <-conf.ArrivedAtFloor
	f.elevatorCommand <- elevatordriver.Stop
	f.status.Dir = common.NoDir
	f.statusSend <- f.status
}

//Handles incomming orders from the scheduler module
func (f *fsm) handleNewOrders(conf Config, order common.Order) {

	//Initializes variables for the statemachine
	targetFloor := order.Floor
	currentFloor := f.status.Floor
	targetDir := order.Dir

	//Elevator out of range
	if (targetDir == common.UpDir && targetFloor >= conf.NumberOfFloors) || (targetDir == common.DownDir && targetFloor <= 0) {
		log.Panic()
	}

	//Clear next order - order is the new order
	f.nextOrder = nil

	switch order.Dir {
	case common.NoDir:
		if orderAbove(order, currentFloor) {
			f.status.Dir = common.UpDir
		} else if order.Floor != currentFloor {
			f.status.Dir = common.DownDir
		}
	case common.DownDir, common.UpDir:
		f.status.Dir = order.Dir
	default:
		log.Panicln("Unknown direction")
	}
	switch f.state {
	case stateMovingDown:
		f.currentOrder = &order
		//Set state to current order dir
		if orderAbove(order, currentFloor) || currentFloor == targetFloor {
			f.transitionToMovingUp(conf)
		}
	case stateMovingUp:
		f.currentOrder = &order
		//Set state to current order dir
		if !orderAbove(order, currentFloor) || currentFloor == targetFloor {
			f.transitionToMovingDown(conf)
		}

	case stateDoorClosed:
		f.currentOrder = &order
		//Set state to current order dir
		if currentFloor == targetFloor {
			f.transitionToDoorOpen(conf)
		} else if orderAbove(order, currentFloor) {
			f.transitionToMovingUp(conf)
		} else {
			f.transitionToMovingDown(conf)
		}
	case stateDoorOpen:
		f.nextOrder = &order
	}

}

//Handles events that occur when reaching a new floow
func (f *fsm) handleAtFloor(conf Config) { //julie
	f.statusSend <- f.status
	switch f.state {
	case stateMovingUp, stateMovingDown:
		if f.shouldStop(f.status.Floor) {
			f.transitionToDoorOpen(conf)
		}
	}
}

//Handles transition from one state to the open door
func (f *fsm) transitionToDoorOpen(conf Config) {
	log.Println("Transition to door open")
	f.elevatorCommand <- elevatordriver.Stop
	f.elevatorCommand <- elevatordriver.OpenDoor
	f.timer.Reset(doorOpenDuration)
	f.status.Moving = false
	f.state = stateDoorOpen

}

//Handles transition from one state to door closed
func (f *fsm) transitionToDoorClosed(conf Config) {
	log.Println("Transition to door closed")
	f.elevatorCommand <- elevatordriver.CloseDoor
	f.status.Moving = false
	if f.currentOrder != nil {
		f.orderCompleted <- *f.currentOrder
		f.currentOrder = nil
	}
	//Handle next order
	f.state = stateDoorClosed
	//Handle next order if it exists
	if f.nextOrder != nil {
		f.handleNewOrders(conf, *f.nextOrder)
	}
}

//handles transition from one state to moving down
func (f *fsm) transitionToMovingDown(conf Config) {
	log.Println("Transition to moving down")
	f.elevatorCommand <- elevatordriver.MoveDown
	f.elevatorCommand <- elevatordriver.CloseDoor
	f.status.Moving = true
	fmt.Println(f.status)
	f.state = stateMovingDown
}

//Handles transition from one state to moving up
func (f *fsm) transitionToMovingUp(conf Config) {
	log.Println("Transition to moving up")
	f.elevatorCommand <- elevatordriver.MoveUp
	f.elevatorCommand <- elevatordriver.CloseDoor
	f.status.Moving = true
	f.state = stateMovingUp
}

//Checking if new order is above or below current floor of the elevator
func orderAbove(order common.Order, floor int) bool {
	targetFloor := order.Floor
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
func (f *fsm) handleTimerElapsed(conf Config) {
	switch f.state {
	case stateDoorOpen:
		f.transitionToDoorClosed(conf)
	}
}
