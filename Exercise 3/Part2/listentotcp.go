package main 

import (
	"fmt"
	"net"
)


func main() {
	buffer := make([]byte, 1024)
	serverAddr, err := net.ResolveTCPAddr("tcp", "10.100.23.242:34933")

	if err != nil {
		fmt.Println(err)
		return
	}

	conn, err := net.DialTCP("tcp", nil, serverAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	
	
	defer conn.Close()

	for {
		message := "Message \x00"
		fmt.Println("Sending: ", message, "\n" )
		conn.Write([]byte(message))

		conn.Read(buffer)
		fmt.Println("Recieved: ", string(buffer), "\n")
		break

	}
}