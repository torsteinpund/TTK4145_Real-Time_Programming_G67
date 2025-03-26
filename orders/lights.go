package orders

import (
	"Driver-go/elevatorDriver"
	. "Driver-go/types"
)


func setAllLights(globalOrderMap GlobalOrderMap, ID string){
	setHallLights(globalOrderMap)
	setCabLights(globalOrderMap[ID])
}

func setLocalLights(orderMatrix OrderMatrix){
	for floor := 0; floor < NUMFLOORS; floor++ {
		for btn := 0; btn < NUMBUTTONTYPE; btn++ {
			state := orderMatrix[floor][btn]
			elevatorDriver.SetButtonLamp(ButtonType(btn), floor, state)
		}
	}
}

func setHallLights(globalOrderMap GlobalOrderMap) {
	for floor := 0; floor < NUMFLOORS; floor++ {
		for btn := 0; btn < NUMHALLBUTTONS; btn++ {
			hasOrder := false
			for _, orderMatrix := range globalOrderMap {
				if orderMatrix[floor][btn] {
					hasOrder = true
					break
				}
			}
			elevatorDriver.SetButtonLamp(ButtonType(btn), floor, hasOrder)
		}
	}
}

func setCabLights(orderMatrix OrderMatrix) {
	for floor := 0; floor < NUMFLOORS; floor++ {
		state := orderMatrix[floor][BT_Cab]
		elevatorDriver.SetButtonLamp(BT_Cab, floor, state)
	}
}	

