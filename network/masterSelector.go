package network

import (
	"Driver-go/network/peers"
	"strconv"
)


func returnMasterID(elevatorIDs []string) string {
	if len(elevatorIDs) == 0 {
		return "No elevator exists" 
	}

	masterID := elevatorIDs[0]
	for _, id := range elevatorIDs {
		if id < masterID {
			masterID = id
		}
	}
	return masterID
}


func determineMaster(masterID string, peersUpdate peers.PeerUpdate, isMasterChannel chan <- bool) string{
	peers := peersUpdate.PeersID

	if peers[0] == masterID{
		isMasterChannel <- true
		return masterID
	}else{
		isMasterChannel <- false
		return "Mayday do not have a master"

	}

}


func UpdateMaster(isMasterChannel chan<- bool, peersUpdate peers.PeerUpdate){
	peers := peersUpdate.PeersID
	newMasterID := peers[0]
	isMasterChannel <- true
}


func UpdateIDs(elevatorIDs []string) (int, string) {

	currentHighestID := elevatorIDs[0]
	for _, id := range elevatorIDs {
		if id > currentHighestID {
			currentHighestID = id
		}
	}
	// Checks if there are any empty IDs and updates this
	highestIDInt, _ := strconv.Atoi(currentHighestID)
	currIndex := 0
	for index, id := range elevatorIDs {
		if id == "" {
			currentHighestID = strconv.Itoa(highestIDInt + 1)
			currIndex = index
			break
		}
	}
	return currIndex, currentHighestID
}


