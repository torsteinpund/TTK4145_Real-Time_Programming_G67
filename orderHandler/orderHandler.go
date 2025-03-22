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
	Ch_clearedFloor      chan int
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
			orderEvent := OrderEvent{ElevatorID: ID, Completed: false, Orders: button}
			newOrderEvent := NetworkMessage{MsgType: "registerorderchannel", MsgData: orderEvent}
			ch.Ch_networkToMaster <- newOrderEvent
			fmt.Println("Order sent to master: ")

		case fromMaster := <-ch.Ch_networkToSlave:
			ordersFromMaster = fromMaster.MsgData.(GlobalOrderMap)
			fmt.Println("Orders from master: ", ordersFromMaster)
			lights.SetLights(ordersFromMaster, ID)
			ch.Ch_localOrders <- ordersFromMaster[ID]
			// fmt.Println("Ordermatrix passed", ordersFromMaster[ID])

		case floor := <-ch.Ch_clearedFloor:
			orders := []ButtonEvent{}
			for btn := 0; btn < NUMBUTTONTYPE; btn++ {
				button := ButtonEvent{Floor: floor, Button: ButtonType(btn)}
				orders = append(orders, button)

			}
			finishedOrder := OrderEvent{ElevatorID: ID, Completed: true, Orders: orders}
			regFinishedOrder := NetworkMessage{MsgType: "registerorderchannel", MsgData: finishedOrder}
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
