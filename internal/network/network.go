package network

import (
	"context"
	"log"
	"net"

	"fmt"

	"github.com/TTK4145/Network-go/network/conn"
)

type config struct {
	id      int
	port    int
	send    interface{}
	receive interface{}
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
