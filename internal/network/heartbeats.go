package network

import (
	"context"
	"log"
	"time"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/common"
)

const heartbInterval = 10 * time.Millisecond
const timeout = 10 * heartbInterval

//HeartbeatConfig contains config params for the heartbeat module
type HeartbeatConfig struct {
	Config
	CostIn          <-chan common.OrderCosts
	CostOut         chan<- common.OrderCosts
	LostElevators   chan<- int
	OnlineElevators chan<- []int
}

//RunHeartbeat is the main entrypoint for heartbeats
func RunHeartbeat(ctx context.Context, conf HeartbeatConfig) {
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
	mapLastHeartbeat := make(map[int]time.Time)

	//Wait for first ordercost from anotherm module
	cost := <-conf.CostIn

	//Start atMostOnce service
	go RunAtMostOnce(ctx, atMostOnceConfig)

	for {
		select {
		case <-ctx.Done():
			return

		case cost = <-conf.CostIn:

		case hbt := <-recvHeartbeatChan:
			idfound := false
			//Send orders cost (includes id) to receiver
			conf.CostOut <- hbt
			for id := range mapLastHeartbeat {
				if id == hbt.ID {
					idfound = true
					break
				}
			}
			//Store timestamp
			mapLastHeartbeat[hbt.ID] = time.Now()
			//If no previous heartbeat exitst - notify channels
			if !idfound {
				//Create online elevators message
				msg := []int{}
				for id := range mapLastHeartbeat {
					msg = append(msg, id)
				}
				conf.OnlineElevators <- msg
			}

		case <-time.After(timeout):
			for id, timestamp := range mapLastHeartbeat {
				if time.Now().Sub(timestamp) > timeout {
					log.Println("Heartbeat timeout with id: ", id)
					delete(mapLastHeartbeat, id)
					conf.LostElevators <- id
				}
			}
		case <-time.After(heartbInterval):
			sendHeartbeatChan <- cost
		}

	}
}
