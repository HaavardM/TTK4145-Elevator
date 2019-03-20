package network

import (
	"context"
	"encoding/json"
	"log"
	"reflect"
	"time"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/utilities"
)

type atLeastOnceMsg struct {
	Ack       bool        `json:"ack"`
	SenderID  int         `json:"sender_id"`
	MessageID string      `json:"message_id"`
	Data      interface{} `json:"data"`
}

//AtLeastOnceConfig contains configuration for the atLeastOnce QoS
type AtLeastOnceConfig struct {
	Config
	NodesOnline <-chan []int
}

//RunAtLeastOnce runs at most once publishing at a certain port
//Service is limited to one datatype per port
func RunAtLeastOnce(ctx context.Context, conf AtLeastOnceConfig) {

	//Store received acks for current messages
	acks := make(map[string]map[int]struct{})
	//Store current publishers with cancel function
	publishers := make(map[string]func())
	//Store current alive nodes
	nodesOnline := []int{}

	//Get type of data sent on input/output channel
	T := reflect.TypeOf(conf.Send).Elem()
	if reflect.TypeOf(conf.Receive).Elem() != T {
		log.Panic("Datatypes for send and receive not consistent")
	}

	//Create channels
	bSend := make(chan atLeastOnceMsg)
	bRecv := make(chan atLeastOnceMsg)
	ret := make(chan atLeastOnceMsg)
	//Get input channel as a type agnostic interface channel
	atleastOnceInput, err := utilities.ReflectChan2InterfaceChan(ctx, reflect.ValueOf(conf.Send))
	recvChan := reflect.ValueOf(conf.Receive)
	if err != nil {
		log.Panicln("Error starting atleastonce: ", err)
	}

	//Start AtMostOnce service
	c := Config{
		Send:    bSend,
		Receive: bRecv,
		ID:      conf.ID,
		Port:    conf.Port,
	}
	go RunAtMostOnce(ctx, c)

	//Wait for new input
	for {
		select {
		case <-ctx.Done():
			return
		//When a new message is ready to send
		case m := <-atleastOnceInput:
			msg := atLeastOnceMsg{
				Ack:       false,
				SenderID:  conf.ID,
				MessageID: "TODO",
				Data:      m,
			}
			//Create child context - cancellable when all acks received
			sendCtx, cancel := context.WithCancel(ctx)
			publishers[msg.MessageID] = cancel
			acks[msg.MessageID] = make(map[int]struct{})
			//Start a new goroutine to send same message at fixed interval
			go sendUntilDone(sendCtx, msg, bSend, ret)
		//When a send
		case r := <-ret:
			//Cleanup
			if _, ok := publishers[r.MessageID]; ok {
				delete(publishers, r.MessageID)
			}

			b, err := json.Marshal(r.Data)
			if err != nil {
				log.Println("error receiving message: ", err)
			}
			v := reflect.New(T)
			err = json.Unmarshal(b, v.Interface())
			if err != nil {
				log.Panicln("Failed unmarshal")
			}
			recvChan.Send(reflect.Indirect(v))
		case m := <-bRecv:
			//Send ack to corresponding goroutine
			if m.Ack {
				if d, ok := acks[m.MessageID]; ok {
					d[m.SenderID] = struct{}{}
				} else {
					log.Println("Received ack for non existent message")
				}
			} else if m.SenderID != conf.ID {
				//Send ACK
				m.Ack = true
				m.SenderID = conf.ID
				bSend <- m
			} else {
				//Message from another node
				//Return the message to client
				ret <- m
			}
		//Set nodesOnline to updated value
		case nodesOnline = <-conf.NodesOnline:
		}

		//Remove completed sends
		for m, d := range acks {
			done := true
			for _, o := range nodesOnline {
				if _, ok := d[o]; !ok {
					done = false
					break
				}
			}

			//Cleanup if removed
			if done {
				if c, ok := publishers[m]; ok {
					if c != nil {
						//Cancel send go routine
						c()
					} else {
						log.Println("Missing cancel function for test")
					}
				} else {
					log.Println("Test name from receivedAcks not in activeMsgs")
				}
				delete(publishers, m)
				delete(acks, m)
			}
		}
	}
}

//Send until context ends
func sendUntilDone(ctx context.Context, content atLeastOnceMsg, send chan<- atLeastOnceMsg, ret chan<- atLeastOnceMsg) {
	timer := time.NewTicker(100 * time.Millisecond)
	defer timer.Stop()
	//While not received all acks
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			send <- content
		}
	}
}
