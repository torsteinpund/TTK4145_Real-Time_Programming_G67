package fsm

// import (
// 	"Driver-go/lights"
// 	"Driver-go/singleElevatorDriver/elevio"
// 	"Driver-go/singleElevatorDriver/requests"
// 	"Driver-go/singleElevatorDriver/timer"
// 	. "Driver-go/types"
// 	"fmt"
// 	"time"
// )

// type FsmChannels struct {
// 	Ch_floorSensor  chan int
// 	Ch_stopButton   chan bool
// 	Ch_obstruction  chan bool
// 	Ch_localLights  chan OrderMatrix
// 	Ch_localOrders  chan OrderMatrix
// 	Ch_toMaster     chan NetworkMessage
// 	Ch_clearedFloor chan int
// 	Ch_stateUpdate  chan Elevator
// }

// func FsmButtonPressed(orderMatrix OrderMatrix, elev Elevator) (OrderMatrix, Elevator) {

// 	dirnBehaviour := requests.RequestsChooseDirection(orderMatrix, elev)

// 	elev.Dirn = dirnBehaviour.Dirn
// 	elev.Behaviour = ElevatorBehaviour(dirnBehaviour.Behaviour)

// 	switch dirnBehaviour.Behaviour {

// 	case EB_DoorOpen:
// 		elevio.SetDoorOpenLamp(true)
// 		timer.TimerStart(elev.Config.DoorOpenDuration)
// 		orderMatrix, elev = requests.RequestsClearAtCurrentFloor(orderMatrix, elev)

// 	case EB_Moving:
// 		elevio.SetMotorDirection(elev.Dirn)

// 	case EB_Idle:
// 	}
// 	return orderMatrix, elev
// }

// func fsmFloorArrival(orderMatrix OrderMatrix, newFloor int, elev Elevator) (OrderMatrix, Elevator, bool) {

// 	elev.Floor = newFloor
// 	reqCleared := false

// 	elevio.SetFloorIndicator(elev.Floor)

// 	switch elev.Behaviour {
// 	case ElevatorBehaviour(EB_Moving):
// 		// Check if the elevator should stop at the current floor
// 		if requests.RequestsShouldStop(orderMatrix, elev) {
// 			// ch_doorOpen <- true
// 			fmt.Println("Door open event")
// 			elevio.SetMotorDirection(MD_Stop)
// 			elevio.SetDoorOpenLamp(true)
// 			orderMatrix, elev = requests.RequestsClearAtCurrentFloor(orderMatrix, elev)
// 			timer.TimerStart(elev.Config.DoorOpenDuration)
// 			orderMatrix = lights.SetCabLights(orderMatrix)
// 			elev.Behaviour = ElevatorBehaviour(EB_DoorOpen)
// 			reqCleared = true
// 		}
// 	default:
// 		// No action
// 		//elevio.SetMotorDirection(MD_Stop)
// 	}

// 	return orderMatrix, elev, reqCleared
// }

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

// func FsmRun(ch_fsm FsmChannels, elev Elevator) {
// 	fmt.Println("FSM Started!")
// 	// inputPollRate := 25 * time.Millisecond // Adjust as needed
// 	orderMatrix := OrderMatrix{}
// 	//prevFloor := elev.Floor
// 	obstructionActive := false
// 	lastKnownDirection := MotorDirection(0)
// 	doorOpenCh := make(chan bool, 200)
// 	doorClose := time.NewTimer(3 * time.Second)
// 	doorClose.Stop()
// 	errorTimeout := time.NewTimer(5 * time.Second)
// 	elevio.SetDoorOpenLamp(false)

// 	for {
// 		select {

// 		case <-doorOpenCh:
// 			println("FSM: Door Open")
// 			elev.Behaviour = EB_DoorOpen
// 			elevio.SetMotorDirection(MD_Stop)
// 			elevio.SetDoorOpenLamp(true)
// 			doorClose.Reset(3 * time.Second)
// 			errorTimeout.Stop()
// 			ch_fsm.Ch_clearedFloor <- elev.Floor

// 		case <-doorClose.C:

// 			if obstructionActive {
// 				doorClose.Reset(3 * time.Second)
// 				break
// 			}
// 			println("FSM: Door Close")
// 			elevio.SetDoorOpenLamp(false)
// 			if orderMatrix == (OrderMatrix{}) {
// 				elev.Behaviour = EB_Idle
// 				errorTimeout.Stop()
// 				break
// 			} else {
// 				dirnpair := requests.RequestsChooseDirection(orderMatrix, elev)
// 				elev.Dirn = dirnpair.Dirn
// 				elev.Behaviour = ElevatorBehaviour(dirnpair.Behaviour)
// 				elevio.SetMotorDirection(elev.Dirn)
// 				errorTimeout.Reset(5 * time.Second)
// 			}

