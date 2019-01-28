package main

import (
	"fmt"
	"net"
	"time"
)

func recieve(){
	buffer := make([]byte, 1024)

	serverAddr, err := net.ResolveUDPAddr("udp", ":20008")
	if err != nil {
		fmt.Println(err)
		return
	}

	
	conn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer conn.Close()


	for {

		n, addr, err := conn.ReadFromUDP(buffer)

		if err != nil {
		fmt.Println(err)
		return
		}

		fmt.Println(string(buffer[0:n]))
		fmt.Println("Recieved from:", addr)
		break
	}
}
func send(){
	time.Sleep(time.Second*1)

	serverAddr, err := net.ResolveUDPAddr("udp", "10.100.23.242:20008")
	if err != nil {
		fmt.Println(err)
		return
	}
	connection, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	connection.Write([]byte("This is a message "))
	if err != nil {
		fmt.Println(err)
		return
	}
}


func main() {
	go recieve()
	go send()
	time.Sleep(time.Second*5)
}
