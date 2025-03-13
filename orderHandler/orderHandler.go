package orderHandler

import (
	// "Driver-go/lights"
	. "Driver-go/types"
	"fmt"
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

type OrderChannels struct {
	LocalOrderChannel       chan OrderMatrix
	LocalLightsChannel      chan OrderMatrix
	OrdersFromMasterChannel chan GlobalOrderMap
	OrdersToMasterChannel   chan NetworkMessage
	ButtonEventChannel      chan ButtonEvent
	FinishedFloorChannel    chan int
	Ch_registerOrder		chan OrderEvent
	Ch_toSlave				chan NetworkMessage
	Ch_toSlaveTest			chan GlobalOrderMap
}

func OrderHandler(ch OrderChannels, ID string) {
	ordersFromMaster := make(GlobalOrderMap)

	for {
		select {
		case buttonEvent := <-ch.ButtonEventChannel:
			button := []ButtonEvent{buttonEvent}
			orderEvent := OrderEvent{ElevatorID: ID, Completed: false, Orders: button}
			// newOrderEvent := NetworkMessage{MsgType: "New OrderEvent", MsgData: orderEvent, Receipient: Master}
			// ch.OrdersToMasterChannel <- newOrderEvent
			ch.Ch_registerOrder <- orderEvent
			fmt.Println("OrderEvent sent to master from orderhandler after buttonEvent")

		case ordersFromMaster = <-ch.Ch_toSlaveTest:
			fmt.Println("OrdersFromMaster received in orderHandler")
			localRequests := ordersFromMaster[ID]
			//ch.LocalOrderChannel <- localRequests
			println("wE MADE IT")
			localLights := localRequests
			
			for _, requests := range ordersFromMaster {
				fmt.Println("OrderHandler: ", requests)
				// localLights = lights.SetCabLights(requests)
			}
			println("wE MADE IT 2")
			ch.LocalLightsChannel <- localLights

		case floor := <-ch.FinishedFloorChannel:
			orders := []ButtonEvent{}
			for btn := 0; btn < NUMBUTTONTYPE; btn++ {
				button := ButtonEvent{Floor: floor, Button: ButtonType(btn)}
				orders = append(orders, button)

			}

			finishedOrder := OrderEvent{ElevatorID: ID, Completed: true, Orders: orders}
			regFinishedOrder := NetworkMessage{MsgType: "Finished OrderEvent", MsgData: finishedOrder, Receipient: Master}
			ch.OrdersToMasterChannel <- regFinishedOrder

		}
	}
}
