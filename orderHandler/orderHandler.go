package orderHandler

import (
	"Driver-go/lights"
	. "Driver-go/types"
	"fmt"
)

func OrderHandler(ID string,
	Ch_localOrders chan<- OrderMatrix,
	Ch_orderEventToMaster chan<- OrderEvent,
	Ch_buttonPress <-chan ButtonEvent,
	Ch_clearedFloor <-chan DirnFloorPair,
	Ch_ordersFromMaster <-chan GlobalOrderMap) {

	ordersFromMaster := GlobalOrderMap{}
	for {
		select {
		case buttonEvent := <-Ch_buttonPress:
			button := []ButtonEvent{buttonEvent}
			fmt.Println("Button pressed: ", button)
			orderEvent := OrderEvent{ElevatorID: ID, Completed: false, Orders: button}
			Ch_orderEventToMaster <- orderEvent
			fmt.Println("Order sent to master: ")

		case ordersFromMaster = <-Ch_ordersFromMaster:
			fmt.Println("Orders from master: ", ordersFromMaster)
			lights.SetLights(ordersFromMaster, ID)
			Ch_localOrders <- ordersFromMaster[ID]
			// fmt.Println("Ordermatrix passed")


		case dirnFloor := <-Ch_clearedFloor:
			lights.SetLights(ordersFromMaster, ID)
			orders := []ButtonEvent{}
			if dirnFloor.Dirn == MD_Down {
				hallButton := ButtonEvent{Floor: dirnFloor.Floor, Button: ButtonType(BT_HallDown)}
				cabButton := ButtonEvent{Floor: dirnFloor.Floor, Button: ButtonType(BT_Cab)}
				orders = append(orders, hallButton)
				orders = append(orders, cabButton)
			} else if dirnFloor.Dirn == MD_Up {
				hallButton := ButtonEvent{Floor: dirnFloor.Floor, Button: ButtonType(BT_HallUp)}
				cabButton := ButtonEvent{Floor: dirnFloor.Floor, Button: ButtonType(BT_Cab)}
				orders = append(orders, hallButton)
				orders = append(orders, cabButton)
			} else {
				for btn := 0; btn < NUMBUTTONTYPE; btn++ {
					button := ButtonEvent{Floor: dirnFloor.Floor, Button: ButtonType(btn)}
					orders = append(orders, button)
				}
			}

			finishedOrder := OrderEvent{ElevatorID: ID, Completed: true, Orders: orders}
			Ch_orderEventToMaster <- finishedOrder

			// case <-Ch_orderCopyRequest:
			// 	orderCopy := NetworkMessage{
			// 		MsgType: "ordercopyresponse",
			// 		MsgData: ordersFromMaster,
			// 	}
			// 	Ch_networkToMaster <- orderCopy

		}
	}
}
