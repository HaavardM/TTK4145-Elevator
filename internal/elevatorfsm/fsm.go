package 
//Elevator_controller

import (
	"fmt"
	"context"
)

const(
	MovingUp State = iota+1
	MovingDown
	DoorOpen
	DoorClosed
	Emergency
)

type Direction int
const (
	UP Direction = iota+1
	DOWN Direction
)

type Order struct{
	Dir Direction
	Floor int
}

type Config struct{
	ElevatorCommand chan <- elevatordriver.Command
	ElevatorEvents <- chan elevatordriver.Event
	Orders <- chan []Order
	ArrivedAtFloor <- chan int
}

type State int

type FSM struct{
	state State
	timer timer.Timer
	elevatorCommand chan <- elevatordriver.Command
}

func newFSM(elevatorCommand chan <- elevatordriver.Command) *FSM{
	temp := &FSM{
		state: DoorClosed, 
		timer: timer.NewTimer(3*time.Second),
		elevatorCommand: elevatorCommand, 
	}
	if !(temp.timer.Stop()){
		<- temp.timer.C
	}
	return temp
}

func run(ctx context.Context, conf Config) {
	fsm := newFSM(conf.ElevatorCommand)
	for{
		select{
		case event := <- conf.ElevatorEvents:
			fsm.handleNewEvent(event)
		case orders := <- conf.Orders:
			fsm.handleNewOrders(orders)
		case floor := <- conf.ArrivedAtFloor:
			fsm.handleAtFloor(floor)
		case <- timer.C:
			fsm.handleTimerElapsed()
		case <- ctx.Done():
			break
		}
	}
}

func (f *FSM) handleNewEvent(event elevatordriver.Event){

}

func (f *FSM) handleNewOrders(orders []Order){
	switch f.state{
		case MovingUp //Oppdater kø
		case MovingDown //Oppdater kø
		case DoorOpen
		case DoorClosed
	}
}

func (f *FSM) handleAtFloor(floor int){
	switch f.state{
		case MovingUp:
			fallthrough
		case MovingDown:
			if shouldStop(floor){
				transitionToDoorOpen()
			}
	}
}

func (f *FSM)transitionToDoorOpen(){
	f.elevatorCommand <- elevatordriver.Stop
	f.elevatorCommand <- elevatordriver.OpenDoor
	f.timer.Reset()
}

func (f *FSM)transitionToDoorClosed(){
	f.elevatorCommand <- elevatordriver.CloseDoor
	f.state = DoorClosed
}

func (f *FSM)handleTimerElapsed(){
	switch f.state{
	case DoorOpen:
		transitionToDoorClosed()
	}
}