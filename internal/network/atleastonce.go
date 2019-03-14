package network

import (
	"context"
	"encoding/json"
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
	ret := make(chan atLeastOnceMsg)

	T := reflect.TypeOf(conf.send).Elem()
	if reflect.TypeOf(conf.receive).Elem() != T {
		log.Panic("Datatypes for send and receive not consistent")
	}

	atleastOnceTx, err := reflectchan2interfacechan(ctx, reflect.ValueOf(conf.send))
	if err != nil {
		log.Panicln("Error starting atleastonce: ", err)
	}

	c := config{
		send:    bSend,
		receive: bRecv,
		id:      conf.id,
		port:    conf.port,
	}
	go RunAtMostOnce(ctx, c)

	for {
		select {
		case <-ctx.Done():
			break
		case m := <-atleastOnceTx:
			msg := atLeastOnceMsg{
				Ack:       false,
				SenderID:  conf.id,
				MessageID: "TODO",
				Data:      m,
			}
			go sendUntilAck(ctx, msg, bSend, bRecv, ret)
		}

		select {
		case <-ctx.Done():
			break
		case r := <-ret:
			b, err := json.Marshal(r.Data)
			if err != nil {
				log.Println("error receiving message: ", err)
			}
			v := reflect.New(T)
			err = json.Unmarshal(b, v.Interface())
			if err != nil {
				log.Panicln("Failed unmarshal")
			}

		}
	}
}

func sendUntilAck(ctx context.Context, content atLeastOnceMsg, send chan<- atLeastOnceMsg, recv <-chan atLeastOnceMsg, ret <-chan atLeastOnceMsg) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(100 * time.Millisecond):
			send <- content

		}
	}

}
