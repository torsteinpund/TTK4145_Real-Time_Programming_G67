package fsm

// import (
//     "fmt"
//     "time"
// 	"Driver-go/elevio"
// 	"Driver-go/lights"
// 	"Driver-go/singleElevatorDriver/requests"
// 	. "Driver-go/types"

// )


// type DoorChannels struct {
// 	Ch_doorOpen chan bool
// 	Ch_toSlave  chan NetworkMessage
// }



// func doorOpen(doorChans DoorChannels) {
//     for {
//         select {
//         case <-doorChans.Ch_doorOpen:
//             fmt.Println("Door open signal received. Starting 3-second timer.")
//             timer := time.NewTimer(3 * time.Second)
// 			fmt.Println("Door open event")
// 			elevio.SetMotorDirection(MD_Stop)
// 			elevio.SetDoorOpenLamp(true)
// 			orders := NetworkMessage{MsgType: "orderupdatechannel", MsgData: globOrderMap, Receipient: All}
//             for {
//                 select {
//                 case <-timer.C:
//                     fmt.Println("3 seconds have passed. Door closing.")
//                     // Send an update to the master/slave system

//                     doorChans.Ch_toSlave <- orders{
//                         // Populate with relevant data
//                     }
//                     return

//                 case ordersFromMaster := <-doorChans.Ch_toSlave:
//                     // Process orders from the master
//                     fmt.Println("Received order from master:", ordersFromMaster)
// 					orders = ordersFromMaster.MsgData.(GlobalOrderMap)
//                     // Optionally handle the order here
                
// 				default:
// 					time.Sleep(10 * time.Millisecond)
// 				}
//             }

//         default:
//             // Optional: Add logic for other cases if needed
//             time.Sleep(10 * time.Millisecond) // Prevent busy-waiting
//         }
//     }
// }