package network

import (
	"Driver-go/network/peers"
	. "Driver-go/types"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type ClientChannels struct {
	Ch_peerUpdate        chan peers.PeersUpdate
	Ch_peerLost          chan string
	Ch_newPeer           chan string
	Ch_isMaster          chan bool
}



type Client struct {
	id          string
	activePeers map[string]peers.Peer
}

// Creates a client object
func NewClient(id string) *Client {
	return &Client{
		id:          id,
		activePeers: make(map[string]peers.Peer),
	}
}

func (c *Client) RunClient(id string, clientChannels ClientChannels,ch_RX RXChannels, Ch_netWorkMsg <-chan NetworkMessage) {
	// currentMasterID := id
	for {
		select {
		case update := <-clientChannels.Ch_peerUpdate:

			peerstatus, peerID := c.updatePeers(update)
			if peerstatus == "lostPeer" {
					delete(c.activePeers, peerID)
					clientChannels.Ch_peerLost <- peerID
					fmt.Println("Peer lost, we made it passed: ", peerID)
				
			} else if peerstatus == "newPeer" {
				
				c.activePeers[peerID] = peers.Peer{ID: peerID}
				clientChannels.Ch_newPeer <- peerID
				fmt.Println("New peer added, we made it passed: ", peerID)
			}
			currentMasterID := updateMaster(c.activePeers, clientChannels.Ch_isMaster)
			if currentMasterID == id {
				clientChannels.Ch_isMaster <- true
			}else{
				clientChannels.Ch_isMaster <- false
			}
		default:

		case networkMsg := <-Ch_netWorkMsg:
			fmt.Println("Network message received")
			msgData, _ := json.Marshal(networkMsg.MsgData)
			msgType := networkMsg.MsgType
			simpleNetworkMessage := SimpleNetworkMsg{MsgType: msgType, MsgData: msgData}
			go DecodeMessage(ch_RX, simpleNetworkMessage)
		}
	}
}

func (c *Client) updatePeers(update peers.PeersUpdate) (string, string) {
	// Updates activePeers
	changedAllPeers := ""
	peerID := ""

	if update.New != "" {
		c.activePeers[update.New] = peers.Peer{ID: update.New}
		changedAllPeers, peerID = "newPeer", update.New
	}
	// Removes lost peers
	for _, lostID := range update.Lost {
		delete(c.activePeers, lostID)
		changedAllPeers, peerID = "lostPeer", lostID
	}
	return changedAllPeers, peerID
}


func checkIfMaster(currentMasterID string, lostPeerID string) bool {
	return currentMasterID == lostPeerID
}


func updateMaster(activePeers map[string]peers.Peer, Ch_isMaster chan<-bool) string {

	peers := []int{}
	// Find the ID of the current master
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

