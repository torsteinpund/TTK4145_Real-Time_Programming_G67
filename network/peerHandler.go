package network

import (
	"Driver-go/network/peers"
	"fmt"
	"sort"
	"strconv"
	"time"
	"Driver-go/network/bcast"

)

type PeerHandler struct {
	id          string
	activePeers map[string]peers.Peer
}

func NewPeerHandler(id string) *PeerHandler {
	return &PeerHandler{
		id:          id,
		activePeers: make(map[string]peers.Peer),
	}
}

func (pH *PeerHandler) PeerHandler(id string,
	Ch_RX RXChannels,
	Ch_peerUpdate <-chan peers.PeersUpdate,
	Ch_peerLost chan<- string,
	Ch_newPeer chan<- string,
	Ch_isMaster chan<- bool) {


	timeout := time.After(100 *time.Millisecond)
	existingPeers := make(map[string]peers.Peer)
	currentMasterID := ""
	masters := make(map[string]peers.Peer)


	aloneOnNetworkLoop:
	for {
		select {
		case update := <-Ch_peerUpdate:

			for _, peerID := range update.PeersID {
				existingPeers[peerID] = peers.Peer{ID: peerID}
			}
		case <-timeout:
			break aloneOnNetworkLoop
		}
	}

	
	masterUpdate := make(chan string)
	masterHbPort := 14343
	fmt.Println("Existing peers ", existingPeers)
	if len(existingPeers) == 1{
		fmt.Println("I am the first peer in this bitch")
		go StartMasterHeartbeat(id,masterHbPort)
		Ch_isMaster <- true
		Ch_newPeer <- id
		currentMasterID = id
		masters[currentMasterID] = peers.Peer{ID: currentMasterID}
		
	}else{
		fmt.Println("Found existing peers...")
		Ch_isMaster <- false
	}

	go StartMasterReceiver(masterHbPort, masterUpdate)
	pH.activePeers = existingPeers
	fmt.Println(pH.activePeers)
	for{
		select{
		case update := <- Ch_peerUpdate:
			peerstatus, peerID := pH.updatePeers(update, id)
			if peerstatus == "lostPeer" {
				if peerID == currentMasterID {
					currentMasterID = updateMaster(pH.activePeers)
					if currentMasterID == id {
						Ch_isMaster <- true
						Ch_peerLost <-peerID
						go StartMasterHeartbeat(id, masterHbPort)
						fmt.Println("I became the MAAAASTER!")
					} else {
						Ch_isMaster <- false
					}
					delete(pH.activePeers, peerID)
				}else{
					if currentMasterID == id{
						Ch_peerLost <- peerID
					}
					delete(pH.activePeers,peerID)
					fmt.Println("Peer lost:", peerID)
				}
				
			}else if peerstatus == 	"newPeer" {
				if currentMasterID == id {
					Ch_isMaster <- true
					Ch_newPeer <- peerID
				} else {
					Ch_isMaster <- false
				}
				pH.activePeers[peerID] = peers.Peer{ID: peerID}
			
				}
		case masterID := <-masterUpdate:
			if masterID == "" {
				currentMasterID = updateMaster(pH.activePeers)
			} else if masterID != currentMasterID {
				fmt.Println("Updated master, the new master is ", masterID)
				currentMasterID = masterID
				if len(masters) > 1 {
					fmt.Println("I am no longer the master")
					masters = make(map[string]peers.Peer)
					currentMasterID = updateMaster(pH.activePeers)
					fmt.Println("The new master is ", currentMasterID)
					masters[currentMasterID] = peers.Peer{ID: currentMasterID}
				}
			}
		}
	}
}


func (pH *PeerHandler) updatePeers(update peers.PeersUpdate, ownID string) (string, string) {
	// Updates activePeers
	fmt.Println("Updating peers")
	changedAllPeers := ""
	peerID := ""
	if update.New != "" {
		pH.activePeers[update.New] = peers.Peer{ID: update.New}
		if update.New != ownID {
			changedAllPeers, peerID = "newPeer", update.New
		}
	}
	// Removes lost peers
	for _, lostID := range update.Lost {
		if lostID != ownID {
			delete(pH.activePeers, lostID)
			fmt.Println("Removed peer: ", lostID)
			changedAllPeers, peerID = "lostPeer", lostID
		}
	}
	return changedAllPeers, peerID

}


func updateMaster(activePeers map[string]peers.Peer) string {
	if len(activePeers) == 0{
		return ""
	}
	peers := []int{}
	for _, peer := range activePeers {
		// parts := strings.Split(peer.ID, ".")
		// if len(parts) == 0 {
		// 	fmt.Println("Invalid peer ID: ", peer.ID)
		// 	return ""
		// }
		// // Use the last part of the IP adress (after last ".")
		// lastPart := parts[len(parts)-1]
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



type MasterHeartBeat struct {
	MasterID string `json:"masterID"`
	Timestamp int64 `json:"timestamp"`
}


func StartMasterHeartbeat(masterID string, port int) {
    ch_heartbeat := make(chan MasterHeartBeat)

 
    go bcast.Transmitter(port, ch_heartbeat)

    ticker := time.NewTicker(500 * time.Millisecond) 
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            heartbeat := MasterHeartBeat{
                MasterID:  masterID,
                Timestamp: time.Now().UnixNano(),
            }
  
            ch_heartbeat <- heartbeat

        }
    }
}

func StartMasterReceiver(port int, ch_masterUpdate chan<- string) {
    ch_heartbeat := make(chan MasterHeartBeat)
    go bcast.Receiver(port, ch_heartbeat)

    // Sjekk for heartbeats kontinuerlig
    timeoutDuration := 1 * time.Second
    lastHeartbeat := time.Now()

    for {
        select {
        case heartbeat := <-ch_heartbeat:
            // Oppdater siste heartbeat tid
			// fmt.Println("MasterIdMottatt: ",heartbeat.MasterID)
            lastHeartbeat = time.Now()
            ch_masterUpdate <- heartbeat.MasterID

        default:
            if time.Since(lastHeartbeat) > timeoutDuration {
                fmt.Println("Ingen master heartbeat mottatt innenfor timeout - master antas nede!")
                ch_masterUpdate <- ""
                lastHeartbeat = time.Now()
            }
            time.Sleep(50 * time.Millisecond)
        }
    }
}