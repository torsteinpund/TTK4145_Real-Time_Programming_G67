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

func SetAllLights(req [elevator.NUMFLOORS][elevator.NUMBUTTONTYPE]int) [elevator.NUMFLOORS][elevator.NUMBUTTONTYPE]int{
	for floor := 0; floor < elevio.NumFloors; floor++ {
		for btn := 0; btn < elevio.NumButtonTypes; btn++ {
			state := req[floor][btn]
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, state == 1)
		}
	}
	return req
}

func FsmOnInitBetweenFloors() (elevator.ElevatorBehaviour, elevio.MotorDirection)  {
	// Sett motoren til å bevege seg nedover
	elevio.SetMotorDirection(elevio.MD_Down)

	// Oppdater heisens retning og oppførsel
	dirn := elevio.MD_Down
	behaviour := elevator.ElevatorBehaviour(elevator.EB_Moving)
	return behaviour, dirn
}


func FsmOnRequestButtonPress(btnFloor int, btnType elevio.ButtonType, elev elevator.Elevator) elevator.Elevator{
	fmt.Printf("\n\nfsmOnRequestButtonPress(%d, %v)\n", btnFloor, btnType)
	//elevatorPrint(elevator)
	switch elev.Behaviour {
	case elevator.ElevatorBehaviour(elevator.EB_DoorOpen):
		if requests.RequestsShouldClearImmediately(elev, btnFloor, btnType) {
			timer.TimerStart(elev.Config.DoorOpenDuration)
		} else {
			elev.Requests[btnFloor][btnType] = 1
		}

	case elevator.ElevatorBehaviour(elevator.EB_Moving):
		elev.Requests[btnFloor][btnType] = 1

	case elevator.ElevatorBehaviour(elevator.EB_Idle):
		elev.Requests[btnFloor][btnType] = 1
		dirnBehaviour := requests.RequestsChooseDirection(elev)
		elev.Dirn = dirnBehaviour.Dirn
		elev.Behaviour = elevator.ElevatorBehaviour(dirnBehaviour.Behaviour)

		switch dirnBehaviour.Behaviour {
		case elevator.EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			timer.TimerStart(elev.Config.DoorOpenDuration)
			elev = requests.RequestsClearAtCurrentFloor(elev, nil)

		case elevator.EB_Moving:
			elevio.SetMotorDirection(elev.Dirn)

		case elevator.EB_Idle:
			// Ingen handling nødvendig
		}
	}

	// Oppdater knappelysene
	elev.Requests = SetAllLights(elev.Requests)

	// Logg den nye tilstanden til heisen
	fmt.Println("\nNew state:")
	//elevatorPrint(elevator)
	return elev
}

func FsmOnFloorArrival(newFloor int, elev elevator.Elevator) elevator.Elevator {
	fmt.Printf("\n\nfsmOnFloorArrival(%d)\n", newFloor)
	//elevatorPrint(elevator)

	// Oppdater heisens nåværende etasje
	elev.Floor = newFloor

	// Oppdater etasjeindikatorenhttps://github.com/TTK4145?q=driver
	elevio.SetFloorIndicator(elev.Floor)

	switch elev.Behaviour {
	case elevator.ElevatorBehaviour(elevator.EB_Moving):
		// Sjekk om heisen skal stoppe i denne etasjen
		if requests.RequestsShouldStop(elev) {
			// Stopp motoren
			elevio.SetMotorDirection(elevio.MD_Stop)

			// Slå på dørlyset
			elevio.SetDoorOpenLamp(true)

			// Rydd forespørsler for nåværende etasje
			elev = requests.RequestsClearAtCurrentFloor(elev, nil)

			// Start timer for å holde dørene åpne
			timer.TimerStart(elev.Config.DoorOpenDuration)

			// Oppdater knappelysene
			elev.Requests = SetAllLights(elev.Requests)

			// Endre heisens oppførsel til "DoorOpen"
			elev.Behaviour = elevator.ElevatorBehaviour(elevator.EB_DoorOpen)
		}

	default:
		// Ingen spesifikk handling for andre oppførsler
	}

	// Logg den nye tilstanden til heisen
	fmt.Println("\nNew state:")
	//elevatorPrint(elevator)
	return elev
}

func FsmOnDoorTimeout(elev elevator.Elevator) elevator.Elevator{
	fmt.Printf("\n\nfsmOnDoorTimeout()\n")
	//elevatorPrint(elevator)

	switch elev.Behaviour {
	case elevator.ElevatorBehaviour(elevator.EB_DoorOpen):
		// Velg neste retning og oppførsel basert på forespørsler
		dirnBehaviour := requests.RequestsChooseDirection(elev)
		elev.Dirn = dirnBehaviour.Dirn
		elev.Behaviour = elevator.ElevatorBehaviour(dirnBehaviour.Behaviour)

		switch elev.Behaviour {
		case elevator.ElevatorBehaviour(elevator.EB_DoorOpen):
			// Start timer på nytt og rydd forespørsler i nåværende etasje
			timer.TimerStart(elev.Config.DoorOpenDuration)
			elev = requests.RequestsClearAtCurrentFloor(elev, nil)
			elev.Requests = SetAllLights(elev.Requests)

		case elevator.ElevatorBehaviour(elevator.EB_Moving), elevator.ElevatorBehaviour(elevator.EB_Idle):
			// Lukk dørene og sett motorretning
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elev.Dirn)
		}

	default:
		// Ingen handling for andre tilstander
	}

	fmt.Println("\nNew state:")
	return elev
}