// 		case receivedOrder := <-ch_fsm.Ch_localOrders:
// 			fmt.Println("Received order:", receivedOrder)
// 			orderMatrix = receivedOrder
// 			switch elev.Behaviour {
// 			case EB_Idle:
// 				if requests.RequestsHere(orderMatrix, elev.Floor) {
// 					doorOpenCh <- true
// 					break
// 				}
// 				dirnPair := requests.RequestsChooseDirection(orderMatrix, elev)
// 				elev.Behaviour = EB_Moving
// 				elev.Dirn = dirnPair.Dirn
// 				elevio.SetMotorDirection(elev.Dirn)
// 				errorTimeout.Reset(5 * time.Second)

// 			case EB_Moving:
// 				break
// 			case EB_DoorOpen:
// 				if requests.RequestsHere(orderMatrix, elev.Floor) {
// 					doorOpenCh <- true
// 					break
// 				}
// 			}
// 			orderMatrix, elev = FsmButtonPressed(orderMatrix, elev)
// 			orderMatrix = lights.SetCabLights(orderMatrix)

// 		case currentFloor := <-ch_fsm.Ch_floorSensor:
// 			elev.Floor = currentFloor
// 			elevio.SetFloorIndicator(elev.Floor)
// 			fmt.Println("Floor sensor event")
// 				switch elev.Behaviour {
// 				case EB_DoorOpen, EB_Idle:
// 					elevio.SetMotorDirection(MD_Stop)
// 					errorTimeout.Stop()
// 				case EB_Moving:
// 					if requests.RequestsShouldStop(orderMatrix, elev) {
// 						elevio.SetMotorDirection(MD_Stop)
// 						doorOpenCh <- true
// 						break
// 					}
// 					if orderMatrix == (OrderMatrix{}) {
// 						elevio.SetMotorDirection(MD_Stop)
// 						elev.Behaviour = EB_Idle
// 						errorTimeout.Stop()
// 						break
// 					}
// 					switch elev.Dirn {
// 					case MD_Down:
// 						if !requests.RequestsAbove(orderMatrix, elev.Floor) {
// 							elev.Dirn = MD_Down
// 							elevio.SetMotorDirection(MD_Down)
// 						}
// 					case MD_Up:
// 						if !requests.RequestsBelow(orderMatrix, elev.Floor) {
// 							elev.Dirn = MD_Up
// 							elevio.SetMotorDirection(MD_Up)
// 						}
// 					}
// 					errorTimeout.Reset(5 * time.Second)
				
// 				//prevFloor = currentFloor
// 				obstructionActive = false
// 				elev.Available = true

// 				// updateState := NetworkMessage{
// 				// 	MsgData:       elev,
// 				// 	Receipient: Master,
// 				// 	MsgType:     "stateupdatech"}
// 				ch_fsm.Ch_stateUpdate <- elev

// 				fmt.Printf("Received floor sensor event: %d\n", currentFloor)
// 			}

// 			// if currentFloor != prevFloor {
// 			// 	fmt.Printf("Arrived at floor %d\n", currentFloor)
// 			// 	clearedFloor := false
// 			// 	orderMatrix, elev, clearedFloor = fsmFloorArrival(orderMatrix, currentFloor, elev)
// 			// 	elevio.SetFloorIndicator(currentFloor) // Update floor indicator lamp

// 			// 	// ch_fsm.Ch_stateUpdate<-elev
// 			// 	if !obstructionActive {
// 			// 		timer.TimerStop()
// 			// 		timer.TimerStart(3.0)
// 			// 		// fmt.Println("ti//Passes the updated statemer started")
// 			// 	}
// 			// 	fmt.Println("Cleared floor", clearedFloor)
// 			// 	if clearedFloor {
// 			// 		ch_fsm.Ch_clearedFloor <- currentFloor
// 			// 	}
// 			// 	fmt.Println("Cleared floor")

// 			// 	prevFloor = currentFloor
// 			// 	obstructionActive = false
// 			// 	elev.Available = true
// 			// 	// updateElevator := NetworkMessage{MsgType: "elevatorupdatechannel", MsgData:  elev,Receipient: Master}
// 			// 	ch_fsm.Ch_stateUpdate <- elev
// 			// 	fmt.Println("Elevator state updated")

// 			// }

// 		case stopPressed := <-ch_fsm.Ch_stopButton:
// 			// Handle stop button event
// 			if stopPressed {
// 				lastKnownDirection = elev.Dirn
// 				fmt.Println("Stop button pressed!")
// 				fmt.Println(lastKnownDirection)
// 				elevio.SetStopLamp(true)
// 				elevio.SetMotorDirection(0)
// 				// stop = true
// 			} else {
// 				fmt.Println("Stop button released!")
// 				elevio.SetStopLamp(false)
// 			}

		
// 		case obstruction := <-ch_fsm.Ch_obstruction:
// 			fmt.Println("Obstruction detected")
// 			if obstruction {
// 				obstructionActive = true
// 				elev.Available = false
// 				// timer.TimerStop()
// 				fmt.Println("obstruction switch")
// 			} else if !obstruction {
// 				obstructionActive = false
// 				elev.Available = true
// 				// timer.TimerStop()
// 				// timer.TimerStart(3.0)
// 				fmt.Println("obstruction switch off")
// 			}

// 			// ch_fsm.Ch_stateUpdate <- elev //Passes the updated state
// 		}
// 	}
// }




