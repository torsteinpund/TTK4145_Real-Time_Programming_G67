package network

import (
	"Driver-go/network/bcast"
	"Driver-go/network/peers"
	"fmt"
	"sort"
	"strconv"
	"time"
)

type masterHeartBeat struct {
	MasterID string `json:"masterID"`
	Timestamp int64 `json:"timestamp"`
}


func PeerHandler(id string,
				 Ch_RX RXChannels,
				 Ch_peerUpdate <-chan peers.PeersUpdate,
				 Ch_peerLost   chan<- string,
				 Ch_newPeer    chan<- string,
				 Ch_isMaster   chan<- bool) {

	fmt.Println("PeerHandler Started!")

	timeout 		:= time.After(250 * time.Millisecond)
	activePeers 	:= make(map[string]peers.Peer)
	ch_currentStop  := make(chan struct{})
	masterUpdate 	:= make(chan string)
	masterHbPort 	:= 14343
	currentMasterID := ""


	aloneOnNetworkCheck:
	for {
		select {
		case update := <-Ch_peerUpdate:
			for _, peerID := range update.PeersID {
				activePeers[peerID] = peers.Peer{ID: peerID}
			}
		case <-timeout:
			break aloneOnNetworkCheck
		}
	}

	if len(activePeers) == 1{
		currentMasterID = id
		Ch_isMaster <- true
		Ch_newPeer <- id
		ch_currentStop = make(chan struct{})
		go startMasterHeartbeat(id, masterHbPort, ch_currentStop)
		
	}else{
		Ch_isMaster <- false
	}

	go startMasterReceiver(masterHbPort, masterUpdate)
	
	for {
		select {
		case update := <- Ch_peerUpdate:
			activePeers, peerstatus, peerID := updatePeers(activePeers, update, id)
			switch peerstatus {
			case "lostPeer":
				fmt.Println("Peer lost: ", peerID)
				if peerID == currentMasterID {
					currentMasterID = updateMaster(activePeers)
					if currentMasterID == id {
						Ch_isMaster <- true
						Ch_peerLost <-peerID
						ch_currentStop = restartHeartbeat(id, masterHbPort, ch_currentStop)

					} else {
						Ch_isMaster <- false
					}

				} else { 
					if currentMasterID == id{
						Ch_isMaster <- true
						Ch_peerLost <- peerID
						ch_currentStop = restartHeartbeat(id, masterHbPort, ch_currentStop)
					}
				}

			case "newPeer":
				if currentMasterID == id {
					Ch_isMaster <- true
					Ch_newPeer <- peerID
				} else {
					Ch_isMaster <- false
				}

			case "backOnNet":
				timeout := time.After(1 * time.Second)
				gettingBackOnNet:
					for {
						select {
						case update := <-Ch_peerUpdate:
							for _, peerID := range update.PeersID {
								activePeers[peerID] = peers.Peer{ID: peerID}
							}
						case masterID := <-masterUpdate:
							currentMasterID = masterID

						case <-timeout:
							break gettingBackOnNet
						}
					}

				if currentMasterID == id {
					Ch_isMaster <- true
				} else {
					Ch_isMaster <- false
				}
			}

		case masterID := <-masterUpdate:
			
			if masterID == "" {
				currentMasterID = updateMaster(activePeers)
			} else if masterID != currentMasterID {
				fmt.Println("Updated master, the new master is ", masterID)
				currentMasterID = masterID

				if currentMasterID == id {
					Ch_isMaster <- true

					if ch_currentStop != nil{
						close(ch_currentStop)
					}
					
					ch_currentStop = make(chan struct{})
					go startMasterHeartbeat(id, masterHbPort, ch_currentStop)

				} else {
					Ch_isMaster <- false

					if ch_currentStop != nil {
						close(ch_currentStop)
						ch_currentStop = nil
					}
				}
			} else {
				currentMasterID = masterID
			}
		}
	}
}


func updatePeers(activePeers map[string]peers.Peer,update peers.PeersUpdate, ownID string) (map[string]peers.Peer, string, string) {
	changedAllPeers := ""
	peerID := ""
	
	if update.New != "" {
			activePeers[update.New] = peers.Peer{ID: update.New}
			if update.New != ownID {
				changedAllPeers, peerID = "newPeer", update.New
			}else{
				changedAllPeers, peerID = "backOnNet", update.New
			}
		}

	for _, lostID := range update.Lost {
		if lostID != ownID {
			delete(activePeers, lostID)
			changedAllPeers, peerID = "lostPeer", lostID
		}
	}

	return activePeers, changedAllPeers, peerID

}


func updateMaster(activePeers map[string]peers.Peer) string {
	if len(activePeers) == 0{
		return ""
	}

	peers := []int{}
	for _, peer := range activePeers {
		int_ID, err := strconv.Atoi(peer.ID)
		if err != nil {
			fmt.Println("Could not convert ID: ", peer.ID, " to int")
			return ""
		}
		peers = append(peers, int_ID)
	}

	sort.Ints(peers)
	currentMasterID := strconv.Itoa(peers[0])
	return currentMasterID
}


func restartHeartbeat(id string, masterHbPort int, currentStop chan struct{}) chan struct{} {
    if currentStop != nil {
        close(currentStop)
    }
    newStop := make(chan struct{})
    go startMasterHeartbeat(id, masterHbPort, newStop)
    return newStop
}


func startMasterHeartbeat(masterId string, port int, ch_stop <-chan struct{}) {
    ch_heartbeatTX := make(chan masterHeartBeat)
    go bcast.Transmitter(port, ch_heartbeatTX)
    ticker := time.NewTicker(500 * time.Millisecond) 
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            heartbeat := masterHeartBeat{
                MasterID:  masterId,
                Timestamp: time.Now().UnixNano(),
            }
            ch_heartbeatTX <- heartbeat

		case <-ch_stop:
			return
        }
    }
}


func startMasterReceiver(port int, ch_masterUpdate chan<- string) {
    ch_heartbeatRX := make(chan masterHeartBeat)
    go bcast.Receiver(port, ch_heartbeatRX)
    timeoutDuration := 1 * time.Second
    lastHeartbeat := time.Now()

    for {
        select {
        case heartbeat := <-ch_heartbeatRX:
            lastHeartbeat = time.Now()
            ch_masterUpdate <- heartbeat.MasterID

        default:
            if time.Since(lastHeartbeat) > timeoutDuration {
                ch_masterUpdate <- ""
                lastHeartbeat = time.Now()
            }
            time.Sleep(50 * time.Millisecond)
        }
    }
}
