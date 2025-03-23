package lights

import (
	"Driver-go/singleElevatorDriver/elevio"
	. "Driver-go/types"
	// "time"
	// "fmt"
)


func SetLights(globalOrderMap GlobalOrderMap, ID string){
	setHallLights(globalOrderMap)
	setCabLights(globalOrderMap[ID])
}



func setHallLights(globalOrderMap GlobalOrderMap) {
	emptyOrderMatrix := OrderMatrix{}
    for _, orderMatrix := range globalOrderMap {
		for floor := 0; floor < NUMFLOORS; floor++ {
			for btn := 0; btn < NUMHALLBUTTONS; btn++ {
				if orderMatrix[floor][btn] {
					emptyOrderMatrix[floor][btn] = true

				}
			}
		}
	}
	for floor := 0; floor < NUMFLOORS; floor++ {
		for btn := 0; btn < NUMHALLBUTTONS; btn++ {
			state := emptyOrderMatrix[floor][btn]
			elevio.SetButtonLamp(ButtonType(btn), floor, state)
		}
	}
}

func setCabLights(orderMatrix OrderMatrix) {
	for floor := 0; floor < NUMFLOORS; floor++ {
		state := orderMatrix[floor][BT_Cab]
		elevio.SetButtonLamp(BT_Cab, floor, state)
	}
}	

