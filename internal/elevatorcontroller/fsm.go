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
	Orders          <-chan []Order
	ArrivedAtFloor  <-chan int
}

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

//Run starts the elevatorcontroller fsm
func Run(ctx context.Context, conf Config) {
	fsm := newFSM(conf.ElevatorCommand)
	fsm.transitionToDoorOpen()
	for {
		select {
		case event := <-conf.ElevatorEvents:
			fsm.handleNewEvent(event)
		case orders := <-conf.Orders:
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

func (f *fsm) handleNewEvent(event elevatordriver.Event) {
//????Hvilke events? Er ikke det bare stop pressed og released? Det skal vi vel ikke ha med
}

func (f *fsm) handleNewOrders(orders []Order) {
addOrders([]Order)
	/*
	Skal vi legge til i liste i det hele tatt? buttonpress i heis må vel legges til, men om vi skal
	kunne bytte retning med én gang en ny ordre kommer er vel ikke det nødvendig for bestilling av heis?
	Lettere måte å gjøre dette på en switch i switch??
	*/
	
	switch Order.Dir {
	case DOWN:
		switch f.state {
		case stateMovingUp: 
			fallthrough
		case stateDoorOpen:
			fallthrough
		case stateDoorClosed:
			f.transitionToMovingDown()
		case stateMovingDown:
			//trenger vel ikke gjøre noe her?
		}
	case UP:
		switch f.state {
		case stateMovingDown: 
			fallthrough
		case stateDoorOpen:
			fallthrough
		case stateDoorClosed:
			f.transitionToMovingUp()
		case stateMovingUp:
			//trenger vel ikke gjøre noe her?
		}
	case NoDirection:
		switch f.state {
		//hva skal prioriteres her da??
		}
	}
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
	f.handleTimerElapsed() //er det dumt å legge den her?
	// LEGGE INN HER AT ORDRE ER FULLFØRT?? Hvor skal det sendes? Event?
	//altså slette fra egen liste (deleteOrder) og sende bekreftelse
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

func (f *fsm) transitionToMovingDown() {
	log.Println("Transition to moving down")
	f.elevatorCommand <- elevatordriver.MoveDown
	f.state = stateMovingDown
}

func (f *fsm) transitionToMovingUp() {
	log.Println("Transition to moving up")
	f.elevatorCommand <- elevatordriver.MoveUp
	f.state = stateMovingUp
}