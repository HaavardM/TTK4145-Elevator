package network

import (
	"context"
	"log"
	"net"
	"reflect"

	"fmt"

	"github.com/TTK4145/Network-go/network/conn"
)

//Config contains common information for starting network modules
type Config struct {
	//ID is the unique id of this node
	ID int
	//Port is the UDP port number to use for communication
	Port int
}

func createConn(port int) (net.PacketConn, *net.UDPAddr, error) {
	conn := conn.DialBroadcastUDP(port)
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("127.255.255.255:%d", port))
	if err != nil {
		log.Panicf("Can't connect to net %s", err)
		return nil, nil, err
	}
	return conn, addr, nil
}

//SendMessage attempts to send a message on a chan
func SendMessage(ctx context.Context, c interface{}, m interface{}) {
	channel := reflect.ValueOf(c)
	msg := reflect.ValueOf(m)

	selectCases := []reflect.SelectCase{
		reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ctx.Done()),
		},
		reflect.SelectCase{
			Dir:  reflect.SelectSend,
			Chan: channel,
			Send: msg,
		},
	}
	reflect.Select(selectCases)
}
