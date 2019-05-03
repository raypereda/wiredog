package alerts_test

import (
	"net/http"
	"testing"
	"time"

	// "github.com/jonboulle/clockwork"
	"code.cloudfoundry.org/clock/fakeclock"

	"github.com/raypereda/wiredog/pkg/alerts"
	"github.com/raypereda/wiredog/pkg/screens"
)

func TestAlerter(t *testing.T) {
	clock := fakeclock.NewFakeClock(time.Now())
	interval := time.Second
	threshold := 2
	hist := alerts.NewHistory(clock, interval)
	screen := screens.New("fake")
	alerter := alerts.New(clock, hist, interval, threshold, screen)

	alerter.Check()
	if alerter.IsActivated() {
		t.Fatal("Alerter should NOT be activated")
	}

	r := &http.Request{}
	alerter.History.Record(r)
	alerter.History.Record(r)
	alerter.Check()
	if alerter.IsActivated() {
		t.Fatal("Alerter should NOT be activated")
	}

	alerter.History.Record(r)
	alerter.Check()
	if !alerter.IsActivated() {
		t.Fatal("Alerter should be activated")
	}

	clock.Increment(2 * time.Second)
	alerter.Check()
	if alerter.IsActivated() {
		t.Fatal("Alerter should have cleared")
	}
}
