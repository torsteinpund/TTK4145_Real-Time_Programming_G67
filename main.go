// test_program.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	. "Driver-go/network/masterSelector"
)

type Data struct {
	ID string `json:"id"`
}

func main() {
	

	data := []Data{
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
		{ID: ""},
	}

	elevatorIDs := make([]string, len(data))
	for i, d := range data {
		elevatorIDs[i] = d.ID
	}

	highestID :=UpdateIDs(elevatorIDs)

	for i := range data {
		data[i].ID = elevatorIDs[i]
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Error marshaling data: %v", err)
	}

	fmt.Println(string(jsonData))
}



