package fsm

import (
	"Driver-go/lights"
	"Driver-go/singleElevatorDriver/elevio"
	"Driver-go/singleElevatorDriver/requests"
	"Driver-go/singleElevatorDriver/timer"
	. "Driver-go/types"
	"fmt"
	"time"
)

type FsmChannels struct {
	Ch_floorSensor  chan int
	Ch_stopButton   chan bool
	Ch_obstruction  chan bool
	Ch_localLights  chan OrderMatrix
	Ch_localOrders  chan OrderMatrix
	Ch_toMaster     chan NetworkMessage
	Ch_clearedFloor chan int
	Ch_stateUpdate  chan Elevator
	Ch_doorOpen     chan bool
}

func executeOrder(orderMatrix OrderMatrix, elev Elevator) (OrderMatrix, Elevator) {

	dirnBehaviour := requests.RequestsChooseDirection(orderMatrix, elev)
	fmt.Println("DirnBehaviour: ", dirnBehaviour)
	elev.Dirn = dirnBehaviour.Dirn
	elev.Behaviour = ElevatorBehaviour(dirnBehaviour.Behaviour)

	switch elev.Behaviour {

	case EB_DoorOpen:
		elevio.SetDoorOpenLamp(true)
		timer.TimerStart(elev.Config.DoorOpenDuration)
		orderMatrix, elev = requests.RequestsClearAtCurrentFloor(orderMatrix, elev)

	case EB_Moving:
		elevio.SetMotorDirection(elev.Dirn)

	case EB_Idle:
	}
	return orderMatrix, elev
}

func fsmFloorArrival(orderMatrix OrderMatrix, newFloor int, elev Elevator) (OrderMatrix, Elevator, bool) {

	elev.Floor = newFloor
	reqCleared := false

	elevio.SetFloorIndicator(elev.Floor)

	switch elev.Behaviour {
	case ElevatorBehaviour(EB_Moving):
		// Check if the elevator should stop at the current floor
		if requests.RequestsShouldStop(orderMatrix, elev) {
		
			fmt.Println("Door open event")
			elevio.SetMotorDirection(MD_Stop)
			elevio.SetDoorOpenLamp(true)
			orderMatrix, elev = requests.RequestsClearAtCurrentFloor(orderMatrix, elev)
			timer.TimerStop()
			timer.TimerStart(elev.Config.DoorOpenDuration)
			orderMatrix = lights.SetCabLights(orderMatrix)
			elev.Behaviour = ElevatorBehaviour(EB_DoorOpen)
			reqCleared = true
		}
	default:
		// No action
		//elevio.SetMotorDirection(MD_Stop)
	}

	return orderMatrix, elev, reqCleared
}

// func fsmDoorTimeout(orderMatrix OrderMatrix, elev Elevator) (OrderMatrix, Elevator) {

// 	switch elev.Behaviour {
// 	case ElevatorBehaviour(EB_DoorOpen):
// 		// Choose direction based on requests
// 		dirnBehaviour := requests.RequestsChooseDirection(orderMatrix, elev)
// 		elev.Dirn = dirnBehaviour.Dirn
// 		elev.Behaviour = ElevatorBehaviour(dirnBehaviour.Behaviour)

// 		switch elev.Behaviour {
// 		case ElevatorBehaviour(EB_DoorOpen):
// 			// Start timer and clear requests
// 			timer.TimerStart(elev.Config.DoorOpenDuration)
// 			orderMatrix, elev = requests.RequestsClearAtCurrentFloor(orderMatrix, elev)
// 			// elev.Requests = lights.SetHallLights(elev.Requests)
// 			orderMatrix = lights.SetCabLights(orderMatrix)

// 		case ElevatorBehaviour(EB_Moving), ElevatorBehaviour(EB_Idle):
// 			// Shut the door and start moving
// 			elevio.SetDoorOpenLamp(false)
// 			elevio.SetMotorDirection(elev.Dirn)
// 		}

// 	default:
// 		// No action
// 	}

// 	return orderMatrix, elev
// }

func fsmDoorTimeout(orderMatrix OrderMatrix, elev Elevator) (OrderMatrix, Elevator) {
	fmt.Println("The behaviour is: ", elev.Behaviour)
	switch elev.Behaviour {
	case ElevatorBehaviour(EB_DoorOpen):
		fmt.Println("Door timeout: Handling door close")
		// Choose direction based on requests
		dirnBehaviour := requests.RequestsChooseDirection(orderMatrix, elev)
		elev.Dirn = dirnBehaviour.Dirn
		elev.Behaviour = ElevatorBehaviour(dirnBehaviour.Behaviour)

		switch elev.Behaviour {
		case ElevatorBehaviour(EB_DoorOpen):
			fmt.Println("Door remains open: Restarting timer")
			// Start timer and clear requests
			timer.TimerStart(elev.Config.DoorOpenDuration)
			orderMatrix, elev = requests.RequestsClearAtCurrentFloor(orderMatrix, elev)
			orderMatrix = lights.SetCabLights(orderMatrix)

		case ElevatorBehaviour(EB_Moving), ElevatorBehaviour(EB_Idle):
			fmt.Println("Closing door and transitioning to next state")
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elev.Dirn)
		}

	default:
		fmt.Println("Door timeout: No action for current behavior")
	}

	return orderMatrix, elev
}

