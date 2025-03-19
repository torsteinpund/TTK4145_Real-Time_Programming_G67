package fsm

import (
    "Driver-go/lights"
    "Driver-go/singleElevatorDriver/elevio"
    "Driver-go/singleElevatorDriver/requests"
    "Driver-go/singleElevatorDriver/timer"
    . "Driver-go/types"
    "fmt"
    "time"
)

type FsmChannels struct {
    Ch_floorSensor      chan int
    Ch_stopButton 		chan bool
    Ch_obstruction   	chan bool
    Ch_stateUpdate 		chan Elevator
    Ch_localLights		chan OrderMatrix
    Ch_localOrders		chan OrderMatrix
    Ch_toMaster 		chan NetworkMessage
    Ch_clearedFloor 	chan int
}

// fsmInit initialiserer heisens tilstand og tilhørende systemer
/*func fsmInit() {
    // Initialiser heisen med standardverdier
    elevator = ElevatorUninitialized()

    // Last inn konfigurasjon fra fil (simulert)
    err := loadConfig("con", &elevator)
    if err != nil {
        fmt.Println("Error loading configuration:", err)
    }

    // Initialiser output-enheten (simulert)
    outputDevice = getOutputDevice()
}*/

func fsmInitBetweenFloors() (ElevatorBehaviour, MotorDirection) {
    // Move the elevator down until it reaches a floor
    
    for{
        elevio.SetMotorDirection(MD_Down)
        if elevio.GetFloor() != -1 {
            break
        }
    }
    // Update the elevator's state
    dirn := MD_Stop
    elevio.SetMotorDirection(MD_Stop)
    behaviour := ElevatorBehaviour(EB_Idle)
    return behaviour, dirn
}

func fsmButtonPressed(elev Elevator) Elevator {
    // fmt.Printf("\n\nfsmOnRequestButtonPress(%d, %v)\n", btnFloor, btnType)

    dirnBehaviour := requests.RequestsChooseDirection(elev)
    
    elev.Dirn = dirnBehaviour.Dirn
    elev.Behaviour = ElevatorBehaviour(dirnBehaviour.Behaviour)

    switch dirnBehaviour.Behaviour {
    
    case EB_DoorOpen:
        elevio.SetDoorOpenLamp(true)
        timer.TimerStart(elev.Config.DoorOpenDuration)
        elev = requests.RequestsClearAtCurrentFloor(elev, nil)

    case EB_Moving:
        elevio.SetMotorDirection(elev.Dirn)

    case EB_Idle:
        //No action
    }

    // Update button lights
    // elev.Requests = lights.SetHallLights(elev.Requests)
    // elev.Requests = lights.SetCabLights(elev.Requests)

    return elev
}

func fsmFloorArrival(newFloor int, elev Elevator, ch_doorOpen chan<- bool) Elevator {
    // fmt.Printf("\n\nfsmOnFloorArrival(%d)\n", newFloor)

    elev.Floor = newFloor

    elevio.SetFloorIndicator(elev.Floor)

    switch elev.Behaviour {
    case ElevatorBehaviour(EB_Moving):
        // Check if the elevator should stop at the current floor
        if requests.RequestsShouldStop(elev) {

            ch_doorOpen <- true
            
            elevio.SetMotorDirection(MD_Stop)

            elevio.SetDoorOpenLamp(true)

            elev = requests.RequestsClearAtCurrentFloor(elev, nil)

            timer.TimerStart(elev.Config.DoorOpenDuration)

            // elev.Requests = lights.SetHallLights(elev.Requests)
            elev.Requests = lights.SetCabLights(elev.Requests)

            elev.Behaviour = ElevatorBehaviour(EB_DoorOpen)
        }

    default:
        // No action
    }

    return elev
}

func fsmDoorTimeout(elev Elevator) Elevator {

    switch elev.Behaviour {
    case ElevatorBehaviour(EB_DoorOpen):
        // Choose direction based on requests
        dirnBehaviour := requests.RequestsChooseDirection(elev)
        elev.Dirn = dirnBehaviour.Dirn
        elev.Behaviour = ElevatorBehaviour(dirnBehaviour.Behaviour)

        switch elev.Behaviour {
        case ElevatorBehaviour(EB_DoorOpen):
            // Start timer and clear requests
            timer.TimerStart(elev.Config.DoorOpenDuration)
            elev = requests.RequestsClearAtCurrentFloor(elev, nil)
            // elev.Requests = lights.SetHallLights(elev.Requests)
            elev.Requests = lights.SetCabLights(elev.Requests)

        case ElevatorBehaviour(EB_Moving), ElevatorBehaviour(EB_Idle):
            // Shut the door and start moving
            elevio.SetDoorOpenLamp(false)
            elevio.SetMotorDirection(elev.Dirn)
        }

    default:
        // No action
    }

    return elev
}

