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

//RunAtLeastOnce runs at most once publishing at a certain port
//Service is limited to one datatype per port
func RunAtLeastOnce(ctx context.Context, conf Config) {
	bSend := make(chan atLeastOnceMsg)
	bRecv := make(chan atLeastOnceMsg)
	ret := make(chan atLeastOnceMsg)

	T := reflect.TypeOf(conf.Send).Elem()
	if reflect.TypeOf(conf.Receive).Elem() != T {
		log.Panic("Datatypes for send and receive not consistent")
	}

	atleastOnceTx, err := utilities.ReflectChan2InterfaceChan(ctx, reflect.ValueOf(conf.Send))
	if err != nil {
		log.Panicln("Error starting atleastonce: ", err)
	}

	c := Config{
		Send:    bSend,
		Receive: bRecv,
		ID:      conf.ID,
		Port:    conf.Port,
	}
	go RunAtMostOnce(ctx, c)

	for {
		select {
		case <-ctx.Done():
			break
		case m := <-atleastOnceTx:
			msg := atLeastOnceMsg{
				Ack:       false,
				SenderID:  conf.ID,
				MessageID: "TODO",
				Data:      m,
			}
			go sendUntilAck(ctx, msg, bSend, bRecv, ret)
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

func sendUntilAck(ctx context.Context, content atLeastOnceMsg, send chan<- atLeastOnceMsg, recv <-chan atLeastOnceMsg, ret chan<- atLeastOnceMsg) {
	received := make(map[int]struct{})
	acksExpected := []int{}
	done := false
	//While not received all acks
	for !done {
		select {
		case <-ctx.Done():
			return
		case <-time.After(100 * time.Millisecond):
			send <- content
		case m := <-recv:
			if m.Ack {
				received[m.SenderID] = struct{}{}
			}
		}
		for _, id := range acksExpected {
			done = true
			if _, present := received[id]; !present {
				done = false
				break
			}
		}
	}
	ret <- content

}
