package main

import (
	// "encoding/json"
	// "fmt"

	//. "Driver-go/network/masterSelector"
	"Driver-go/master"
	//"os/exec"
	"Driver-go/network/client"
	"Driver-go/network/peers"
	"Driver-go/singleElevatorDriver/elevio"
	"Driver-go/singleElevatorDriver/fsm"
	// "Driver-go/network/bcast"
	// "Driver-go/network/conn"
	// "Driver-go/network/localip"
	. "Driver-go/types"
)


func main(){
	elevator := elevio.InitElevator(NUMFLOORS, NUMBUTTONTYPE, Elevator{})
	masterChannels := master.MasterChannels{
		IsMasterChannel: 			make(chan bool),
		PeerLostChannel: 			make(chan string),
		ToSlavesChannel: 			make(chan NetworkMessage),
		RegisterOrderChannel: 		make(chan OrderEvent),
		StateUpdateChannel: 		make(chan Elevator),
		OrderCopyResponseChannel: 	make(chan GlobalOrderMap),
		RegisteredPeerChannel: 		make(chan string),
	}

	// peerChannels := peers.PeerChannels{
	// 	PeerUpdateChannel: 			make(chan peers.PeersUpdate),
	// 	PeerLostChannel: 			make(chan string), 
	// 	PeerNewChannel: 			make(chan string),
	// 	RegisteredNewPeerChannel: 	make(chan string),
	// }

	clientChannels := client.ClientChannels{
		InputChannel: 				make(chan NetworkMessage),
		OutputChannel: 				make(chan NetworkMessage),
		PeerUpdateChannel: 			make(chan peers.PeersUpdate),
		PeerLostChannel: 			make(chan string),
		PeerNewChannel: 			make(chan string),
		IsMasterChannel: 			make(chan bool),
		RegisteredNewPeerChannel: 	make(chan string),
	}

	client := client.NewClient(elevator.ID)

	go master.RunMaster(elevator.ID, masterChannels)
	go client.RunClient(elevator.ID,
		clientChannels.InputChannel,
		clientChannels.OutputChannel,
		clientChannels.PeerUpdateChannel,
		clientChannels.PeerLostChannel,
		clientChannels.PeerNewChannel,
		clientChannels.IsMasterChannel,
		clientChannels.RegisteredNewPeerChannel,)
	go fsm.FsmRun(masterChannels.StateUpdateChannel)
	
    // input := master.AllElevators{
    //     GlobalOrders: [NUMFLOORS][NUMHALLBUTTONS]bool{{false, false}, {true, false}, {false, false}, {false, true}},
    //     States: map[string]master.StateSingleElevator{
    //         "one": master.StateSingleElevator{
    //             ElevatorBehaviour:       "moving",
    //             Floor:          2,
    //             Direction:      "up",
	// 			Available: true,
    //             CabOrders:    [NUMFLOORS]bool{false, false, false, true},
				
    //         },
    //         "two": master.StateSingleElevator{
    //             ElevatorBehaviour:       "idle",
    //             Floor:          0,
    //             Direction:      "stop",
	// 			Available: true,
    //             CabOrders:    [NUMFLOORS]bool{false, false, false, false},
				
    //         },
    //     },
    // }

	// hraExecutable := "hall_request_assigner"

    // jsonBytes, err := json.Marshal(input)
    // if err != nil {
    //     fmt.Println("json.Marshal error: ", err)
    //     return
    // }
    
    // ret, err := exec.Command("../TTK4145_Real-Time_Programming_G67/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
    // if err != nil {
    //     fmt.Println("exec.Command error: ", err)
    //     fmt.Println(string(ret))
    //     return
    // }
    
    // output := new(map[string][NUMFLOORS][NUMHALLBUTTONS]bool)
    // err = json.Unmarshal(ret, &output)
    // if err != nil {
    //     fmt.Println("json.Unmarshal error: ", err)
    //     return
    // }
        
    // fmt.Printf("output: \n")
    // for k, v := range *output {
    //     fmt.Printf("%6v :  %+v\n", k, v)
    // }
}


