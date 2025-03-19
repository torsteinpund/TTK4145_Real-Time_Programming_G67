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
	Ch_floorSensor      chan int
	Ch_stopButton 		chan bool
	Ch_obstruction   	chan bool
	Ch_localLights		chan OrderMatrix
	Ch_localOrders		chan OrderMatrix
	Ch_toMaster 		chan NetworkMessage
	Ch_clearedFloor 	chan int
}


func FsmInitBetweenFloors() (ElevatorBehaviour, MotorDirection) {
	// Move the elevator down until it reaches a floor
	
	for{
		elevio.SetMotorDirection(MD_Down)
		if elevio.GetFloor() != -1 {
			break
		}
	}
	// Update the elevator's state
	dirn := MD_Stop
	elevio.SetMotorDirection(MD_Stop)
	behaviour := ElevatorBehaviour(EB_Idle)
	return behaviour, dirn
}

func FsmButtonPressed(elev Elevator, ch_doorOpen chan<- bool) Elevator {

	

	switch elev.Behaviour {
	
	case EB_DoorOpen:
		if requests.RequestsHere(elev.Requests, elev.Floor) {
			ch_doorOpen <- true
			break
		}

	case EB_Moving:
		break

	case EB_Idle:
		if requests.RequestsHere(elev.Requests, elev.Floor) {
			ch_doorOpen <- true
			break
		}
		dirnBehaviour := requests.RequestsChooseDirection(elev)
	
		elev.Dirn = dirnBehaviour.Dirn
		elev.Behaviour = ElevatorBehaviour(dirnBehaviour.Behaviour)
		elevio.SetMotorDirection(elev.Dirn)

	}

	return elev
	
}




func fsmFloorArrival(newFloor int, elev Elevator, ch_doorOpen chan<- bool) Elevator {

	elev.Floor = newFloor

	elevio.SetFloorIndicator(elev.Floor)

	switch elev.Behaviour {
	case ElevatorBehaviour(EB_Moving):
		// Check if the elevator should stop at the current floor
		if requests.RequestsShouldStop(elev) {
			ch_doorOpen <- true
			break
		}

		// Do not necessairly need this
		// switch elev.Dirn {
		// case MD_Up:
		// 	if !requests.RequestsAbove(elev.Requests, elev.Floor) {
		// 		elev.Dirn = MD_Down
		// 		elevio.SetMotorDirection(MD_Down)
		// 	}
			
		// case MD_Down:
		// 	if !requests.RequestsBelow(elev.Requests, elev.Floor) {
		// 		elev.Dirn = MD_Stop
		// 		elevio.SetMotorDirection(MD_Up)
		// 	}
		// }

	default:
		// No action
		//elevio.SetMotorDirection(MD_Stop)
	}

	return elev
}

func fsmDoorTimeout(elev Elevator) Elevator {

	switch elev.Behaviour {
	case ElevatorBehaviour(EB_DoorOpen):
		// Choose direction based on requests
		dirnBehaviour := requests.RequestsChooseDirection(elev)
		elev.Dirn = dirnBehaviour.Dirn
		elev.Behaviour = ElevatorBehaviour(dirnBehaviour.Behaviour)

		switch elev.Behaviour {
		case ElevatorBehaviour(EB_DoorOpen):
			// Start timer and clear requests
			timer.TimerStart(elev.Config.DoorOpenDuration)
			elev = requests.RequestsClearAtCurrentFloor(elev, nil)
			// elev.Requests = lights.SetHallLights(elev.Requests)
			elev.Requests = lights.SetCabLights(elev.Requests)

		case ElevatorBehaviour(EB_Moving), ElevatorBehaviour(EB_Idle):
			// Shut the door and start moving
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elev.Dirn)
		}

	default:
		// No action
	}

	return elev
}


