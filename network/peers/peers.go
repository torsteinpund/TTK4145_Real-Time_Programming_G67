package peers

import (
	"Driver-go/network/conn"
	"fmt"
	"net"
	"sort"
	"time"
)

type Peer struct {
	ID   	string
	Ip  string
	UDPPort int
	lastSeen time.Time
}

type PeersUpdate struct {
	PeersID []string
	New   string
	Lost  []string
}

const interval = 15 * time.Millisecond
const timeout = 500 * time.Millisecond

func Transmitter(port int, id string, transmitEnable <-chan bool) {


	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))

	enable := true
	for {
		select {
		case enable = <-transmitEnable:
		case <-time.After(interval):
		}
		if enable {
			conn.WriteTo([]byte(id), addr)
		}
	}
}

func Receiver(port int, peerUpdateCh chan<- PeersUpdate) {

	var buf [1024]byte
	var p PeersUpdate
	lastSeen := make(map[string]time.Time)

	conn := conn.DialBroadcastUDP(port)

	for {
		updated := false

		conn.SetReadDeadline(time.Now().Add(interval))
		n, _, _ := conn.ReadFrom(buf[0:])

		id := string(buf[:n])

		// Adding new connection
		p.New = ""
		if id != "" {
			if _, idExists := lastSeen[id]; !idExists {
				p.New = id
				updated = true
			}

			lastSeen[id] = time.Now()
		}
		
		// Removing dead connection

		p.Lost = make([]string, 0)
		for k, v := range lastSeen {
			if time.Since(v) > timeout {
				updated = true
				p.Lost = append(p.Lost, k)
				delete(lastSeen, k)
			}
		}

		// Sending update
		if updated {
			p.PeersID = make([]string, 0, len(lastSeen))

			for k, _ := range lastSeen {
				p.PeersID = append(p.PeersID, k)
			}

			sort.Strings(p.PeersID)
			sort.Strings(p.Lost)
			peerUpdateCh <- p
		}
	}
}
