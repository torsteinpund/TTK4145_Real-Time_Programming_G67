package fsm

import (
	"Driver-go/elevio"
	"Driver-go/requests"
	"Driver-go/timer"
	. "Driver-go/types"
	"fmt"
)

// fsmInit initialiserer heisens tilstand og tilhørende systemer
/*func fsmInit() {
	// Initialiser heisen med standardverdier
	elevator = ElevatorUninitialized()

	// Last inn konfigurasjon fra fil (simulert)
	err := loadConfig("con", &elevator)
	if err != nil {
		fmt.Println("Error loading configuration:", err)
	}

	// Initialiser output-enheten (simulert)
	outputDevice = getOutputDevice()
}*/

func SetAllLocalLights(req [NUMFLOORS][NUMBUTTONTYPE]int) [NUMFLOORS][NUMBUTTONTYPE]int {
	for floor := 0; floor < NUMFLOORS; floor++ {
		for btn := 0; btn < NUMBUTTONTYPE; btn++ {
			state := req[floor][btn]
			elevio.SetButtonLamp(ButtonType(btn), floor, state == 1)
		}
	}
	return req
}

func FsmInitBetweenFloors() (ElevatorBehaviour, MotorDirection) {
	// Move the elevator down until it reaches a floor
	elevio.SetMotorDirection(MD_Down)

	// Update the elevator's state
	dirn := MD_Down
	behaviour := ElevatorBehaviour(EB_Moving)
	return behaviour, dirn
}

func FsmButtonPressed(btnFloor int, btnType ButtonType, elev Elevator) Elevator {
	fmt.Printf("\n\nfsmOnRequestButtonPress(%d, %v)\n", btnFloor, btnType)

	switch elev.Behaviour {
	case ElevatorBehaviour(EB_DoorOpen):
		if requests.RequestsShouldClearImmediately(elev, btnFloor, btnType) {
			timer.TimerStart(elev.Config.DoorOpenDuration)
		} else {
			elev.Requests[btnFloor][btnType] = 1
		}

	case ElevatorBehaviour(EB_Moving):
		elev.Requests[btnFloor][btnType] = 1

	case ElevatorBehaviour(EB_Idle):
		elev.Requests[btnFloor][btnType] = 1
		dirnBehaviour := requests.RequestsChooseDirection(elev)
		elev.Dirn = dirnBehaviour.Dirn
		elev.Behaviour = ElevatorBehaviour(dirnBehaviour.Behaviour)

		switch dirnBehaviour.Behaviour {
		case EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			timer.TimerStart(elev.Config.DoorOpenDuration)
			elev = requests.RequestsClearAtCurrentFloor(elev, nil)

		case EB_Moving:
			elevio.SetMotorDirection(elev.Dirn)

		case EB_Idle:
			//No action
		}
	}

	// Update button lights
	elev.Requests = SetAllLocalLights(elev.Requests)

	return elev
}

func FsmFloorArrival(newFloor int, elev Elevator) Elevator {
	fmt.Printf("\n\nfsmOnFloorArrival(%d)\n", newFloor)

	elev.Floor = newFloor

	elevio.SetFloorIndicator(elev.Floor)

	switch elev.Behaviour {
	case ElevatorBehaviour(EB_Moving):
		// Check if the elevator should stop at the current floor
		if requests.RequestsShouldStop(elev) {
			elevio.SetMotorDirection(MD_Stop)

			elevio.SetDoorOpenLamp(true)

			elev = requests.RequestsClearAtCurrentFloor(elev, nil)

			timer.TimerStart(elev.Config.DoorOpenDuration)

			elev.Requests = SetAllLocalLights(elev.Requests)

			elev.Behaviour = ElevatorBehaviour(EB_DoorOpen)
		}

	default:
		// No action
	}

	return elev
}

func FsmDoorTimeout(elev Elevator) Elevator {

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
			elev.Requests = SetAllLocalLights(elev.Requests)

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
