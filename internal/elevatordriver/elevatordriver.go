package elevatordriver

import (
	"errors"

	"golang.org/x/net/context"

	"github.com/TTK4145/driver-go/elevio"
)

//Command used to send elevator commands
type Command int

//Event used to send events from elevator
type Event int

//LightType used to differentiate between types of lights
type LightType int

//LightState is used to request a light state
type LightState struct {
	Type  LightType
	State bool
	Floor int
}

const (
	/******************Commands***************************/

	//CloseDoor closes elevator door
	CloseDoor Command = iota + 1
	//OpenDoor opens elevator door
	OpenDoor
	//MoveUp starts motor with direction up
	MoveUp
	//MoveDown starts motor with direction down
	MoveDown
	//Stop stops the motor
	Stop

	/****************************Button Types************************/

	//UpButtonLight is the floor order buttons upwards
	UpButtonLight LightType = iota + 1
	//DownButtonLight is the floor order buttons downwards
	DownButtonLight
	//InternalButtonLight is the internal order buttons
	InternalButtonLight
	//AllLights used to set all lights at floor
	AllLights
)

//Config contains neccessary configuration for the elevator driver
type Config struct {
	Address        string
	NumberOfFloors int
	Commands       <-chan Command
	SetStatusLight <-chan LightState
	ArrivedAtFloor chan<- int
	OnButtonPress  chan<- elevio.ButtonEvent
}

//Run runs the elevator driver module
func Run(ctx context.Context, config Config) {
	arrivedAtFloor := make(chan int)
	//Initialize elevio module
	elevio.Init(config.Address, config.NumberOfFloors)
	//Start button poller
	go elevio.PollButtons(config.OnButtonPress)
	//Start floor sensor poller
	go elevio.PollFloorSensor(arrivedAtFloor)

	//Initalize to a stop state
	handleNewCommand(Stop)

	//Run infite loop until context finishes
	for {
		select {
		case c := <-config.Commands:
			handleNewCommand(c)
		case l := <-config.SetStatusLight:
			handleNewLightState(l)
		case f := <-arrivedAtFloor:
			elevio.SetFloorIndicator(f)
			config.ArrivedAtFloor <- f
		case <-ctx.Done():
			break
		}
	}
}

//Adapts the incoming commands for the hardware
func handleNewCommand(cmd Command) error {
	switch cmd {
	case CloseDoor:
		elevio.SetDoorOpenLamp(false)
	case OpenDoor:
		elevio.SetDoorOpenLamp(true)
	case MoveUp:
		elevio.SetMotorDirection(elevio.MD_Up)
	case MoveDown:
		elevio.SetMotorDirection(elevio.MD_Down)
	case Stop:
		elevio.SetMotorDirection(elevio.MD_Stop)
	default:
		return errors.New("ElevatorDriver: Command not recognized")
	}
	return nil
}

//Handles change in light state(on/off)
func handleNewLightState(light LightState) error {
	switch light.Type {
	case UpButtonLight:
		elevio.SetButtonLamp(elevio.BT_HallUp, light.Floor, light.State)
	case DownButtonLight:
		elevio.SetButtonLamp(elevio.BT_HallDown, light.Floor, light.State)
	case InternalButtonLight:
		elevio.SetButtonLamp(elevio.BT_Cab, light.Floor, light.State)
	case AllLights:
		elevio.SetButtonLamp(elevio.BT_HallUp, light.Floor, light.State)
		elevio.SetButtonLamp(elevio.BT_HallDown, light.Floor, light.State)
		elevio.SetButtonLamp(elevio.BT_Cab, light.Floor, light.State)
	default:
		return errors.New("Unrecognized light type")
	}
	return nil
}
