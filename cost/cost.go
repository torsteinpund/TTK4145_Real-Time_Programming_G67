package cost

import(
	. "Driver-go/types"
	"Driver-go/singleElevatorDriver/requests"
)



func TimeToServeRequest(e_old Elevator, receivedCh <-chan ButtonEvent) float64 {
	e := e_old
	buttonEvent := <-receivedCh
	b := buttonEvent.Button
	f := buttonEvent.Floor

	e.Requests[f][b] = true
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

// func assignHallRequests(input AllElevators) GlobalOrderMap {
// 	// Initialiser output for hver heis med en matrise (NUMFLOORS x NUMHALLBUTTONS) satt til false.
// 	globalOrderMap := GlobalOrderMap{}
// 	for id := range input.States {
// 		matrix := OrderMatrix{}
// 		globalOrderMap[id] = matrix
// 	}

// 	// For hver etasje og for hver hall-knapp (opp og ned), hvis det er en aktiv forespørsel,
// 	// finn den heisen med lavest "kostnad" og tildel denne forespørselen.
// 	for floor := 0; floor < NUMFLOORS; floor++ {
// 		for btn := 0; btn < NUMHALLBUTTONS; btn++ {
// 			if input.GlobalOrders[floor][btn] {
// 				bestElevator := ""
// 				bestCost := math.MaxFloat64
// 				for id, state := range input.States {
// 					c := ComputeCost(state, floor, btn)
// 					if c < bestCost {
// 						bestCost = c
// 						bestElevator = id
// 					}
// 				}
// 				if bestElevator != "" {
// 					matrix := globalOrderMap[bestElevator]
// 					matrix[floor][btn] = true
// 					globalOrderMap[bestElevator] = matrix

// 				}
// 			}
// 		}
// 	}
	
// 	return globalOrderMap
// }



// func ComputeCost(elevator StateSingleElevator, requestFloor int, button int) float64 {
// 	// Grunnkostnad basert på avstand (absolutt forskjell i etasjer)
// 	cost := math.Abs(float64(elevator.Floor - requestFloor))

// 	// Bonus: Hvis heisen er idle, trekk litt fra kostnaden
// 	if strings.ToLower(elevator.ElevatorBehaviour) == "idle" {
// 		cost -= 0.5
// 	}

// 	// Hvis heisens retning stemmer overens med forespurt knapp, trekk også litt fra
// 	if button == 0 && strings.ToLower(elevator.Direction) == "up" {
// 		cost -= 0.2
// 	}
// 	if button == 1 && strings.ToLower(elevator.Direction) == "down" {
// 		cost -= 0.2
// 	}

// 	return cost
// }