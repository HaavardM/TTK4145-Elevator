package network

import (
	"fmt"
	"log"
	"net"

	"github.com/TTK4145/Network-go/network/conn"
)

//Config contains common information for starting network modules
type Config struct {
	//ID is the unique id of this node
	ID int
	//Port is the UDP port number to use for communication
	Port int
}

//createConn creates an UDP broadcast connection and finds the connection address
func createConn(port int) (net.PacketConn, *net.UDPAddr, error) {
	conn := conn.DialBroadcastUDP(port)
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))
	if err != nil {
		log.Panicf("Can't connect to net %s", err)
		return nil, nil, err
	}
	return conn, addr, nil
}
