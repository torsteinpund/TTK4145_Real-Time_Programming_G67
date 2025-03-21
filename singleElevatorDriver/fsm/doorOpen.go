package fsm

import (
    
    "time"
	"Driver-go/singleElevatorDriver/elevio"
	// "Driver-go/lights"
	// "Driver-go/singleElevatorDriver/requests"
	// . "Driver-go/types"

)


type DoorChannels struct {
	Ch_doorOpen chan bool
	Ch_obstruction chan bool
	Ch_doorClosed chan bool
}


type DoorState int

const (
	DoorClosed DoorState = iota
	DoorInCountdown
	DoorObstructed
)

func HandleDoor(doorChannels DoorChannels, ch_obstructionFSM chan<- bool) {
	elevio.SetDoorOpenLamp(false)

	isObstructed := false
	doorState := DoorClosed
	doorTimer := time.NewTimer(time.Hour)
	doorTimer.Stop()

	for {
		select {
		case isObstructed = <-doorChannels.Ch_obstruction:
			if !isObstructed && doorState == DoorObstructed {
				elevio.SetDoorOpenLamp(false)
				doorChannels.Ch_doorClosed <- true
				doorState = DoorClosed
			}
			if isObstructed {
				ch_obstructionFSM <- true
			} else {
				
				ch_obstructionFSM <- false
			}
		case <-doorChannels.Ch_doorOpen:
			if isObstructed {
				ch_obstructionFSM <- true
			}
			switch doorState {
			case DoorClosed:
				elevio.SetDoorOpenLamp(true)
				doorTimer = time.NewTimer(3 * time.Second)
				doorState = DoorInCountdown
			case DoorInCountdown:
				doorTimer = time.NewTimer(3 * time.Second)
			case DoorObstructed:
				doorTimer = time.NewTimer(3 * time.Second)
				doorState = DoorInCountdown
			default:
				panic("Door state not implemented")
			}
		case <-doorTimer.C:
			if doorState != DoorInCountdown {
				panic("Door timer expired in wrong state")
			}
			if isObstructed {
				doorState = DoorObstructed
			} else {
				elevio.SetDoorOpenLamp(false)
				doorChannels.Ch_doorClosed <- true
				doorState = DoorClosed
			}
		}
	}
}
