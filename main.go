package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/TTK4145-students-2019/project-thefuturezebras/pkg/network"

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
	//TopicHeartbeat is used to detect other nodes
	TopicHeartbeat
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
	orderCompleted := make(chan common.Order)

	//Make these buffered to avoid blocking on send
	//We do not require the scheduler and elevatorcontroller to be in perfect sync,
	//but the order of the messages sent on these channels must be correct
	order := make(chan common.Order, 1)
	elevatorInfo := make(chan common.ElevatorStatus, 1) //julie

	topicNewOrderSend := make(chan scheduler.SchedulableOrder)
	topicNewOrderRecv := make(chan scheduler.SchedulableOrder)
	topicNewOrderExpectedAcks := make(chan []int)
	topicOrderCompleteSend := make(chan scheduler.SchedulableOrder)
	topicOrderCompleteRecv := make(chan scheduler.SchedulableOrder)
	topicOrderCompleteExpectedAcks := make(chan []int)

	costSend := make(chan common.OrderCosts, 1)
	costRecv := make(chan common.OrderCosts, 1)
	workerLost := make(chan int, 1)

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
		ElevatorStatus:  elevatorInfo, //julie
	}

	topicNewOrderConf := network.AtLeastOnceConfig{
		Config: network.Config{
			Port: conf.BasePort + TopicNewOrder,
			ID:   conf.ElevatorID,
		},
		Send:        topicNewOrderSend,
		Receive:     topicNewOrderRecv,
		NodesOnline: topicNewOrderExpectedAcks,
	}

	topicOrderCompletedConf := network.AtLeastOnceConfig{
		Config: network.Config{
			Port: conf.BasePort + TopicOrderComplete,
			ID:   conf.ElevatorID,
		},
		Send:        topicOrderCompleteSend,
		Receive:     topicOrderCompleteRecv,
		NodesOnline: topicOrderCompleteExpectedAcks,
	}

	heartbeatConf := network.HeartbeatConfig{
		Config: network.Config{
			ID:   conf.ElevatorID,
			Port: conf.BasePort + TopicHeartbeat,
		},
		CostIn:        costSend,
		CostOut:       costRecv,
		LostElevators: workerLost,
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
		FilePath:           conf.FilePath,
		WorkerLost:         workerLost,
	}

	//Launch modules
	go elevatordriver.Run(ctx, elevatorConf)
	go elevatorcontroller.Run(ctx, controllerConf)

	//Create two AtLeastOnce topics
	go network.RunAtLeastOnce(ctx, topicNewOrderConf)
	go network.RunAtLeastOnce(ctx, topicOrderCompletedConf)

	//Create heartbeat module
	go network.RunHeartbeat(ctx, heartbeatConf, topicNewOrderExpectedAcks, topicOrderCompleteExpectedAcks)

	//Wait for scheduler to complete
	waitGroup.Add(1)
	go scheduler.Run(ctx, &waitGroup, schedulerConf)

	/*
		go func() {
			time.Sleep(10 * time.Second)
			log.Panic("SOME PANIC")
		}()
	*/

	//Handle signals to get a graceful shutdown
	sig := make(chan os.Signal)
	go handleSignals(sig, cancel)
	signal.Notify(sig, os.Interrupt, os.Kill)
	//Wait for shutdown
	<-ctx.Done()

	//Wait for important goroutines to exit
	waitGroup.Wait()
	os.Exit(0)
}

func handleSignals(sig <-chan os.Signal, cancelCtx func()) {
	if cancelCtx == nil {
		log.Panicln("Invalid cancel function")
	}
	<-sig
	cancelCtx()
}
