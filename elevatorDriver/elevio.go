package elevatorDriver

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

	go pollButtons(ch_hardware.Ch_buttonPress)
	go pollFloorSensor(ch_hardware.Ch_floorSensor)
	go pollStopButton(ch_hardware.Ch_stopButton)
	go pollObstructionSwitch(ch_hardware.Ch_obstruction)
}

func InitElevator(numFloors int, numButtonTypes int, elev Elevator, id string) Elevator {
	if numFloors > NUMFLOORS || numButtonTypes > NUMBUTTONTYPE {
		fmt.Println("Error: Configuration exceeds allowed array size.")
		return Elevator{}
	}

	elev = Elevator{
		ID:        id,
		Floor:     -1,
		Dirn:      MD_Stop,
		Behaviour: ElevatorBehaviour(EB_Idle),
		Available: true,
		Config: struct {
			DoorOpenDuration    float64
			TimeBetweenFloors   float64
		}{
			DoorOpenDuration:   3.0,
			TimeBetweenFloors:  2.0,
		},
	}

	if initialFloor := getFloor(); initialFloor == -1 {
		fmt.Println("Elevator is between floors on startup. Running initialization...")
		elev.Behaviour, elev.Dirn = initBetweenFloors()
	}
	elev.Floor = getFloor()

	fmt.Println("Elevator initialized:")
	return elev
}

func initBetweenFloors() (ElevatorBehaviour, MotorDirection) {
	for {
		setMotorDirection(MD_Down)
		if getFloor() != -1 {
			break
		}
	}
	dirn := MD_Stop
	setMotorDirection(MD_Stop)
	behaviour := ElevatorBehaviour(EB_Idle)
	return behaviour, dirn
}

func setMotorDirection(dir MotorDirection) {
	write([4]byte{1, byte(dir), 0, 0})
}

func SetButtonLamp(button ButtonType, floor int, value bool) {
	write([4]byte{2, byte(button), byte(floor), toByte(value)})
}

func setFloorIndicator(floor int) {
	write([4]byte{3, byte(floor), 0, 0})
}

func setDoorOpenLamp(value bool) {
	write([4]byte{4, toByte(value), 0, 0})
}

func setStopLamp(value bool) {
	write([4]byte{5, toByte(value), 0, 0})
}

func pollButtons(receiver chan<- ButtonEvent) {
	prev := make([][3]bool, NUMFLOORS)
	for {
		time.Sleep(_pollRate)
		for f := 0; f < NUMFLOORS; f++ {
			for b := ButtonType(0); b < 3; b++ {
				v := getButton(b, f)
				if v != prev[f][b] && v {
					receiver <- ButtonEvent{Floor: f, Button: ButtonType(b)}
				}
				prev[f][b] = v
			}
		}
	}
}

func pollFloorSensor(receiver chan<- int) {
	prev := -1
	for {
		time.Sleep(_pollRate)
		v := getFloor()
		if v != prev && v != -1 {
			receiver <- v
		}
		prev = v
	}
}

func pollStopButton(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := getStop()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func pollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := getObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func getButton(button ButtonType, floor int) bool {
	a := read([4]byte{6, byte(button), byte(floor), 0})
	return toBool(a[1])
}

func getFloor() int {
	a := read([4]byte{7, 0, 0, 0})
	if a[1] != 0 {
		return int(a[2])
	} else {
		return -1
	}
}

func getStop() bool {
	a := read([4]byte{8, 0, 0, 0})
	return toBool(a[1])
}

func getObstruction() bool {
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
