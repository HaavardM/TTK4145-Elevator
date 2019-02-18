package main

import (
	"context"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/elevatorcontroller"
	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/elevatordriver"
	"github.com/TTK4145/driver-go/elevio"
)

func main() {
	arrivedAtFloor := make(chan int)
	elevatorCommand := make(chan elevatordriver.Command)
	elevatorEvents := make(chan elevatordriver.Event)
	onButtonPress := make(chan elevio.ButtonEvent)
	lightState := make(chan elevatordriver.LightState)
	elevatorConf := elevatordriver.Config{
		Address:        "localhost:15657",
		NumberOfFloors: 4,
		ArrivedAtFloor: arrivedAtFloor,
		Commands:       elevatorCommand,
		Events:         elevatorEvents,
		OnButtonPress:  onButtonPress,
		SetStatusLight: lightState,
	}
	orders := make(chan []elevatorcontroller.Order)
	controllerConf := elevatorcontroller.Config{
		ElevatorCommand: elevatorCommand,
		ElevatorEvents:  elevatorEvents,
		Orders:          orders,
		ArrivedAtFloor:  arrivedAtFloor,
	}
	ctx := context.Background()
	go elevatordriver.Run(ctx, elevatorConf)
	elevatorcontroller.Run(ctx, controllerConf)
}