func FsmRun(ch_fsm FsmChannels, elev Elevator) {
	fmt.Println("FSM Started!")
	//
	// Polling rate configuration
	inputPollRate := 25 * time.Millisecond // Adjust as needed
	orderMatrix := OrderMatrix{}

	// Initialize system state
	prevFloor := elev.Floor
	//timerActive := false
	//var timerEndTime float64
	obstructionActive := false
	clearedFloor := false
	lastKnownDirection := MotorDirection(0)
	elevio.SetDoorOpenLamp(false)
	doorClose := time.NewTimer(3 * time.Second)
	doorClose.Stop()
	errorTimeout := time.NewTimer(5 * time.Second)

	// Main event loop
	for {
		select {
		case <-doorClose.C:

			if obstructionActive {
				doorClose.Reset(3 * time.Second)
				break
			}
			println("FSM: Door Close")
			elevio.SetDoorOpenLamp(false)
			if orderMatrix == (OrderMatrix{}) {
				elev.Behaviour = EB_Idle
				errorTimeout.Stop()
				break
			} else {
				dirnpair := requests.RequestsChooseDirection(orderMatrix, elev)
				elev.Dirn = dirnpair.Dirn
				elev.Behaviour = ElevatorBehaviour(dirnpair.Behaviour)
				elevio.SetMotorDirection(elev.Dirn)
				errorTimeout.Reset(5 * time.Second)
			}
			

		case receivedOrder := <-ch_fsm.Ch_localOrders:
			orderMatrix = receivedOrder
			orderMatrix = lights.SetCabLights(orderMatrix)
			if elev.Behaviour == ElevatorBehaviour(EB_DoorOpen) {
				break
			}
			orderMatrix, elev = executeOrder(orderMatrix, elev)
			fmt.Println(elev.Behaviour)

		case currentFloor := <-ch_fsm.Ch_floorSensor:
			fmt.Printf("Received floor sensor event: %d\n", currentFloor)

			if currentFloor != prevFloor {
				clearedFloor = false
				fmt.Println("OrderMatrix before fsmFloorArrival", orderMatrix)
				orderMatrix, elev, clearedFloor = fsmFloorArrival(orderMatrix, currentFloor, elev)

				// elevio.SetFloorIndicator(currentFloor) // Update floor indicator lamp
				if clearedFloor {
					doorClose.Reset(3 * time.Second)
				}

				// if elev.Behaviour == ElevatorBehaviour(EB_DoorOpen) && !obstructionActive {
				// 	timer.TimerStop()
				// 	timer.TimerStart(3.0)
				// 	// fmt.Println("ti//Passes the updated statemer started")
				// }
				fmt.Println("Cleared floor", clearedFloor)

				prevFloor = currentFloor
				// obstructionActive = false
				elev.Available = true
				// updateElevator := NetworkMessage{MsgType: "elevatorupdatechannel", MsgData:  elev,Receipient: Master}
				ch_fsm.Ch_stateUpdate <- elev
				fmt.Println("Elevator state updated")
				fmt.Println("Behaviour after State update: ", elev.Behaviour)

			}

		case stopPressed := <-ch_fsm.Ch_stopButton:
			// Handle stop button event
			if stopPressed {
				lastKnownDirection = elev.Dirn
				fmt.Println("Stop button pressed!")
				fmt.Println(lastKnownDirection)
				elevio.SetStopLamp(true)
				elevio.SetMotorDirection(0)
				// stop = true
			} else {
				fmt.Println("Stop button released!")
				elevio.SetStopLamp(false)
			}

		case <-time.After(inputPollRate):
			// Periodic tasks (check timer)

			if timer.TimerTimedOut() {
				fmt.Println("Door timeout occurred. Elevator stopped", elev.Behaviour)
				// updateElevator := NetworkMessage{MsgType: "elevatorupdatechannel", MsgData: elev, Receipient: Master}
				// ch_fsm.Ch_toMaster <- updateElevator
				// fmt.Println("Door timeout occurred.")
				orderMatrix, elev = fsmDoorTimeout(orderMatrix, elev)
				fmt.Println("Passed FSMdoorTimeout")
				timer.TimerStop() // Reset the timer after timeout handling
			}

		case obstruction := <-ch_fsm.Ch_obstruction:
			fmt.Println("Obstruction detected")
			if obstruction {
				// obstructionActive = true
				elev.Available = false
				timer.TimerStop()
				elevio.SetDoorOpenLamp(true)
				fmt.Println("obstruction switch")

			} else if !obstruction {
				// obstructionActive = false
				elev.Available = true
				timer.TimerStart(3.0)
				fmt.Println("obstruction switch off")
			}

			

			// ch_fsm.Ch_stateUpdate <- elev //Passes the updated state
		}
	}

}
