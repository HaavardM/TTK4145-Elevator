package network

import (
	"context"
	"log"
	"reflect"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/utilities"
)

//RunAtMostOnce runs at most once publishing at a certain port
//Service is limited to one datatype per port
//We use reflection to allow multiple channel types. The network module does not care what the user want to send.
func RunAtMostOnce(ctx context.Context, conf Config) {
	//Create channels
	atMostOnceTx, err := utilities.ReflectChan2InterfaceChan(ctx, reflect.ValueOf(conf.Send))
	if err != nil {
		log.Panicln("Error starting AtMostOnce: ", err)
	}
	atMostOnceRx := make(chan interface{})
	defer close(atMostOnceRx)

	//Get datatype of send element
	T := reflect.TypeOf(conf.Send).Elem()
	if reflect.TypeOf(conf.Receive).Elem() != T {
		log.Panicf("Inconsistent types in AtMostOnce")
	}

	template := reflect.New(T).Interface()
	//Launch transmitter and receiver
	go broadcastTransmitter(ctx, conf.Port, conf.ID, atMostOnceTx)
	go broadcastReceiver(ctx, conf.Port, conf.ID, atMostOnceRx, template)

	//Create reflect select statement
	out := reflect.SelectCase{
		Dir:  reflect.SelectSend,
		Chan: reflect.ValueOf(conf.Receive),
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
