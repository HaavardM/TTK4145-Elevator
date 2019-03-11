package network

import (
	"context"
	"log"
	"reflect"
	"time"
)

type atLeastOnceMsg struct {
	Ack       bool        `json:"ack"`
	SenderID  int         `json:"sender_id"`
	MessageID string      `json:"message_id"`
	Data      interface{} `json:"data"`
}

//RunAtLeastOnce runs at most once publishing at a certain port
//Service is limited to one datatype per port
func RunAtLeastOnce(ctx context.Context, conf config) {
	bSend := make(chan atLeastOnceMsg)
	bRecv := make(chan atLeastOnceMsg)

	c := config{
		send:    bSend,
		receive: bRecv,
		id:      conf.id,
	}
	go RunAtMostOnce(ctx, c)

	T := reflect.TypeOf(conf.send)
	if reflect.TypeOf(conf.receive) != T {
		log.Panicln("Incompatible channel types")
	}

}

func sendAndWaitForAck(ctx context.Context, content atLeastOnceMsg, send chan<- atLeastOnceMsg, recv <-chan atLeastOnceMsg, ret <-chan atLeastOnceMsg) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(100 * time.Millisecond):
			send <- content

		}
	}

}