func FsmRun(ch_fsm FsmChannels, elev Elevator) {
	fmt.Println("FSM Started!")
	if initialFloor := elevio.GetFloor(); initialFloor == -1 {
		fmt.Println("Elevator is between floors on startup. Running initialization...")
		elev.Behaviour, elev.Dirn = FsmInitBetweenFloors()
	}

	// Polling rate configuration
	inputPollRate := 25 * time.Millisecond // Adjust as needed


	// Initialize system state
	prevFloor := -1
	//timerActive := false
	//var timerEndTime float64
	doorOpenCh := make(chan bool)
	obstructionActive := false
	lastKnownDirection := MotorDirection(0)
	imAliveSignal := time.NewTimer(1 * time.Second)
	// stop := false

	// Main event loop
	for {
		select {

		case <-doorOpenCh:
			fmt.Println("Door open event")
			elevio.SetMotorDirection(MD_Stop)
			elevio.SetDoorOpenLamp(true)
			elev = requests.RequestsClearAtCurrentFloor(elev, nil)
			timer.TimerStart(elev.Config.DoorOpenDuration)
			elev.Requests = lights.SetCabLights(elev.Requests)
			elev.Behaviour = ElevatorBehaviour(EB_DoorOpen)
		
		case receivedOrder := <- ch_fsm.Ch_localOrders:
			fmt.Println("Received order")
			elev.Requests = receivedOrder
			// fmt.Println(elev.Requests)
			elev = FsmButtonPressed(elev, doorOpenCh)

		case currentFloor := <-ch_fsm.Ch_floorSensor:
			// Handle floor sensor event
			fmt.Printf("Received floor sensor event: %d\n", currentFloor)

			if currentFloor != prevFloor {
				fmt.Printf("Arrived at floor %d\n", currentFloor)
				elev = fsmFloorArrival(currentFloor, elev, doorOpenCh)
				elevio.SetFloorIndicator(currentFloor) // Update floor indicator lamp
				fmt.Println("After FSMFLOORARRIVAL," ,elev.Requests)
				// ch_fsm.Ch_stateUpdate<-elev
				if !obstructionActive {
					timer.TimerStop()
					timer.TimerStart(3.0)
					// fmt.Println("ti//Passes the updated statemer started")
				}
				// Stop and restart the timer when arriving at a floor
				// Set door timeout to 3 seconds
			}
			prevFloor = currentFloor
			obstructionActive = false
			elev.Avaliable = true
			updateElevator := NetworkMessage{MsgType: "elevatorupdatechannel", MsgData: elev, Receipient: Master}
			ch_fsm.Ch_toMaster <- updateElevator



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
				fmt.Println("Door timeout occurred. Elevator stopped")
				elev.Avaliable = false
				updateElevator := NetworkMessage{MsgType: "elevatorupdatechannel", MsgData: elev, Receipient: Master}
				ch_fsm.Ch_toMaster <- updateElevator
				// fmt.Println("Door timeout occurred.")
				// elev = fsmDoorTimeout(elev)
				//timer.TimerStop() // Reset the timer after timeout handling
			}
		case obstruction := <-ch_fsm.Ch_obstruction:
			fmt.Println("Obstruction detected")
			if obstruction {
				obstructionActive = true
				elev.Avaliable = false
				timer.TimerStop()
				fmt.Println("obstruction switch")
			} else if !obstruction {
				obstructionActive = false
				elev.Avaliable = true
				timer.TimerStop()
				timer.TimerStart(3.0)
				fmt.Println("obstruction switch off")
			}


		case setLights := <-ch_fsm.Ch_localLights:
			elev.Requests = lights.SetCabLights(setLights)

		case <-imAliveSignal.C:
			updateElevator := NetworkMessage{MsgType: "elevatorupdatechannel", MsgData: elev, Receipient: Master}
			imAliveSignal.Reset(1 * time.Second)
			ch_fsm.Ch_toMaster <- updateElevator
		}


		// ch_fsm.Ch_stateUpdate <- elev //Passes the updated state
	}
}


