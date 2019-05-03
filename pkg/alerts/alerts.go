// Package alerts provides alerts based on history of events.
package alerts

import (
	"fmt"
	"log"
	"sync"
	"time"

	// "github.com/jonboulle/clockwork"
	"code.cloudfoundry.org/clock"
	"github.com/raypereda/wiredog/pkg/screens"
)

// Alerter is based on number of recent requests.
type Alerter struct {
	sync.Mutex
	isActivated   bool
	History       *History
	Events        []*Event
	checkInterval time.Duration
	threshold     int
	clock         clock.Clock
	screen        *screens.Screen
}

// New creates a new alerter.
func New(c clock.Clock, h *History, i time.Duration, t int, s *screens.Screen) *Alerter {
	return &Alerter{
		History:       h,
		checkInterval: i,
		threshold:     t,
		clock:         c,
		screen:        s,
	}
}

// Run starts the process of periodically processing the alert.
func (a *Alerter) Run() {
	ticker := time.NewTicker(a.checkInterval)
	// TODO: check if need to sleep for first interval
	for {
		<-ticker.C
		a.Check()
	}
}

// IsActivated ...
func (a *Alerter) IsActivated() bool {
	a.Lock()
	defer a.Unlock()
	return a.isActivated
}

// Check activates and clears the alert as needed
func (a *Alerter) Check() {
	a.Lock()
	defer a.Unlock()
	numRequests := a.History.RecentRequestCount()
	log.Println("number of requests:", numRequests)
	if !a.isActivated {
		if numRequests > a.threshold {
			msg := fmt.Sprintf("High traffic generated an alert - hits = %d", numRequests)
			if a.screen != nil { // for easier testing
				a.screen.LogPrintln(msg)
				a.screen.Render()
			}
			a.isActivated = true
		}
	} else {
		if numRequests <= a.threshold {
			msg := "Alert recovered; back to normal traffic"
			if a.screen != nil { // for easier testing
				a.screen.LogPrintln(msg)
				a.screen.Render()
			}
			a.isActivated = false
		}
	}
}
