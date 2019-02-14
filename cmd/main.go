package main

import (
	"context"
	"time"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/elevatordriver"
	"github.com/TTK4145/driver-go/elevio"
)

func main() {
	arrivedAtFloor := make(chan int)
	elevatorCommand := make(chan elevatordriver.Command)
	elevatorEvents := make(chan elevatordriver.Event)
	onButtonPress := make(chan elevio.ButtonEvent)
	lightState := make(chan elevatordriver.LightState)
	conf := elevatordriver.Config{
		Address:        "localhost:8080",
		NumberOfFloors: 4,
		ArrivedAtFloor: arrivedAtFloor,
		Commands:       elevatorCommand,
		Events:         elevatorEvents,
		OnButtonPress:  onButtonPress,
		SetStatusLight: lightState,
	}
	ctx := context.Background()
	go elevatordriver.Run(ctx, conf)
	dir := elevatordriver.MoveUp
	for {
		select {
		case f := <-arrivedAtFloor:
			lightState <- elevatordriver.LightState{Type: elevatordriver.AllLights, State: false, Floor: f}
			elevatorCommand <- elevatordriver.Stop
			if f == conf.NumberOfFloors-1 {
				dir = elevatordriver.MoveDown
			} else if f == 0 {
				dir = elevatordriver.MoveUp
			}
		case <-time.After(5 * time.Second):
			elevatorCommand <- dir
		case b := <-onButtonPress:
			switch b.Button {
			case elevio.BT_HallDown:
				lightState <- elevatordriver.LightState{Type: elevatordriver.DownButtonLight, State: true, Floor: b.Floor}
			case elevio.BT_HallUp:
				lightState <- elevatordriver.LightState{Type: elevatordriver.UpButtonLight, State: true, Floor: b.Floor}
			case elevio.BT_Cab:
				lightState <- elevatordriver.LightState{Type: elevatordriver.InternalButtonLight, State: true, Floor: b.Floor}
			}
		}
	}

}
