package requests

import (
	. "Driver-go/types"
)



func RequestsAbove(req [NUMFLOORS][NUMBUTTONTYPE]bool, floor int) bool {
	for i := floor + 1; i < NUMFLOORS; i++ {
		for j := 0; j < NUMBUTTONTYPE; j++ {
			if req[i][j] { 
				return true
			}
		}
	}
	return false
}

func RequestsBelow(req [NUMFLOORS][NUMBUTTONTYPE]bool, floor int) bool {
	for i := 0; i < floor; i++ {
		for j := 0; j < NUMBUTTONTYPE; j++ {
			if req[i][j] { 
				return true
			}
		}
	}
	return false
}


func RequestsHere(req [NUMFLOORS][NUMBUTTONTYPE]bool, floor int) bool {
	for j := 0; j < NUMBUTTONTYPE; j++ {
		if req[floor][j] { 
			return true
		}
	}
	return false
}


func RequestsChooseDirection(elev Elevator) DirnBehaviourPair {
	switch elev.Dirn {
	case MD_Up:
		if RequestsAbove(elev.Requests, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Up, Behaviour:EB_Moving}
		} else if RequestsHere(elev.Requests, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Down, Behaviour:EB_DoorOpen}
		} else if RequestsBelow(elev.Requests, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Down, Behaviour:EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn:MD_Stop, Behaviour:EB_Idle}
		}
	case MD_Down:
		if RequestsBelow(elev.Requests, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Down, Behaviour:EB_Moving}
		} else if RequestsHere(elev.Requests, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Up, Behaviour:EB_DoorOpen}
		} else if RequestsAbove(elev.Requests, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Up, Behaviour:EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn:MD_Stop, Behaviour:EB_Idle}
		}
	case MD_Stop:
		if RequestsHere(elev.Requests, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Stop, Behaviour:EB_DoorOpen}
		} else if RequestsAbove(elev.Requests, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Up, Behaviour:EB_Moving}
		} else if RequestsBelow(elev.Requests, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Down, Behaviour:EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn:MD_Stop, Behaviour:EB_Idle}
		}
	default:
		return DirnBehaviourPair{Dirn:MD_Stop, Behaviour:EB_Idle}
	}
}

func RequestsShouldStop(elev Elevator) bool {
	switch elev.Dirn {
	case MD_Down:
		return elev.Requests[elev.Floor][BT_HallDown] ||
			elev.Requests[elev.Floor][BT_Cab]  ||
			!RequestsBelow(elev.Requests, elev.Floor)
	case MD_Up:
		return elev.Requests[elev.Floor][BT_HallUp]  ||
			elev.Requests[elev.Floor][BT_Cab]  ||
			!RequestsAbove(elev.Requests, elev.Floor)
	case MD_Stop:
		fallthrough 
	default:
		return true
	}
}

func RequestsShouldClearImmediately(elev Elevator, btnFloor int, btnType ButtonType) bool {
	switch elev.Config.ClearRequestVariant {
	case CV_All:
		return elev.Floor == btnFloor
	case CV_InDirn:
		return elev.Floor == btnFloor &&
			(
				(elev.Dirn == MD_Up && btnType == BT_HallUp) ||
				(elev.Dirn == MD_Down && btnType == BT_HallDown) ||
				elev.Dirn == MD_Stop ||
				btnType == BT_Cab)
	default:
		return false
	}
}

func RequestsClearAtCurrentFloor(elev Elevator, onClearedRequest func(ButtonType, int)) Elevator {
	switch elev.Config.ClearRequestVariant {
	case CV_All:
		for btn := 0; btn < NUMBUTTONTYPE; btn++ {
			if elev.Requests[elev.Floor][btn]  {
				elev.Requests[elev.Floor][btn] = false
				if onClearedRequest != nil {
					onClearedRequest(ButtonType(btn), elev.Floor)
				}
			}
		}


	case CV_InDirn:
		elev.Requests[elev.Floor][BT_Cab] = false

		switch elev.Dirn {
		case MD_Up:
			if !RequestsAbove(elev.Requests, elev.Floor) && !elev.Requests[elev.Floor][BT_Cab]{
				elev.Requests[elev.Floor][BT_Cab] = false
			}
			elev.Requests[elev.Floor][BT_Cab] = false

		case MD_Down:
			if !RequestsBelow(elev.Requests, elev.Floor) && !elev.Requests[elev.Floor][BT_Cab]{
				elev.Requests[elev.Floor][BT_Cab] = false
			}
			elev.Requests[elev.Floor][BT_Cab] = false

		case MD_Stop:
			fallthrough
		default:
			elev.Requests[elev.Floor][BT_HallUp] = false
			elev.Requests[elev.Floor][BT_HallDown] = false
		}
	}
	return elev 
}
