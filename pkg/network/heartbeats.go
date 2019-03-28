package network

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/common"
)

const heartbInterval = 50 * time.Millisecond
const timeout = 10 * heartbInterval

//HeartbeatConfig contains config params for the heartbeat module
type HeartbeatConfig struct {
	Config
	CostIn          <-chan common.OrderCosts
	CostOut         chan<- common.OrderCosts
	LostElevators   chan<- int
	OnlineElevators chan<- []int
}

type stampedHeartbeat struct {
	timestamp time.Time
	hbt       common.OrderCosts
}

//RunHeartbeat is the main entrypoint for heartbeats
func RunHeartbeat(ctx context.Context, conf HeartbeatConfig, onlineElevators ...chan<- []int) {
	sendHeartbeatChan := make(chan common.OrderCosts)
	recvHeartbeatChan := make(chan common.OrderCosts)
	defer close(sendHeartbeatChan)
	defer close(recvHeartbeatChan)

	atMostOnceConfig := AtMostOnceConfig{
		Config:  conf.Config,
		Send:    sendHeartbeatChan,
		Receive: recvHeartbeatChan,
	}
	//Store last received heartbeats
	mapLastHeartbeat := make(map[int]stampedHeartbeat)

	//Wait for first ordercost from anotherm module
	cost := <-conf.CostIn

	timeoutTimer := time.NewTicker(timeout)

	//Start atMostOnce service
	go RunAtMostOnce(ctx, atMostOnceConfig)

	for {
		select {
		case <-ctx.Done():
			return

		case cost = <-conf.CostIn:

		case hbt := <-recvHeartbeatChan:
			_, idfound := mapLastHeartbeat[hbt.ID]

			//Send orders cost (includes id) to receiver
			if !idfound || !reflect.DeepEqual(hbt, mapLastHeartbeat[hbt.ID].hbt) {
				fmt.Printf("New cost %v\n", hbt)
				conf.CostOut <- hbt
			}
			//Store timestamp
			mapLastHeartbeat[hbt.ID] = stampedHeartbeat{
				timestamp: time.Now(),
				hbt:       hbt,
			}
			//If no previous heartbeat exitst - notify channels
			if !idfound {
				//Publish online elevators list
				go publishNodesOnline(mapLastHeartbeat, onlineElevators...)
				fmt.Printf("New node detected %d\n", hbt.ID)
			}

		case <-timeoutTimer.C:
			for id, hbt := range mapLastHeartbeat {
				if time.Now().Sub(hbt.timestamp) > timeout {
					delete(mapLastHeartbeat, id)
					conf.LostElevators <- id
					//Published updated list of online elevators
					go publishNodesOnline(mapLastHeartbeat, onlineElevators...)
					fmt.Printf("Disconnected node detected %d\n", id)
				}
			}
		case <-time.After(heartbInterval):
			sendHeartbeatChan <- cost
		}

	}
}

func publishNodesOnline(mapLastHeartbeat map[int]stampedHeartbeat, sends ...chan<- []int) {
	for _, c := range sends {
		msg := []int{}
		for id := range mapLastHeartbeat {
			msg = append(msg, id)
		}
		c <- msg
	}
}
