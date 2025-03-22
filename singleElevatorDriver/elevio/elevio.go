package elevio

import (
	. "Driver-go/types"
	"fmt"
	"net"
	"sync"
	"time"
)

const _pollRate = 20 * time.Millisecond

var _initialized bool = false
var _mtx sync.Mutex
var _conn net.Conn

type HardwareChannels struct {
	Ch_buttonPress chan ButtonEvent
	Ch_floorSensor chan int
	Ch_stopButton  chan bool
	Ch_obstruction chan bool
}

func InitHardwareConnection(addr string, ch_hardware HardwareChannels) {
	if _initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	_mtx = sync.Mutex{}
	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	_initialized = true

	go PollButtons(ch_hardware.Ch_buttonPress)
	go PollFloorSensor(ch_hardware.Ch_floorSensor)
	go PollStopButton(ch_hardware.Ch_stopButton)
	go PollObstructionSwitch(ch_hardware.Ch_obstruction)
}

func ElevatorUninitialized() Elevator {
	return Elevator{
		Floor:     -1,
		Dirn:      MD_Stop,
		Behaviour: ElevatorBehaviour(EB_Idle),
		Config: struct {
			ClearRequestVariant ClearRequestVariant
			DoorOpenDuration    float64
			TimeBetweenFloors   float64
		}{
			ClearRequestVariant: CV_All,
			DoorOpenDuration:    3.0,
			TimeBetweenFloors:   2.0,
		},
	}
}

func InitElevator(numFloors int, numButtonTypes int, elev Elevator, id string) Elevator {
	if numFloors > NUMFLOORS || numButtonTypes > NUMBUTTONTYPE {
		fmt.Println("Error: Configuration exceeds allowed array size.")
		return Elevator{}
	}

	elev = Elevator{
		// Start on an invalid floor
		ID:        id,
		Floor:     -1,
		Dirn:      MD_Stop,
		Behaviour: ElevatorBehaviour(EB_Idle),
		Available: true,
		Config: struct {
			ClearRequestVariant ClearRequestVariant
			DoorOpenDuration    float64
			TimeBetweenFloors   float64
		}{
			ClearRequestVariant: CV_InDirn,
			DoorOpenDuration:    3.0,
			TimeBetweenFloors:   2.0,
		},
	}

	if initialFloor := GetFloor(); initialFloor == -1 {
		fmt.Println("Elevator is between floors on startup. Running initialization...")
		elev.Behaviour, elev.Dirn = initBetweenFloors()
	}
	elev.Floor = GetFloor()

	fmt.Println("Elevator initialized:")
	return elev
}

func initBetweenFloors() (ElevatorBehaviour, MotorDirection) {
	// Move the elevator down until it reaches a floor

	for {
		SetMotorDirection(MD_Down)
		if GetFloor() != -1 {
			break
		}
	}
	// Update the elevator's state
	dirn := MD_Stop
	SetMotorDirection(MD_Stop)
	behaviour := ElevatorBehaviour(EB_Idle)
	return behaviour, dirn
}

func SetMotorDirection(dir MotorDirection) {
	write([4]byte{1, byte(dir), 0, 0})
}

func SetButtonLamp(button ButtonType, floor int, value bool) {
	write([4]byte{2, byte(button), byte(floor), toByte(value)})
}

func SetFloorIndicator(floor int) {
	write([4]byte{3, byte(floor), 0, 0})
}

func SetDoorOpenLamp(value bool) {
	write([4]byte{4, toByte(value), 0, 0})
}

func SetStopLamp(value bool) {
	write([4]byte{5, toByte(value), 0, 0})
}

func PollButtons(receiver chan<- ButtonEvent) {

	prev := make([][3]bool, NUMFLOORS)
	for {
		time.Sleep(_pollRate)
		for f := 0; f < NUMFLOORS; f++ {
			for b := ButtonType(0); b < 3; b++ {
				v := GetButton(b, f)
				if v != prev[f][b] && v {
					receiver <- ButtonEvent{Floor: f, Button: ButtonType(b)}
					// print("Button pressed: ")
				}
				prev[f][b] = v
			}
		}
	}
}

func PollFloorSensor(receiver chan<- int) {
	prev := -1
	for {
		time.Sleep(_pollRate)
		v := GetFloor()
		if v != prev && v != -1 {
			receiver <- v
		}
		prev = v
	}
}

func PollStopButton(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := GetStop()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func PollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := GetObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func GetButton(button ButtonType, floor int) bool {
	a := read([4]byte{6, byte(button), byte(floor), 0})
	return toBool(a[1])
}

func GetFloor() int {
	a := read([4]byte{7, 0, 0, 0})
	if a[1] != 0 {
		return int(a[2])
	} else {
		return -1
	}
}

func GetStop() bool {
	a := read([4]byte{8, 0, 0, 0})
	return toBool(a[1])
}

func GetObstruction() bool {
	a := read([4]byte{9, 0, 0, 0})
	return toBool(a[1])
}

func read(in [4]byte) [4]byte {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	var out [4]byte
	_, err = _conn.Read(out[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	return out
}

func write(in [4]byte) {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
}

func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}
