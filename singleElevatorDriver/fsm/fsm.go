package fsm

import (
	"Driver-go/singleElevatorDriver/elevio"
	"Driver-go/singleElevatorDriver/requests"
	. "Driver-go/types"
	"fmt"
	"time"
)

func Fsm(Ch_floorSensor <-chan int,
	Ch_stopButton <-chan bool,
	Ch_obstruction <-chan bool,
	Ch_localOrders <-chan OrderMatrix,
	Ch_clearedFloor chan<- DirnFloorPair,
	Ch_stateUpdate chan<- Elevator,
	elev Elevator) {

	fmt.Println("FSM Started!")
	orderMatrix := OrderMatrix{}
	obstructionActive := false
	doorOpenCh := make(chan bool, 200)
	doorClose := time.NewTimer(3 * time.Second)
	doorClose.Stop()
	errorTimeout := time.NewTimer(5 * time.Second)
	periodicStateUpdate := time.NewTicker(1 * time.Second)
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
				if requests.RequestsHere(orderMatrix, elev.Floor) && lastDirn == MD_Stop {
					doorOpenCh <- true
					break
				}
				errorTimeout.Reset(5 * time.Second)
			}

		case currentFloor := <-Ch_floorSensor:
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
					errorTimeout.Stop()
					break
				}
				errorTimeout.Reset(5 * time.Second)
			case EB_DoorOpen:
				elevio.SetMotorDirection(MD_Stop)
				errorTimeout.Stop()
			case EB_Idle:
				elevio.SetMotorDirection(MD_Stop)
				errorTimeout.Stop()
			}
			elev.Available = true
			Ch_stateUpdate <- elev

		case <-doorOpenCh:
			elev.Behaviour = EB_DoorOpen
			elev.Dirn = MD_Stop
			elevio.SetMotorDirection(MD_Stop)
			elevio.SetDoorOpenLamp(true)
			doorClose.Reset(3 * time.Second)
			errorTimeout.Stop()
			orderMatrix = requests.RequestsClearAtCurrentFloor(orderMatrix, elev, lastDirn)
			Ch_clearedFloor <- DirnFloorPair{Dirn: lastDirn, Floor: elev.Floor}

		case <-doorClose.C:
			if obstructionActive {
				doorClose.Reset(3 * time.Second)
				break
			}
			elevio.SetDoorOpenLamp(false)
			doorClose.Stop()
			elev.Behaviour = EB_Idle
			if emptyOrderMatrix(orderMatrix) {
				elev.Behaviour = EB_Idle
				lastDirn = elev.Dirn
				errorTimeout.Stop()
				break
			} else {
				dirnBehaviour := requests.RequestsChooseDirection(orderMatrix, elev, lastDirn)
				elev.Dirn = dirnBehaviour.Dirn
				lastDirn = elev.Dirn
				elev.Behaviour = ElevatorBehaviour(dirnBehaviour.Behaviour)
				elevio.SetMotorDirection(elev.Dirn)
				errorTimeout.Reset(5 * time.Second)
			}

		case stopPressed := <-Ch_stopButton:
			if stopPressed {
				lastDirn = elev.Dirn
				fmt.Println("Stop button pressed!")
				fmt.Println(lastDirn)
				elevio.SetStopLamp(true)
				elevio.SetMotorDirection(MD_Stop)
			} else {
				fmt.Println("Stop button released!")
				elevio.SetStopLamp(false)
				elevio.SetMotorDirection(lastDirn)
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

		case <-periodicStateUpdate.C:
			periodicStateUpdate.Stop()
			Ch_stateUpdate <- elev
			periodicStateUpdate.Reset(1 * time.Second)


		case <-errorTimeout.C:
			fmt.Println("Error timeout!Elevator behav: ", elev.Behaviour, "elevID: ", elev.ID)


		default:
			if (elev.Behaviour == EB_Idle || elev.Behaviour == EB_Moving) && elev.Available {
				newDirPair := requests.RequestsChooseDirection(orderMatrix, elev, elev.Dirn)
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
					lastDirn = elev.Dirn
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
				return false
			}
		}
	}
	return true
}
