package alerts_test

import (
	"net/http"
	"testing"
	"time"

	// This clockwork did not work out well.
	// "github.com/jonboulle/clockwork"
	"code.cloudfoundry.org/clock/fakeclock"
	"github.com/raypereda/wiredog/pkg/alerts"
)

func TestCacheRequests(t *testing.T) {
	clk := fakeclock.NewFakeClock(time.Now())
	history := alerts.NewHistory(clk, time.Second)

	r := &http.Request{}
	history.Record(r)
	history.Record(r)

	c := history.RecentRequestCount()
	if c != 2 {
		t.Error("Expected 2 recent requests, got", c)
	}

	clk.IncrementBySeconds(1)
	// history.Tidy() automatically called.

	history.Record(r)
	c = history.RecentRequestCount()
	if c != 1 {
		t.Error("Expected 1 recent requests, got", c)
	}

	clk.IncrementBySeconds(2)
	// history.Tidy() automatically called.

	c = history.RecentRequestCount()
	if c != 0 {
		t.Error("Expected 0 recent requests, got", c)
	}

}
