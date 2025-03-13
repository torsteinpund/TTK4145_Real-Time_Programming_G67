package main

import (
	// "encoding/json"

	//. "Driver-go/network/masterSelector"
	"Driver-go/master"
	"Driver-go/orderHandler"
	"fmt"

	//"os/exec"
	// "Driver-go/network/client"
	// "Driver-go/network/peers"
	"Driver-go/singleElevatorDriver/elevio"
	"Driver-go/singleElevatorDriver/fsm"

	// "Driver-go/network/bcast"
	// "Driver-go/network/conn"
	// "Driver-go/network/localip"
	"Driver-go/lights"
	. "Driver-go/types"
)

func main() {
    elevio.InitHardwareConnection("localhost:15657")
	elevator := elevio.InitElevator(NUMFLOORS, NUMBUTTONTYPE, Elevator{})
    fmt.Println("Elevator initialized DONE")
	masterChannels := master.MasterChannels{
		Ch_isMaster:          make(chan bool),
		Ch_peerLost:          make(chan string),
		Ch_toSlave:           make(chan NetworkMessage),
		Ch_registerOrder:     make(chan OrderEvent),
		Ch_stateUpdate:       make(chan Elevator),
		Ch_orderCopyResponse: make(chan GlobalOrderMap),
		Ch_registeredPeer:    make(chan string),
		Ch_toSlaveTest:       make(chan GlobalOrderMap),
	}

	fsmChannels := fsm.FsmChannels{
		Ch_buttonPress: make(chan ButtonEvent),
		Ch_floorSensor: make(chan int),
		Ch_stopButton:  make(chan bool),
		Ch_obstruction: make(chan bool),
		Ch_stateUpdate: masterChannels.Ch_stateUpdate,
	}

	// peerChannels := peers.PeerChannels{
	// 	PeerUpdateChannel: 			make(chan peers.PeersUpdate),
	// 	PeerLostChannel: 			make(chan string),
	// 	PeerNewChannel: 			make(chan string),
	// 	RegisteredNewPeerChannel: 	make(chan string),
	// }

	// clientChannels := client.ClientChannels{
	// 	InputChannel: 				make(chan NetworkMessage),
	// 	OutputChannel: 				make(chan NetworkMessage),
	// 	PeerUpdateChannel: 			make(chan peers.PeersUpdate),
	// 	PeerLostChannel: 			make(chan string),
	// 	PeerNewChannel: 			make(chan string),
	// 	IsMasterChannel: 			make(chan bool),
	// 	RegisteredNewPeerChannel: 	make(chan string),
	// }

	orderChannels := orderHandler.OrderChannels{
		LocalOrderChannel:       make(chan OrderMatrix),
		LocalLightsChannel:       make(chan OrderMatrix),
		OrdersFromMasterChannel: make(chan GlobalOrderMap),
		OrdersToMasterChannel:   make(chan NetworkMessage),
		ButtonEventChannel:      fsmChannels.Ch_buttonPress,
		FinishedFloorChannel:    make(chan int),
		Ch_registerOrder:        masterChannels.Ch_registerOrder,
		Ch_toSlave:              masterChannels.Ch_toSlave,
		Ch_toSlaveTest:          masterChannels.Ch_toSlaveTest,
	}
    elevio.SetButtonLamp(ButtonType(1), 0, true)
	// client := client.NewClient(elevator.ID)
    
	go master.RunMaster(elevator.ID, masterChannels)
	// go client.RunClient(elevator.ID,clientChannels)
	go fsm.FsmRun(fsmChannels, elevator)
	go orderHandler.OrderHandler(orderChannels, elevator.ID)
    go lights.SetHallLights(orderChannels.LocalLightsChannel)
	go func() {
		masterChannels.Ch_registeredPeer <- elevator.ID
	}()
	select {}

}
