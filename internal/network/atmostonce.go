package network

import (
	"context"
	"log"
	"reflect"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/utilities"
)

type atMostOnceMsg struct {
	SenderID int         `json:"sender_id"`
	Data     interface{} `json:"data"`
}

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

	//Get channel from reflect
	outChan := reflect.ValueOf(conf.Receive)

	//Create template used for Unmarshalling
	template := atMostOnceMsg{
		SenderID: conf.ID,
		Data:     reflect.New(T).Interface(),
	}
	//Launch transmitter and receiver
	go broadcastTransmitter(ctx, conf.Port, conf.ID, atMostOnceTx)
	go broadcastReceiver(ctx, conf.Port, conf.ID, atMostOnceRx, template)

	for {
		select {
		case <-ctx.Done():
			return
		case m := <-atMostOnceRx:
			if v, ok := m.(atMostOnceMsg); ok {
				valuePtr := reflect.ValueOf(v.Data)      //Pointer type
				outChan.Send(reflect.Indirect(valuePtr)) //Get actual value
			}
		}

	}
}
