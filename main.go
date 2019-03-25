package main

import (
	"context"
	"fmt"
	"time"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/configuration"
	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/elevatorcontroller"
	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/elevatordriver"
	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/network"
	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/common"
	"github.com/TTK4145/driver-go/elevio"
)

const (
	TopicNewOrder int = iota + 1
	TopicOrderComplete
	TopicCurrentOrders
)

type Test struct {
	M string
	N string
}

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
	order := make(chan common.Order)
	orderCompleted := make(chan common.Order)
	elevatorInfo := make(chan common.Elevatorstatus)					//julie

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
		Order:           order,
		ArrivedAtFloor:  arrivedAtFloor,
		NumberOfFloors:  conf.Floors,
		OrderCompleted:  orderCompleted,	
		ElevatorInfo:	elevatorInfo,									//julie 
	}

	//Launch modules
	go elevatordriver.Run(ctx, elevatorConf)
	go elevatorcontroller.Run(ctx, controllerConf)
	go elevatorcontroller.Test(ctx, controllerConf) 				///juliehei
	go newButtonPressed(ctx, onButtonPress, order)

	/*************************TEST CODE***********************/
	atLeastOnceSend := make(chan string)
	atLeastOnceRecv := make(chan string)
	nodesOnline := make(chan []int)
	//go utilities.ConstantPublisher(ctx, nodesOnline, []int{1, 2})
	go func() {
		nodesOnline <- []int{1, 2}
	}()
	/*atMostOnceConf := network.AtMostOnceConfig{
		Config: network.Config{
			Port: conf.BasePort + TopicNewOrder,
			ID:   conf.ElevatorID,
		},
		Send:    atMostOnceSend,
		Receive: atMostOnceRecv,
	}
	go network.RunAtMostOnce(ctx, atMostOnceConf)
	*/
	atLeastOnceConf := network.AtLeastOnceConfig{
		Config: network.Config{
			Port: conf.BasePort + TopicOrderComplete,
			ID:   conf.ElevatorID,
		},
		Send:        atLeastOnceSend,
		Receive:     atLeastOnceRecv,
		NodesOnline: nodesOnline,
	}

	go network.RunAtLeastOnce(ctx, atLeastOnceConf)
	//atLeastOnceSend <- fmt.Sprintf("Hello from %d", conf.NetworkID)
	atLeastOnceSend <- "HeiPåDeg"
	for {
		select {
		case m := <-atLeastOnceRecv:
			atLeastOnceSend <- fmt.Sprintf("Hello again ALO from %d", conf.ElevatorID)
			<-time.After(time.Second)
			fmt.Println(m)
		}
	}

	//Wait for completion
	<-ctx.Done()
}

func newButtonPressed(ctx context.Context, onButtonPress <-chan elevio.ButtonEvent, elevatorOrder chan<- common.Order) {
	for {
		select {
		case b := <-onButtonPress:
			order := common.Order{}
			switch b.Button {
			case elevio.BT_HallDown:
				order = common.Order{
					Dir:   common.DownDir,
					Floor: b.Floor,
				}
			case elevio.BT_HallUp:
				order = common.Order{
					Dir:   common.UpDir,
					Floor: b.Floor,
				}
			case elevio.BT_Cab:
				order = common.Order{
					Dir:   common.NoDir,
					Floor: b.Floor,
				}
			}
			elevatorOrder <- order
		case <-ctx.Done():
			break
		}
	}
}
