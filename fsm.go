package main

import (
	"Driver-go/elevio"
	"fmt"
)


type ElevOutputDevice struct{}

func (d ElevOutputDevice) SetMotorDirection(dir elevio.MotorDirection) {
	elevio.SetMotorDirection(dir)
}

func (d ElevOutputDevice) SetButtonLamp(button elevio.ButtonType, floor int, value bool) {
	elevio.SetButtonLamp(button, floor, value)
}

func (d ElevOutputDevice) SetFloorIndicator(floor int) {
	elevio.SetFloorIndicator(floor)
}

func (d ElevOutputDevice) SetDoorOpenLamp(value bool) {
	elevio.SetDoorOpenLamp(value)
}

func (d ElevOutputDevice) SetStopLamp(value bool) {
	elevio.SetStopLamp(value)
}




var (
	elevator     Elevator
	outputDevice ElevOutputDevice
)

// fsmInit initialiserer heisens tilstand og tilhørende systemer
/*func fsmInit() {
	// Initialiser heisen med standardverdier
	elevator = ElevatorUninitialized()

	// Last inn konfigurasjon fra fil (simulert)
	err := loadConfig("elevator.con", &elevator)
	if err != nil {
		fmt.Println("Error loading configuration:", err)
	}

	// Initialiser output-enheten (simulert)
	outputDevice = getOutputDevice()
}*/

func setAllLights(e Elevator) {
	for floor := 0; floor < elevio.NumFloors; floor++ {
		for btn := 0; btn < elevio.NumButtonTypes; btn++ {
			state := e.Requests[floor][btn]
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, state == 1)
		}
	}
}

func fsmOnInitBetweenFloors() {
	// Sett motoren til å bevege seg nedover
	elevio.SetMotorDirection(elevio.MD_Down)

	// Oppdater heisens retning og oppførsel
	elevator.Dirn = elevio.MD_Down
	elevator.Behaviour = ElevatorBehaviour(EB_Moving)
}


func fsmOnRequestButtonPress(btnFloor int, btnType elevio.ButtonType) {
	fmt.Printf("\n\nfsmOnRequestButtonPress(%d, %v)\n", btnFloor, btnType)
	//elevatorPrint(elevator)

	switch elevator.Behaviour {
	case ElevatorBehaviour(EB_DoorOpen):
		if requestsShouldClearImmediately(elevator, btnFloor, btnType) {
			timerStart(elevator.Config.DoorOpenDurationS)
		} else {
			elevator.Requests[btnFloor][btnType] = 1
		}

	case ElevatorBehaviour(EB_Moving):
		elevator.Requests[btnFloor][btnType] = 1

	case ElevatorBehaviour(EB_Idle):
		elevator.Requests[btnFloor][btnType] = 1
		dirnBehaviour := requestsChooseDirection(elevator)
		elevator.Dirn = dirnBehaviour.Dirn
		elevator.Behaviour = ElevatorBehaviour(dirnBehaviour.Behaviour)

		switch dirnBehaviour.Behaviour {
		case EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			timerStart(elevator.Config.DoorOpenDurationS)
			elevator = requestsClearAtCurrentFloor(elevator)

		case EB_Moving:
			elevio.SetMotorDirection(elevator.Dirn)

		case EB_Idle:
			// Ingen handling nødvendig
		}
	}

	// Oppdater knappelysene
	setAllLights(elevator)

	// Logg den nye tilstanden til heisen
	fmt.Println("\nNew state:")
	//elevatorPrint(elevator)
}

func fsmOnFloorArrival(newFloor int) {
	fmt.Printf("\n\nfsmOnFloorArrival(%d)\n", newFloor)
	//elevatorPrint(elevator)

	// Oppdater heisens nåværende etasje
	elevator.Floor = newFloor

	// Oppdater etasjeindikatoren
	elevio.SetFloorIndicator(elevator.Floor)

	switch elevator.Behaviour {
	case ElevatorBehaviour(EB_Moving):
		// Sjekk om heisen skal stoppe i denne etasjen
		if requestsShouldStop(elevator) {
			// Stopp motoren
			elevio.SetMotorDirection(elevio.MD_Stop)

			// Slå på dørlyset
			elevio.SetDoorOpenLamp(true)

			// Rydd forespørsler for nåværende etasje
			elevator = requestsClearAtCurrentFloor(elevator)

			// Start timer for å holde dørene åpne
			timerStart(elevator.Config.DoorOpenDurationS)

			// Oppdater knappelysene
			setAllLights(elevator)

			// Endre heisens oppførsel til "DoorOpen"
			elevator.Behaviour = ElevatorBehaviour(EB_DoorOpen)
		}

	default:
		// Ingen spesifikk handling for andre oppførsler
	}

	// Logg den nye tilstanden til heisen
	fmt.Println("\nNew state:")
	//elevatorPrint(elevator)
}

func fsmOnDoorTimeout() {
	fmt.Printf("\n\nfsmOnDoorTimeout()\n")
	//elevatorPrint(elevator)

	switch elevator.Behaviour {
	case ElevatorBehaviour(EB_DoorOpen):
		// Velg neste retning og oppførsel basert på forespørsler
		dirnBehaviour := requestsChooseDirection(elevator)
		elevator.Dirn = dirnBehaviour.Dirn
		elevator.Behaviour = ElevatorBehaviour(dirnBehaviour.Behaviour)

		switch elevator.Behaviour {
		case ElevatorBehaviour(EB_DoorOpen):
			// Start timer på nytt og rydd forespørsler i nåværende etasje
			timerStart(elevator.Config.DoorOpenDurationS)
			elevator = requestsClearAtCurrentFloor(elevator)
			setAllLights(elevator)

		case ElevatorBehaviour(EB_Moving), ElevatorBehaviour(EB_Idle):
			// Lukk dørene og sett motorretning
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevator.Dirn)
		}

	default:
		// Ingen handling for andre tilstander
	}

	fmt.Println("\nNew state:")
	//elevatorPrint(elevator)
}
