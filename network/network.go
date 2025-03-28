package network

import (
	. "Driver-go/types"
	"net"
	"strings"
	"time"
)

type RXChannels struct {
	Ch_stateUpdate        chan Elevator       `addr:"rx_elevatorupdatechannel"`
	Ch_registerOrder      chan OrderEvent     `addr:"rx_registerorderchannel"`
	Ch_ordersFromMaster   chan GlobalOrderMap `addr:"rx_ordersfrommaster"`
}

type TXChannels struct {
	Ch_stateUpdate        chan Elevator       `addr:"tx_elevatorupdatechannel"`
	Ch_orderEventToMaster chan OrderEvent     `addr:"tx_registerorderchannel"`
	Ch_ordersFromMaster   chan GlobalOrderMap `addr:"tx_ordersfrommaster"`
}

func getLocalIP() (string, error) {
	var localIP string
	if localIP == "" {
		conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
		if err != nil {
			return "", err
		}
		defer conn.Close()
		localIP = strings.Split(conn.LocalAddr().String(), ":")[0]
	}
	return localIP, nil
}

func checkConnection() bool {
	_, err := getLocalIP()
	return err == nil
}

func PollNetworkConnection(networkConnection chan<- bool) {
	var wasDisconnected bool
	networkConnection <- checkConnection()
	for {
		if !checkConnection() {
			if !wasDisconnected {
				networkConnection <- false
				wasDisconnected = true
			}
		} else {
			if wasDisconnected {
				networkConnection <- true
				wasDisconnected = false
			}
		}
		time.Sleep(2 * time.Second)
	}
}
