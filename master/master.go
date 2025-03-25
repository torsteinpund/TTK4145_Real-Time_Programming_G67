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
	"Driver-go/backup"
)

// StateSingleElevator represents the state of a single elevator
type StateSingleElevator struct {
	ElevatorBehaviour string `json:"behaviour"`
	Floor             int    `json:"floor"`
	Direction         string `json:"direction"`
	Available         bool
	CabOrders         [NUMFLOORS]bool `json:"cabRequests"`
}

func unitializedSingleStateElevator() StateSingleElevator {
	return StateSingleElevator{
		ElevatorBehaviour: "idle",
		Floor:             1,
		Direction:         "stop",
		Available:         true,
		CabOrders:         [NUMFLOORS]bool{},
	}
}

// Elevators represents the global state of all elevators, including global orders
// and the state of each individual elevator.
type AllElevators struct {
	GlobalOrders [NUMFLOORS][NUMHALLBUTTONS]bool `json:"hallRequests"`
	//The key to States is the local IP address of the elevator
	States map[string]StateSingleElevator `json:"states"`
}

func Master(ID string,
	Ch_isMaster <-chan bool,
	Ch_peerLost <-chan string,
	Ch_ordersFromMaster chan<- GlobalOrderMap,
	Ch_registerOrder <-chan OrderEvent,
	Ch_stateUpdate <-chan Elevator,
	Ch_orderCopyResponse <-chan GlobalOrderMap,
	Ch_orderCopyRequest chan<- bool,
	Ch_newPeer <-chan string,) {
	fmt.Println("Running master...")

	allElevatorStates := map[string]StateSingleElevator{}
	hallOrders := [NUMFLOORS][NUMHALLBUTTONS]bool{}
	lastGlobaleOrderMap := GlobalOrderMap{}

	for {
		select {

		case lostPeer := <-Ch_peerLost:
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
			fmt.Println("Updated orders after peer loss: ", updatedOrders)
			lastGlobaleOrderMap = updatedOrders
			Ch_ordersFromMaster <- updatedOrders

		case newPeer := <-Ch_newPeer:
			fmt.Println("Master has registered a new peer: ", newPeer)
			elevator, exists := allElevatorStates[newPeer]
			if !exists {
				elevator = unitializedSingleStateElevator()
				elevator.Available = true
				allElevatorStates[newPeer] = elevator
			} else {
				elevator.Available = true
				allElevatorStates[newPeer] = elevator
			}

			updatedOrders := reAssignOrders(hallOrders, allElevatorStates)
			lastGlobaleOrderMap = updatedOrders
			fmt.Println("Master has reassigned the new peer")
			Ch_ordersFromMaster <- updatedOrders
			fmt.Println("Master has sent the updated orders to the slave")

		case newOrderEvent := <-Ch_registerOrder:
			fmt.Println("Master has received a new order event")
			elevatorID := newOrderEvent.ElevatorID
			_, exist := allElevatorStates[elevatorID]
			if !exist {
				println("M: No client with ID: ", elevatorID)
				break
			}
			for _, order := range newOrderEvent.Orders {
				switch order.Button {
				case BT_HallUp:
					hallOrders[order.Floor][order.Button] = !newOrderEvent.Completed
				case BT_HallDown:
					hallOrders[order.Floor][order.Button] = !newOrderEvent.Completed
				case BT_Cab:
					elevator := allElevatorStates[elevatorID]
					elevator.CabOrders[order.Floor] = !newOrderEvent.Completed
					allElevatorStates[elevatorID] = elevator
				}
			}
			updatedGlobalOrders := reAssignOrders(hallOrders, allElevatorStates)
			lastGlobaleOrderMap = updatedGlobalOrders
			Ch_ordersFromMaster <- updatedGlobalOrders

		case masterCheck := <-Ch_isMaster:
			fmt.Println("Master has received a check if master")
			if masterCheck {
				Ch_orderCopyRequest <- true //If the master is still running, the ordercopy is passed through the ToslavesChannel.

			} else {
				fmt.Println("Mayday, Mayday. Getting sucked into the matrix: " + ID + " is getting ready to work for free")
				stuckInTheMatrix:
				for {
					select {
					case masterCheck := <-Ch_isMaster:
						if masterCheck {
							Ch_orderCopyRequest <- true
							time.Sleep(500 * time.Millisecond)
							fmt.Println("Master waking up")
							break stuckInTheMatrix
						}

					// case <-Ch_newPeer:

					// case <-Ch_peerLost:
						

					case <-Ch_registerOrder:

					case <-Ch_stateUpdate:

					// case orderCopyResp := <-Ch_orderCopyResponse:
					// 	copyGlobalOrderMap = orderCopyResp

					default:
						time.Sleep(10 * time.Millisecond)

					}
				}
			}
		case state := <-Ch_stateUpdate:
			// reassign := false
			elevator, exist := allElevatorStates[state.ID]
			cabOrders := [NUMFLOORS]bool{}

			if exist {
				cabOrders = elevator.CabOrders
				// reassign = elevator.Available != state.Available //If the elevator is not available, we should reassign the order.
			}

			allElevatorStates[state.ID] = StateSingleElevator{
				state.Behaviour.ToString(),
				state.Floor,
				state.Dirn.ToString(),
				state.Available,
				cabOrders}

			updatedOrders := reAssignOrders(hallOrders, allElevatorStates)

			if backup.CheckIfUpdatedGlobalOrderMap(updatedOrders, lastGlobaleOrderMap) {
				lastGlobaleOrderMap = updatedOrders
				Ch_ordersFromMaster <- updatedOrders
				
			}
			// fmt.Println("AllElevatorStates: ", allElevatorStates)

		case orderCopy := <-Ch_orderCopyResponse:
			fmt.Println("Master has received an order copy response", orderCopy)
			for elevatorID, orderMatrix := range orderCopy { //Loops through every elevator
				for floor, row := range orderMatrix {
					for button, isOrder := range row {
						switch ButtonType(button) {
						case BT_HallUp, BT_HallDown:
							hallOrders[floor][button] = hallOrders[floor][button] || isOrder
						case BT_Cab:
							elevator, exist := allElevatorStates[elevatorID]
							if !exist {
								cabOrders := [NUMFLOORS]bool{}
								cabOrders[floor] = isOrder
								allElevatorStates[elevatorID] = StateSingleElevator{
									"idle",
									0,
									"down",
									true,
									cabOrders}

							} else {
								elevator.CabOrders[floor] = elevator.CabOrders[floor] || isOrder
								allElevatorStates[elevatorID] = elevator
							}
						}
					}
				}
			}

			updatedOrders := reAssignOrders(hallOrders, allElevatorStates)
			fmt.Println("Master has reassigned the orders after receiving an order copy response", updatedOrders)
			Ch_ordersFromMaster <- updatedOrders

		default:
			// fmt.Println("Master is waiting for a message")
			// time.Sleep(100 * time.Millisecond)
		}
	}
}

