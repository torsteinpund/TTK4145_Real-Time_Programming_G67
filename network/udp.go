package network

import (
	"fmt"
	"net"
	"time"
)

type Data struct{
	Msg string
	Addr string
}


func receiveUDP(){
	buffer := make([]byte, 1024) //create a buffer with size 1024 byte (message size)

	Addr := net.UDPAddr{ //create a socket with port 3000 and IP address 0.0.0.0
		Port: 30000,
		IP: net.ParseIP("0.0.0.0"),
	}
	connection, err := net.ListenUDP("udp", &Addr) //listens to the UDP socket from the IP and address of ADDR
	if err != nil { //error handler
		fmt.Println("Something wrong with the UDP connection:", err)
		return
	}
	defer connection.Close() //scheduling makes sure the socket closes no matter what, even if an error occurs
	for { //Reads from connected UDP and returns number of bytes in data, port and address of client, and if any error
		numbersOfBytes, remoteAddress, err := connection.ReadFromUDP(buffer) 
		if err != nil{ //error handler
			fmt.Println("Could not read data into the buffer:", err)
			continue
		} 
		fmt.Printf("The data received from: %s %d %s\n ", remoteAddress.IP, remoteAddress.Port, string(buffer[:numbersOfBytes]))
	}
}

func sendUDP(){
	receiverAddr := net.UDPAddr{ //create a socket with port 3000 and IP address 0.0.0.0
		Port: 20002,
		IP: net.ParseIP("10.100.23.204"),
	}
	connection, err := net.DialUDP("udp", nil, &receiverAddr)
	if err != nil {
		fmt.Println("Could not send to UPD port", receiverAddr.Port, "error: ", err) 
	}
	defer connection.Close()
	message := []byte("Hello, chopper")
	for{
		_, err = connection.Write(message)
		if err != nil {
			fmt.Println("Error sending message:", err)
			continue
		}
		fmt.Println("message sendt")
		time.Sleep(4*time.Second)
	}
}