package alerts

import (
	"log"
	"net/http"
	"sync"
	"time"

	// "github.com/jonboulle/clockwork"
	// "code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/clock"
)

// History stores a list of recent requests.
type History struct {
	sync.Mutex
	requests []*request
	duration time.Duration
	clock    clock.Clock
}

// NewHistory creates a new instance of History.
// A clock is passed in to facilate testing.
func NewHistory(c clock.Clock, d time.Duration) *History {
	h := &History{
		duration: d,
		clock:    c,
	}
	go h.periodicallyTidy()
	return h
}

// Record processes the HTTP request.
func (h *History) Record(req *http.Request) {
	h.Lock()
	defer h.Unlock()

	r := &request{
		request:    req,
		recordTime: h.clock.Now(),
	}
	h.requests = append(h.requests, r)
}

func (h *History) periodicallyTidy() {
	ticker := h.clock.NewTicker(time.Second)
	for {
		<-ticker.C()
		h.Tidy()
	}
}

// Tidy deletes requests that are old.
func (h *History) Tidy() {
	h.Lock()
	defer h.Unlock()

	if len(h.requests) == 0 {
		return
	}
	// keep two time periods to keep steady RecentRequestCount
	startTime := h.clock.Now().Add(-2 * h.duration)
	var skip int
	for _, request := range h.requests {
		if request.recordTime.After(startTime) {
			break
		}
		skip++
	}
	log.Printf("trimmed %d requests from recent history\n", skip)
	h.requests = h.requests[skip:]
}

// RecentRequestCount returns number of requests since last watch interval.
func (h *History) RecentRequestCount() int {
	h.Lock()
	defer h.Unlock()

	startTime := h.clock.Now().Add(-h.duration)
	var count int

	// the most recent requests are on the end
	for i := len(h.requests) - 1; i >= 0; i-- {
		request := h.requests[i]
		// for _, request := range h.requests {
		if request.recordTime.After(startTime) {
			count++
		} else { // requests are ordered by record time and
			break // serialized with a channel
		}
	}
	return count
}

// request represents a request with a timestamp.
type request struct {
	request    *http.Request
	recordTime time.Time
}
