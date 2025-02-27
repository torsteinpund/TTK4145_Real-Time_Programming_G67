package orderHandler

import (
	. "Driver-go/types"
	"Driver-go/requests"
)

func timeToServeRequest(e_old Elevator, receivedCh <-chan ButtonEvent) float64 {
	e := e_old
	buttenEvent := <-receivedCh
	b := buttenEvent.Button
	f := buttenEvent.Floor

	e.Requests[f][b] = 1
	arrivedAtRequest := false

	ifEqual := func(inner_b ButtonType, inner_f int) {
		if inner_b == b && inner_f == f {
			arrivedAtRequest = true
		}
	}

	duration := 0.0

	switch e.Behaviour {
	case ElevatorBehaviour(EB_Idle):
		e.Dirn = requests.RequestsChooseDirection(e).Dirn
		if e.Dirn == MD_Stop {
			return duration
		}
	case ElevatorBehaviour(EB_Moving):
		duration += e.Config.TimeBetweenFloors / 2
		e.Floor += int(e.Dirn)
	case ElevatorBehaviour(EB_DoorOpen):
		duration -= e.Config.DoorOpenDuration / 2
	}

	for {
		if requests.RequestsShouldStop(e) {
			e = requests.RequestsClearAtCurrentFloor(e, ifEqual)
			if arrivedAtRequest {
				return duration
			}
			duration += e.Config.DoorOpenDuration
			e.Dirn = requests.RequestsChooseDirection(e).Dirn
		}
		e.Floor += int(e.Dirn)
		duration += e.Config.TimeBetweenFloors
	}
}
