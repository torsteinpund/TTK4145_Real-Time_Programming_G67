package orderHandler

import (
	"Driver-go/lights"
	. "Driver-go/types"
	"fmt"
	"encoding/json"
    "os"
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
				fmt.Println("Sending ordercipies to master", ordersFromMaster)
				Ch_orderCopyResponse <- ordersFromMaster

		}
	}
}


func writeCabOrdersToFile(orders GlobalOrderMap) error {
    // Vi lager en midlertidig mappe som kun lagrer cab-orders (slice med bool for hvert nivå)
    backup := make(map[string][]bool)
    for id, orderMatrix := range orders {
        cabOrders := make([]bool, NUMFLOORS)
        for floor := 0; floor < NUMFLOORS; floor++ {
            cabOrders[floor] = orderMatrix[floor][BT_Cab]
        }
        backup[id] = cabOrders
    }
    data, err := json.MarshalIndent(backup, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile("caborders.txt", data, 0644)
}


func ReadCabOrdersFromFile() (GlobalOrderMap, error) {
    data, err := os.ReadFile("caborders.txt")
    if err != nil {
        return nil, err
    }
    // Midlertidig map: ID -> []bool (cab orders)
    var backup map[string][]bool
    err = json.Unmarshal(data, &backup)
    if err != nil {
        return nil, err
    }
    // Konverter til GlobalOrderMap (hvor vi lager et OrderMatrix med bare cab-knappen aktivert)
    globalOrders := make(GlobalOrderMap)
    for id, cabOrders := range backup {
        var orderMatrix OrderMatrix // alle verdier er false som standard
        for floor := 0; floor < NUMFLOORS && floor < len(cabOrders); floor++ {
            orderMatrix[floor][BT_Cab] = cabOrders[floor]
        }
        globalOrders[id] = orderMatrix
    }
    return globalOrders, nil
}
