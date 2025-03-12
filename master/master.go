package master

import (
	. "Driver-go/types"
	"encoding/json"
	"fmt"
	"os/exec"
	//"Driver-go/cost"
	// "math"
	// "strings"
	"time"
	
)

type MasterChannels struct {
	IsMasterChannel      chan bool
	PeerLostChannel      chan string
	ToSlavesChannel      chan NetworkMessage
	RegisterOrderChannel chan OrderEvent
	StateUpdateChannel   chan Elevator
	OrderCopyResponseChannel  chan GlobalOrderMap 
	RegisteredPeerChannel chan string
}

// StateSingleElevator represents the state of a single elevator
type StateSingleElevator struct {
	ElevatorBehaviour string `json:"behaviour"`
	Floor             int    `json:"floor"`
	Direction         string `json:"direction"`
	Available         bool
	CabOrders         [NUMFLOORS]bool `json:"cabRequests"`
}


// Elevators represents the global state of all elevators, including global orders
// and the state of each individual elevator.
type AllElevators struct {
	GlobalOrders [NUMFLOORS][NUMHALLBUTTONS]bool	`json:"hallRequests"`
	//The key to States is the local IP address of the elevator
	States       map[string]StateSingleElevator		`json:"states"`
}

func RunMaster(ID string, channel MasterChannels) {
	fmt.Println("Running master...")

	allElevatorStates := map[string]StateSingleElevator{}
	hallOrders := [NUMFLOORS][NUMHALLBUTTONS]bool{}

	orderCopy := NetworkMessage{
		MsgType:    "Broadcast message",
		Receipient: All,
		MsgData:    true,
	}

	channel.ToSlavesChannel <- orderCopy

	for {
		select {

		case lostPeer := <-channel.PeerLostChannel:
			elevator, exist := allElevatorStates[lostPeer]
			fmt.Println("Houston, we have a problem! Master has lost a peer")
			if !exist {
				elevator = StateSingleElevator{}
				elevator.Available = false	

				allElevatorStates[lostPeer] = elevator
			} else {
				elevator.Available = false
				allElevatorStates[lostPeer] = elevator
			}
		

			updatedOrders := reAssignOrders(hallOrders, allElevatorStates)

			channel.ToSlavesChannel <- updatedOrders

		case newPeer := <-channel.RegisteredPeerChannel:
			fmt.Println("Master has registered a new peer: ", newPeer)
			elevator, exists := allElevatorStates[newPeer]
			if !exists {
				elevator = StateSingleElevator{}
				elevator.Available = true
				allElevatorStates[newPeer] = elevator
			} else {
				elevator.Available = true
				allElevatorStates[newPeer] = elevator
			}
			updatedOrders := reAssignOrders(hallOrders, allElevatorStates)
			channel.ToSlavesChannel <- updatedOrders


		case newOrderEvent := <- channel.RegisterOrderChannel:
			elevatorID := newOrderEvent.ElevatorID
			_, exist := allElevatorStates[elevatorID]
			if !exist {
				println("M: No client with ID: ", elevatorID)
				break
			}
			for _, order := range newOrderEvent.Orders{
				switch order.Button {
				case BT_HallUp, BT_HallDown:
					hallOrders[order.Floor][order.Button] = !newOrderEvent.Completed

				case BT_Cab:
					elevator := allElevatorStates[elevatorID]
					elevator.CabOrders[order.Floor] = !newOrderEvent.Completed
					allElevatorStates[elevatorID] = elevator
				}
			}
			updatedGlobalOrders := reAssignOrders(hallOrders,allElevatorStates)
			channel.ToSlavesChannel <- updatedGlobalOrders


		case masterCheck:= <- channel.IsMasterChannel:
			if masterCheck{
				channel.ToSlavesChannel <- orderCopy //If the master is still running, the ordercopy is passed through the ToslavesChannel.
			}else{
				fmt.Println("Mayday, Mayday. The master elevator: " + ID +"is shutting down")
			findNewMaster:
				for{
					select{
					case masterCheck:= <-channel.IsMasterChannel:
						if masterCheck {
							channel.ToSlavesChannel <- orderCopy
							time.Sleep(500*time.Millisecond)
							fmt.Println("Master waking up")
							break findNewMaster
						}
					}
				}
			}
		case state := <- channel.StateUpdateChannel:
			reassign := false
			elevator,exist := allElevatorStates[state.ID]

			cabOrders := [NUMFLOORS]bool{}
			if exist {
				cabOrders = elevator.CabOrders
				reassign = elevator.Available != state.Avaliable //If the elevator is not available, we should reassign the order.
			}

			allElevatorStates[state.ID] = StateSingleElevator{
				state.Behaviour.ToString(),
				state.Floor,
				state.Dirn.ToString(),
				state.Avaliable,
				cabOrders}
			if reassign {
				updatedOrders := reAssignOrders(hallOrders, allElevatorStates)
				channel.ToSlavesChannel <- updatedOrders
			}
		case orderCopy := <-channel.OrderCopyResponseChannel:
			for elevatorID,orderMatrix := range orderCopy { //Loops through every elevator 
				for floor, row := range orderMatrix {
					for button, isOrder := range row {
							switch ButtonType(button) {
							case BT_HallUp,BT_HallDown:
								hallOrders[floor][button] = hallOrders[floor][button] || isOrder
							case BT_Cab:
								elevator,exist := allElevatorStates[elevatorID]
								if !exist {
									cabOrders := [NUMFLOORS]bool{}
									cabOrders[floor] = isOrder
									allElevatorStates[elevatorID] = StateSingleElevator{
										"idle",
										0,
										"down",
										true,
										cabOrders}

								}else{
									elevator.CabOrders[floor] = elevator.CabOrders[floor] || isOrder
									allElevatorStates[elevatorID] = elevator
								}	
							}
					}
				}
			}
			updatedOrders := reAssignOrders(hallOrders,allElevatorStates)
			channel.ToSlavesChannel <- updatedOrders
		
		}
	}	
}



