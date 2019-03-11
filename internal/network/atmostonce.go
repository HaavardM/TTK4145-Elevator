package network

import (
	"context"
	"log"
	"reflect"
)

//RunAtMostOnce runs at most once publishing at a certain port
//Service is limited to one datatype per port
func RunAtMostOnce(ctx context.Context, conf config) {

	//Create channels
	atMostOnceTx := make(chan interface{})
	atMostOnceRx := make(chan interface{})
	defer close(atMostOnceRx)

	//Get datatype of send element
	T := reflect.TypeOf(conf.send)
	if reflect.TypeOf(conf.receive) != T {
		log.Panicf("Inconsistent types in AtMostOnce")
	}

	//Launch transmitter and receiver
	go broadcastTransmitter(ctx, conf.port, atMostOnceTx, T)
	go broadcastReceiver(ctx, conf.port, atMostOnceRx, T)

	//Create reflect select statement
	out := reflect.SelectCase{
		Dir:  reflect.SelectSend,
		Chan: reflect.ValueOf(conf.receive),
	}

	//Convert between channel types
	go reflect2chan(ctx, reflect.ValueOf(conf.send), atMostOnceTx)

	//Wait for completion
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-atMostOnceRx:
			out.Send = reflect.ValueOf(m)
			reflect.Select([]reflect.SelectCase{out})
		}

	}
}
