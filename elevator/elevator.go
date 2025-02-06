package elevator

import (
	"Driver-go/elevio"
	"fmt"
)


type ElevatorBehaviour int

const (
	NUMFLOORS  = 4 // Juster antall etasjer etter behov
	NUMBUTTONTYPE = 3 // Juster antall knappetyper etter behov
)

const (
	EB_Idle elevio.Behaviour = iota
	EB_DoorOpen
	EB_Moving
)

type ClearRequestVariant int

const (
	CV_All ClearRequestVariant = iota
	CV_InDirn
)


type Elevator struct {
	Floor     int                     // Nåværende etasje
	Dirn      elevio.MotorDirection               // Heisens retning
	Requests  [NUMFLOORS][NUMBUTTONTYPE]int // Forespørsler (to-dimensjonalt array)
	Behaviour ElevatorBehaviour       // Heisens nåværende oppførsel

	Config struct { // Konfigurasjon for heisen
		ClearRequestVariant ClearRequestVariant
		DoorOpenDurationS   float64
	}
}

func ElevatorUninitialized() Elevator {
	return Elevator{
		Floor:     -1,
		Dirn:      elevio.MD_Stop,
		Behaviour: ElevatorBehaviour(EB_Idle),
		Config: struct {
			ClearRequestVariant ClearRequestVariant
			DoorOpenDurationS   float64
		}{
			ClearRequestVariant: CV_All,
			DoorOpenDurationS:   3.0,
		},
		Requests: [NUMFLOORS][NUMBUTTONTYPE]int{}, // Initialiser forespørslene til 0
	}
}


func InitElevator(numFloors int, numButtonTypes int, elev Elevator) Elevator {
	if numFloors > NUMFLOORS || numButtonTypes > NUMBUTTONTYPE {
		fmt.Println("Error: Configuration exceeds allowed array size.")
		return Elevator{}
	}

	elev = Elevator{
		Floor:    -1,                      // Start utenfor en definert etasje
		Dirn:     elevio.MD_Stop,          // Heisen starter som stoppet
		Behaviour: ElevatorBehaviour(EB_Idle),                // Heisen starter i "Idle"-tilstand
		Config: struct {                   // Konfigurasjon
			ClearRequestVariant ClearRequestVariant
			DoorOpenDurationS   float64
		}{
			ClearRequestVariant: CV_All,  // Standard klareringsvariant
			DoorOpenDurationS:   3.0,    // Standard tid dørene forblir åpne
		},
	}

	// Initialiser forespørselsmatrisen med nuller
	for i := 0; i < NUMFLOORS; i++ {
		for j := 0; j < NUMBUTTONTYPE; j++ {
			elev.Requests[i][j] = 0
		}
	}

	fmt.Println("Elevator initialized:")
	fmt.Printf("%+v\n", elev)
	return elev
}


func CheckValidElevatorStates(elev Elevator) bool{
	if elev.Behaviour>2 || elev.Behaviour<0{
		return false
	}
	return true
}