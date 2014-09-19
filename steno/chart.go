package main

import (
	//	"encoding/csv"
	//	"fmt"
	//	"github.com/bcampbell/badger"
	//	"github.com/gorilla/mux"
	"net/http"
	//	"net/url"
	//	"path"
	"regexp"
	//	"strings"
	"time"
)

var datePat = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)

// getDate extracts a yyyy-mm-dd date from anywhere in a string
func getDate(s string) time.Time {
	const dateForm = "2006-01-02"
	t, _ := time.Parse(dateForm, datePat.FindString(s))
	return t
}

// dateExtents returns the minimum and maximum publication dates from a
// list of articles. Articles with missing dates are discounted.
func dateExtents(arts []*Article) (time.Time, time.Time) {
	if len(arts) == 0 {
		panic("no articles")
	}
	minDate := time.Date(9999, time.December, 31, 0, 0, 0, 0, time.UTC)
	maxDate := time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC)
	for _, art := range arts {
		// determine min/max dates
		pubDate := getDate(art.Published)
		if pubDate.IsZero() {
			continue
		}
		if pubDate.Before(minDate) {
			minDate = pubDate
		}
		if pubDate.After(maxDate) {
			maxDate = pubDate
		}
	}

	return minDate, maxDate
}

type bar struct {
	Date time.Time
	Val  int
}

func handleBarChart(w http.ResponseWriter, req *http.Request) {
	var arts []*Article
	var queryErr error
	queryString := req.FormValue("q")
	arts, queryErr = search(queryString)
	if queryErr != nil {
		http.Error(w, "error: "+queryErr.Error(), http.StatusInternalServerError)
		return
	}

	// collect counts per day and determine extent (min/max date)
	counts := map[time.Time]int{}
	minDate := time.Date(9999, time.December, 31, 0, 0, 0, 0, time.UTC)
	maxDate := time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC)
	for _, art := range arts {
		t := getDate(art.Published)
		counts[t]++
		if t.IsZero() {
			continue
		}
		if t.Before(minDate) {
			minDate = t
		}
		if t.After(maxDate) {
			maxDate = t
		}
	}

	//
	bars := []bar{}
	for t := minDate; !t.After(maxDate); t = t.AddDate(0, 0, 1) {
		b := bar{t, counts[t]}
		bars = append(bars, b)
	}

	//fmt.Println(bars)
	tmpl := tmpls.MustGet("barchart")

	params := struct {
		Data []bar
		//		Err   error
		Query string
	}{
		bars,
		//		err,
		queryString,
	}
	tmpl.Execute(w, params)
}
