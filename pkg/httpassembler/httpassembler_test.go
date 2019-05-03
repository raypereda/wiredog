package httpassembler_test

import (
	"net/http"
	"testing"
	"time"

	asm "github.com/raypereda/wiredog/pkg/httpassembler"
)

func TestAssembler(t *testing.T) {
	requests := make(chan *http.Request)
	// Here are the other devices on my laptop: lo, enp0s31f6
	asm, err := asm.NewHTTPAssembler("wlp2s0", requests)
	if err != nil {
		t.Fatal(err)
	}
	go asm.Run()

	_, err = http.Get("http://example.com")
	if err != nil {
		t.Fatal(err)
	}

	timer := time.NewTimer(3 * time.Second)
	for {
		select {
		case <-timer.C:
			t.Fatal("http request timed out")
		case request := <-requests:
			if request.Host == "example.com" {
				return
			}
		}
	}
}
