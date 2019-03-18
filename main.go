package main

import (
	"context"
	"fmt"
	"time"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/utilities"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/configuration"
	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/elevatorcontroller"
	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/elevatordriver"
	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/network"
	"github.com/TTK4145/driver-go/elevio"
)

const (
	TopicNewOrder int = iota + 1
	TopicOrderComplete
	TopicCurrentOrders
)

func main() {
	//Main context
	ctx := context.Background()

	//Get configration
	conf := configuration.GetConfig()

	//Create neccessary channels for the elevator
	arrivedAtFloor := make(chan int)
	elevatorCommand := make(chan elevatordriver.Command)
	elevatorEvents := make(chan elevatordriver.Event)
	onButtonPress := make(chan elevio.ButtonEvent)
	lightState := make(chan elevatordriver.LightState)
	orders := make(chan []elevatorcontroller.Order)

	//Create elevator configuration
	elevatorConf := elevatordriver.Config{
		Address:        fmt.Sprintf("localhost:%d", conf.ElevatorPort),
		NumberOfFloors: conf.Floors,
		ArrivedAtFloor: arrivedAtFloor,
		Commands:       elevatorCommand,
		Events:         elevatorEvents,
		OnButtonPress:  onButtonPress,
		SetStatusLight: lightState,
	}

	//Create elevator controller configuration
	controllerConf := elevatorcontroller.Config{
		ElevatorCommand: elevatorCommand,
		ElevatorEvents:  elevatorEvents,
		Orders:          orders,
		ArrivedAtFloor:  arrivedAtFloor,
	}

	//Launch modules
	go elevatordriver.Run(ctx, elevatorConf)
	go elevatorcontroller.Run(ctx, controllerConf)
	go newButtonPressed(ctx, onButtonPress, orders)

	/*************************TEST CODE***********************/
	atLeastOnceSend := make(chan string)
	atMostOnceSend := make(chan string)
	atLeastOnceRecv := make(chan string)
	atMostOnceRecv := make(chan string)
	nodesOnline := make(chan []int)
	go utilities.ConstantPublisher(ctx, nodesOnline, []int{1})

	atMostOnceConf := network.Config{
		Port:    conf.BasePort + TopicNewOrder,
		ID:      conf.NetworkID,
		Send:    atMostOnceSend,
		Receive: atMostOnceRecv,
	}
	go network.RunAtMostOnce(ctx, atMostOnceConf)
	atLeastOnceConf := network.AtLeastOnceConfig{
		Config: network.Config{
			Port:    conf.BasePort + TopicOrderComplete,
			ID:      conf.NetworkID,
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

	//Wait for completion
	<-ctx.Done()
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
