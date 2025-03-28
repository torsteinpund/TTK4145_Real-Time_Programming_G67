package master

import (
	. "Driver-go/types"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type StateSingleElevator struct {
	ElevatorBehaviour string          `json:"behaviour"`
	Floor             int             `json:"floor"`
	Direction         string 		  `json:"direction"`
	Available         bool
	CabOrders         [NUMFLOORS]bool `json:"cabRequests"`
}

type AllElevators struct {
	GlobalOrders	  HallOrders		 			 `json:"hallRequests"`
	AllElevatorStates map[string]StateSingleElevator `json:"states"`
}


func Master(ID string,
			Ch_isMaster           <-chan bool,
			Ch_peerLost 		  <-chan string,
			Ch_ordersFromMaster   chan<- GlobalOrderMap,
			Ch_registerOrder 	  <-chan OrderEvent,
			Ch_stateUpdate 		  <-chan Elevator,
			Ch_globalOrderCopy    <-chan GlobalOrderMap,
			Ch_orderCopyRequest   chan<- bool,
			Ch_newPeer 			  <-chan string) {

	fmt.Println("Master Started!")
	allElevatorStates 	:= map[string]StateSingleElevator{}
	hallOrders 			:= HallOrders{}
	lastGlobalOrderMap := GlobalOrderMap{}

	for {
		select {

		case lostPeer := <-Ch_peerLost:
			elevator, exist := allElevatorStates[lostPeer]
			fmt.Println("Houston, we have a problem! Master has lost a peer")
			if !exist {
				elevator = unitializedSingleStateElevator()
				elevator.Available = false
				allElevatorStates[lostPeer] = elevator
			} else {
				elevator.Available = false
				allElevatorStates[lostPeer] = elevator
			}
			
			updatedOrders := reAssignOrders(hallOrders, allElevatorStates)
			lastGlobalOrderMap = updatedOrders
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

		case newOrderEvent := <-Ch_registerOrder:
			elevatorID := newOrderEvent.ElevatorID
			_, exist := allElevatorStates[elevatorID]
			if !exist {
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
			lastGlobalOrderMap = updatedGlobalOrders
			Ch_ordersFromMaster <- updatedGlobalOrders

		case masterCheck := <-Ch_isMaster:
			if masterCheck {
				Ch_orderCopyRequest <- true
			} else {
				slaveLoop:
				for {
					select {
					case masterCheck := <-Ch_isMaster:
						if masterCheck {
							Ch_orderCopyRequest <- true
							time.Sleep(500 * time.Millisecond)
							fmt.Println("Master waking up")
							break slaveLoop
						}

					case <-Ch_registerOrder:
					case <-Ch_stateUpdate:
					case <-Ch_globalOrderCopy:
						// Ensures draining of channels
					}
				}
			}

		case newState := <-Ch_stateUpdate:
			elevator, exist := allElevatorStates[newState.ID]
			cabOrders := [NUMFLOORS]bool{}

			if exist {
				cabOrders = elevator.CabOrders
			}

			allElevatorStates[newState.ID] = StateSingleElevator{
				newState.Behaviour.ToString(),
				newState.Floor,
				newState.Dirn.ToString(),
				newState.Available,
				cabOrders}

			updatedOrders := reAssignOrders(hallOrders, allElevatorStates)

			if checkIfUpdatedGlobalOrderMap(updatedOrders, lastGlobalOrderMap) {
				lastGlobalOrderMap = updatedOrders
				Ch_ordersFromMaster <- updatedOrders
			}

		case orderCopy := <-Ch_globalOrderCopy:
			allElevatorStates, hallOrders = updateAllElevators(hallOrders, orderCopy, allElevatorStates)
			updatedOrders := reAssignOrders(hallOrders, allElevatorStates)
			Ch_ordersFromMaster <- updatedOrders
		}
	}
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


func reAssignOrders(hallOrders HallOrders, allElevatorStates map[string]StateSingleElevator) GlobalOrderMap {
	unavailableElevators  := []string{}
	availableElevatorsMap := map[string]StateSingleElevator{}

	for elevatorID, elevatorState := range allElevatorStates {
		if !elevatorState.Available {
			unavailableElevators = append(unavailableElevators, elevatorID)
		} else {
			availableElevatorsMap[elevatorID] = elevatorState
		}
	}

	globOrderMap := GlobalOrderMap{}
	if len(availableElevatorsMap) > 0 {
		allElevators := AllElevators{GlobalOrders: hallOrders, AllElevatorStates: availableElevatorsMap}
		globOrderMap = hallAssignerExec(allElevators)
		if globOrderMap == nil {
			globOrderMap = GlobalOrderMap{}
		}
	}

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


func updateAllElevators(hallOrders HallOrders, orderCopy GlobalOrderMap, allElevatorStates map[string]StateSingleElevator) (map[string]StateSingleElevator, HallOrders){
	for elevatorID, orderMatrix := range orderCopy {
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
						allElevatorStates[elevatorID] = StateSingleElevator{"idle", 0, "down", true, cabOrders}
					} else {
						elevator.CabOrders[floor] = elevator.CabOrders[floor] || isOrder
						allElevatorStates[elevatorID] = elevator
					}
				}
			}
		}
	}
	return allElevatorStates,hallOrders
}


func checkIfUpdatedGlobalOrderMap(updatedOrders GlobalOrderMap, lastGlobaleOrderMap GlobalOrderMap) bool {
	for elevatorID, orders := range updatedOrders {
		for floor, row := range orders {
			for button, isOrder := range row {
				if isOrder != lastGlobaleOrderMap[elevatorID][floor][button] {
					return true
				}
			}
		}
	}
	return false
}