func reAssignOrders(hallOrders [NUMFLOORS][NUMHALLBUTTONS]bool, allElevatorStates map[string]StateSingleElevator) NetworkMessage {

	unavailableElevators := []string{}
	elevatorMap := map[string]StateSingleElevator{}

	//Checks availability for all elevators, and appends them in either an unavaliable list or an elevatormap
	for elevatorID, elevatorState := range allElevatorStates {
		if !elevatorState.Available {
			unavailableElevators = append(unavailableElevators, elevatorID)
		} else {
			elevatorMap[elevatorID] = elevatorState
		}
	}

	//Calculates which available elevators should take the hallorders of the lost peer
	allElevators := AllElevators{GlobalOrders: hallOrders, States: elevatorMap}
	globOrderMap := hallAssignerExec(allElevators)

	
	//Add the cab-calls of the lost peer to the orderlist so it can be reminded of them when it returns
	for _, elevatorID := range unavailableElevators {
		orders := OrderMatrix{}
		for floor := range orders {
			orders[floor][BT_Cab] = allElevatorStates[elevatorID].CabOrders[floor]
			
		}
		globOrderMap[elevatorID] = orders
	}

	updatedOrders := NetworkMessage{MsgType: "Updated globalorders", MsgData: globOrderMap, Receipient: All}

	return updatedOrders
}







func hallAssignerExec(input AllElevators) GlobalOrderMap{
	hraExecutable := "hall_request_assigner"

    jsonBytes, err := json.Marshal(input)
    if err != nil {
        fmt.Println("json.Marshal error: ", err)
        return nil
    }
    
    ret, err := exec.Command("../TTK4145_Real-Time_Programming_G67/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
    if err != nil {
        fmt.Println("exec.Command error: ", err)
        fmt.Println(string(ret))
        return nil
    }
    
    output := GlobalOrderMap{}
    err = json.Unmarshal(ret, &output)
    if err != nil {
        fmt.Println("json.Unmarshal error: ", err)
        return nil
    }
        
    // fmt.Printf("output: \n")
    // for k, v := range *output {
    //     fmt.Printf("%6v :  %+v\n", k, v)
    // }

	return output
	
}




func getElevatorIDs(states map[string]StateSingleElevator) []string {
	var ids []string
	for id := range states {
		ids = append(ids, id)
	}
	return ids
}