func reAssignOrders(hallOrders [NUMFLOORS][NUMHALLBUTTONS]bool, allElevatorStates map[string]StateSingleElevator) GlobalOrderMap {
	// fmt.Println("Reassigning orders")
	// fmt.Println("HallOrders: ", hallOrders)
	// fmt.Println("AllElevatorStates: ", allElevatorStates)
	unavailableElevators := []string{}
	availableElevatorsMap := map[string]StateSingleElevator{}

	//Checks availability for all elevators, and appends them in either an unavaliable list or an elevatormap
	for elevatorID, elevatorState := range allElevatorStates {
		if !elevatorState.Available {
			unavailableElevators = append(unavailableElevators, elevatorID)
		} else {
			availableElevatorsMap[elevatorID] = elevatorState
		}
	}

	//Calculates which available elevators should take the hallorders of the lost peer
	allElevators := AllElevators{GlobalOrders: hallOrders, States: availableElevatorsMap}
	globOrderMap := hallAssignerExec(allElevators)

	//Add the cab-calls of the lost peer to the orderlist so it can be reminded of them when it returns
	for _, elevatorID := range unavailableElevators {
		orders := OrderMatrix{}
		for floor := range orders {
			orders[floor][BT_Cab] = allElevatorStates[elevatorID].CabOrders[floor]

		}
		globOrderMap[elevatorID] = orders
	}

	return globOrderMap
}

func hallAssignerExec(input AllElevators) GlobalOrderMap {
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

	return output

}

func getElevatorIDs(states map[string]StateSingleElevator) []string {
	var ids []string
	for id := range states {
		ids = append(ids, id)
	}
	return ids
}

func getElevatorCabOrders(ordermatrix OrderMatrix) [NUMFLOORS]bool {
	var cabOrders [NUMFLOORS]bool
	for i := range NUMFLOORS {
		cabOrders[i] = ordermatrix[i][BT_Cab]
	}
	return cabOrders
}
