package network

import (
	"context"
	"log"
	"net"
	"time"

	"fmt"

	"github.com/TTK4145/Network-go/network/conn"
)

type config struct {
	id      int
	port    int
	send    interface{}
	receive interface{}
}

type Message struct {
	MSG string `json:"msg"`
}

func Run(ctx context.Context) {
	s := make(chan Message)
	r := make(chan Message)
	conf := config{
		id:      1,
		port:    18843,
		send:    s,
		receive: r,
	}

	go RunAtMostOnce(ctx, conf)
	for {
		select {
		case <-time.After(1 * time.Second):
			s <- Message{MSG: "Hello there"}
		case ret := <-r:
			fmt.Println("Ret: ", ret)
		}
	}
}

func createConn(port int) (net.PacketConn, *net.UDPAddr, error) {
	conn := conn.DialBroadcastUDP(port)
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))
	if err != nil {
		log.Panicf("Can't connect to net %s", err)
		return nil, nil, err
	}
	return conn, addr, nil
}

//SendMessage attempts to send a message on a chan
func SendMessage(ctx context.Context, c chan<- interface{}, m interface{}) {
	select {
	case <-ctx.Done():
		break
	case c <- m:
		break
	}
}
