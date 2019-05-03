package stats_test

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/raypereda/wiredog/pkg/stats"
)

func ExampleGetTopSections() {
	m := map[string]int{
		"google.com/sectionA": 60,
		"google.com/sectionB": 70,
		"google.com/sectionC": 80,
		"google.com/sectionD": 90,
		"google.com":          100,
		"amazon.com/sectionA": 55,
		"amazon.com/sectionB": 65,
		"amazon.com/sectionC": 75,
		"amazon.com/sectionD": 85,
		"amazon.com":          100,
	}

	s := stats.New()
	s.SectionCounts = m

	topSections := s.GetTopSections()

	fmt.Print(strings.Join(topSections, "\n"))
	// Output:
	// 100 amazon.com
	// 100 google.com
	//  90 google.com/sectionD
	//  85 amazon.com/sectionD
	//  80 google.com/sectionC
	//  75 amazon.com/sectionC
	//  70 google.com/sectionB
	//  65 amazon.com/sectionB
	//  60 google.com/sectionA
	//  55 amazon.com/sectionA
}

func ExampleGetSection() {
	urls := []string{
		"http://datadog.com",
		"http://datadog.com/",
		"http://datadog.com/section/",
		"http://datadog.com/section/misc",
	}
	for _, u := range urls {
		url, _ := url.Parse(u)
		section := stats.GetSection(url)
		fmt.Println(u, "==>", section)
	}
	// Output:
	// http://datadog.com ==> datadog.com
	// http://datadog.com/ ==> datadog.com
	// http://datadog.com/section/ ==> datadog.com/section
	// http://datadog.com/section/misc ==> datadog.com/section
}
