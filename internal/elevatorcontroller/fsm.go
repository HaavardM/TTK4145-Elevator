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

//Direction used to define preferred elevator direction
type Direction int

const (
	//UP direction
	UP Direction = iota + 1
	//DOWN direction
	DOWN
)

//Order contains information about an elevator order
type Order struct {
	Dir   Direction
	Floor int
}

//Config used to configure the fsm
type Config struct {
	ElevatorCommand chan<- elevatordriver.Command
	ElevatorEvents  <-chan elevatordriver.Event
	Orders          <-chan []Order
	ArrivedAtFloor  <-chan int
}

type state int

type fsm struct {
	state           state
	timer           *time.Timer
	elevatorCommand chan<- elevatordriver.Command
}

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

func Run(ctx context.Context, conf Config) {
	fsm := newFSM(conf.ElevatorCommand)
	fsm.transitionToDoorOpen()
	for {
		select {
		case event := <-conf.ElevatorEvents:
			fsm.handleNewEvent(event)
		case orders := <-conf.Orders:
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

func (f *fsm) handleNewEvent(event elevatordriver.Event) {

}

func (f *fsm) handleNewOrders(orders []Order) {
	switch f.state {
	case stateMovingUp: //Oppdater kø
	case stateMovingDown: //Oppdater kø
	case stateDoorOpen:
	case stateDoorClosed:
	}
}

func shouldStop(floor int) bool {
	return true
}

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

func (f *fsm) transitionToDoorOpen() {
	log.Println("Transition to door open")
	f.elevatorCommand <- elevatordriver.Stop
	f.elevatorCommand <- elevatordriver.OpenDoor
	//TODO Avoid hardcoded duration
	f.timer.Reset(3 * time.Second)
	f.state = stateDoorOpen
}

func (f *fsm) transitionToDoorClosed() {
	log.Println("Transition to door closed")
	f.elevatorCommand <- elevatordriver.CloseDoor
	f.state = stateDoorClosed
}

func (f *fsm) handleTimerElapsed() {
	switch f.state {
	case stateDoorOpen:
		f.transitionToDoorClosed()
	}
}
