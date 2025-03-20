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
}

func FsmButtonPressed(orderMatrix OrderMatrix, elev Elevator) (OrderMatrix, Elevator) {

	dirnBehaviour := requests.RequestsChooseDirection(orderMatrix, elev)

	elev.Dirn = dirnBehaviour.Dirn
	elev.Behaviour = ElevatorBehaviour(dirnBehaviour.Behaviour)

	switch dirnBehaviour.Behaviour {

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

func fsmFloorArrival(orderMatrix OrderMatrix, newFloor int, elev Elevator) (OrderMatrix, Elevator) {

	elev.Floor = newFloor

	elevio.SetFloorIndicator(elev.Floor)

	switch elev.Behaviour {
	case ElevatorBehaviour(EB_Moving):
		// Check if the elevator should stop at the current floor
		if requests.RequestsShouldStop(orderMatrix, elev) {
			// ch_doorOpen <- true
			fmt.Println("Door open event")
			elevio.SetMotorDirection(MD_Stop)
			elevio.SetDoorOpenLamp(true)
			orderMatrix, elev = requests.RequestsClearAtCurrentFloor(orderMatrix, elev)
			timer.TimerStart(elev.Config.DoorOpenDuration)
			orderMatrix = lights.SetCabLights(orderMatrix)
			elev.Behaviour = ElevatorBehaviour(EB_DoorOpen)
		}
	default:
		// No action
		//elevio.SetMotorDirection(MD_Stop)
	}

	return orderMatrix, elev
}

func fsmDoorTimeout(orderMatrix OrderMatrix, elev Elevator) (OrderMatrix, Elevator) {

	switch elev.Behaviour {
	case ElevatorBehaviour(EB_DoorOpen):
		// Choose direction based on requests
		dirnBehaviour := requests.RequestsChooseDirection(orderMatrix, elev)
		elev.Dirn = dirnBehaviour.Dirn
		elev.Behaviour = ElevatorBehaviour(dirnBehaviour.Behaviour)

		switch elev.Behaviour {
		case ElevatorBehaviour(EB_DoorOpen):
			// Start timer and clear requests
			timer.TimerStart(elev.Config.DoorOpenDuration)
			orderMatrix, elev = requests.RequestsClearAtCurrentFloor(orderMatrix, elev)
			// elev.Requests = lights.SetHallLights(elev.Requests)
			orderMatrix = lights.SetCabLights(orderMatrix)

		case ElevatorBehaviour(EB_Moving), ElevatorBehaviour(EB_Idle):
			// Shut the door and start moving
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elev.Dirn)
		}

	default:
		// No action
	}

	return orderMatrix, elev
}

func FsmRun(ch_fsm FsmChannels, elev Elevator) {
    fmt.Println("FSM Started!")

    // Polling rate configuration
    inputPollRate := 25 * time.Millisecond // Adjust as needed
    orderMatrix := OrderMatrix{}

    // Initialize system state
    prevFloor := -1
    obstructionActive := false

    // Main event loop
    for {
        select {
        case obstruction := <-ch_fsm.Ch_obstruction:
            fmt.Println("Obstruction detected")
            if obstruction {
                obstructionActive = true
                elev.Avaliable = false
                timer.TimerStop()
                elevio.SetDoorOpenLamp(true) // Ensure the door stays open
                fmt.Println("Obstruction active, door remains open")
            } else {
                obstructionActive = false
                elev.Avaliable = true
                timer.TimerStart(3.0) // Restart the timer
                fmt.Println("Obstruction cleared, restarting timer")
            }

        case receivedOrder := <-ch_fsm.Ch_localOrders:
            fmt.Println("Received order:", receivedOrder)
            orderMatrix = receivedOrder
            orderMatrix, elev = FsmButtonPressed(orderMatrix, elev)
            orderMatrix = lights.SetCabLights(orderMatrix)

        case currentFloor := <-ch_fsm.Ch_floorSensor:
            fmt.Printf("Received floor sensor event: %d\n", currentFloor)

            if currentFloor != prevFloor {
                fmt.Printf("Arrived at floor %d\n", currentFloor)
                orderMatrix, elev = fsmFloorArrival(orderMatrix, currentFloor, elev)
                elevio.SetFloorIndicator(currentFloor) // Update floor indicator lamp

                if elev.Behaviour == EB_DoorOpen && !obstructionActive {
                    timer.TimerStop()
                    timer.TimerStart(3.0) // Start the door timer
                }
            }
            prevFloor = currentFloor
            obstructionActive = false
            elev.Avaliable = true
            ch_fsm.Ch_stateUpdate <- elev // Send updated state

        case stopPressed := <-ch_fsm.Ch_stopButton:
            if stopPressed {
                fmt.Println("Stop button pressed!")
                elevio.SetStopLamp(true)
                elevio.SetMotorDirection(MD_Stop)
            } else {
                fmt.Println("Stop button released!")
                elevio.SetStopLamp(false)
            }

        case <-time.After(inputPollRate):
            // Periodic tasks (check timer)
            if timer.TimerTimedOut() {
                fmt.Println("Door timeout occurred. Closing door.")
                elevio.SetDoorOpenLamp(false) // Close the door
                orderMatrix, elev = fsmDoorTimeout(orderMatrix, elev)
                timer.TimerStop() // Reset the timer after timeout handling

                // If no further orders, transition to idle
                if elev.Behaviour == EB_Idle {
                    fmt.Println("Elevator is idle. Waiting for new orders.")
                }
            }
        }
    }
}