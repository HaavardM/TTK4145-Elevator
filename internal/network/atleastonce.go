package network

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/rs/xid"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/utilities"
)

type atLeastOnceMsg struct {
	Ack       bool        `json:"ack"`
	SenderID  int         `json:"sender_id"`
	MessageID string      `json:"message_id"`
	Data      interface{} `json:"data"`
}

//IDSet is a set of ids (ints)
//Implemented as a hashmap
//Value is an empty struct to save memory (empty structs use no memory)
type IDSet map[int]struct{}

//AtLeastOnceConfig contains configuration for the atLeastOnce QoS
type AtLeastOnceConfig struct {
	Config
	Send        interface{}
	Receive     interface{}
	NodesOnline <-chan []int
}

//RunAtLeastOnce runs at most once publishing at a certain port
//Service is limited to one datatype per port
func RunAtLeastOnce(ctx context.Context, conf AtLeastOnceConfig) {

	//Store received acks for current messages
	acks := make(map[string]IDSet)
	//Store current publishers with cancel function
	publishers := make(map[string]func())
	//Store current alive nodes
	nodesOnline := []int{}

	msgCounter := 0

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
	c := AtMostOnceConfig{
		Send:    bSend,
		Receive: bRecv,
		Config:  conf.Config,
	}
	go RunAtMostOnce(ctx, c)

	//Wait for new input
	for {
		select {
		case <-ctx.Done():
			return
		//When a new message is ready to send
		case m := <-atleastOnceInput:
			msgCounter++
			msg := atLeastOnceMsg{
				Ack:       false,
				SenderID:  conf.ID,
				MessageID: fmt.Sprintf("%s:%d", xid.New().String(), msgCounter),
				Data:      m,
			}
			//Create child context - cancellable when all acks received
			sendCtx, cancel := context.WithCancel(ctx)
			publishers[msg.MessageID] = cancel
			acks[msg.MessageID] = make(IDSet)
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
			go recvChan.Send(reflect.Indirect(v))
		case m := <-bRecv:
			//Send ack to corresponding goroutine
			if m.Ack {
				if idSet, ok := acks[m.MessageID]; ok {
					idSet[m.SenderID] = struct{}{}
				} else {
					log.Printf("Received ack for non existent message %v\n", m)
				}
			} else {
				//Send ACK
				m.Ack = true
				m.SenderID = conf.ID
				bSend <- m
				go SendMessage(ctx, ret, m)
			}
		//Set nodesOnline to updated value
		case nodesOnline = <-conf.NodesOnline:
		}

		//Remove completed sends
		for m, d := range acks {
			done := true
			for _, o := range nodesOnline {
				if o == conf.ID {
					continue
				}
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
	timer := time.NewTicker(1000 * time.Millisecond)
	defer timer.Stop()
	//While not received all acks
	done := false
	for !done {
		select {
		case <-ctx.Done():
			done = true
		case <-timer.C:
			send <- content
		}
	}
	ret <- content
}
