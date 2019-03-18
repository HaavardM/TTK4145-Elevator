package network

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/TTK4145-students-2019/project-thefuturezebras/internal/utilities"
)

//broadcastReceiver receives JSON messages from a UDP broadcast port and unmarshalls into template
func broadcastReceiver(ctx context.Context, port int, id int, message chan<- interface{}, template interface{}) {
	noConn := make(chan error)
	conn, _, err := createConn(port)
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

		//Copy template
		msg := template

		//Create message template
		err = json.Unmarshal(buf[0:n], &msg)
		if err != nil {
			log.Println("Invalid json format in message")
		} else {
			message <- msg
		}

	}
}

//
func broadcastTransmitter(ctx context.Context, port int, id int, message <-chan interface{}) {
	noConn := make(chan error)
	transmitQueue := utilities.RChan2RWChan(ctx, message)
	conn, addr, err := createConn(port)
	defer conn.Close()
	if err != nil {
		noConn <- err
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
			data, err := json.Marshal(m)
			if err != nil {
				log.Println("Couldn't marshal message ", err)
				continue
			}
			_, err = conn.WriteTo(data, addr)
			if err != nil {
				log.Println("Failed to write - attempting reconnect")
				noConn <- err
				//Do not skip message
				go SendMessage(ctx, transmitQueue, m)
			}
		}
	}
}
