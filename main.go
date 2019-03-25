package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/network"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/common"
	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/configuration"
	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/elevatorcontroller"
	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/elevatordriver"
	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/scheduler"
	"github.com/TTK4145/driver-go/elevio"
)

const (
	//TopicNewOrder is a AtLeastOnceTopic used to send new orders
	TopicNewOrder int = iota + 1
	//TopicOrderComplete is an AtLeastOnceTopic used to send order complete msgs
	TopicOrderComplete
)

func main() {
	//Main context
	ctx, cancel := context.WithCancel(context.Background())
	waitGroup := sync.WaitGroup{}

	//Get configration
	conf := configuration.GetConfig()

	//Create neccessary channels for the elevator
	arrivedAtFloor := make(chan int)
	elevatorCommand := make(chan elevatordriver.Command)
	onButtonPress := make(chan elevio.ButtonEvent)
	lightState := make(chan elevatordriver.LightState)
	order := make(chan common.Order)
	orderCompleted := make(chan common.Order)
	elevatorInfo := make(chan common.ElevatorStatus) //julie

	topicNewOrderSend := make(chan scheduler.SchedulableOrder)
	topicNewOrderRecv := make(chan scheduler.SchedulableOrder)
	topicOrderCompleteSend := make(chan scheduler.SchedulableOrder)
	topicOrderCompleteRecv := make(chan scheduler.SchedulableOrder)

	costSend := make(chan common.OrderCosts)
	costRecv := make(chan common.OrderCosts)

	//Create elevator configuration
	elevatorConf := elevatordriver.Config{
		Address:        fmt.Sprintf("localhost:%d", conf.ElevatorPort),
		NumberOfFloors: conf.Floors,
		ArrivedAtFloor: arrivedAtFloor,
		Commands:       elevatorCommand,
		OnButtonPress:  onButtonPress,
		SetStatusLight: lightState,
	}

	//Create elevator controller configuration
	controllerConf := elevatorcontroller.Config{
		ElevatorCommand: elevatorCommand,
		Order:           order,
		ArrivedAtFloor:  arrivedAtFloor,
		NumberOfFloors:  conf.Floors,
		OrderCompleted:  orderCompleted,
		ElevatorInfo:    elevatorInfo, //julie
	}

	newOrderNodesOnline := make(chan []int)
	topicNewOrderConf := network.AtLeastOnceConfig{
		Config: network.Config{
			Port: conf.BasePort + TopicNewOrder,
			ID:   conf.ElevatorID,
		},
		Send:        topicNewOrderSend,
		Receive:     topicNewOrderRecv,
		NodesOnline: newOrderNodesOnline,
	}

	orderCompletedNodesOnline := make(chan []int)
	topicOrderCompletedConf := network.AtLeastOnceConfig{
		Config: network.Config{
			Port: conf.BasePort + TopicOrderComplete,
			ID:   conf.ElevatorID,
		},
		Send:        topicOrderCompleteSend,
		Receive:     topicOrderCompleteRecv,
		NodesOnline: orderCompletedNodesOnline,
	}

	schedulerConf := scheduler.Config{
		NumFloors:          conf.Floors,
		NewOrderRecv:       topicNewOrderRecv,
		NewOrderSend:       topicNewOrderSend,
		OrderCompletedRecv: topicOrderCompleteRecv,
		OrderCompletedSend: topicOrderCompleteSend,
		ElevStatus:         elevatorInfo,
		ElevatorID:         conf.ElevatorID,
		ElevButtonPressed:  onButtonPress,
		ElevCompletedOrder: orderCompleted,
		Lights:             lightState,
		CostsSend:          costSend,
		CostsRecv:          costRecv,
		ElevExecuteOrder:   order,
		FolderPath:         conf.FolderPath,
	}

	//Launch modules
	go elevatordriver.Run(ctx, elevatorConf)
	go elevatorcontroller.Run(ctx, controllerConf)

	//Create two AtLeastOnce topics
	waitGroup.Add(2)
	go network.RunAtLeastOnce(ctx, &waitGroup, topicNewOrderConf)
	go network.RunAtLeastOnce(ctx, &waitGroup, topicOrderCompletedConf)

	//Wait for scheduler to complete
	waitGroup.Add(1)
	go scheduler.Run(ctx, &waitGroup, schedulerConf)

	go func() {
		newOrderNodesOnline <- []int{1, 2}
		orderCompletedNodesOnline <- []int{1, 2}
	}()

	go func() {
		time.Sleep(10 * time.Second)
		log.Panic("SOME PANIC")
	}()

	//Handle signals to get a graceful shutdown
	sig := make(chan os.Signal)
	go handleSignals(sig, cancel)
	signal.Notify(sig, os.Interrupt, os.Kill)
	//Wait for shutdown
	<-ctx.Done()

	//Wait for important goroutines to exit
	waitGroup.Wait()
}

func handleSignals(sig <-chan os.Signal, cancelCtx func()) {
	if cancelCtx == nil {
		log.Panicln("Invalid cancel function")
	}
	<-sig
	cancelCtx()
}
