package orderHandler

import (
	// "Driver-go/lights"
	"Driver-go/lights"
	. "Driver-go/types"
	"fmt"
	// "time"
)

type OrderChannels struct {
	Ch_localOrders       chan OrderMatrix
	Ch_localLights       chan OrderMatrix
	Ch_orderFromMaster   chan GlobalOrderMap
	Ch_networkToMaster   chan NetworkMessage
	Ch_buttonPress       chan ButtonEvent
	Ch_clearedFloor      chan ClearedFloorInfo
	Ch_registerOrder     chan OrderEvent
	Ch_networkToSlave    chan NetworkMessage
	Ch_orderCopyResponse chan GlobalOrderMap
	Ch_orderCopyRequest  chan bool
}



func OrderHandler(ch OrderChannels, ID string) {

	ordersFromMaster := GlobalOrderMap{}
	for {
		select {
		case buttonEvent := <-ch.Ch_buttonPress:
			button := []ButtonEvent{buttonEvent}
			fmt.Println("Button pressed: ", button)
			completed := [NUMBUTTONTYPE]bool{false,false,false}
			orderEvent := OrderEvent{ElevatorID: ID, Completed: completed, Orders: button}
			// newOrderEvent := NetworkMessage{MsgType: "registerorderchannel", MsgData: orderEvent}
			ch.Ch_registerOrder<- orderEvent
			fmt.Println("Order sent to master: ")

		case fromMaster := <-ch.Ch_orderFromMaster:
			// ordersFromMaster = fromMaster.MsgData.(GlobalOrderMap)
			fmt.Println("Orders from master: ", ordersFromMaster, "ID: ", ID)
			lights.SetLights(fromMaster, ID)
			ch.Ch_localOrders <- fromMaster[ID]
			// fmt.Println("Ordermatrix passed", ordersFromMaster[ID])

		case clearedFloorInfo := <-ch.Ch_clearedFloor:
			orders := []ButtonEvent{}
			fmt.Println("Last known direction", clearedFloorInfo.LastKnownDirection)

			completed := [NUMBUTTONTYPE]bool{}

			for btn := 0; btn < NUMBUTTONTYPE; btn++ {
				button := ButtonEvent{Floor: clearedFloorInfo.Floor, Button: ButtonType(btn)}
				if clearedFloorInfo.LastKnownDirection == MD_Down && ButtonType(btn) == BT_HallUp {
					completed[btn] = false
					orders = append(orders, button)
					
				}
				if clearedFloorInfo.LastKnownDirection == MD_Up && ButtonType(btn) == BT_HallDown{
					completed[btn] = false
					orders = append(orders, button)
				}else{
					orders = append(orders, button)
					completed[btn] = true
				}
				// button := ButtonEvent{Floor: clearedFloorInfo.Floor, Button: ButtonType(btn)}


			}
			fmt.Println("Orders from cleared floor", orders)
			finishedOrder := OrderEvent{ElevatorID: ID, Completed: completed, Orders: orders}
			regFinishedOrder := NetworkMessage{MsgType: "registerorderchannel", MsgData: finishedOrder,}
			fmt.Println("Finished cleared floor: ", regFinishedOrder)
			ch.Ch_networkToMaster <- regFinishedOrder

			// case <-ch.Ch_orderCopyRequest:
			// 	orderCopy := NetworkMessage{
			// 		MsgType: "ordercopyresponse",
			// 		MsgData: ordersFromMaster,
			// 	}
			// 	ch.Ch_networkToMaster <- orderCopy

		}
	}
}
