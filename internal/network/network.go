package network

import (
	"context"
	"log"
	"net"

	"fmt"

	"github.com/TTK4145/Network-go/network/conn"
)

//Config contains common information for starting network modules
type Config struct {
	//ID is the unique id of this node
	ID int
	//Port is the UDP port number to use for communication
	Port int
	//Send is a channel to use for sending to network: Must be a channel!!
	Send interface{}
	//Receive is a channel used to receive from network. Must be same type as Send!
	Receive interface{}
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
