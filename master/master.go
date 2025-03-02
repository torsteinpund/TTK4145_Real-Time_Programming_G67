package master

import (
	. "Driver-go/types"
	//"encoding/json"
	"fmt"
	//"os/exec"
	//"Driver-go/cost"
	"math"
	"strings"
)

type MasterChannels struct {
	IsMasterChannel      chan bool
	PeerLostChannel      chan string
	ToSlavesChannel      chan NetworkMessage
	RegisterOrderChannel chan OrderEvent
	StateUpdateChannel   chan Elevator
}

// StateSingleElevator represents the state of a single elevator
type StateSingleElevator struct {
	Floor             int    `json:"floor"`
	Direction         string `json:"direction"`
	ElevatorBehaviour string `json:"behaviour"`
	Avaliable         bool
	CabOrders         [NUMFLOORS]bool `json:"caborders"`
}

type AssignmentMap map[string][][]bool

// Elevators represents the global state of all elevators, including global orders
// and the state of each individual elevator.
type Elevators struct {
	GlobalOrders [NUMFLOORS][NUMBUTTONTYPE - 1]bool
	States       map[string]StateSingleElevator
}

func RunMaster() {
	fmt.Println("Running master...")

	for {
		select {
		// Add your cases here
		}
	}
}

func AssignHallRequests(input Elevators) AssignmentMap {
	// Initialiser output for hver heis med en matrise (NUMFLOORS x NUMHALLBUTTONS) satt til false.
	assignments := make(AssignmentMap)
	for id := range input.States {
		matrix := make([][]bool, NUMFLOORS)
		for i := 0; i < NUMFLOORS; i++ {
			matrix[i] = make([]bool, NUMBUTTONTYPE-1)
		}
		assignments[id] = matrix
	}

	// For hver etasje og for hver hall-knapp (opp og ned), hvis det er en aktiv forespørsel,
	// finn den heisen med lavest "kostnad" og tildel denne forespørselen.
	for floor := 0; floor < NUMFLOORS; floor++ {
		for btn := 0; btn < NUMBUTTONTYPE-1; btn++ {
			if input.GlobalOrders[floor][btn] {
				bestElevator := ""
				bestCost := math.MaxFloat64
				for id, state := range input.States {
					c := ComputeCost(state, floor, btn)
					if c < bestCost {
						bestCost = c
						bestElevator = id
					}
				}
				if bestElevator != "" {
					assignments[bestElevator][floor][btn] = true
				}
			}
		}
	}
	return assignments
}

func ComputeCost(elevator StateSingleElevator, requestFloor int, button int) float64 {
	// Grunnkostnad basert på avstand (absolutt forskjell i etasjer)
	cost := math.Abs(float64(elevator.Floor - requestFloor))

	// Bonus: Hvis heisen er idle, trekk litt fra kostnaden
	if strings.ToLower(elevator.ElevatorBehaviour) == "idle" {
		cost -= 0.5
	}

	// Hvis heisens retning stemmer overens med forespurt knapp, trekk også litt fra
	if button == 0 && strings.ToLower(elevator.Direction) == "up" {
		cost -= 0.2
	}
	if button == 1 && strings.ToLower(elevator.Direction) == "down" {
		cost -= 0.2
	}

	return cost
}
