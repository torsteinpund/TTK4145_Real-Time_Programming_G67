package main

import (
	"Driver-go/elevio"
	"fmt"
	"time"
)

func main() {
	fmt.Println("Started!")

	// Initialize elevator with hardware connection
	numFloors := 4
	elevio.Init("localhost:15657", numFloors) // Connect to hardware server

	initElevator(numFloors, elevio.NumButtonTypes)

	// Check if elevator starts between floors
	if initialFloor := elevio.GetFloor(); initialFloor == -1 {
		fmt.Println("Elevator is between floors on startup. Running initialization...")
		fsmOnInitBetweenFloors()
	} else {
		// If the elevator starts at a valid floor, initialize its state
		fsmOnFloorArrival(initialFloor)
	}

	// Polling rate configuration
	inputPollRate := 25 * time.Millisecond // Adjust as needed

	// Event channels for hardware events
	buttonPressCh := make(chan elevio.ButtonEvent)
	floorSensorCh := make(chan int)
	stopButtonCh := make(chan bool)
	obstructionSwitchCh := make(chan bool)

	// Start polling goroutines
	go elevio.PollButtons(buttonPressCh)
	go elevio.PollFloorSensor(floorSensorCh)
	go elevio.PollStopButton(stopButtonCh)
	go elevio.PollObstructionSwitch(obstructionSwitchCh)

	// Initialize system state
	prevFloor := -1
	//timerActive := false
	//var timerEndTime float64

	obstructionActive := false
	lastKnownDirection := elevio.MotorDirection(0)
	stop := false

	// Main event loop
	for {
		select {
		case buttonEvent := <-buttonPressCh:
			// Handle button press event

			if stop {
				elevio.SetMotorDirection(lastKnownDirection)
				stop = false
			}

			fmt.Printf("Button pressed at floor %d, button type %d\n", buttonEvent.Floor, buttonEvent.Button)
			fsmOnRequestButtonPress(buttonEvent.Floor, buttonEvent.Button)

		case currentFloor := <-floorSensorCh:
			// Handle floor sensor event

			if currentFloor != prevFloor {
				fmt.Printf("Arrived at floor %d\n", currentFloor)
				fsmOnFloorArrival(currentFloor)
				elevio.SetFloorIndicator(currentFloor) // Update floor indicator lamp

				if !obstructionActive {
					timerStop()
					timerStart(3.0)
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
				lastKnownDirection = elevator.Dirn
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
			if timerTimedOut() {
				fmt.Println("Door timeout occurred.")
				fsmOnDoorTimeout()
				timerStop() // Reset the timer after timeout handling
			}
		case obstruction := <-obstructionSwitchCh:
			if obstruction {
				obstructionActive = true
				timerStop()
				fmt.Println("obstruction switch")
			} else if !obstruction {
				obstructionActive = false
				timerStop()
				timerStart(3.0)
				fmt.Println("obstruction switch off")
			}
		}
	}
}

/*package main

import (
	"Driver-go/elevio"
	"fmt"
	"time"
)

type InputDevice struct{}

// RequestButton returnerer statusen til en knapp (trykket eller ikke)
func (d InputDevice) RequestButton(floor, button int) int {
	// Simuler knappestatus (returner 0 eller 1)
	// I en ekte implementasjon ville du lese fra maskinvare
	return 0 // Endre dette til faktisk logikk
}

// FloorSensor returnerer nåværende etasje som sensoren oppdager
func (d InputDevice) FloorSensor() int {
	// Simuler etasjeavlesning
	// Returner -1 hvis ingen etasje oppdages
	return -1 // Endre dette til faktisk logikk
}

// GetInputDevice returnerer en instans av InputDevice
func GetInputDevice() InputDevice {
	return InputDevice{}
}


func main() {


	fmt.Println("Started!")
	numFloors := 4

	initElevator(elevio.NumFloors, elevio.NumButtonTypes)

	elevio.Init("localhost:15657", numFloors)

	// Initialiser polling-rate
	inputPollRateMs := 25
	loadConfig("elevator.con", &inputPollRateMs)

	// Hent input-enhet
	input := GetInputDevice()

	// Sjekk initialisering mellom etasjer
	if input.FloorSensor() == -1 {
		fsmOnInitBetweenFloors()
	}

	// Statisk lagring for tidligere tilstander
	prevRequestButtons := make([][]int, elevio.NumFloors)
	for i := range prevRequestButtons {
		prevRequestButtons[i] = make([]int, elevio.NumButtonTypes)
	}
	prevFloor := -1

	// Kanaler for håndtering av events
	buttonPressCh := make(chan struct{})
	floorSensorCh := make(chan struct{})
	timerCh := make(chan struct{})

	// Start gorutiner for eventhåndtering
	go func() {
		for {
			// Knappetrykk
			for f := 0; f < elevio.NumFloors; f++ {
				for b := 0; b < elevio.NumButtonTypes; b++ {
					v := input.RequestButton(f, b)
					if v != 0 && v != prevRequestButtons[f][b] {
						buttonPressCh <- struct{}{}
						fsmOnRequestButtonPress(f, elevio.ButtonType(b))
					}
					prevRequestButtons[f][b] = v
				}
			}
			time.Sleep(time.Duration(inputPollRateMs) * time.Millisecond)
		}
	}()

	go func() {
		for {
			// Etasjesensor
			currentFloor := input.FloorSensor()
			if currentFloor != -1 && currentFloor != prevFloor {
				floorSensorCh <- struct{}{}
				fsmOnFloorArrival(currentFloor)
			}
			prevFloor = currentFloor
			time.Sleep(time.Duration(inputPollRateMs) * time.Millisecond)
		}
	}()

	go func() {
		for {
			// Timer
			if timerTimedOut() {
				timerCh <- struct{}{}
				timerStop()
				fsmOnDoorTimeout()
			}
			time.Sleep(time.Duration(inputPollRateMs) * time.Millisecond)
		}
	}()

	// Hovedløkke med select-case
	for {
		select {
		case <-buttonPressCh:
			// Knappetrykk håndteres i gorutinen
		case <-floorSensorCh:
			// Etasjehåndtering skjer i gorutinen
		case <-timerCh:
			// Timerhåndtering skjer i gorutinen
		}
	}
}

// Ekstra hjelpefunksjoner

// loadConfig simulerer lasting av konfigurasjonsverdier
func loadConfig(filename string, inputPollRateMs *int) {
	// Simuler lasting fra en fil
	fmt.Printf("Loading configuration from %s\n", filename)
	*inputPollRateMs = 25 // Standardverdi
}


/*func main() {

	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	var d elevio.MotorDirection = elevio.MD_Up
	elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go timerActivate(5)

	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			elevio.SetButtonLamp(a.Button, a.Floor, true)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			if a == numFloors-1 {
				d = elevio.MD_Down
			} else if a == 0 {
				d = elevio.MD_Up
			}
			elevio.SetMotorDirection(d)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(d)
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		}
	}
}
*/
