package requests

import (
	"Driver-go/elevio"
	"Driver-go/elevator"
)

//All channels have underscore in names, and variables have capital letters to devide words

type DirnBehaviourPair struct {
	Dirn      elevio.MotorDirection
	Behaviour elevio.Behaviour
}

type ClearRequestCallback func(button elevio.ButtonType, floor int)


func RequestsAbove(req [elevator.NUMFLOORS][elevator.NUMBUTTONTYPE]int, floor int) bool {
	// Iterer gjennom etasjene over nåværende etasje
	for i := floor + 1; i < elevio.NumFloors; i++ {
		// Sjekk alle knappeforespørsler for hver etasje
		for j := 0; j < elevio.NumButtonTypes; j++ {
			if req[i][j] == 1 { // Hvis det finnes en aktiv forespørsel
				return true
			}
		}
	}
	return false
}

func RequestsBelow(req [elevator.NUMFLOORS][elevator.NUMBUTTONTYPE]int, floor int) bool {
	// Iterer gjennom etasjene under nåværende etasje
	for i := 0; i < floor; i++ {
		// Sjekk alle knappeforespørsler for hver etasje
		for j := 0; j < elevio.NumButtonTypes; j++ {
			if req[i][j] == 1 { // Hvis det finnes en aktiv forespørsel
				return true
			}
		}
	}
	return false
}

func RequestsHere(req [elevator.NUMFLOORS][elevator.NUMBUTTONTYPE]int, floor int) bool {
	// Sjekk alle knappeforespørsler for den nåværende etasjen
	for j := 0; j < elevio.NumButtonTypes; j++ {
		if req[floor][j] == 1 { // Hvis det finnes en aktiv forespørsel
			return true
		}
	}
	return false
}


func RequestsChooseDirection(elev elevator.Elevator) DirnBehaviourPair {
	switch elev.Dirn {
	case elevio.MD_Up:
		if RequestsAbove(elev.Requests, elev.Floor) {
			return DirnBehaviourPair{elevio.MD_Up, elevator.EB_Moving}
		} else if RequestsHere(elev.Requests, elev.Floor) {
			return DirnBehaviourPair{elevio.MD_Down, elevator.EB_DoorOpen}
		} else if RequestsBelow(elev.Requests, elev.Floor) {
			return DirnBehaviourPair{elevio.MD_Down, elevator.EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
		}
	case elevio.MD_Down:
		if RequestsBelow(elev.Requests, elev.Floor) {
			return DirnBehaviourPair{elevio.MD_Down, elevator.EB_Moving}
		} else if RequestsHere(elev.Requests, elev.Floor) {
			return DirnBehaviourPair{elevio.MD_Up, elevator.EB_DoorOpen}
		} else if RequestsAbove(elev.Requests, elev.Floor) {
			return DirnBehaviourPair{elevio.MD_Up, elevator.EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
		}
	case elevio.MD_Stop:
		if RequestsHere(elev.Requests, elev.Floor) {
			return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_DoorOpen}
		} else if RequestsAbove(elev.Requests, elev.Floor) {
			return DirnBehaviourPair{elevio.MD_Up, elevator.EB_Moving}
		} else if RequestsBelow(elev.Requests, elev.Floor) {
			return DirnBehaviourPair{elevio.MD_Down, elevator.EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
		}
	default:
		return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
	}
}

func RequestsShouldStop(elev elevator.Elevator) bool {
	switch elev.Dirn {
	case elevio.MD_Down:
		return elev.Requests[elev.Floor][elevio.BT_HallDown] == 1 ||
			elev.Requests[elev.Floor][elevio.BT_Cab] == 1 ||
			!RequestsBelow(elev.Requests, elev.Floor)
	case elevio.MD_Up:
		return elev.Requests[elev.Floor][elevio.BT_HallUp] == 1 ||
			elev.Requests[elev.Floor][elevio.BT_Cab] == 1 ||
			!RequestsAbove(elev.Requests, elev.Floor)
	case elevio.MD_Stop:
		fallthrough // Gå til default
	default:
		return true
	}
}

func RequestsShouldClearImmediately(elev elevator.Elevator, btnFloor int, btnType elevio.ButtonType) bool {
	switch elev.Config.ClearRequestVariant {
	case elevator.CV_All:
		// Fjern forespørselen hvis heisen er i samme etasje
		return elev.Floor == btnFloor
	case elevator.CV_InDirn:
		// Fjern forespørselen hvis heisen er i samme etasje og:
		return elev.Floor == btnFloor &&
			(
				(elev.Dirn == elevio.MD_Up && btnType == elevio.BT_HallUp) ||
				(elev.Dirn == elevio.MD_Down && btnType == elevio.BT_HallDown) ||
				elev.Dirn == elevio.MD_Stop ||
				btnType == elevio.BT_Cab)
	default:
		// Ikke fjern forespørselen
		return false
	}
}

func RequestsClearAtCurrentFloor(elev elevator.Elevator, onClearedRequest func(elevio.ButtonType, int)) elevator.Elevator {
	switch elev.Config.ClearRequestVariant {
	case elevator.CV_All:
		// Fjern alle forespørsler i den nåværende etasjen
		for btn := 0; btn < elevio.NumButtonTypes; btn++ {
			if elev.Requests[elev.Floor][btn] == 1 {
				elev.Requests[elev.Floor][btn] = 0
				if onClearedRequest != nil {
					onClearedRequest(elevio.ButtonType(btn), elev.Floor)
				}
			}
		}


	case elevator.CV_InDirn:
		// Fjern forespørsler fra innsiden av heisen
		elev.Requests[elev.Floor][elevio.BT_Cab] = 0

		switch elev.Dirn {
		case elevio.MD_Up:
			if !RequestsAbove(elev.Requests, elev.Floor) && elev.Requests[elev.Floor][elevio.BT_Cab] == 0 {
				elev.Requests[elev.Floor][elevio.BT_Cab] = 0
			}
			elev.Requests[elev.Floor][elevio.BT_Cab] = 0

		case elevio.MD_Down:
			if !RequestsBelow(elev.Requests, elev.Floor) && elev.Requests[elev.Floor][elevio.BT_Cab] == 0 {
				elev.Requests[elev.Floor][elevio.BT_Cab] = 0
			}
			elev.Requests[elev.Floor][elevio.BT_Cab] = 0

		case elevio.MD_Stop:
			fallthrough
		default:
			elev.Requests[elev.Floor][elevio.BT_HallUp] = 0
			elev.Requests[elev.Floor][elevio.BT_HallDown] = 0
		}
	}
	return elev 
}
