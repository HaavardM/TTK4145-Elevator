package network

import (
	"context"
	"encoding/json"
	"log"
	"reflect"
	"time"

	"github.com/TTK4145-students-2019/project-thefuturezebras/pkg/utilities"
)

type broadcastMsg struct {
	SenderID int         `json:"sender_id"`
	Data     interface{} `json:"data"`
}

//broadcastReceiver receives JSON messages from a UDP broadcast port and unmarshalls into template
func broadcastReceiver(ctx context.Context, port int, id int, message chan<- interface{}, T reflect.Type) {
	noConn := make(chan error)
	defer close(noConn)
	conn, _, err := createConn(port)
	//Close connection on exit
	defer conn.Close()

	if err != nil {
		noConn <- err
	}
	var buf [1024]byte
	for {
		select {
		case <-noConn:
			//Wait before retry
			<-time.After(1 * time.Second)
			//Close old
			conn.Close()
			//Dial new
			conn, _, err = createConn(port)
			if err != nil {
				noConn <- err
			}
		case <-ctx.Done():
			return
		default:
		}

		n, _, err := conn.ReadFrom(buf[0:])
		if err != nil {
			log.Println("Failed to read - reconnecting")
			noConn <- err
			continue
		}

		//Create message template
		msg := &broadcastMsg{
			SenderID: -1,
			Data:     reflect.New(T).Interface(),
		}
		//fmt.Printf("%s\n", buf[0:n])
		err = json.Unmarshal(buf[0:n], msg)
		if err != nil {
			log.Println(err)
		}
		if msg.SenderID != id || msg.SenderID < 0 {
			if msg.Data != nil {
				go utilities.SendMessage(ctx, message, msg.Data)
			}
		}

	}
}

//broadcastTransmitter transmits JSONs messages to a UDP broadcast port
func broadcastTransmitter(ctx context.Context, port int, id int, message <-chan interface{}) {
	noConn := make(chan error)
	transmitQueue := utilities.RChan2RWChan(ctx, message)
	conn, addr, err := createConn(port)
	defer conn.Close()
	if err != nil {
		go func() {
			noConn <- err
		}()
	}

	for {
		select {
		case <-noConn:
			//Wait before retry
			<-time.After(1 * time.Second)
			//Close old
			conn.Close()
			//Dial new
			conn, addr, err = createConn(port)
			if err != nil {
				noConn <- err
			}
		case <-ctx.Done():
			return
		case m := <-transmitQueue:
			data, err := json.Marshal(broadcastMsg{
				Data:     m,
				SenderID: id,
			},
			)
			if err != nil {
				log.Println("Couldn't marshal message ", err)
				continue
			}
			_, err = conn.WriteTo(data, addr)
			if err != nil {
				log.Println("Failed to write - attempting reconnect")
				go utilities.SendMessage(ctx, noConn, err)
				//Do not skip message
				go utilities.SendMessage(ctx, transmitQueue, m)
			}
		}
	}
}
