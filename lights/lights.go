package lights

import (
	"Driver-go/singleElevatorDriver/elevio"
	. "Driver-go/types"
)


func SetAllLights(globalOrderMap GlobalOrderMap, ID string){
	setHallLights(globalOrderMap)
	setCabLights(globalOrderMap[ID])
}


func SetLocalLights(orderMatrix OrderMatrix){
	for floor := 0; floor < NUMFLOORS; floor++ {
		for btn := 0; btn < NUMBUTTONTYPE; btn++ {
			state := orderMatrix[floor][btn]
			elevio.SetButtonLamp(ButtonType(btn), floor, state)
		}
	}
}


func setHallLights(globalOrderMap GlobalOrderMap) {
	for floor := 0; floor < NUMFLOORS; floor++ {
		for btn := 0; btn < NUMHALLBUTTONS; btn++ {
			hasOrder := false
			// Sjekk om noen av orderMatrixene har en bestilling for denne etasjen og knappen
			for _, orderMatrix := range globalOrderMap {
				if orderMatrix[floor][btn] {
					hasOrder = true
					break
				}
			}
			elevio.SetButtonLamp(ButtonType(btn), floor, hasOrder)
		}
	}
}


func setCabLights(orderMatrix OrderMatrix) {
	for floor := 0; floor < NUMFLOORS; floor++ {
		state := orderMatrix[floor][BT_Cab]
		elevio.SetButtonLamp(BT_Cab, floor, state)
	}
}	

