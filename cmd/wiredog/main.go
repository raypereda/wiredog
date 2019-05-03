// Wiredog is a command for network monitoring.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"code.cloudfoundry.org/clock"
	ui "github.com/gizak/termui/v3"
	"github.com/raypereda/wiredog/pkg/alerts"
	asm "github.com/raypereda/wiredog/pkg/httpassembler"
	"github.com/raypereda/wiredog/pkg/screens"
	"github.com/raypereda/wiredog/pkg/stats"
)

var (
	device        string
	hitThreshold  int
	logDirectory  string
	alertInterval time.Duration
	testRequests  bool
)

func settings() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Alert Interval    : %s\n", alertInterval)
	fmt.Fprintf(&sb, "Hit Threshold     : %d\n", hitThreshold)
	fmt.Fprintf(&sb, "Network Interface : %s\n", device)
	absDirectory, err := filepath.Abs(logDirectory)
	if err != nil {
		log.Fatalf("error getting absolute path for %s: %v", logDirectory, err)
	}
	logDirectory = absDirectory
	fmt.Fprintf(&sb, "Log Directory     : %8s\n", logDirectory)
	return sb.String()
}

// Usage prints usage, overwrites default
var Usage = func() {
	fmt.Fprintln(flag.CommandLine.Output(), "Wiredog is a tool for monitoring network traffic.")
	fmt.Fprintln(flag.CommandLine.Output())
	fmt.Fprintln(flag.CommandLine.Output(), "Only local HTTP requests are studied.")
	fmt.Fprintln(flag.CommandLine.Output(), "Administrator permission is required to access network devices.")
	fmt.Fprintln(flag.CommandLine.Output(), "Consider using sudo on Linux or get admin rights on Windows.")
	fmt.Fprintln(flag.CommandLine.Output())
	fmt.Fprintln(flag.CommandLine.Output(), "Usage: sudo \\path\\to\\wiredog [flags]")
	fmt.Fprintln(flag.CommandLine.Output())
	flag.PrintDefaults()
}

func main() {
	// other network devices on my laptop: lo, enp0s31f6, but they don't work in promiscuous mode
	flag.StringVar(&device, "d", "wlp2s0", "network device name")
	flag.IntVar(&hitThreshold, "h", 2, "threshold of HTTP request hits")
	flag.StringVar(&logDirectory, "log", ".", "log directory")
	flag.DurationVar(&alertInterval, "p", 2*time.Minute, "time period between traffic checks")
	flag.BoolVar(&testRequests, "t", false, "generates test HTTP requests; sets h=2 and p=2s")
	flag.Usage = Usage
	flag.Parse()

	logFile := logDirectory + "/wiredog.log"
	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	if testRequests {
		hitThreshold = 5
		alertInterval = 2 * time.Second
		go makeTestRequests()
	}

	requests := make(chan *http.Request) // receiver of requests
	asm, err := asm.NewHTTPAssembler(device, requests)
	if err != nil {
		log.Fatalln("error connecting to network devices", err)
	}
	go asm.Run()

	s := screens.New(settings())

	clock := clock.NewClock()
	hist := alerts.NewHistory(clock, alertInterval)
	// BUGFIX: avoid sharing screen state with alert, need to rethink
	alert := alerts.New(clock, hist, alertInterval, hitThreshold, s)
	go alert.Run()

	metrics := stats.New()
	go func() {
		for {
			r := <-requests
			metrics.Tally(r) // collects statistics
			hist.Record(r)   // keep history for alerter
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			<-ticker.C
			s.UpdateTopN(metrics.GetTopSections())
			s.Render()
		}
	}()

	go func() {
		ticker := time.NewTicker(alertInterval)
		for {
			<-ticker.C
			count := float64(hist.RecentRequestCount())
			s.AddRequestCount(count)
			s.Render()
		}
	}()

	s.Start()
	log.Println("Stopped monitoring.")
}

func get(url string) {
	_, _ = http.Get(url)
}

func makeTestRequests() {
	time.Sleep(1 * time.Second) // allow alerts to setup
	get("http://example1.com")
	get("http://example2.com/sectionA/")
	get("http://example2.com/sectionA/")
	get("http://example3.com")
	get("http://example3.com")
	get("http://example3.com")

	for {
		time.Sleep(time.Second)
		get("http://example4.com")
		get("http://example4.com")
		get("http://example4.com")
		get("http://example4.com")
		time.Sleep(time.Second)
		get("http://example5.com")
		get("http://example5.com")
		get("http://example5.com")
		get("http://example5.com")
		get("http://example5.com")
	}
}
