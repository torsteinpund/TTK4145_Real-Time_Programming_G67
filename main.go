package main

import (
	// "encoding/json"

	//. "Driver-go/network/masterSelector"
	"Driver-go/master"
	"Driver-go/orderHandler"
	"fmt"

	// "time"

	"Driver-go/network/peers"
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
	// id = "13"
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
		Ch_stateUpdate:       make(chan Elevator),
		Ch_registerOrder:     make(chan OrderEvent, 10),
		Ch_orderCopyResponse: make(chan GlobalOrderMap),
		Ch_ordersFromMaster:  make(chan GlobalOrderMap),
		Ch_orderCopyRequest:  make(chan bool),
	}

	txChannels := network.RXChannels{
		Ch_stateUpdate:      make(chan Elevator),
		Ch_registerOrder:    make(chan OrderEvent, 10),
		Ch_ordersFromMaster: make(chan GlobalOrderMap),
	}

	masterChannels := master.MasterChannels{
		Ch_isMaster:         make(chan bool),
		Ch_peerLost:         make(chan string),
		Ch_networkToSlave:   Ch_netWorkMsg,
		Ch_registerOrder:    rxChannels.Ch_registerOrder,
		Ch_stateUpdate:      txChannels.Ch_stateUpdate,
		Ch_orderCopy:        rxChannels.Ch_orderCopyResponse,
		Ch_newPeer:          make(chan string),
		Ch_ordersFromMaster: txChannels.Ch_ordersFromMaster,
	}

	fsmChannels := fsm.FsmChannels{
		Ch_floorSensor:     hardwareChannels.Ch_floorSensor,
		Ch_stopButton:      hardwareChannels.Ch_stopButton,
		Ch_obstruction:     hardwareChannels.Ch_obstruction,
		Ch_localLights:     make(chan OrderMatrix),
		Ch_localOrders:     make(chan OrderMatrix),
		Ch_networkToMaster: Ch_netWorkMsg,
		Ch_clearedFloor:    make(chan ClearedFloorInfo),
		Ch_stateUpdate:     txChannels.Ch_stateUpdate,
	}

	clientChannels := network.ClientChannels{
		Ch_peerUpdate: make(chan peers.PeersUpdate),
		Ch_peerLost:   make(chan string),
		Ch_newPeer:    masterChannels.Ch_newPeer,
		Ch_isMaster:   masterChannels.Ch_isMaster,
	}

	orderChannels := orderHandler.OrderChannels{
		Ch_localOrders:       fsmChannels.Ch_localOrders,
		Ch_localLights:       fsmChannels.Ch_localLights,
		Ch_clearedFloor:      fsmChannels.Ch_clearedFloor,
		Ch_orderFromMaster:   rxChannels.Ch_ordersFromMaster,
		Ch_networkToSlave:    masterChannels.Ch_networkToSlave,
		Ch_networkToMaster:   Ch_netWorkMsg,
		Ch_buttonPress:       hardwareChannels.Ch_buttonPress,
		Ch_registerOrder:     txChannels.Ch_registerOrder,
		Ch_orderCopyResponse: rxChannels.Ch_orderCopyResponse,
		Ch_orderCopyRequest:  rxChannels.Ch_orderCopyRequest,
	}

	// doorChannels := fsm.DoorChannels{
	// 	Ch_doorOpen: fsmChannels.Ch_doorOpen,
	// 	Ch_toSlave: masterChannels.Ch_toSlave,
	// }

	// go peers.Transmitter(19191, )

	// elevio.SetButtonLamp(ButtonType(1), 0, true)

	network.InitNettwork(rxChannels, Ch_netWorkMsg, 18191, id, Ch_peerTxEnable, clientChannels)
	Ch_peerTxEnable <- true
	go bcast.Transmitter(19191,
		txChannels.Ch_stateUpdate,      // State updates to transmit
		txChannels.Ch_registerOrder,    // Orders to transmit
		txChannels.Ch_ordersFromMaster) // Master orders to transmit

	go bcast.Receiver(19191,
		rxChannels.Ch_stateUpdate,      // State updates to receive
		rxChannels.Ch_registerOrder,    // Orders to receive
		rxChannels.Ch_ordersFromMaster) // Master orders to receive
	// for{
	// 	txChannels.Ch_stateUpdate <- elevator

	// 	// test<- elevator
	// 	fmt.Println("Test: ")

	// 	read:= <- rxChannels.Ch_stateUpdate

	// 	fmt.Println("Read: ", read)

	// 	time.Sleep(2*time.Second)

	// }

	// txChannels.Ch_stateUpdate <- elevator

	// // test<- elevator
	// fmt.Println("Test: ")

	// read:= <- rxChannels.Ch_stateUpdate

	// fmt.Println("Read: ", read)
	fmt.Println("equal?: ,",rxChannels.Ch_registerOrder == masterChannels.Ch_registerOrder)
	// client := network.NewClient(id)
	go master.RunMaster(elevator.ID, masterChannels)
	// go client.RunClient(elevator.ID, clientChannels)
	go fsm.FsmRun(fsmChannels, elevator)
	go orderHandler.OrderHandler(orderChannels, elevator.ID)
	// go lights.SetHallLights(orderChannels.Ch_localLights)

	// go func() {
	// 	masterChannels.Ch_newPeer <- elevator.ID

	// }()

	// timetest := 5 * time.Second

	// for {
	// 	select {
	// 	case p := <-rxChannels.Ch_stateUpdate:
	// 		fmt.Println("Received from network: ", p.ID)
	// 		fmt.Println("Received from networjk: ", p.Floor)
	// // 		// case <-time.After(timetest):
	// // 		// 	masterChannels.Ch_isMaster <- false
	// // 		// 	fmt.Println("I am not master")
	// // 		// 	time.Sleep(1 * time.Second)
	// // 		// 	globalOrderMap := GlobalOrderMap{
	// // 		// 		"12": OrderMatrix{
	// // 		// 			{true, false, false}, // Etasje 0: Opp, Ned, Kabin
	// // 		// 			{false, true, false}, // Etasje 1: Opp, Ned, Kabin
	// // 		// 			{false, false, true}, // Etasje 2: Opp, Ned, Kabin
	// // 		// 		},
	// // 		// 		"elevator2": OrderMatrix{
	// // 		// 			{false, false, true}, // Etasje 0: Opp, Ned, Kabin
	// // 		// 			{true, false, false}, // Etasje 1: Opp, Ned, Kabin
	// // 		// 			{false, true, false}, // Etasje 2: Opp, Ned, Kabin
	// // 		// 		},
	// // 		// 	}
	// // 		// 	netMsg := NetworkMessage{
	// // 		// 		MsgType: "ordersfrommaster",
	// // 		// 		MsgData: globalOrderMap,
	// // 		// 	}
	// // 		// 	masterChannels.Ch_networkToSlave <- netMsg
	// // 		// default:

	// 	}
	go func ()  {
		select {
		case p := <-rxChannels.Ch_registerOrder:	
			fmt.Println("Received order: ", p)
		}
	}()
	// }
	select {}
}
