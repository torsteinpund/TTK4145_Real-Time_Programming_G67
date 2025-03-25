package backup

import (
	"encoding/json"
	"os"
	. "Driver-go/types"
)

// var Filename = "backup.json"

func WriteCabOrdersToFile(orders GlobalOrderMap, filename string) error {
	// Vi lager en midlertidig mappe der vi kun lagrer cab‑ordrene (en bool-slice per heis)
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

	return os.WriteFile(filename, data, 0644)
}

func ReadCabOrdersFromFile(filename string) (GlobalOrderMap, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Midlertidig map: ID -> []bool (kun cab‑ordrer)
	var backup map[string][]bool
	err = json.Unmarshal(data, &backup)
	if err != nil {
		return nil, err
	}

	orders := make(GlobalOrderMap)
	for id, cabOrders := range backup {
		var orderMatrix OrderMatrix // Alle verdier er false som standard
		for floor := 0; floor < NUMFLOORS && floor < len(cabOrders); floor++ {
			orderMatrix[floor][BT_Cab] = cabOrders[floor]
		}
		orders[id] = orderMatrix
	}
	return orders, nil
}

func IsGlobalOrderMapEmpty(orders GlobalOrderMap) bool {
    for _, orderMatrix := range orders {
        for floor := 0; floor < NUMFLOORS; floor++ {
            for btn := 0; btn < NUMBUTTONTYPE; btn++ {
                if orderMatrix[floor][btn] {
                    return false
                }
            }
        }
    }
    return true
}

func CheckIfUpdatedGlobalOrderMap(updatedOrders GlobalOrderMap, lastGlobaleOrderMap GlobalOrderMap) bool {
	for elevatorID, orders := range updatedOrders {
		for floor, row := range orders {
			for button, isOrder := range row {
				if isOrder != lastGlobaleOrderMap[elevatorID][floor][button] {
					return true
				}
			}
		}
	}
	return false
}

