package orders

import (
	. "Driver-go/types"
	"fmt"
)

func OrderHandler(ID string,
				  Ch_localOrders 		chan<- LocalOrder,
				  Ch_orderEventToMaster chan<- OrderEvent,
				  Ch_orderCopyToMaster  chan<- GlobalOrderMap,
				  Ch_buttonPress 		<-chan ButtonEvent,
				  Ch_clearedFloor 		<-chan DirnFloorPair,
				  Ch_ordersFromMaster 	<-chan GlobalOrderMap,
				  Ch_orderCopyRequest 	<-chan bool,
				  Ch_networkConnected 	<-chan bool) {

	fmt.Println("OrderHandler Started!")
	
	ordersFromMaster   := GlobalOrderMap{}
	localOrderMatrix   := OrderMatrix{}
	connectedToNetwork := false

	for {
		select {

		case buttonEvent := <-Ch_buttonPress:
			button := []ButtonEvent{buttonEvent}
			orderEvent := OrderEvent{ElevatorID: ID, Completed: false, Orders: button}

			if !connectedToNetwork {
				localOrderMatrix, ordersFromMaster = addOrderToLocalMatrix(orderEvent, ordersFromMaster, localOrderMatrix)
				setLocalLights(localOrderMatrix)
				Ch_localOrders <- LocalOrder{OrderMatrix: localOrderMatrix, NetworkConnection: connectedToNetwork}
			
			} else {
				if !orderEventInGlobalOrderMap(orderEvent, ordersFromMaster) {
					Ch_orderEventToMaster <- orderEvent
				}
			}

		case ordersFromMaster = <-Ch_ordersFromMaster:
			setAllLights(ordersFromMaster, ID)
			Ch_localOrders <- LocalOrder{OrderMatrix: ordersFromMaster[ID], NetworkConnection: connectedToNetwork}

		case dirnFloor := <-Ch_clearedFloor:
			if !connectedToNetwork {
				localOrderMatrix = clearLocalOrderMatrix(dirnFloor, localOrderMatrix)
				setLocalLights(localOrderMatrix)
				} else {
					orders := []ButtonEvent{}
					orders = clearFloor(dirnFloor, orders)
					finishedOrder := OrderEvent{ElevatorID: ID, Completed: true, Orders: orders}
					Ch_orderEventToMaster <- finishedOrder
				}

		case <-Ch_orderCopyRequest:
			Ch_orderCopyToMaster <- ordersFromMaster

		case networkStatus := <-Ch_networkConnected:
			connectedToNetwork = networkStatus
		}
	}
}


func orderEventInGlobalOrderMap(orderEvent OrderEvent, globalOrderMap GlobalOrderMap) bool {
    for _, order := range orderEvent.Orders {
		switch order.Button{
		case ButtonType(BT_Cab):
			elevatorOrders, exists := globalOrderMap[orderEvent.ElevatorID]
            if !exists {
                return false
            }

            if orderEvent.Completed {
                if elevatorOrders[order.Floor][order.Button] {
                    return false
                }
            } else {
                if !elevatorOrders[order.Floor][order.Button] {
                    return false
                }
            }

		case ButtonType(BT_HallDown), ButtonType(BT_HallUp):
			found := false
            for _, orderMatrix := range globalOrderMap {
                if orderMatrix[order.Floor][order.Button] {
                    found = true
                    break
                }
            }

            if orderEvent.Completed && found{
                if found {
                    return false
                }
            } else {
                if !found {
                    return false
                }
            }
		}
    }
    return true
}


func addOrderToLocalMatrix(orderEvent OrderEvent, globalOrderMap GlobalOrderMap, localOrderMatrix OrderMatrix) (OrderMatrix, GlobalOrderMap) {

    elevatorOrders, exists := globalOrderMap[orderEvent.ElevatorID]
    if !exists {
        elevatorOrders = OrderMatrix{} 
    }

    for floor, buttonMap := range elevatorOrders {
        for button, isOrder := range buttonMap {
            if button == int(BT_Cab) && isOrder {
				elevatorOrders[floor][button] = false
                localOrderMatrix[floor][button] = true
            }
        }
    }

    for _, order := range orderEvent.Orders {
        if order.Button == ButtonType(BT_Cab) {
            localOrderMatrix[order.Floor][order.Button] = true
        }
		
        if order.Button == ButtonType(BT_HallUp) || order.Button == ButtonType(BT_HallDown) {
            localOrderMatrix[order.Floor][order.Button] = true
        }
    }

	globalOrderMap[orderEvent.ElevatorID] = elevatorOrders

    return localOrderMatrix, globalOrderMap
}


func clearLocalOrderMatrix(clearedFloor DirnFloorPair, localOrderMatrix OrderMatrix) OrderMatrix {
	switch clearedFloor.Dirn {
	case MD_Up:
		localOrderMatrix[clearedFloor.Floor][ButtonType(BT_HallUp)] = false
		localOrderMatrix[clearedFloor.Floor][ButtonType(BT_Cab)] = false

	case MD_Down:
		localOrderMatrix[clearedFloor.Floor][ButtonType(BT_HallDown)] = false
		localOrderMatrix[clearedFloor.Floor][ButtonType(BT_Cab)] = false

	default:
		for button := 0; button < NUMBUTTONTYPE; button++ {
			localOrderMatrix[clearedFloor.Floor][ButtonType(button)] = false
		}
	}
	return localOrderMatrix
}


func clearFloor(dirnFloor DirnFloorPair, orders []ButtonEvent) []ButtonEvent{
	switch dirnFloor.Dirn{
	case MD_Down:
		hallButton := ButtonEvent{Floor: dirnFloor.Floor, Button: ButtonType(BT_HallDown)}
		cabButton  := ButtonEvent{Floor: dirnFloor.Floor, Button: ButtonType(BT_Cab)}
		orders = append(orders, hallButton)
		orders = append(orders, cabButton)
		hallButton1 := ButtonEvent{Floor: dirnFloor.Floor, Button: ButtonType(BT_HallUp)}
		if dirnFloor.Floor == 0 {
			orders = append(orders, hallButton1)
		}

	case MD_Up:
		hallButton := ButtonEvent{Floor: dirnFloor.Floor, Button: ButtonType(BT_HallUp)}
		cabButton  := ButtonEvent{Floor: dirnFloor.Floor, Button: ButtonType(BT_Cab)}
		orders = append(orders, hallButton)
		orders = append(orders, cabButton)
		hallButton1 := ButtonEvent{Floor: dirnFloor.Floor, Button: ButtonType(BT_HallDown)}
		if dirnFloor.Floor == (NUMFLOORS - 1) {
			orders = append(orders, hallButton1)
		}

	default:
		for btn := 0; btn < NUMBUTTONTYPE; btn++ {
			button := ButtonEvent{Floor: dirnFloor.Floor, Button: ButtonType(btn)}
			orders = append(orders, button)
		}
	}
	return orders
}
