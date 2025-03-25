package network

import (
	"Driver-go/network/peers"
	"fmt"
	"sort"
	"strconv"
	"strings"
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
	ch_RX RXChannels,
	Ch_peerUpdate <-chan peers.PeersUpdate,
	Ch_peerLost chan<- string,
	Ch_newPeer chan<- string,
	Ch_isMaster chan<- bool) {


	timeout := time.After(1000 *time.Millisecond)
	existingPeers := make(map[string]peers.Peer)
	currentMasterID := ""



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
	if len(existingPeers) == 1{
		fmt.Println("I am the first peer in this bitch")
		go StartMasterHeartbeat(id,14343)
		Ch_isMaster <- true
		Ch_newPeer <- id
		currentMasterID = id
		
	}else{
		fmt.Println("Fant eksisterende peers")
		Ch_isMaster <- false
		Ch_newPeer <- id
	}

	go StartMasterReceiver(14343, masterUpdate)
	pH.activePeers = existingPeers
	fmt.Println(pH.activePeers)
	for{
		select{
		case update := <- Ch_peerUpdate:
			peerstatus, peerID := pH.updatePeers(update)
			if peerstatus == "lostPeer" {
				if peerID == currentMasterID {
					currentMasterID = updateMaster(pH.activePeers)
					if currentMasterID == id {
						Ch_isMaster <- true
						go StartMasterHeartbeat(id, 14343)
						fmt.Println("Jeg ble ny master!")
					} else {
						Ch_isMaster <- false
					}
					delete(pH.activePeers, peerID)
					Ch_peerLost <-peerID
				}else{
					delete(pH.activePeers,peerID)
					Ch_peerLost <- peerID
					fmt.Println("Peer mistet:", peerID)
				}
				
			}else if peerstatus == 	"newPeer" {
				if currentMasterID != id{
					Ch_isMaster <- false
				}
				pH.activePeers[peerID] = peers.Peer{ID: peerID}
				Ch_newPeer <- peerID
				}
		case masterID := <-masterUpdate:
			if masterID == "" {
				currentMasterID = updateMaster(pH.activePeers)
			} else if masterID != currentMasterID {
				fmt.Println("Oppdatert master status: nå er master", masterID)
				currentMasterID = masterID
			}
		}
	}
}


func (pH *PeerHandler) updatePeers(update peers.PeersUpdate) (string, string) {
	// Updates activePeers
	changedAllPeers := ""
	peerID := ""
	if update.New != "" {
		pH.activePeers[update.New] = peers.Peer{ID: update.New}
		changedAllPeers, peerID = "newPeer", update.New
	}
	// Removes lost peers
	for _, lostID := range update.Lost {

		delete(pH.activePeers, lostID)
		fmt.Println("Fjernet tapt peer:", lostID)
		changedAllPeers, peerID = "lostPeer", lostID
	}
	return changedAllPeers, peerID

}


func updateMaster(activePeers map[string]peers.Peer) string {

	peers := []int{}
	for _, peer := range activePeers {
		parts := strings.Split(peer.ID, ".")
		if len(parts) == 0 {
			fmt.Println("Invalid peer ID:", peer.ID)
			return ""
		}
		// Use the last part of the IP adress (after last ".")
		lastPart := parts[len(parts)-1]
		int_ID, err := strconv.Atoi(lastPart)
		if err != nil {
			fmt.Println("Could not convert ID to int")
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
    // Opprett en kanal for heartbeat meldinger
    hbCh := make(chan MasterHeartBeat)

    // Start transmitteren på den dedikerte porten
    go bcast.Transmitter(port, hbCh)

    ticker := time.NewTicker(500 * time.Millisecond) // Heartbeat hvert 500ms
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            hb := MasterHeartBeat{
                MasterID:  masterID,
                Timestamp: time.Now().UnixNano(),
            }
            // Send heartbeat-meldingen til transmitter-kanalen
            hbCh <- hb

        }
    }
}

func StartMasterReceiver(port int, masterUpdateCh chan<- string) {
    // Opprett en kanal som skal motta heartbeat-meldinger
    hbCh := make(chan MasterHeartBeat)
    go bcast.Receiver(port, hbCh)

    // Sjekk for heartbeats kontinuerlig
    timeoutDuration := 1 * time.Second
    lastHeartbeat := time.Now()

    for {
        select {
        case hb := <-hbCh:
            // Oppdater siste heartbeat tid
			fmt.Println("MasterIdMottatt: ",hb.MasterID)
            lastHeartbeat = time.Now()
            masterUpdateCh <- hb.MasterID

        default:
            if time.Since(lastHeartbeat) > timeoutDuration {
                fmt.Println("Ingen master heartbeat mottatt innenfor timeout - master antas nede!")
                masterUpdateCh <- ""
                lastHeartbeat = time.Now()
            }
            time.Sleep(50 * time.Millisecond)
        }
    }
}