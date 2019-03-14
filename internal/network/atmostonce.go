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
	atMostOnceTx, err := reflectchan2interfacechan(ctx, reflect.ValueOf(conf.send))
	if err != nil {
		log.Panicln("Error starting AtMostOnce: ", err)
	}
	atMostOnceRx := make(chan interface{})
	defer close(atMostOnceRx)

	//Get datatype of send element
	T := reflect.TypeOf(conf.send).Elem()
	if reflect.TypeOf(conf.receive).Elem() != T {
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

	//Wait for completion
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-atMostOnceRx:
			valuePtr := reflect.ValueOf(m)            //Pointer type
			out.Send = reflect.Indirect(valuePtr)     //Get actual value
			reflect.Select([]reflect.SelectCase{out}) //Send on channel
		}

	}
}
