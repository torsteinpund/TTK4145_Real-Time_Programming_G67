package main

import (
	"Driver-go/master"
	"Driver-go/network"
	"Driver-go/network/bcast"
	"Driver-go/network/peers"
	"Driver-go/orderHandler"
	"Driver-go/singleElevatorDriver/elevio"
	"Driver-go/singleElevatorDriver/fsm"
	. "Driver-go/types"
	"fmt"
	"flag"
)

func main() {
	fmt.Println("Hello, World!")


    // Define command-line flags
    var id string
    var port string
    flag.StringVar(&id, "id", "", "The ID of the elevator")
    flag.StringVar(&port, "port", "19091", "The port for the elevator hardware connection")

	// Standard port is 15657
    // Parse command-line flags
    flag.Parse()

    if id == "" {
        fmt.Println("ID is required")
        return
    }

    fmt.Println("ID: ", id)
    fmt.Println("Port: ", port)

	peerDetectionPort := 18191
	bcastPort := 19191

	// id := network.SetID()
	

	Ch_txEnable     := make(chan bool)
	Ch_isMaster     := make(chan bool)
	Ch_peerLost     := make(chan string)
	Ch_newPeer      := make(chan string)
	Ch_localOrders  := make(chan OrderMatrix)
	Ch_clearedFloor := make(chan DirnFloorPair,10)
	Ch_peerUpdate   := make(chan peers.PeersUpdate)
	Ch_orderCopyResponse := make(chan GlobalOrderMap)
	Ch_orderCopyRequest := make(chan bool)
	

	hardwareChannels := elevio.HardwareChannels{
		Ch_buttonPress: 	make(chan ButtonEvent),
		Ch_floorSensor: 	make(chan int),
		Ch_stopButton:  	make(chan bool),
		Ch_obstruction: 	make(chan bool),
	}

	rxChannels := network.RXChannels{
		Ch_stateUpdate:      make(chan Elevator),
		Ch_registerOrder:    make(chan OrderEvent),
		Ch_ordersFromMaster: make(chan GlobalOrderMap),
		Ch_orderCopyRequest: Ch_orderCopyRequest,
	}

	txChannels := network.TXChannels{
		Ch_stateUpdate:        make(chan Elevator),
		Ch_orderEventToMaster: make(chan OrderEvent),
		Ch_ordersFromMaster:   make(chan GlobalOrderMap),
	}

	elevio.InitHardwareConnection("localhost:"+port, hardwareChannels)
	elevator := elevio.InitElevator(NUMFLOORS, NUMBUTTONTYPE, Elevator{}, id)
	pH := network.NewPeerHandler(id)
	
	go peers.Transmitter(peerDetectionPort, id, Ch_txEnable)
	go peers.Receiver(peerDetectionPort, Ch_peerUpdate)
	Ch_txEnable <- true

	go pH.PeerHandler(id, 
					  rxChannels, 
					  Ch_peerUpdate, 
					  Ch_peerLost, 
					  Ch_newPeer, 
					  Ch_isMaster)

	go bcast.Transmitter(bcastPort, 
						 txChannels.Ch_stateUpdate, 
						 txChannels.Ch_orderEventToMaster, 
						 txChannels.Ch_ordersFromMaster)

	go bcast.Receiver(bcastPort, 
					  rxChannels.Ch_stateUpdate, 
					  rxChannels.Ch_registerOrder, 
					  rxChannels.Ch_ordersFromMaster)

	go master.Master(elevator.ID, 
					 Ch_isMaster, 
					 Ch_peerLost, 
					 txChannels.Ch_ordersFromMaster, 
					 rxChannels.Ch_registerOrder, 
					 rxChannels.Ch_stateUpdate, 
					 Ch_orderCopyResponse,
					 Ch_orderCopyRequest,
					 Ch_newPeer)

	go fsm.Fsm(hardwareChannels.Ch_floorSensor, 
			   hardwareChannels.Ch_stopButton, 
			   hardwareChannels.Ch_obstruction, 
			   Ch_localOrders, 
			   Ch_clearedFloor, 
			   txChannels.Ch_stateUpdate, 
			   elevator)

	go orderHandler.OrderHandler(elevator.ID,
								 Ch_localOrders, 
								 txChannels.Ch_orderEventToMaster, 
								 Ch_orderCopyResponse,
								 hardwareChannels.Ch_buttonPress, 
								 Ch_clearedFloor, 
								 rxChannels.Ch_ordersFromMaster,
								 Ch_orderCopyRequest)

	select {}
}
