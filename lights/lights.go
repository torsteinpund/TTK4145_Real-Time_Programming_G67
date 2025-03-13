package lights

import (
	"Driver-go/singleElevatorDriver/elevio"
	. "Driver-go/types"
	"time"
	"fmt"
)

func SetHallLights(ch_Lightschannel chan OrderMatrix) {
    for {
        select {
        case req, ok := <-ch_Lightschannel:
			fmt.Println("SetHallLights: ", req)
            if !ok {
                fmt.Println("ok NOT") // If the channel is closed, exit the function.
            }
            for floor := 0; floor < NUMFLOORS; floor++ {
                for btn := 0; btn < NUMBUTTONTYPE-1; btn++ {
                    state := req[floor][btn]
                    elevio.SetButtonLamp(ButtonType(btn), floor, state)
                }
            }
			fmt.Println("SetHallLights: Done")
        default:
			
            time.Sleep(10 * time.Millisecond) 
    }
}
}

func SetCabLights(req [NUMFLOORS][NUMBUTTONTYPE]bool) [NUMFLOORS][NUMBUTTONTYPE]bool {
	for floor := 0; floor < NUMFLOORS; floor++ {
		state := req[floor][2]
		elevio.SetButtonLamp(BT_Cab, floor, state)
	}
	return req
}
