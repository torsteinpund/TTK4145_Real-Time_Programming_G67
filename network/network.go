package network

import (
	"Driver-go/network/peers"
	. "Driver-go/types"

	// "fmt"

	// "net"
	"strings"
)

type RXChannels struct {
    Ch_stateUpdate     		chan Elevator       	`addr:"rx_elevatorupdatechannel"`
    Ch_registerOrder   		chan OrderEvent     	`addr:"rx_registerorderchannel"`
    Ch_orderCopyResponse  	chan GlobalOrderMap 	`addr:"rx_ordercopyresponse"`
    Ch_orderCopyRequest 	chan bool 				`addr:"rx_ordercopyrequest"`
    Ch_ordersFromMaster 	chan GlobalOrderMap 	`addr:"rx_ordersfrommaster"`
}

type TXChannels struct {
    Ch_stateUpdate     		chan Elevator       	`addr:"tx_elevatorupdatechannel"`
    Ch_orderEventToMaster   		chan OrderEvent     	`addr:"tx_registerorderchannel"`
    Ch_ordersFromMaster 	chan GlobalOrderMap 	`addr:"tx_ordersfrommaster"`
}

func InitNettwork(ch_RX RXChannels, detectionPort int, id string, ch_transmitEnable <-chan bool, ch_isMaster chan<- bool, ch_peerLost chan<- string, ch_newPeer chan<- string) {

	ch_peerUpdate := make(chan peers.PeersUpdate)
	go peers.Transmitter(detectionPort, id, ch_transmitEnable)
	go peers.Receiver(detectionPort, ch_peerUpdate)

	c := NewClient(id) // This should maybe be in the main and passed as an argument instead
	go c.RunClient(id, ch_RX, ch_peerUpdate, ch_peerLost, ch_newPeer, ch_isMaster)

}

func GetID(ipAdress string) string {
	parts := strings.Split(ipAdress, ".")

	return parts[len(parts)-1]
}
