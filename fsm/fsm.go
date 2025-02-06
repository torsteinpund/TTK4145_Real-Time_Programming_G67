package fsm

import (
	"Driver-go/elevio"
	"Driver-go/elevator"
	"Driver-go/timer"
	"Driver-go/requests"
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

func SetAllLights(e elevator.Elevator) {
	for floor := 0; floor < elevio.NumFloors; floor++ {
		for btn := 0; btn < elevio.NumButtonTypes; btn++ {
			state := e.Requests[floor][btn]
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, state == 1)
		}
	}
}

func FsmOnInitBetweenFloors(e elevator.Elevator) {
	// Sett motoren til å bevege seg nedover
	elevio.SetMotorDirection(elevio.MD_Down)

	// Oppdater heisens retning og oppførsel
	e.Dirn = elevio.MD_Down
	e.Behaviour = elevator.ElevatorBehaviour(elevator.EB_Moving)
}


func FsmOnRequestButtonPress(btnFloor int, btnType elevio.ButtonType) {
	fmt.Printf("\n\nfsmOnRequestButtonPress(%d, %v)\n", btnFloor, btnType)
	//elevatorPrint(elevator)

	switch elevator.Elev.Behaviour {
	case elevator.ElevatorBehaviour(elevator.EB_DoorOpen):
		if requests.RequestsShouldClearImmediately(elevator.Elev, btnFloor, btnType) {
			timer.TimerStart(elevator.Elev.Config.DoorOpenDurationS)
		} else {
			elevator.Elev.Requests[btnFloor][btnType] = 1
		}

	case elevator.ElevatorBehaviour(elevator.EB_Moving):
		elevator.Elev.Requests[btnFloor][btnType] = 1

	case elevator.ElevatorBehaviour(elevator.EB_Idle):
		elevator.Elev.Requests[btnFloor][btnType] = 1
		dirnBehaviour := requests.RequestsChooseDirection(elevator.Elev)
		elevator.Elev.Dirn = dirnBehaviour.Dirn
		elevator.Elev.Behaviour = elevator.ElevatorBehaviour(dirnBehaviour.Behaviour)

		switch dirnBehaviour.Behaviour {
		case elevator.EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			timer.TimerStart(elevator.Elev.Config.DoorOpenDurationS)
			elevator.Elev = requests.RequestsClearAtCurrentFloor(elevator.Elev)

		case elevator.EB_Moving:
			elevio.SetMotorDirection(elevator.Elev.Dirn)

		case elevator.EB_Idle:
			// Ingen handling nødvendig
		}
	}

	// Oppdater knappelysene
	SetAllLights(elevator.Elev)

	// Logg den nye tilstanden til heisen
	fmt.Println("\nNew state:")
	//elevatorPrint(elevator)
}

func FsmOnFloorArrival(newFloor int) {
	fmt.Printf("\n\nfsmOnFloorArrival(%d)\n", newFloor)
	//elevatorPrint(elevator)

	// Oppdater heisens nåværende etasje
	elevator.Elev.Floor = newFloor

	// Oppdater etasjeindikatoren
	elevio.SetFloorIndicator(elevator.Elev.Floor)

	switch elevator.Elev.Behaviour {
	case elevator.ElevatorBehaviour(elevator.EB_Moving):
		// Sjekk om heisen skal stoppe i denne etasjen
		if requests.RequestsShouldStop(elevator.Elev) {
			// Stopp motoren
			elevio.SetMotorDirection(elevio.MD_Stop)

			// Slå på dørlyset
			elevio.SetDoorOpenLamp(true)

			// Rydd forespørsler for nåværende etasje
			elevator.Elev = requests.RequestsClearAtCurrentFloor(elevator.Elev)

			// Start timer for å holde dørene åpne
			timer.TimerStart(elevator.Elev.Config.DoorOpenDurationS)

			// Oppdater knappelysene
			SetAllLights(elevator.Elev)

			// Endre heisens oppførsel til "DoorOpen"
			elevator.Elev.Behaviour = elevator.ElevatorBehaviour(elevator.EB_DoorOpen)
		}

	default:
		// Ingen spesifikk handling for andre oppførsler
	}

	// Logg den nye tilstanden til heisen
	fmt.Println("\nNew state:")
	//elevatorPrint(elevator)
}

func FsmOnDoorTimeout() {
	fmt.Printf("\n\nfsmOnDoorTimeout()\n")
	//elevatorPrint(elevator)

	switch elevator.Elev.Behaviour {
	case elevator.ElevatorBehaviour(elevator.EB_DoorOpen):
		// Velg neste retning og oppførsel basert på forespørsler
		dirnBehaviour := requests.RequestsChooseDirection(elevator.Elev)
		elevator.Elev.Dirn = dirnBehaviour.Dirn
		elevator.Elev.Behaviour = elevator.ElevatorBehaviour(dirnBehaviour.Behaviour)

		switch elevator.Elev.Behaviour {
		case elevator.ElevatorBehaviour(elevator.EB_DoorOpen):
			// Start timer på nytt og rydd forespørsler i nåværende etasje
			timer.TimerStart(elevator.Elev.Config.DoorOpenDurationS)
			elevator.Elev = requests.RequestsClearAtCurrentFloor(elevator.Elev)
			SetAllLights(elevator.Elev)

		case elevator.ElevatorBehaviour(elevator.EB_Moving), elevator.ElevatorBehaviour(elevator.EB_Idle):
			// Lukk dørene og sett motorretning
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevator.Elev.Dirn)
		}

	default:
		// Ingen handling for andre tilstander
	}

	fmt.Println("\nNew state:")
	//elevatorPrint(elevator)
}
