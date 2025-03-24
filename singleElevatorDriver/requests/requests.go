package requests

import (
	. "Driver-go/types"
	// "fmt"
)



func RequestsAbove(orderMatrix OrderMatrix, floor int) bool {
	for i := floor + 1; i < NUMFLOORS; i++ {
		for j := 0; j < NUMBUTTONTYPE; j++ {
			if orderMatrix[i][j] { 
				return true
			}
		}
	}
	return false
}

func RequestsBelow(orderMatrix OrderMatrix, floor int) bool {
	for i := 0; i < floor; i++ {
		for j := 0; j < NUMBUTTONTYPE; j++ {
			if orderMatrix[i][j] { 
				return true
			}
		}
	}
	return false
}


func RequestsHere(orderMatrix OrderMatrix, floor int) bool {
	for j := 0; j < NUMBUTTONTYPE; j++ {
		if orderMatrix[floor][j] { 
			return true
		}
	}
	return false
}


func RequestsChooseDirection(orderMatrix OrderMatrix, elev Elevator, dirn MotorDirection) DirnBehaviourPair {
	switch dirn {
	case MD_Up:
		if RequestsAbove(orderMatrix, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Up, Behaviour:EB_Moving}
		} else if RequestsHere(orderMatrix, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Stop, Behaviour:EB_DoorOpen}
		} else if RequestsBelow(orderMatrix, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Down, Behaviour:EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn:MD_Stop, Behaviour:EB_Idle}
		}
	case MD_Down:
		if RequestsBelow(orderMatrix, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Down, Behaviour:EB_Moving}
		} else if RequestsHere(orderMatrix, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Stop, Behaviour:EB_DoorOpen}
		} else if RequestsAbove(orderMatrix, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Up, Behaviour:EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn:MD_Stop, Behaviour:EB_Idle}
		}
	case MD_Stop:
		if RequestsHere(orderMatrix, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Stop, Behaviour:EB_DoorOpen}
		} else if RequestsAbove(orderMatrix, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Up, Behaviour:EB_Moving}
		} else if RequestsBelow(orderMatrix, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Down, Behaviour:EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn:MD_Stop, Behaviour:EB_Idle}
		}
	default:
		return DirnBehaviourPair{Dirn:MD_Stop, Behaviour:EB_Idle}
	}
}

func RequestsShouldStop(orderMatrix OrderMatrix, elev Elevator) bool {
	switch elev.Dirn {
	case MD_Down:
		return  orderMatrix[elev.Floor][BT_HallDown] ||
				orderMatrix[elev.Floor][BT_Cab]  ||
				!RequestsBelow(orderMatrix, elev.Floor)
	case MD_Up:
		return  orderMatrix[elev.Floor][BT_HallUp]  ||
				orderMatrix[elev.Floor][BT_Cab]  ||
				!RequestsAbove(orderMatrix, elev.Floor)
	case MD_Stop:
		return  orderMatrix[elev.Floor][BT_HallUp]  ||
				orderMatrix[elev.Floor][BT_HallDown]  ||
				orderMatrix[elev.Floor][BT_Cab] 
	default:
		return false
	}
}

func RequestsShouldClearImmediately(orderMatrix OrderMatrix, elev Elevator, btnFloor int, btnType ButtonType) bool {
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

func RequestsClearAtCurrentFloor(orderMatrix OrderMatrix, elev Elevator, dirn MotorDirection) (OrderMatrix, MotorDirection) {
	switch elev.Config.ClearRequestVariant {
	case CV_All:
		for btn := 0; btn < NUMBUTTONTYPE; btn++ {
			if orderMatrix[elev.Floor][btn]  {
				orderMatrix[elev.Floor][btn] = false
			}
		}


	case CV_InDirn:
		orderMatrix[elev.Floor][BT_Cab] = false

		switch dirn {
		case MD_Up:
			orderMatrix[elev.Floor][BT_HallUp] = false
			if !RequestsAbove(orderMatrix, elev.Floor) {
				orderMatrix[elev.Floor][BT_HallDown] = false
				dirn = MD_Stop
			}

		case MD_Down:
			orderMatrix[elev.Floor][BT_HallDown] = false
			if !RequestsBelow(orderMatrix, elev.Floor) {
				orderMatrix[elev.Floor][BT_HallUp] = false
				dirn = MD_Stop
			}
		case MD_Stop:
			orderMatrix[elev.Floor][BT_HallUp] = false
			orderMatrix[elev.Floor][BT_HallDown] = false
		}
	}
	return orderMatrix, dirn
}
