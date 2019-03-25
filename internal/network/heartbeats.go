package network

import(
		"time"
		"net"
		"log"
		"reflect"
		"github.com/TTK4145-students-2019/project-thefuturezebras/internal/common"
)

const heartbInterval = 10*time.Millisecond
const timeout = 10*heartbInterval

type costConfig struct{
	Config
	OrderCostIn <-chan common.OrderCosts
	OrderCostOut chan<- common.OrderCosts
	LostElevators chan<- common.OrderCosts.ID
	OnlineElevators chan<- []common.OrderCosts.ID
}

func runHeartbeat(ctx context.Context, conf costConfig){ 
	sendHeartbeatChan := chan common.OrderCosts
	recvHeartbeatChan := chan common.OrderCosts
	defer close(common.OrderCosts)

	atMostOnceConfig := AtMostOnceConfig {
		Config: conf.Config,
		Send: sendHeartbeatChan,
		Receive: recvHeartbeatChan,
	}

	go RunAtMostOnce(ctx, atMostOnceConfig) 

	mapLastHeartbeat := make(map[string]time.Time)

	cost:= <-conf.OrderCostIn

	onlineElevatorList:= []common.OrderCosts.ID

	for{
		idfound:=false
		select{
		case <-ctx.Done():
			return

		case rcvHeartb := <-recvHeartbeatChan: 
			OrderCostOut<- recvHeartb
			for id, _ := range mapLastHeartbeat{
				if id == recvHeartb.ID {
					idInMap = id
					idfound = true
					break
				}	
			}
			if idfound{
				mapLastHeartbeat[idInMap] = time.Now()
			}else{
				mapLastHeartbeat[rcvHeartb.ID] = time.Now()
				onlineElevatorList = append(onlineElevatorList, rcvHeartb.ID)
			}
			OnlineElevators<- onlineElevatorList

		case <- time.After(timeout):
			for id, timestamp := range mapLastHeartbeat{
				if time.Now().Sub(timestamp) > timeout{
					log.Println("Heartbeat timeout with id: ", id)	
					delete(mapLastHeartbeat,id)
					delete(onlineElevatorList,id)
					LostElevators<- id
				}
			}
		case <- time.After(heartbInterval):
			sendHeartbeatChan <- sendHeartb

		case cost = <- costConf.OrderCostIn:
				sendHeartbeatChan<- cost
	}
}