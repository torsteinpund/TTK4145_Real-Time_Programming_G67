package orderHandler

import (
	"Driver-go/lights"
	. "Driver-go/types"
	"fmt"
	"Driver-go/backup"
)

func OrderHandler(ID string,
	Ch_localOrders chan<- OrderMatrix,
	Ch_orderEventToMaster chan<- OrderEvent,
	Ch_orderCopyResponse chan<- GlobalOrderMap,
	Ch_buttonPress <-chan ButtonEvent,
	Ch_clearedFloor <-chan DirnFloorPair,
	Ch_ordersFromMaster <-chan GlobalOrderMap,
	Ch_orderCopyRequest <-chan bool) {

	ordersFromMaster := GlobalOrderMap{}
	for {
		select {
		case buttonEvent := <-Ch_buttonPress:
			button := []ButtonEvent{buttonEvent}
			orderEvent := OrderEvent{ElevatorID: ID, Completed: false, Orders: button}
			Ch_orderEventToMaster <- orderEvent


		case ordersFromMaster = <-Ch_ordersFromMaster:
			fmt.Println("Received orders from master", ordersFromMaster)
			lights.SetLights(ordersFromMaster, ID)
			Ch_localOrders <- ordersFromMaster[ID]


		case dirnFloor := <-Ch_clearedFloor:
			orders := []ButtonEvent{}
			if dirnFloor.Dirn == MD_Down {
				hallButton := ButtonEvent{Floor: dirnFloor.Floor, Button: ButtonType(BT_HallDown)}
				cabButton := ButtonEvent{Floor: dirnFloor.Floor, Button: ButtonType(BT_Cab)}
				orders = append(orders, hallButton)
				orders = append(orders, cabButton)
				hallButton1 := ButtonEvent{Floor: dirnFloor.Floor, Button: ButtonType(BT_HallUp)}
				if(dirnFloor.Floor == 0){
					orders = append(orders, hallButton1)
				}
			} else if dirnFloor.Dirn == MD_Up {
				hallButton := ButtonEvent{Floor: dirnFloor.Floor, Button: ButtonType(BT_HallUp)}
				cabButton := ButtonEvent{Floor: dirnFloor.Floor, Button: ButtonType(BT_Cab)}
				orders = append(orders, hallButton)
				orders = append(orders, cabButton)
				hallButton1 := ButtonEvent{Floor: dirnFloor.Floor, Button: ButtonType(BT_HallDown)}
				if(dirnFloor.Floor == (NUMFLOORS-1)){
					orders = append(orders, hallButton1)
				}
			} else {
				for btn := 0; btn < NUMBUTTONTYPE; btn++ {
					button := ButtonEvent{Floor: dirnFloor.Floor, Button: ButtonType(btn)}
					orders = append(orders, button)
				}
			}

			finishedOrder := OrderEvent{ElevatorID: ID, Completed: true, Orders: orders}
			Ch_orderEventToMaster <- finishedOrder


			case <-Ch_orderCopyRequest:
				fmt.Println("OrderHandler: Mottok kopi-forespørsel. Skriver backup av cab-ordrene...")
				if err := backup.WriteCabOrdersToFile(ordersFromMaster, "caborders_backup.json"); err != nil {
					fmt.Println("Feil ved skriving av backup:", err)
				} else {
					fmt.Println("Backup skrevet til caborders_backup.json")
				}
				fmt.Println("Sender order copy response til master:", ordersFromMaster)
				Ch_orderCopyResponse <- ordersFromMaster
		}
	}
}


