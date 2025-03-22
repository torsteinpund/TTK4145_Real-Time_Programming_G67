package network

import (
	"Driver-go/network/peers"
	. "Driver-go/types"

	// "fmt"

	// "net"
	"strings"
)

type RXChannels struct {
	Ch_stateUpdate     		chan Elevator       	`addr:"elevatorupdatechannel"`
	Ch_registerOrder   		chan OrderEvent     	`addr:"registerorderchannel"`
	Ch_orderCopyResponse  	chan GlobalOrderMap 	`addr:"ordercopyresponse"`
	Ch_orderCopyRequest 	chan bool 				`addr:"ordercopyrequest"`
	Ch_ordersFromMaster 	chan GlobalOrderMap 	`addr:"ordersfrommaster"`
}

type TXChannels struct {
	Ch_stateUpdate     		chan Elevator       	`addr:"elevatorupdatechannel"`
	Ch_registerOrder   		chan OrderEvent     	`addr:"registerorderchannel"`
	Ch_ordersFromMaster 	chan GlobalOrderMap 	`addr:"ordersfrommaster"`
}

func InitNettwork(ch_RX RXChannels, Ch_netWorkMsg <-chan NetworkMessage, detectionPort int, id string, ch_transmitEnable <-chan bool, ch_Client ClientChannels) {
	// Initialize client
	// peerUpdateChannel := make(chan peers.PeersUpdate)

	go peers.Transmitter(detectionPort, id, ch_transmitEnable)
	go peers.Receiver(detectionPort, ch_Client.Ch_peerUpdate)

	c := NewClient(id) // This should maybe be in the main and passed as an argument instead
	go c.RunClient(id,ch_Client,ch_RX, Ch_netWorkMsg)

}

func GetID(ipAdress string) string {
	parts := strings.Split(ipAdress, ".")

	return parts[len(parts)-1]
}
