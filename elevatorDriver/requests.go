package elevatorDriver

import (
	. "Driver-go/types"
)

func requestsAbove(orderMatrix OrderMatrix, floor int) bool {
	for i := floor + 1; i < NUMFLOORS; i++ {
		for j := 0; j < NUMBUTTONTYPE; j++ {
			if orderMatrix[i][j] { 
				return true
			}
		}
	}
	return false
}

func requestsBelow(orderMatrix OrderMatrix, floor int) bool {
	for i := 0; i < floor; i++ {
		for j := 0; j < NUMBUTTONTYPE; j++ {
			if orderMatrix[i][j] { 
				return true
			}
		}
	}
	return false
}

func requestsHere(orderMatrix OrderMatrix, floor int) bool {
	for j := 0; j < NUMBUTTONTYPE; j++ {
		if orderMatrix[floor][j] { 
			return true
		}
	}
	return false
}

func requestsChooseDirection(orderMatrix OrderMatrix, elev Elevator, dirn MotorDirection) DirnBehaviourPair {
	switch dirn {
	case MD_Up:
		if requestsAbove(orderMatrix, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Up, Behaviour:EB_Moving}
		} else if requestsHere(orderMatrix, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Stop, Behaviour:EB_DoorOpen}
		} else if requestsBelow(orderMatrix, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Down, Behaviour:EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn:MD_Stop, Behaviour:EB_Idle}
		}
	case MD_Down:
		if requestsBelow(orderMatrix, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Down, Behaviour:EB_Moving}
		} else if requestsHere(orderMatrix, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Stop, Behaviour:EB_DoorOpen}
		} else if requestsAbove(orderMatrix, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Up, Behaviour:EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn:MD_Stop, Behaviour:EB_Idle}
		}
	case MD_Stop:
		if requestsHere(orderMatrix, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Stop, Behaviour:EB_DoorOpen}
		} else if requestsAbove(orderMatrix, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Up, Behaviour:EB_Moving}
		} else if requestsBelow(orderMatrix, elev.Floor) {
			return DirnBehaviourPair{Dirn:MD_Down, Behaviour:EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn:MD_Stop, Behaviour:EB_Idle}
		}
	default:
		return DirnBehaviourPair{Dirn:MD_Stop, Behaviour:EB_Idle}
	}
}

func requestsShouldStop(orderMatrix OrderMatrix, elev Elevator) bool {
	switch elev.Dirn {
	case MD_Down:
		return  orderMatrix[elev.Floor][BT_HallDown] ||
				orderMatrix[elev.Floor][BT_Cab]  ||
				!requestsBelow(orderMatrix, elev.Floor)
	case MD_Up:
		return  orderMatrix[elev.Floor][BT_HallUp]  ||
				orderMatrix[elev.Floor][BT_Cab]  ||
				!requestsAbove(orderMatrix, elev.Floor)
	case MD_Stop:
		return  orderMatrix[elev.Floor][BT_HallUp]  ||
				orderMatrix[elev.Floor][BT_HallDown]  ||
				orderMatrix[elev.Floor][BT_Cab] 
	default:
		return false
	}
}

func requestsClearAtCurrentFloor(orderMatrix OrderMatrix, elev Elevator, dirn MotorDirection) (OrderMatrix, MotorDirection){
		orderMatrix[elev.Floor][BT_Cab] = false

		switch dirn {
		case MD_Up:
			orderMatrix[elev.Floor][BT_HallUp] = false
			if !requestsAbove(orderMatrix, elev.Floor) {
				orderMatrix[elev.Floor][BT_HallDown] = false
				dirn = MD_Stop
			}

		case MD_Down:
			orderMatrix[elev.Floor][BT_HallDown] = false
			if !requestsBelow(orderMatrix, elev.Floor) {
				orderMatrix[elev.Floor][BT_HallUp] = false
				dirn = MD_Stop
			}
		case MD_Stop:
			orderMatrix[elev.Floor][BT_HallUp] = false
			orderMatrix[elev.Floor][BT_HallDown] = false
		}
	return orderMatrix, dirn
}

func emptyOrderMatrix(orderMatrix OrderMatrix) bool {
	for _, row := range orderMatrix {
		for _, value := range row {
			if value {
				return false
			}
		}
	}
	return true
}
