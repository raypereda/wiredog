package alerts

import (
	"time"
)

// Event is message with a time stamp
type Event struct {
	message string
	time    time.Time // when added
}

// AddEvent adds an event to store.
func AddEvent(events []*Event, message string) {
	e := &Event{
		message: message,
		time:    time.Now(),
	}
	events = append(events, e)
}
