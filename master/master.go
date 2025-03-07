package master

import (
	. "Driver-go/types"
	//"encoding/json"
	"fmt"
	//"os/exec"
	//"Driver-go/cost"
	"math"
	"strings"
	"time"
	
)

type MasterChannels struct {
	IsMasterChannel      chan bool
	PeerLostChannel      chan string
	ToSlavesChannel      chan NetworkMessage
	RegisterOrderChannel chan OrderEvent
	StateUpdateChannel   chan Elevator
	OrderCopyResponseCh  chan GlobalOrderMap 
}

// StateSingleElevator represents the state of a single elevator
type StateSingleElevator struct {
	Floor             int    `json:"floor"`
	Direction         string `json:"direction"`
	ElevatorBehaviour string `json:"behaviour"`
	Available         bool
	CabOrders         [NUMFLOORS]bool `json:"caborders"`
}


// Elevators represents the global state of all elevators, including global orders
// and the state of each individual elevator.
type AllElevators struct {
	GlobalOrders [NUMFLOORS][NUMBUTTONTYPE - 1]bool
	States       map[string]StateSingleElevator
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
				fmt.Println("Mayday, Mayday. The master elevator: " + ID +"is shutting the fuck down")
			findNewMaster:
				for{
					select{
					case masterCheck:= <-channel.IsMasterChannel:
						if masterCheck {
							channel.ToSlavesChannel <- orderCopy
							time.Sleep(500*time.Millisecond)
							fmt.Println("Master waking the fuck up")
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
				state.Floor,
				state.Dirn.ToString(),
				state.Behaviour.ToString(),
				state.Avaliable,
				cabOrders}
			if reassign {
				updatedOrders := reAssignOrders(hallOrders, allElevatorStates)
				channel.ToSlavesChannel <- updatedOrders
			}
		case orderCopy := <-channel.OrderCopyResponseCh:
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
										0,
										"down",
										"idle",
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
	globOrderMap := assignHallRequests(allElevators)

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


func assignHallRequests(input AllElevators) GlobalOrderMap {
	// Initialiser output for hver heis med en matrise (NUMFLOORS x NUMHALLBUTTONS) satt til false.
	globalOrderMap := GlobalOrderMap{}
	for id := range input.States {
		matrix := OrderMatrix{}
		globalOrderMap[id] = matrix
	}

	// For hver etasje og for hver hall-knapp (opp og ned), hvis det er en aktiv forespørsel,
	// finn den heisen med lavest "kostnad" og tildel denne forespørselen.
	for floor := 0; floor < NUMFLOORS; floor++ {
		for btn := 0; btn < NUMHALLBUTTONS; btn++ {
			if input.GlobalOrders[floor][btn] {
				bestElevator := ""
				bestCost := math.MaxFloat64
				for id, state := range input.States {
					c := ComputeCost(state, floor, btn)
					if c < bestCost {
						bestCost = c
						bestElevator = id
					}
				}
				if bestElevator != "" {
					matrix := globalOrderMap[bestElevator]
					matrix[floor][btn] = true
					globalOrderMap[bestElevator] = matrix

				}
			}
		}
	}
	return globalOrderMap
}

func ComputeCost(elevator StateSingleElevator, requestFloor int, button int) float64 {
	// Grunnkostnad basert på avstand (absolutt forskjell i etasjer)
	cost := math.Abs(float64(elevator.Floor - requestFloor))

	// Bonus: Hvis heisen er idle, trekk litt fra kostnaden
	if strings.ToLower(elevator.ElevatorBehaviour) == "idle" {
		cost -= 0.5
	}

	// Hvis heisens retning stemmer overens med forespurt knapp, trekk også litt fra
	if button == 0 && strings.ToLower(elevator.Direction) == "up" {
		cost -= 0.2
	}
	if button == 1 && strings.ToLower(elevator.Direction) == "down" {
		cost -= 0.2
	}

	return cost
}

func getElevatorIDs(states map[string]StateSingleElevator) []string {
	var ids []string
	for id := range states {
		ids = append(ids, id)
	}
	return ids
}