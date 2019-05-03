// Package stats collects statistics of HTTP requests.
package stats

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
)

// Metrics stores HTTP metrics for a stream of requests.
type Metrics struct {
	sync.Mutex
	SectionCounts map[string]int
	// TODO: metrics like body length, first line of body
}

// New creates a new set of metrics.
func New() *Metrics {
	return &Metrics{
		SectionCounts: make(map[string]int),
	}
}

// Tally adds a request to the stats
func (m *Metrics) Tally(req *http.Request) {
	m.Lock()
	defer m.Unlock()
	req.URL.Host = req.Host
	section := GetSection(req.URL)
	m.SectionCounts[section]++ // TODO: add more metrics
}

// GetSection returns the "section", which is defined as the host and
// up to the first part of the path before the second /.
func GetSection(url *url.URL) string {
	path := url.EscapedPath()
	if path == "" || path == "/" {
		return url.Host
	}
	section := strings.SplitN(path, "/", 3)[1]
	return url.Host + "/" + section
}

// entry represents a key and value entry
type entry struct {
	name  string
	count int
}

// getTopNEntries sorts the map by value then returns top N
func getTopNEntries(m map[string]int, topN int) []*entry {
	var entries []*entry
	for k, v := range m {
		entries = append(entries, &entry{k, v})
	}
	isMore := func(i, j int) bool {
		if entries[i].count == entries[j].count {
			return entries[i].name < entries[j].name
		}
		return entries[i].count > entries[j].count
	}
	sort.Slice(entries, isMore)
	if topN > len(entries) {
		return entries
	}
	return entries[:topN]
}

// GetTopSections returns the top 10 sections
func (m *Metrics) GetTopSections() []string {
	m.Lock()
	defer m.Unlock()
	entries := getTopNEntries(m.SectionCounts, 10)
	var top []string
	for _, entry := range entries {
		s := fmt.Sprintf("%3d %s", entry.count, entry.name)
		top = append(top, s)
	}
	return top
}
