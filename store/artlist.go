package store

import (
	"strconv"
	"strings"
)

// ArtList is a list of article IDs from a Store.
type ArtList []ArtID

func (arts ArtList) StringList() string {
	frags := make([]string, len(arts))
	for idx, id := range arts {
		frags[idx] = strconv.Itoa(int(id))
	}
	return strings.Join(frags, ",")
}

func (arts *ArtList) Days() []string {
	return []string{}
	/*XYZZY*/
	/*
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
	*/
}

func (arts *ArtList) Pubs() []string {
	return []string{}
	/*XYZZY*/
	/*
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
	*/
}

func (arts ArtList) Subtract(other ArtList) ArtList {
	lookup := map[ArtID]struct{}{}
	for _, id := range other {
		lookup[id] = struct{}{}
	}
	out := make(ArtList, 0, len(arts)-len(other))
	for _, id := range arts {
		if _, got := lookup[id]; !got {
			out = append(out, id)
		}
	}
	return out
}

func (arts ArtList) Intersection(other ArtList) ArtList {
	lookup := map[ArtID]struct{}{}
	for _, id := range other {
		lookup[id] = struct{}{}
	}
	out := ArtList{}
	for _, id := range arts {
		if _, got := lookup[id]; got {
			out = append(out, id)
		}
	}
	return out
}

func (arts ArtList) Union(other ArtList) ArtList {
	lookup := map[ArtID]struct{}{}
	for _, id := range other {
		lookup[id] = struct{}{}
	}
	for _, id := range arts {
		lookup[id] = struct{}{}
	}
	out := make(ArtList, 0, len(lookup))
	for id, _ := range lookup {
		out = append(out, id)
	}
	return out
}
