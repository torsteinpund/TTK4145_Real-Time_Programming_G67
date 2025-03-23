package fsm

import (
	"Driver-go/singleElevatorDriver/elevio"
	"Driver-go/singleElevatorDriver/requests"
	. "Driver-go/types"
	"fmt"
	"time"
)

// type FsmChannels struct {
// 	Ch_floorSensor     chan int
// 	Ch_stopButton      chan bool
// 	Ch_obstruction     chan bool
// 	Ch_localOrders     chan OrderMatrix
// 	Ch_networkToMaster chan NetworkMessage
// 	Ch_clearedFloor    chan DirnFloorPair
// 	// Ch_stateUpdate     chan Elevator
// }




func FsmRun(Ch_floorSensor     <-chan 	int,
			Ch_stopButton      <-chan 	bool,
			Ch_obstruction     <-chan 	bool,
			Ch_localOrders     <-chan 	OrderMatrix,
			// Ch_networkToMaster chan<- 	NetworkMessage,
			Ch_clearedFloor    chan<- 	DirnFloorPair,
			Ch_stateUpdate     chan<- Elevator,
			elev Elevator){

	fmt.Println("FSM Started!")
	// inputPollRate := 25 * time.Millisecond // Adjust as needed
	// go HandleDoor(doorchannels, Ch_obstruction)
	orderMatrix := OrderMatrix{}
	//prevFloor := elev.Floor
	obstructionActive := false
	doorOpenCh := make(chan bool, 200)
	doorClose := time.NewTimer(3 * time.Second)
	doorClose.Stop()
	// errorTimeout := time.NewTimer(5 * time.Second)
	elevio.SetDoorOpenLamp(false)
	lastDirn := elev.Dirn



	for {
		select {
		case receivedOrder := <-Ch_localOrders:
			orderMatrix = receivedOrder
			switch elev.Behaviour {
			case EB_Idle:
				if requests.RequestsHere(orderMatrix, elev.Floor) {
					doorOpenCh <- true
					break
				}

			case EB_Moving:
				break

			case EB_DoorOpen:
				if requests.RequestsHere(orderMatrix, elev.Floor) {
					doorOpenCh <- true
					break
				}
				break
			}

		case currentFloor := <-Ch_floorSensor:
			// fmt.Println("Arrived at floor", currentFloor)
			elev.Floor = currentFloor
			elevio.SetFloorIndicator(elev.Floor)

			switch elev.Behaviour {
			case EB_Moving:
				if requests.RequestsShouldStop(orderMatrix, elev) {
					lastDirn = elev.Dirn
					elev.Behaviour = EB_DoorOpen
					elev.Dirn = MD_Stop
					doorOpenCh <- true
					break
				}

				if emptyOrderMatrix(orderMatrix) {
					elev.Behaviour = EB_Idle
					elevio.SetMotorDirection(MD_Stop)
					break
				}

			case EB_DoorOpen:
				elevio.SetMotorDirection(MD_Stop)

			case EB_Idle:
				elevio.SetMotorDirection(MD_Stop)

			}
			elev.Available = true
			// netMsg := NetworkMessage{MsgType: "elevatorupdatechannel", MsgData: elev}
			// Ch_networkToMaster <- netMsg
			Ch_stateUpdate <- elev

		case <-doorOpenCh:
			elev.Behaviour = EB_DoorOpen
			elev.Dirn = MD_Stop
			elevio.SetMotorDirection(MD_Stop)
			elevio.SetDoorOpenLamp(true)
			doorClose.Reset(3 * time.Second)
			fmt.Println("Order matrix before clearing at current floor:", orderMatrix)
			orderMatrix = requests.RequestsClearAtCurrentFloor(orderMatrix, elev, lastDirn)
			fmt.Println("Order matrix after clearing at current floor:", orderMatrix)
			Ch_clearedFloor <- DirnFloorPair{Dirn: lastDirn, Floor: elev.Floor}

		case <-doorClose.C:

			if obstructionActive {
				doorClose.Reset(3 * time.Second)
				break
			}
			elevio.SetDoorOpenLamp(false)
			doorClose.Stop()
			lastDirn = elev.Dirn
			elev.Behaviour = EB_Idle
			// fmt.Println("Door closed")
			if emptyOrderMatrix(orderMatrix) {
				elev.Behaviour = EB_Idle
				break
			} else {
				dirnBehaviour := requests.RequestsChooseDirection(orderMatrix, elev)
				elev.Dirn = dirnBehaviour.Dirn
				elev.Behaviour = ElevatorBehaviour(dirnBehaviour.Behaviour)
				elevio.SetMotorDirection(elev.Dirn)
			}
			// Ch_clearedFloor <- elev.Floor

		case stopPressed := <-Ch_stopButton:
			// Handle stop button event
			if stopPressed {
				lastDirn = elev.Dirn
				fmt.Println("Stop button pressed!")
				fmt.Println(lastDirn)
				elevio.SetStopLamp(true)
				elevio.SetMotorDirection(MD_Stop)
				// stop = true
			} else {
				fmt.Println("Stop button released!")
				elevio.SetStopLamp(false)
			}

		case obstruction := <-Ch_obstruction:
			fmt.Println("Obstruction detected")
			if obstruction {
				obstructionActive = true
				elev.Available = false
				fmt.Println("obstruction switch")
			} else if !obstruction {
				obstructionActive = false
				elev.Available = true
				fmt.Println("obstruction switch off")
			}

		default:
			if (elev.Behaviour == EB_Idle || elev.Behaviour == EB_Moving) && elev.Available {
				newDirPair := requests.RequestsChooseDirection(orderMatrix, elev)
				if newDirPair.Dirn != elev.Dirn {
					switch newDirPair.Dirn {
					case MD_Stop:
						elev.Behaviour = EB_Idle
						elevio.SetMotorDirection(MD_Stop)

					case MD_Up:
						elev.Behaviour = EB_Moving
						elevio.SetMotorDirection(MD_Up)

					case MD_Down:
						elev.Behaviour = EB_Moving
						elevio.SetMotorDirection(MD_Down)
					}
					elev.Dirn = newDirPair.Dirn

				}
			}

			time.Sleep(10 * time.Millisecond)

		}
	}

}

func emptyOrderMatrix(orderMatrix [NUMFLOORS][NUMBUTTONTYPE]bool) bool {
	for _, row := range orderMatrix {
		for _, value := range row {
			if value {
				return false // Found a `true`, so not all are `false`
			}
		}
	}
	return true // All values are `false`
}
