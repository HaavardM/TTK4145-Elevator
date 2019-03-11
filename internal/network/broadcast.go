package network

import (
	"context"
	"encoding/json"
	"log"
	"reflect"
	"time"
)

func broadcastReceiver(ctx context.Context, port int, message chan<- interface{}, T reflect.Type) {
	noConn := make(chan error)
	conn, _, err := createConn(port)
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
			err := conn.Close()
			if err != nil {
				log.Panicf("Can't close broadcast")
			}
			break
		}

		n, _, err := conn.ReadFrom(buf[0:])
		if err != nil {
			log.Println("Failed to read - reconnecting")
			noConn <- err
			continue
		}

		v := reflect.New(T)
		//TODO Error handling?
		json.Unmarshal(buf[0:n], &v)
		message <- v

	}
}

func broadcastTransmitter(ctx context.Context, port int, message chan interface{}, T reflect.Type) {
	noConn := make(chan error)
	conn, addr, err := createConn(port)
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
			err := conn.Close()
			if err != nil {
				log.Panicf("Can't close broadcast")
			}
			break

		case m := <-message:
			if reflect.TypeOf(m) != T {
				log.Panicf("Type published not consistet with receiver")
			}
			data, err := json.Marshal(m)
			if err != nil {
				log.Println("Couldn't marshal message ", err)
			}
			_, err = conn.WriteTo(data, addr)
			if err != nil {
				log.Println("Failed to write - attempting reconnect")
				noConn <- err
				//Do not skip message
				go SendMessage(ctx, message, m)
			}
		}
	}
}
