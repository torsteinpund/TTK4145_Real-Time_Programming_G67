package elevatorDriver

import (
	. "Driver-go/types"
	"fmt"
	"time"
)

func Fsm(Ch_floorSensor  <-chan int,
		 Ch_stopButton   <-chan bool,
		 Ch_obstruction  <-chan bool,
		 Ch_localOrders  <-chan LocalOrder,
		 Ch_clearedFloor chan<- DirnFloorPair,
		 Ch_stateUpdate  chan<- Elevator,
		 elev Elevator) {

	fmt.Println("FSM Started!")

	orderMatrix 		 := OrderMatrix{}
	obstructionActive 	 := false
	lastDirn 			 := elev.Dirn
	networkConnected 	 := false
	ch_doorOpen 		 := make(chan bool, 200)
	doorClose 			 := time.NewTimer(3 * time.Second)
	errorTimeout 		 := time.NewTimer(5 * time.Second)
	periodicStateUpdate  := time.NewTimer(1 * time.Second)
	doorClose.Stop()
	errorTimeout.Stop()
	setDoorOpenLamp(false)

	for {
		select {
		case receivedOrder := <-Ch_localOrders:
			orderMatrix = receivedOrder.OrderMatrix
			networkConnected = receivedOrder.NetworkConnection
			switch elev.Behaviour {
			case EB_Idle:
				if ordersHere(orderMatrix, elev.Floor) {
					ch_doorOpen <- true
					break
				}
				if !emptyOrderMatrix(orderMatrix){
					errorTimeout.Reset(5 * time.Second)
				}


			case EB_Moving:
				break

			case EB_DoorOpen:
				if ordersHere(orderMatrix, elev.Floor) && lastDirn == MD_Stop {
					ch_doorOpen <- true
					break
				}
				errorTimeout.Reset(5 * time.Second)
			}

		case currentFloor := <-Ch_floorSensor:
			elev.Floor = currentFloor
			setFloorIndicator(elev.Floor)

			switch elev.Behaviour {
			case EB_Moving:
				if shouldStop(orderMatrix, elev) {
					lastDirn = elev.Dirn
					elev.Behaviour = EB_DoorOpen
					elev.Dirn = MD_Stop
					ch_doorOpen <- true
					break
				}

				if emptyOrderMatrix(orderMatrix) {
					elev.Behaviour = EB_Idle
					setMotorDirection(MD_Stop)
					errorTimeout.Stop()
					break
				}
				errorTimeout.Reset(5 * time.Second)

			case EB_DoorOpen:
				setMotorDirection(MD_Stop)
				errorTimeout.Stop()

			case EB_Idle:
				setMotorDirection(MD_Stop)
				errorTimeout.Stop()
			}
		
			elev.Available = true
			if networkConnected {
				Ch_stateUpdate <- elev
			}

		case <-ch_doorOpen:
			elev.Behaviour = EB_DoorOpen
			elev.Dirn = MD_Stop
			setMotorDirection(MD_Stop)
			setDoorOpenLamp(true)
			doorClose.Reset(3 * time.Second)
			errorTimeout.Reset(5 * time.Second)
			orderMatrix, lastDirn = clearOrderAtCurrentFloor(orderMatrix, elev.Floor, lastDirn)
			Ch_clearedFloor <- DirnFloorPair{Dirn: lastDirn, Floor: elev.Floor}

		case <-doorClose.C:
			if obstructionActive {
				doorClose.Reset(3 * time.Second)
				errorTimeout.Reset(5 * time.Second)
				break
			}

			setDoorOpenLamp(false)
			doorClose.Stop()
			elev.Behaviour = EB_Idle

			if emptyOrderMatrix(orderMatrix) {
				elev.Behaviour = EB_Idle
				lastDirn = elev.Dirn
				errorTimeout.Stop()
			} else {
				dirnBehaviour := chooseDirection(orderMatrix, elev.Floor, lastDirn)
				elev.Dirn = dirnBehaviour.Dirn
				lastDirn = elev.Dirn
				elev.Behaviour = ElevatorBehaviour(dirnBehaviour.Behaviour)
				setMotorDirection(elev.Dirn)
				errorTimeout.Reset(5 * time.Second)
			}

		case stopPressed := <-Ch_stopButton:
			if stopPressed {
				lastDirn = elev.Dirn
				fmt.Println(lastDirn)
				setStopLamp(true)
				setMotorDirection(MD_Stop)
			} else {
				setStopLamp(false)
				setMotorDirection(lastDirn)
			}

		case obstruction := <-Ch_obstruction:
			if obstruction {
				obstructionActive = true
				elev.Available = false
			} else {
				obstructionActive = false
				elev.Available = true
			}

		case <-periodicStateUpdate.C:
			if networkConnected {
				Ch_stateUpdate <- elev
			}
			periodicStateUpdate.Reset(1 * time.Second)

		case <-errorTimeout.C:
			errorTimeout.Stop()
			fmt.Println("Error timeout!")
			setDoorOpenLamp(false)
			elev.Available = false

			if networkConnected {
				Ch_stateUpdate <- elev
			}

			elev.Floor, elev.Behaviour = initAfterErrorTimeout(elev.Dirn, Ch_localOrders, orderMatrix, elev.Behaviour)
			elev.Available = true

			if(orderMatrix[elev.Floor][BT_Cab]){
				ch_doorOpen<-true
			}

			if networkConnected {
				Ch_stateUpdate <- elev
			}

		default:
			if (elev.Behaviour == EB_Idle || elev.Behaviour == EB_Moving) && elev.Available {
				newDirnPair := chooseDirection(orderMatrix, elev.Floor, elev.Dirn)
				if newDirnPair.Dirn != elev.Dirn {
					switch newDirnPair.Dirn {
					case MD_Stop:
						elev.Behaviour = EB_Idle
						if getFloor() == -1 {
							break
						}
						if ordersHere(orderMatrix, elev.Floor) {
							ch_doorOpen <- true
							break
						}
						setMotorDirection(MD_Stop)

					case MD_Up:
						elev.Behaviour = EB_Moving
						setMotorDirection(MD_Up)

					case MD_Down:
						elev.Behaviour = EB_Moving
						setMotorDirection(MD_Down)
					}

					lastDirn = elev.Dirn
					elev.Dirn = newDirnPair.Dirn
				}
			}
		}
	}
}


