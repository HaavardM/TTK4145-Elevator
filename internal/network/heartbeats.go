package network

import(
		"../conn"
		"time"
		"sort"
		"net"
		"fmt"
		"log"
)

type Cost []int

type ElevatorState struct { //channel????? til scheduler
	ID int
	CostUp Cost
	CostDown Cost
}

type networkMessage struct{
	ID 		int `json:"id"`
	Cost 	[]int `json:"cost"` //burde ikke cost være delt i to deler?
}

const interval = 10*time.Millisecond
const timeout = 10*interval

type heartbeatConfig struct{
	ExternalCost <-chan Cost
	InternalCost chan<- Cost
}

func runHeartbeat(ctx context.Context, conf Config, heartConf <-chan heartbeatConfig, id int){
	sendHeartbeatChan := chan networkMessage
	recvHeartbeatChan := chan networkMessage
	recvCostChan := chan Cost
	defer close(networkMessage)

	go RunAtMostOnce(ctx, conf) 

	lastHeartbeat := make(map[string]time.Time)

	sendmsg:= networkMessage{
		ID: id
		Cost: <-recvCostChan 
	}

	for{
		update := false
		select{
		case <-ctx.Done():
			return
		case rcvmsg := <-recvHeartbeatChan:
			idfound := false
			for id, _ := range lastHeartbeat{
				if id = rcvmsg.id {
					idfound = true
				} 
			}
			if !idfound {
				log.Println("New heartbeat found with id: ", id)
				update = true
			}
			lastHeartbeat[rcvmsg.id] = time.Now() 
		case <- time.After(timeout):
			for id, timestamp := range lastHeartbeat{
				if time.Now().Sub(timestamp) > timeout{
					log.Println("Heartbeat timeout with id: ", id)	
					delete(lastHeartbeat,id)
					update = true
				}
			}
		case <- time.After(interval):
			sendHeartbeatChan <- sendmsg
		case cost := <- heartConf.InternalCost: ///HVOR KOMMER COSTEN FRA?????
			sendmsg.Cost = cost 
		}
		if update {
			msg := HeartbeatUpdate {}
			externalCost := <- ExternalCost
			for externalid, costup, costdown := externalCost{ //fra annen modul
				//if externalid = id {
					//do the for loop below
				}
			}
				
			for id, _ := lastHeartbeat {
				msg.Heartbeats = append(msg.Heartbeats, ElevatorState {
					ID: id,
					CostUp: costup,
					CostDown: costdown,
				})
			}
			heartConf <- msg //til external ??
		}
	}
}