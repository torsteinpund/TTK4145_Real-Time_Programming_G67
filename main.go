package main

import (
	. "Driver-go/types"
	"Driver-go/elevio"
	"Driver-go/fsm"
	"Driver-go/network/bcast"
	"Driver-go/timer"
	"fmt"
	"time"
)

func main() {
	fmt.Println("Started!")

	// Initialize elevator with hardware connection
	numFloors := 4
	elevio.InitHardwareConnection("localhost:15657", ) // Connect to hardware server
	emptyElev := Elevator{}
	elev := elevio.InitElevator(numFloors, NUMBUTTONTYPE, emptyElev)

	// Check if elevator starts between floors
	if initialFloor := elevio.GetFloor(); initialFloor == -1 {
		fmt.Println("Elevator is between floors on startup. Running initialization...")
		elev.Behaviour, elev.Dirn = fsm.FsmInitBetweenFloors()
	} else {
		// If the elevator starts at a valid floor, initialize its state
		elev = fsm.FsmFloorArrival(initialFloor, elev)
	}

	// Polling rate configuration
	inputPollRate := 25 * time.Millisecond // Adjust as needed

	// Event channels for hardware events
	buttonPressCh := make(chan ButtonEvent)
	floorSensorCh := make(chan int)
	stopButtonCh := make(chan bool)
	obstructionSwitchCh := make(chan bool)
	receivedButtonPressCh := make(chan ButtonEvent)
	// Start polling goroutines
	go elevio.PollButtons(buttonPressCh)
	go elevio.PollFloorSensor(floorSensorCh)
	go elevio.PollStopButton(stopButtonCh)
	go elevio.PollObstructionSwitch(obstructionSwitchCh)

	go bcast.Transmitter(15000, buttonPressCh)
	go bcast.Receiver(15000, receivedButtonPressCh)

	// Initialize system state
	prevFloor := -1
	//timerActive := false
	//var timerEndTime float64

	obstructionActive := false
	lastKnownDirection := MotorDirection(0)
	stop := false

	// Main event loop
	for {
		select {
		case buttonEvent := <-receivedButtonPressCh:
			// Handle button press event
			fmt.Printf("Received: %#v\n", buttonEvent.Floor)
			if stop {
				elevio.SetMotorDirection(lastKnownDirection)
				stop = false
			}
			fmt.Println("Button pressed!")

			fmt.Printf("Button pressed at floor %d, button type %d\n", buttonEvent.Floor, buttonEvent.Button)
			elev = fsm.FsmButtonPressed(buttonEvent.Floor, buttonEvent.Button, elev)

		case currentFloor := <-floorSensorCh:
			// Handle floor sensor event

			if currentFloor != prevFloor {
				fmt.Printf("Arrived at floor %d\n", currentFloor)
				elev = fsm.FsmFloorArrival(currentFloor, elev)
				elevio.SetFloorIndicator(currentFloor) // Update floor indicator lamp

				if !obstructionActive {
					timer.TimerStop()
					timer.TimerStart(3.0)
					fmt.Println("timer started")
				}
				// Stop and restart the timer when arriving at a floor
				// Set door timeout to 3 seconds
			}
			prevFloor = currentFloor
			obstructionActive = false

		case stopPressed := <-stopButtonCh:
			// Handle stop button event
			if stopPressed {
				lastKnownDirection = elev.Dirn
				fmt.Println("Stop button pressed!")
				fmt.Println(lastKnownDirection)
				elevio.SetStopLamp(true)
				elevio.SetMotorDirection(0)
				stop = true
			} else {
				fmt.Println("Stop button released!")
				elevio.SetStopLamp(false)
			}

		case <-time.After(inputPollRate):
			// Periodic tasks (check timer)
			if timer.TimerTimedOut() {
				fmt.Println("Door timeout occurred.")
				elev = fsm.FsmDoorTimeout(elev)
				timer.TimerStop() // Reset the timer after timeout handling
			}
		case obstruction := <-obstructionSwitchCh:
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
	}
}
