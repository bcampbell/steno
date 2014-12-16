package main

import (
	"sort"
)

type ArtList []*Article

func (arts *ArtList) Days() []string {
	found := map[string]struct{}{}
	for _, art := range *arts {
		found[art.Day()] = struct{}{}
	}
	out := make([]string, 0, len(found))
	for day, _ := range found {
		out = append(out, day)
	}
	sort.Strings(out)
	return out
}

func (arts *ArtList) Pubs() []string {
	found := map[string]struct{}{}
	for _, art := range *arts {
		found[art.Pub] = struct{}{}
	}
	out := make([]string, 0, len(found))
	for pub, _ := range found {
		out = append(out, pub)
	}
	sort.Strings(out)
	return out
}

func (arts ArtList) Subtract(other ArtList) ArtList {
	lookup := map[*Article]struct{}{}
	for _, art := range other {
		lookup[art] = struct{}{}
	}
	out := make(ArtList, 0, len(arts)-len(other))
	for _, art := range arts {
		if _, got := lookup[art]; !got {
			out = append(out, art)
		}
	}
	return out
}

//***************************
// support for ArtList sorting

type By func(p1, p2 *Article) bool

func (by By) Sort(arts ArtList) {
	ps := &artSorter{
		arts: arts,
		by:   by,
	}
	sort.Sort(ps)
}

type artSorter struct {
	arts ArtList
	by   func(p1, p2 *Article) bool
}

// Len is part of sort.Interface.
func (s *artSorter) Len() int {
	return len(s.arts)
}

// Swap is part of sort.Interface.
func (s *artSorter) Swap(i, j int) {
	s.arts[i], s.arts[j] = s.arts[j], s.arts[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *artSorter) Less(i, j int) bool {
	return s.by(s.arts[i], s.arts[j])
}
