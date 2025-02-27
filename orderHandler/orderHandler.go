package orderHandler

import (
	"Driver-go/lights"
	"Driver-go/requests"
	. "Driver-go/types"
	// "net"
)

// func SetAllLocalLights(req [NUMFLOORS][NUMBUTTONTYPE]int) [NUMFLOORS][NUMBUTTONTYPE]int {
// 	for floor := 0; floor < NUMFLOORS; floor++ {
// 		for btn := 0; btn < NUMBUTTONTYPE; btn++ {
// 			state := req[floor][btn]
// 			elevio.SetButtonLamp(ButtonType(btn), floor, state == 1)
// 		}
// 	}
// 	return req
// }

type orderChannels struct {
	LocalOrderChannel       chan RequestsMatrix
	LocalLightsChannel      chan RequestsMatrix
	OrdersFromMasterChannel chan GlobalOrderMap
	OrdersToMasterChannel   chan NetworkMessage
	ButtenEventChannel      chan ButtonEvent
}

func timeToServeRequest(e_old Elevator, receivedCh <-chan ButtonEvent) float64 {
	e := e_old
	buttenEvent := <-receivedCh
	b := buttenEvent.Button
	f := buttenEvent.Floor

	e.Requests[f][b] = true
	arrivedAtRequest := false

	ifEqual := func(inner_b ButtonType, inner_f int) {
		if inner_b == b && inner_f == f {
			arrivedAtRequest = true
		}
	}

	duration := 0.0

	switch e.Behaviour {
	case ElevatorBehaviour(EB_Idle):
		e.Dirn = requests.RequestsChooseDirection(e).Dirn
		if e.Dirn == MD_Stop {
			return duration
		}
	case ElevatorBehaviour(EB_Moving):
		duration += e.Config.TimeBetweenFloors / 2
		e.Floor += int(e.Dirn)
	case ElevatorBehaviour(EB_DoorOpen):
		duration -= e.Config.DoorOpenDuration / 2
	}

	for {
		if requests.RequestsShouldStop(e) {
			e = requests.RequestsClearAtCurrentFloor(e, ifEqual)
			if arrivedAtRequest {
				return duration
			}
			duration += e.Config.DoorOpenDuration
			e.Dirn = requests.RequestsChooseDirection(e).Dirn
		}
		e.Floor += int(e.Dirn)
		duration += e.Config.TimeBetweenFloors
	}
}

func orderHandler(elev Elevator, ch orderChannels, ID string) {
	ordersFromMaster := make(GlobalOrderMap)

	for {
		select {
		case buttonEvent := <-ch.ButtenEventChannel:
			button := []ButtonEvent{buttonEvent}
			orderEvent := OrderEvent{ElevatorID: ID, Completed: false, Orders: button}
			newOrderEvent := NetworkMessage{MsgType: "OrderEvent", MsgData: orderEvent, Role: "Master"}
			ch.OrdersToMasterChannel <- newOrderEvent

		case ordersFromMaster = <-ch.OrdersFromMasterChannel:
			localRequests := ordersFromMaster[ID]
			ch.LocalOrderChannel <- localRequests
			localLights := localRequests

			for _, requests := range ordersFromMaster {
				localLights = lights.SetHallLights(requests)

			}

			ch.LocalLightsChannel <- localLights
		}
	}
}
