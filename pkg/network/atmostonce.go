package network

import (
	"context"
	"log"
	"reflect"

	"github.com/TTK4145-students-2019/project-thefuturezebras/pkg/utilities"
)

//AtMostOnceConfig is a configuration struct used to configure atMostOnce
type AtMostOnceConfig struct {
	Config
	//Send is the channel used to send data to the network. Must be of same type as Recv
	Send interface{}
	//Receive is the channel used to receive data from the network. Must be of same type as Send
	Receive interface{}
}

//RunAtMostOnce runs at most once publishing at a certain port
//Service is limited to one datatype per port
//We use reflection to allow multiple channel types. The network module does not care what the user want to send.
func RunAtMostOnce(ctx context.Context, conf AtMostOnceConfig) {
	//Signal done when ready to exit
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

	//Get channel from reflect
	outChan := reflect.ValueOf(conf.Receive)

	//Wait for broadcast goroutines
	//Create template used for Unmarshalling
	//Launch transmitter and receiver
	go broadcastTransmitter(ctx, conf.Port, conf.ID, atMostOnceTx)
	go broadcastReceiver(ctx, conf.Port, conf.ID, atMostOnceRx, T)

	//Wait for completion
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-atMostOnceRx:
			valuePtr := reflect.ValueOf(m)           //Pointer type
			outChan.Send(reflect.Indirect(valuePtr)) //Get actual value
		}

	}
}
