// Package screens manages a terminal screen.
package screens

import (
	"fmt"
	"log"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

// topN sets the top *N* websites
const topN = 10

// Screen is a terminal interface
type Screen struct {
	sparkline *widgets.Sparkline
	topN      *widgets.List
	messages  *widgets.List
	grid      *ui.Grid
}

// New creates a new screen
func New(settings string) *Screen {
	if settings == "fake" {
		return nil
	}
	p1 := widgets.NewParagraph()

	p1.Title = "WireDog - monitors local HTTP network traffic"
	p1.Border = false

	p2 := widgets.NewParagraph()
	p2.Title = "Settings"
	p2.Text = settings

	sl := widgets.NewSparkline()
	sl.LineColor = ui.ColorGreen
	slg := widgets.NewSparklineGroup(sl)
	slg.Title = "Traffic Trend"

	l1 := widgets.NewList()
	l1.Title = "Top " + string(topN) + " Website By Hits"
	l1.WrapText = false

	l2 := widgets.NewList()
	l2.Title = "messages"
	l2.TextStyle = ui.NewStyle(ui.ColorClear, ui.ColorClear, ui.ModifierClear) // for scrolling
	l2.WrapText = false
	l2.ScrollDown()

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	grid.Set(
		ui.NewRow(0.05,
			ui.NewCol(1, p1),
		),
		ui.NewRow(0.25,
			ui.NewCol(1.0/2, p2),
			ui.NewCol(1.0/2, slg),
		),
		ui.NewRow(0.7/2,
			ui.NewCol(1.0, l1),
		),
		ui.NewRow(0.7/2,
			ui.NewCol(1.0, l2),
		),
	)
	return &Screen{
		sparkline: sl,
		topN:      l1,
		messages:  l2,
		grid:      grid,
	}
}

// UpdateTopN updates the top N sections
func (s *Screen) UpdateTopN(rows []string) {
	s.topN.Rows = rows
}

// AddRequestCount add to the sparkline data
func (s *Screen) AddRequestCount(d float64) {
	s.sparkline.Data = append(s.sparkline.Data, d)
	// experimented with value at full screen on laptop
	lastN := 83
	if len(s.sparkline.Data) > lastN { // plot most recent
		skip := len(s.sparkline.Data) - lastN
		s.sparkline.Data = s.sparkline.Data[skip:]
	}
}

// LogPrintln adds to the standard logger and the message wideget
func (s *Screen) LogPrintln(msg string) {
	log.Println(msg)
	now := time.Now() // TODO: share exact time with log and messages widget
	line := fmt.Sprintf("%s %s", now.Format("1968-05-29 15:04:05"), msg)
	
	s.messages.Rows = append(s.messages.Rows, line)
	s.messages.ScrollBottom()
}

// Render renders the whole screen
func (s *Screen) Render() {
	ui.Render(s.grid)
}

// Start kicks off rendering loop
func (s *Screen) Start() {
	s.LogPrintln("Started monitoring.")
	s.messages.ScrollBottom()
	ui.Render(s.grid)
	previousKey := ""
	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		case "j", "<Down>":
			s.messages.ScrollDown()
		case "k", "<Up>":
			s.messages.ScrollUp()
		case "<C-d>":
			s.messages.ScrollHalfPageDown()
		case "<C-u>":
			s.messages.ScrollHalfPageUp()
		case "<C-f>":
			s.messages.ScrollPageDown()
		case "<C-b>":
			s.messages.ScrollPageUp()
		case "g":
			if previousKey == "g" {
				s.messages.ScrollTop()
			}
		case "<Home>":
			s.messages.ScrollTop()
		case "G", "<End>":
			s.messages.ScrollBottom()
		case "<Resize>":
			payload := e.Payload.(ui.Resize)
			s.grid.SetRect(0, 0, payload.Width, payload.Height)
			ui.Clear()
			ui.Render(s.grid)
		}
		if previousKey == "g" {
			previousKey = ""
		} else {
			previousKey = e.ID
		}
		ui.Render(s.grid)
	}
}
