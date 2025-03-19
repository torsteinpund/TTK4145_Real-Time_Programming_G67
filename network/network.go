package network

import(
	. "Driver-go/types"
	"Driver-go/network/client"
	"Driver-go/network/peers"
	// "net"
)

type RXChannels struct {
	ElevatorUpdateChannel 	chan Elevator			`addr:"elevatorupdatechannel"`
	OrderUpdateChannel 		chan OrderMatrix		`addr:"orderupdatechannel"`
	RegisterOrderChannel 	chan OrderEvent			`addr:"registerorderchannel"`
	OrderCopyResponse 		chan GlobalOrderMap		`addr:"ordercopyresponse"`
	OrdersFromMaster 		chan GlobalOrderMap	    `addr:"ordersfrommaster"`
}

func InitNettwork(ch_RX RXChannels, port int, id string, ch_transmitEnable <-chan bool, ch_Client client.ClientChannels) {
	// Initialize client
	peerUpdateChannel := make(chan peers.PeersUpdate)
	

	go peers.Transmitter(port, id, ch_transmitEnable)
	go peers.Receiver(port, peerUpdateChannel)

	c := client.NewClient(id) // This should maybe be in the main and passed as an argument instead
	go c.RunClient(id, ch_Client)

}