func FsmRun(ch_fsm FsmChannels, elev Elevator) {
    fmt.Println("FSM Started!")

    // Initialize elevator with hardware connection
    // numFloors := 4
    // elevio.InitHardwareConnection("localhost:15657") // Connect to hardware server
    // emptyElev := Elevator{}
    // elev := elevio.InitElevator(numFloors, NUMBUTTONTYPE, emptyElev)

    // Check if elevator starts between floors
    // if initialFloor := elevio.GetFloor(); initialFloor == -1 {
    // 	fmt.Println("Elevator is between floors on startup. Running initialization...")
    // 	elev.Behaviour, elev.Dirn = fsmInitBetweenFloors()
    // } else {
    // 	// If the elevator starts at a valid floor, initialize its state
    // 	elev = fsmFloorArrival(initialFloor, elev)
    // }

    // Polling rate configuration
    inputPollRate := 25 * time.Millisecond // Adjust as needed

    // Start polling goroutines
    // go elevio.PollButtons(ch_fsm.Ch_buttonPress)
    // go elevio.PollFloorSensor(ch_fsm.Ch_floorSensor)
    // go elevio.PollStopButton(ch_fsm.Ch_stopButton)
    // go elevio.PollObstructionSwitch(ch_fsm.Ch_obstruction)

    // Initialize system state
    prevFloor := -1
    //timerActive := false
    //var timerEndTime float64

    obstructionActive := false
    lastKnownDirection := MotorDirection(0)
    // stop := false

    // Main event loop
    for {
        select {
        // case buttonEvent := <-ch_fsm.Ch_buttonPress:
        // 	// Handle button press event
        // 	fmt.Printf("Received: %#v\n", buttonEvent.Floor)
        // 	if stop {
        // 		elevio.SetMotorDirection(lastKnownDirection)
        // 		stop = false
        // 	}

        // 	fmt.Printf("Button pressed at floor %d, button type %d\n", buttonEvent.Floor, buttonEvent.Button)
        // 	elev = fsmButtonPressed(buttonEvent.Floor, buttonEvent.Button, elev)
        case receivedOrder := <- ch_fsm.Ch_localOrders:
            fmt.Println("Received order")
            elev.Requests = receivedOrder
            // fmt.Println(elev.Requests)
            elev = fsmButtonPressed(elev)

        case currentFloor := <-ch_fsm.Ch_floorSensor:
            // Handle floor sensor event
            fmt.Printf("Received floor sensor event: %d\n", currentFloor)

            if currentFloor != prevFloor {
                fmt.Printf("Arrived at floor %d\n", currentFloor)
                elev = fsmFloorArrival(currentFloor, elev)
                elevio.SetFloorIndicator(currentFloor) // Update floor indicator lamp
                fmt.Println("After FSMFLOORARRIVAL," ,elev.Requests)
                // ch_fsm.Ch_stateUpdate<-elev
                if !obstructionActive {
                    timer.TimerStop()
                    timer.TimerStart(3.0)
                    // fmt.Println("ti//Passes the updated statemer started")
                }
                // Stop and restart the timer when arriving at a floor
                // Set door timeout to 3 seconds
            }
            prevFloor = currentFloor
            obstructionActive = false

        case stopPressed := <-ch_fsm.Ch_stopButton:
            // Handle stop button event
            if stopPressed {
                lastKnownDirection = elev.Dirn
                fmt.Println("Stop button pressed!")
                fmt.Println(lastKnownDirection)
                elevio.SetStopLamp(true)
                elevio.SetMotorDirection(0)
                // stop = true
            } else {
                fmt.Println("Stop button released!")
                elevio.SetStopLamp(false)
            }

        case <-time.After(inputPollRate):
            // Periodic tasks (check timer)
            if timer.TimerTimedOut() {
                // fmt.Println("Door timeout occurred.")
                elev = fsmDoorTimeout(elev)
                timer.TimerStop() // Reset the timer after timeout handling
            }
        case obstruction := <-ch_fsm.Ch_obstruction:
            if obstruction {
                obstructionActive = true
                timer.TimerStop()
                fmt.Println("obstruction switch")
            } else if !obstruction {
                obstructionActive = false
                timer.TimerStop()
                timer.TimerStart(3.0)
                fmt.Println("obstruction switch off")
            }
        }
        ch_fsm.Ch_stateUpdate <- elev //Passes the updated state
    }
}