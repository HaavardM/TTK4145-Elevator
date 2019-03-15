package main

import (
	"context"
	"fmt"
	"time"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/utilities"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/elevatorcontroller"
	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/elevatordriver"
	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/network"
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
	go newButtonPressed(ctx, onButtonPress, orders)
	go elevatorcontroller.Run(ctx, controllerConf)

	atLeastOnceSend := make(chan string)
	atMostOnceSend := make(chan string)
	atLeastOnceRecv := make(chan string)
	atMostOnceRecv := make(chan string)
	nodesOnline := make(chan []int)
	go utilities.ConstantPublisher(ctx, nodesOnline, []int{1, 2})

	atMostOnceConf := network.Config{
		Port:    2000,
		ID:      1,
		Send:    atMostOnceSend,
		Receive: atMostOnceRecv,
	}
	go network.RunAtMostOnce(ctx, atMostOnceConf)

	atLeastOnceConf := network.AtLeastOnceConfig{
		Config: network.Config{
			Port:    2001,
			ID:      1,
			Send:    atLeastOnceSend,
			Receive: atLeastOnceRecv,
		},
		NodesOnline: nodesOnline,
	}

	go network.RunAtLeastOnce(ctx, atLeastOnceConf)

	atLeastOnceSend <- "Hello atLeastOnce"
	atMostOnceSend <- "Hello atMostOnce"
	for {
		select {
		case m := <-atLeastOnceRecv:
			atLeastOnceSend <- "Hello again ALO"
			fmt.Println(m)
		case <-time.After(1 * time.Second):
			atMostOnceSend <- "Hello again AMO"
		}
	}

}

func newButtonPressed(ctx context.Context, onButtonPress <-chan elevio.ButtonEvent, elevatorOrders chan<- []elevatorcontroller.Order) {
	for {
		select {
		case b := <-onButtonPress:
			order := elevatorcontroller.Order{}
			switch b.Button {
			case elevio.BT_HallDown:
				order = elevatorcontroller.Order{
					Dir:   elevatorcontroller.DOWN,
					Floor: b.Floor,
				}
			case elevio.BT_HallUp:
				order = elevatorcontroller.Order{
					Dir:   elevatorcontroller.UP,
					Floor: b.Floor,
				}
			case elevio.BT_Cab:
				order = elevatorcontroller.Order{
					Dir:   elevatorcontroller.NoDirection,
					Floor: b.Floor,
				}
			}
			elevatorOrders <- []elevatorcontroller.Order{order}
		case <-ctx.Done():
			break
		}
	}
}
