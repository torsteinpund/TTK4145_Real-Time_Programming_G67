package network

import (
	"Driver-go/network/peers"
	"fmt"
	"sort"
	"strconv"
	"strings"
	// "time"
)

type PeerHandler struct {
	id          string
	stopCh      chan struct{}
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

	for {
		select {
		case update := <-Ch_peerUpdate:

			peerstatus, peerID := pH.updatePeers(update)
			if peerstatus == "lostPeer" {
				currentMasterID := updateMaster(pH.activePeers, Ch_isMaster, peerID)
				if currentMasterID == id {
					Ch_isMaster <- true
				} else {
					Ch_isMaster <- false
				}
				delete(pH.activePeers, peerID)
				Ch_peerLost <- peerID
				fmt.Println("Peer lost, we made it passed: ", peerID)

			} else if peerstatus == "newPeer" {
				currentMasterID := updateMaster(pH.activePeers, Ch_isMaster, peerID)
				if currentMasterID == id {
					Ch_isMaster <- true
				} else {
					Ch_isMaster <- false
				}
				pH.activePeers[peerID] = peers.Peer{ID: peerID}
				Ch_newPeer <- peerID
				fmt.Println("New peer added, we made it passed: ", peerID)
			}
				

			// peerstatus, peerID := pH.updatePeers(update)
			// if peerstatus == "lostPeer" && peerID == currentMasterID {
			// 	currentMasterID = updateMaster(pH.activePeers, Ch_isMaster, peerID)
			// 	if currentMasterID == id {
			// 		Ch_isMaster <- true
			// 	} else {
			// 		Ch_isMaster <- false
			// 	}
			// 	delete(pH.activePeers, peerID)
			// 	Ch_peerLost <- peerID
			// 	fmt.Println("Peer lost, we made it passed: ", peerID)

			// } else if peerstatus == "newPeer" && len(pH.activePeers) == 0 {
			// 	currentMasterID = updateMaster(pH.activePeers, Ch_isMaster, peerID)
			// 	if currentMasterID == id {
			// 		Ch_isMaster <- true
			// 	}
			// 	pH.activePeers[peerID] = peers.Peer{ID: peerID}
			// 	Ch_newPeer <- peerID
			// 	fmt.Println("New peer added, we made it passed: ", peerID)
			// } else if peerstatus == "lostPeer"{
			// 	delete(pH.activePeers, peerID)
			// 	Ch_peerLost <- peerID
			// 	fmt.Println("Peer lost, we made it passed: ", peerID)

			// } else if peerstatus == "newPeer"{
			// 	currentMasterID = updateMaster(pH.activePeers, Ch_isMaster, peerID)
			// 	Ch_isMaster <- false
			// 	pH.activePeers[peerID] = peers.Peer{ID: peerID}
			// 	Ch_newPeer <- peerID
			// 	fmt.Println("New peer added, we made it passed: ", peerID)
			//  }
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

func checkIfMaster(currentMasterID string, lostPeerID string) bool {
	return currentMasterID == lostPeerID
}

func updateMaster(activePeers map[string]peers.Peer, Ch_isMaster chan<- bool, ID string) string {

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
