package main

import (
	"encoding/json"
	"fmt"

	//. "Driver-go/network/masterSelector"
	"Driver-go/master"
	"os/exec"
	. "Driver-go/types"
)


func main(){

    


    input := master.AllElevators{
        GlobalOrders: [NUMFLOORS][NUMHALLBUTTONS]bool{{false, false}, {true, false}, {false, false}, {false, true}},
        States: map[string]master.StateSingleElevator{
            "one": master.StateSingleElevator{
                ElevatorBehaviour:       "moving",
                Floor:          2,
                Direction:      "up",
				Available: true,
                CabOrders:    [NUMFLOORS]bool{false, false, false, true},
				
            },
            "two": master.StateSingleElevator{
                ElevatorBehaviour:       "idle",
                Floor:          0,
                Direction:      "stop",
				Available: true,
                CabOrders:    [NUMFLOORS]bool{false, false, false, false},
				
            },
        },
    }

	hraExecutable := "hall_request_assigner"

    jsonBytes, err := json.Marshal(input)
    if err != nil {
        fmt.Println("json.Marshal error: ", err)
        return
    }
    
    ret, err := exec.Command("../TTK4145_Real-Time_Programming_G67/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
    if err != nil {
        fmt.Println("exec.Command error: ", err)
        fmt.Println(string(ret))
        return
    }
    
    output := new(map[string][NUMFLOORS][NUMHALLBUTTONS]bool)
    err = json.Unmarshal(ret, &output)
    if err != nil {
        fmt.Println("json.Unmarshal error: ", err)
        return
    }
        
    fmt.Printf("output: \n")
    for k, v := range *output {
        fmt.Printf("%6v :  %+v\n", k, v)
    }
}


