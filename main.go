package main

import (
	// "encoding/json"

	//. "Driver-go/network/masterSelector"
	"Driver-go/master"
	"Driver-go/orderHandler"
	"fmt"

	// "Driver-go/network/peers"
	"Driver-go/singleElevatorDriver/elevio"
	"Driver-go/singleElevatorDriver/fsm"

	// "time"

	// "Driver-go/network/conn"
	"Driver-go/network"
	"Driver-go/network/localip"

	"Driver-go/network/bcast"
	// "Driver-go/lights"
	. "Driver-go/types"
)

func main() {
	fmt.Println("Hello, World!")

	hardwareChannels := elevio.HardwareChannels{
		Ch_buttonPress: make(chan ButtonEvent),
		Ch_floorSensor: make(chan int),
		Ch_stopButton:  make(chan bool),
		Ch_obstruction: make(chan bool),
	}

	var IP string
	var id string
	if IP == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		IP = fmt.Sprintf(localIP)
		id = network.GetID(IP)
	}
	
	elevio.InitHardwareConnection("localhost:15657", hardwareChannels)
	elevator := elevio.InitElevator(NUMFLOORS, NUMBUTTONTYPE, Elevator{}, id)

	Ch_peerTxEnable := make(chan bool)
	Ch_isMaster := make(chan bool)
	Ch_peerLost := make(chan string)
	Ch_newPeer := make(chan string)
	Ch_localOrders := make(chan OrderMatrix)
	Ch_clearedFloor := make(chan DirnFloorPair)
	// Ch_orderCopyResponse := make(chan GlobalOrderMap)

	rxChannels := network.RXChannels{
		Ch_stateUpdate:       make(chan Elevator),
		Ch_registerOrder:     make(chan OrderEvent),
		Ch_ordersFromMaster:  make(chan GlobalOrderMap),
		Ch_orderCopyRequest:  make(chan bool),
	}

	txChannels := network.TXChannels{
		Ch_stateUpdate:      make(chan Elevator),
		Ch_orderEventToMaster:    make(chan OrderEvent),
		Ch_ordersFromMaster: make(chan GlobalOrderMap),
	}


	network.InitNettwork(rxChannels, 18191, id, Ch_peerTxEnable, Ch_isMaster, Ch_peerLost, Ch_newPeer)
	Ch_peerTxEnable <- true
	go bcast.Transmitter(19191, txChannels.Ch_stateUpdate, txChannels.Ch_orderEventToMaster, txChannels.Ch_ordersFromMaster)
	go bcast.Receiver(19191, rxChannels.Ch_stateUpdate, rxChannels.Ch_registerOrder, rxChannels.Ch_ordersFromMaster)


	go master.RunMaster(elevator.ID, Ch_isMaster, Ch_peerLost, txChannels.Ch_ordersFromMaster, rxChannels.Ch_registerOrder, rxChannels.Ch_stateUpdate, Ch_newPeer)
	go fsm.FsmRun(hardwareChannels.Ch_floorSensor, hardwareChannels.Ch_stopButton, hardwareChannels.Ch_obstruction, Ch_localOrders, Ch_clearedFloor, txChannels.Ch_stateUpdate, elevator)
	go orderHandler.OrderHandler(elevator.ID, Ch_localOrders, txChannels.Ch_orderEventToMaster, hardwareChannels.Ch_buttonPress, Ch_clearedFloor, rxChannels.Ch_ordersFromMaster)
	

	select{}
}
