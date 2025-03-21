package orderHandler

import (
	// "Driver-go/lights"
	. "Driver-go/types"
	"fmt"
	"time"
)

type OrderChannels struct {
	Ch_localOrders          chan OrderMatrix
	Ch_localLights          chan OrderMatrix
	Ch_orderFromMaster 		chan GlobalOrderMap
	Ch_toMaster   			chan NetworkMessage
	Ch_buttonPress      	chan ButtonEvent
	Ch_clearedFloor    		chan int
	Ch_registerOrder        chan OrderEvent
	Ch_toSlave              chan NetworkMessage
	Ch_orderCopyResponse	chan GlobalOrderMap
}

func OrderHandler(ch OrderChannels, ID string) {
	orderCopyTime := 2 * time.Second
	ordersFromMaster := GlobalOrderMap{}
	for {
		select {
		case buttonEvent := <-ch.Ch_buttonPress:
			button := []ButtonEvent{buttonEvent}
			orderEvent := OrderEvent{ElevatorID: ID, Completed: false, Orders: button}
			newOrderEvent := NetworkMessage{MsgType: "orderupdatechannel", MsgData: orderEvent}
			ch.Ch_toMaster <- newOrderEvent
			// ch.Ch_registerOrder <- orderEvent

		case fromMaster := <-ch.Ch_toSlave:
			
			ordersFromMaster = fromMaster.MsgData.(GlobalOrderMap)
			ch.Ch_localOrders <- ordersFromMaster[ID]
			fmt.Println("Ordermatrix passed", ordersFromMaster[ID])

		case floor := <-ch.Ch_clearedFloor:
			orders := []ButtonEvent{}
			for btn := 0; btn < NUMBUTTONTYPE; btn++ {
				button := ButtonEvent{Floor: floor, Button: ButtonType(btn)}
				orders = append(orders, button)

			}

			finishedOrder := OrderEvent{ElevatorID: ID, Completed: true, Orders: orders}
			regFinishedOrder := NetworkMessage{MsgType: "orderupdatechannel", MsgData: finishedOrder}
			fmt.Println("Finished cleared floor: ", regFinishedOrder)
			ch.Ch_registerOrder <- finishedOrder

		
		case <-time.After(orderCopyTime):
			orderCopy := NetworkMessage{
				MsgType: "ordercopyresponse",
				MsgData: ordersFromMaster,
			}
			ch.Ch_toMaster <- orderCopy

		}
	}
}
