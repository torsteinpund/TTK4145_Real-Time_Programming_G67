package client

import (
	"Driver-go/network/peers"
	"Driver-go/types"
	"fmt"
	"sort"
	"strconv"
	"strings"
	//"Driver-go/network/bcast"
)

type ClientChannels struct {
	InputChannel             chan types.NetworkMessage
	OutputChannel            chan types.NetworkMessage
	PeerUpdateChannel        chan peers.PeersUpdate
	PeerLostChannel          chan string
	PeerNewChannel           chan string
	IsMasterChannel          chan bool
	RegisteredNewPeerChannel chan string
	
}

type Client struct {
	id          string
	stopCh      chan struct{}
	activePeers map[string]peers.Peer
}

// Creates a client object
func NewClient(id string) *Client {
	return &Client{
		id:          id,
		stopCh:      make(chan struct{}),
		activePeers: make(map[string]peers.Peer),
	}
}

func (c *Client) RunClient(currentMasterID string, clientChannels ClientChannels) {

	for {
		select {
		case msg := <-clientChannels.InputChannel:
			fmt.Println("Mottatt network-melding:", msg)

		case update := <-clientChannels.PeerUpdateChannel:
			fmt.Println("Peer-oppdatering mottatt:", update)

			peerstatus, peerID := c.updatePeers(update)
			if peerstatus == "lostPeer" {
				if checkIfMaster(currentMasterID, peerID) {
					clientChannels.IsMasterChannel <- false
					newMasterID := updateMaster(c.activePeers)
					if newMasterID != "" {
						currentMasterID = newMasterID
						clientChannels.IsMasterChannel <- true
						clientChannels.PeerLostChannel <- peerID
					}
				} else {
					clientChannels.PeerLostChannel <- peerID
				}
			} else if peerstatus == "newPeer" {
				clientChannels.OutputChannel <- types.NetworkMessage{MsgType: "Registered new peer", MsgData: peerID, Receipient: types.All}
				clientChannels.RegisteredNewPeerChannel <- peerID
			}

		case <-c.stopCh:
			fmt.Println("Client stoppes...")
			return
		}
	}
}

func (c *Client) updatePeers(update peers.PeersUpdate) (string, string) {
	// Updates activePeers
	changedAllPeers := ""
	peerID := ""

	if update.New != "" {
		c.activePeers[update.New] = peers.Peer{ID: update.New}
		fmt.Println("Ny peer lagt til:", update.New)
		changedAllPeers, peerID = "newPeer", update.New
	}
	// Removes lost peers
	for _, lostID := range update.Lost {

		delete(c.activePeers, lostID)
		fmt.Println("Fjernet tapt peer:", lostID)
		changedAllPeers, peerID = "lostPeer", lostID
	}
	return changedAllPeers, peerID

}

func (c *Client) Stop() {
	close(c.stopCh)
}

func checkIfMaster(currentMasterID string, lostPeerID string) bool {
	return currentMasterID == lostPeerID
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
	fmt.Println("New master is: ", currentMasterID)
	return currentMasterID
}
