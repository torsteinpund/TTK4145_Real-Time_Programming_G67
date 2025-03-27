package elevatorDriver

import (
	. "Driver-go/types"
)

func ordersAbove(orderMatrix OrderMatrix, floor int) bool {
	for i := floor + 1; i < NUMFLOORS; i++ {
		for j := 0; j < NUMBUTTONTYPE; j++ {
			if orderMatrix[i][j] {
				return true
			}
		}
	}
	return false
}

func ordersBelow(orderMatrix OrderMatrix, floor int) bool {
	for i := 0; i < floor; i++ {
		for j := 0; j < NUMBUTTONTYPE; j++ {
			if orderMatrix[i][j] {
				return true
			}
		}
	}
	return false
}

func ordersHere(orderMatrix OrderMatrix, floor int) bool {
	for j := 0; j < NUMBUTTONTYPE; j++ {
		if orderMatrix[floor][j] {
			return true
		}
	}
	return false
}

func chooseDirection(orderMatrix OrderMatrix, floor int, dirn MotorDirection) DirnBehaviourPair {
	switch dirn {
	case MD_Up:
		if ordersAbove(orderMatrix, floor) {
			return DirnBehaviourPair{Dirn: MD_Up, Behaviour: EB_Moving}
		} else if ordersHere(orderMatrix, floor) {
			return DirnBehaviourPair{Dirn: MD_Stop, Behaviour: EB_DoorOpen}
		} else if ordersBelow(orderMatrix, floor) {
			return DirnBehaviourPair{Dirn: MD_Down, Behaviour: EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn: MD_Stop, Behaviour: EB_Idle}
		}
	case MD_Down:
		if ordersBelow(orderMatrix, floor) {
			return DirnBehaviourPair{Dirn: MD_Down, Behaviour: EB_Moving}
		} else if ordersHere(orderMatrix, floor) {
			return DirnBehaviourPair{Dirn: MD_Stop, Behaviour: EB_DoorOpen}
		} else if ordersAbove(orderMatrix, floor) {
			return DirnBehaviourPair{Dirn: MD_Up, Behaviour: EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn: MD_Stop, Behaviour: EB_Idle}
		}
	case MD_Stop:
		if ordersHere(orderMatrix, floor) {
			return DirnBehaviourPair{Dirn: MD_Stop, Behaviour: EB_DoorOpen}
		} else if ordersAbove(orderMatrix, floor) {
			return DirnBehaviourPair{Dirn: MD_Up, Behaviour: EB_Moving}
		} else if ordersBelow(orderMatrix, floor) {
			return DirnBehaviourPair{Dirn: MD_Down, Behaviour: EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn: MD_Stop, Behaviour: EB_Idle}
		}
	default:
		return DirnBehaviourPair{Dirn: MD_Stop, Behaviour: EB_Idle}
	}
}

func shouldStop(orderMatrix OrderMatrix, elev Elevator) bool {
	switch elev.Dirn {
	case MD_Down:
		return orderMatrix[elev.Floor][BT_HallDown] ||
			orderMatrix[elev.Floor][BT_Cab] ||
			!ordersBelow(orderMatrix, elev.Floor)
	case MD_Up:
		return orderMatrix[elev.Floor][BT_HallUp] ||
			orderMatrix[elev.Floor][BT_Cab] ||
			!ordersAbove(orderMatrix, elev.Floor)
	case MD_Stop:
		return orderMatrix[elev.Floor][BT_HallUp] ||
			orderMatrix[elev.Floor][BT_HallDown] ||
			orderMatrix[elev.Floor][BT_Cab]
	default:
		return false
	}
}

func clearOrderAtCurrentFloor(orderMatrix OrderMatrix, floor int, dirn MotorDirection) (OrderMatrix, MotorDirection) {
	orderMatrix[floor][BT_Cab] = false

	switch dirn {
	case MD_Up:
		orderMatrix[floor][BT_HallUp] = false
		if !ordersAbove(orderMatrix, floor) {
			orderMatrix[floor][BT_HallDown] = false
			dirn = MD_Stop
		}

	case MD_Down:
		orderMatrix[floor][BT_HallDown] = false
		if !ordersBelow(orderMatrix, floor) {
			orderMatrix[floor][BT_HallUp] = false
			dirn = MD_Stop
		}
	case MD_Stop:
		orderMatrix[floor][BT_HallUp] = false
		orderMatrix[floor][BT_HallDown] = false
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

func initAfterErrorTimeout(dirn MotorDirection, ch_localorders <-chan LocalOrder, orderMatrix OrderMatrix, elevBehaviour ElevatorBehaviour)(int, ElevatorBehaviour){
	setMotorDirection(dirn)
	var floor int
	timeOutLoop:
	for{
		select{
		case <-ch_localorders: 
		//Ensures draining of orderHandler
		
		default:
			if getFloor()!=-1 {
				floor = getFloor()
				if emptyOrderMatrix(orderMatrix){
					setMotorDirection(MD_Stop)
					elevBehaviour = EB_Idle
				}
				break timeOutLoop
			}
		}
	}
	return floor, elevBehaviour
}

