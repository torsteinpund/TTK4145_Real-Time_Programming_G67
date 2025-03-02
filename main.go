// test_program.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
)

// CombinedElevators representerer input til hall request assigneren.
type CombinedElevators struct {
	HallRequests [NUMFLOORS][NUMHALLBUTTONS]bool      `json:"hallRequests"`
	States       map[string]SingleElevatorState        `json:"states"`
}

// SingleElevatorState representerer tilstanden til en enkelt heis slik som assigneren forventer.
type SingleElevatorState struct {
	Behaviour   string   `json:"behaviour"`   // "idle", "moving", eller "doorOpen"
	Floor       int      `json:"floor"`
	Direction   string   `json:"direction"`   // "up", "down" eller "stop"
	CabRequests []bool   `json:"cabRequests"` // Én bool per etasje
}

// AssignmentMap er outputen til assigneringsfunksjonen.
type AssignmentMap map[string][][]bool

// Konstanter slik som brukt i README.md
const (
	NUMFLOORS      = 4
	NUMHALLBUTTONS = 2 // For hall-knapper: opp og ned
)

// assignHallRequests fordeler hall-forespørsler til de ulike heisene
// basert på en enkel kostnadsberegning.
func assignHallRequests(input CombinedElevators) AssignmentMap {
	// Initialiser output: For hver heis lages en matrise med NUMFLOORS rader og NUMHALLBUTTONS kolonner.
	assignments := make(AssignmentMap)
	for id := range input.States {
		matrix := make([][]bool, NUMFLOORS)
		for i := 0; i < NUMFLOORS; i++ {
			matrix[i] = make([]bool, NUMHALLBUTTONS)
		}
		assignments[id] = matrix
	}

	// For hver etasje og for hver hall-knapp, dersom det er en aktiv forespørsel,
	// finn den heisen med lavest "kostnad" og tildel forespørselen.
	for floor := 0; floor < NUMFLOORS; floor++ {
		for btn := 0; btn < NUMHALLBUTTONS; btn++ {
			if input.HallRequests[floor][btn] {
				bestElevator := ""
				bestCost := 1e9
				for id, state := range input.States {
					cost := computeCost(state, floor, btn)
					if cost < bestCost {
						bestCost = cost
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

// computeCost beregner en enkel kostnad for en heis til å betjene en hall-request.
// I dette eksempelet brukes avstanden i etasjer, med bonuser for heiser som er idle
// eller har riktig retning i forhold til knappen (btn 0 er "up", btn 1 er "down").
func computeCost(elevator SingleElevatorState, requestFloor int, button int) float64 {
	cost := abs(float64(elevator.Floor - requestFloor))

	// Bonus for idle heis
	if elevator.Behaviour == "idle" {
		cost -= 0.5
	}

	// Bonus dersom retning stemmer overens med knappen
	if button == 0 && elevator.Direction == "up" {
		cost -= 0.2
	}
	if button == 1 && elevator.Direction == "down" {
		cost -= 0.2
	}

	return cost
}

// abs returnerer absoluttverdien av et tall.
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func main() {
	// Eksempel-JSON som i README.md
	inputJSON := `
{
    "hallRequests": [[false, false], [true, false], [true, false], [false, true]],
    "states": {
        "one": {
            "behaviour": "moving",
            "floor": 2,
            "direction": "up",
            "cabRequests": [false, false, true, true]
        },
        "two": {
            "behaviour": "idle",
            "floor": 0,
            "direction": "stop",
            "cabRequests": [false, false, false, false]
        }
    }
}`

	var input CombinedElevators
	if err := json.Unmarshal([]byte(inputJSON), &input); err != nil {
		log.Fatalf("Feil ved parsing av JSON: %v", err)
	}

	// Kjør assigneringsalgoritmen
	assignments := assignHallRequests(input)

	// Konverter output til JSON for lesbar utskrift
	output, err := json.MarshalIndent(assignments, "", "    ")
	if err != nil {
		log.Fatalf("Feil ved generering av output JSON: %v", err)
	}

	fmt.Println("Assignment output:")
	fmt.Println(string(output))
}
