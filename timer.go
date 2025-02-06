package main

import (
	"time"
)

var (
	timerEndTime float64
	timerActive  bool
)

// getWallTime returnerer veggklokken som et flyttall
func getWallTime() float64 {
	now := time.Now()
	return float64(now.Unix()) + float64(now.Nanosecond())*1e-9
}

// timerStart starter en timer med en spesifisert varighet i sekunder
func timerStart(duration float64) {
	timerEndTime = getWallTime() + duration
	timerActive = true
}

// timerStop stopper den aktive timeren
func timerStop() {
	timerActive = false
}

// timerTimedOut sjekker om timeren har gått ut
func timerTimedOut() bool {
	return timerActive && getWallTime() > timerEndTime
}



/*var timerActive bool = false

func timerActivate(duration time.Duration) {
	timerActive = true

	timer := time.NewTimer(duration * time.Second)
	fmt.Println("Timer started")
	<-timer.C
	fmt.Println("Timer stopped")
	timerActive = false
}
*/