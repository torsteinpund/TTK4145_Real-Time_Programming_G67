package main

import (
	// "encoding/json"

	//. "Driver-go/network/masterSelector"
	"Driver-go/master"
	"Driver-go/orderHandler"
	"fmt"

	"Driver-go/network/peers"
	"Driver-go/singleElevatorDriver/elevio"
	"Driver-go/singleElevatorDriver/fsm"

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

	fmt.Println("ID: ", id)

	elevio.InitHardwareConnection("localhost:15657", hardwareChannels)
	elevator := elevio.InitElevator(NUMFLOORS, NUMBUTTONTYPE, Elevator{}, id)

	Ch_netWorkMsg := make(chan NetworkMessage)
	Ch_peerTxEnable := make(chan bool)

	// // If the elevator starts at a valid floor, initialize its state
	// elevator = fsm.FsmFloorArrival(elevio.GetFloor(), elevator)
	// fmt.Println("Elevator initialized DONE")
	// fmt.Println(elevator.Avaliable)

	rxChannels := network.RXChannels{
		Ch_stateUpdate:   make(chan Elevator),
		Ch_orderUpdate:   make(chan OrderMatrix),
		Ch_registerOrder: make(chan OrderEvent),
		Ch_orderCopyResponse:  make(chan GlobalOrderMap),
		Ch_ordersFromMaster: make(chan GlobalOrderMap),
	}

	masterChannels := master.MasterChannels{
		Ch_isMaster:          make(chan bool),
		Ch_peerLost:          make(chan string),
		Ch_toSlave:           Ch_netWorkMsg,
		Ch_registerOrder:     rxChannels.Ch_registerOrder,
		Ch_stateUpdate:       rxChannels.Ch_stateUpdate,
		Ch_orderCopyResponse: rxChannels.Ch_orderCopyResponse,
		Ch_newPeer:           make(chan string),
	}

	fsmChannels := fsm.FsmChannels{
		Ch_floorSensor:  hardwareChannels.Ch_floorSensor,
		Ch_stopButton:   hardwareChannels.Ch_stopButton,
		Ch_obstruction:  hardwareChannels.Ch_obstruction,
		Ch_localLights:  make(chan OrderMatrix),
		Ch_localOrders:  make(chan OrderMatrix),
		Ch_toMaster:     Ch_netWorkMsg,
		Ch_clearedFloor: make(chan int),
		Ch_stateUpdate:  rxChannels.Ch_stateUpdate,
	}

	clientChannels :=  network.ClientChannels{
		Ch_peerUpdate: make(chan peers.PeersUpdate),
		Ch_peerLost:   make(chan string),
		Ch_newPeer:    masterChannels.Ch_newPeer,
		Ch_isMaster:   masterChannels.Ch_isMaster,
	}

	orderChannels := orderHandler.OrderChannels{
		Ch_localOrders:     	 fsmChannels.Ch_localOrders,
		Ch_localLights:     	 fsmChannels.Ch_localLights,
		Ch_clearedFloor:    	 fsmChannels.Ch_clearedFloor,
		Ch_orderFromMaster: 	 masterChannels.Ch_orderCopyResponse,
		Ch_toSlave:         	 masterChannels.Ch_toSlave,
		Ch_toMaster:        	 Ch_netWorkMsg,
		Ch_buttonPress:     	 hardwareChannels.Ch_buttonPress,
		Ch_registerOrder:   	 rxChannels.Ch_registerOrder,
		Ch_orderCopyResponse: 	 rxChannels.Ch_orderCopyResponse,
	}

	// doorChannels := fsm.DoorChannels{
	// 	Ch_doorOpen: fsmChannels.Ch_doorOpen,
	// 	Ch_toSlave: masterChannels.Ch_toSlave,
	// }

	// go peers.Transmitter(19191, )

	// elevio.SetButtonLamp(ButtonType(1), 0, true)

	network.InitNettwork(rxChannels, Ch_netWorkMsg, 19191, id, Ch_peerTxEnable, clientChannels)
	Ch_peerTxEnable <- true
	go bcast.Transmitter(19191, rxChannels.Ch_stateUpdate, rxChannels.Ch_orderUpdate, rxChannels.Ch_registerOrder, rxChannels.Ch_ordersFromMaster)
	go bcast.Receiver(19191, rxChannels.Ch_stateUpdate, rxChannels.Ch_orderUpdate, rxChannels.Ch_registerOrder, rxChannels.Ch_ordersFromMaster)
	

	go master.RunMaster(elevator.ID, masterChannels)
	// go client.RunClient(elevator.ID,clientChannels)
	go fsm.FsmRun(fsmChannels, elevator)
	go orderHandler.OrderHandler(orderChannels, elevator.ID)
	// go lights.SetHallLights(orderChannels.Ch_localLights)

	go func() {
		masterChannels.Ch_newPeer <- elevator.ID

	}()
	for {
		select {
		case p := <-rxChannels.Ch_stateUpdate:
			fmt.Println("Received from network: ", p.ID)
			fmt.Println("Received from networjk: ", p.Floor)

		default:

		}

	}
}
