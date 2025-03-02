package orderHandler

import (
	"Driver-go/lights"
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
	FinishedFloorChannel 	chan int
}



func orderHandler(ch orderChannels, ID string) {
	ordersFromMaster := make(GlobalOrderMap)

	for {
		select {
		case buttonEvent := <-ch.ButtenEventChannel:
			button := []ButtonEvent{buttonEvent}
			orderEvent := OrderEvent{ElevatorID: ID, Completed: false, Orders: button}
			newOrderEvent := NetworkMessage{MsgType: "New OrderEvent", MsgData: orderEvent, Role: "Master"}
			ch.OrdersToMasterChannel <- newOrderEvent

		case ordersFromMaster = <-ch.OrdersFromMasterChannel:
			localRequests := ordersFromMaster[ID]
			ch.LocalOrderChannel <- localRequests
			localLights := localRequests

			for _, requests := range ordersFromMaster {
				localLights = lights.SetHallLights(requests)

			}

			ch.LocalLightsChannel <- localLights

		case floor := <- ch.FinishedFloorChannel:
			orders := []ButtonEvent{}
			for btn := 0; btn < NUMBUTTONTYPE; btn++{
				button := ButtonEvent{Floor: floor, Button: ButtonType(btn)}
				orders = append(orders, button)

			}

			finishedOrder := OrderEvent{ElevatorID: ID, Completed: true, Orders: orders}
			regFinishedOrder := NetworkMessage{MsgType: "Finished OrderEvent", MsgData: finishedOrder, Role: "Master"}
			ch.OrdersToMasterChannel <- regFinishedOrder
			
		}
	}
}
