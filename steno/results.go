package main

import (
	"fmt"
	"regexp"
	"semprini/steno/steno/store"
	"strings"
)

type Results struct {
	Query string
	arts  store.ArtList
	Len   int

	FacetLen int
	facets   []*Facet
}

func NewResults(db *store.Store, query string) (*Results, error) {

	var arts store.ArtList
	var err error
	if query == "" {
		arts, err = db.AllArts()
	} else {
		arts, err = db.Search(query)
	}
	if err != nil {
		return nil, err
	}

	res := Results{Query: query}
	res.setArts(arts)
	return &res, nil
}

func (res *Results) setArts(arts store.ArtList) {
	res.arts = arts
	res.Len = len(res.arts)

	// calc facets
	tab := map[string]int{}
	for _, art := range res.arts {
		tab[art.Pub]++
	}
	res.facets = []*Facet{}
	for txt, cnt := range tab {
		res.facets = append(res.facets, &Facet{txt, cnt})
	}
	res.FacetLen = len(res.facets)
}

// return terms from current query, for highlighting
// eg if query is `chaos headline:"climate change"`, this fn
// should return "chaos","climate change". or something.
// TODO: should really fall out of query parsing...
func (res *Results) HighlightTerms() string {
	// ultrashonky hack for now.
	q := res.Query
	stripFields := regexp.MustCompile("[a-zA-Z]+:")
	stripPunct := regexp.MustCompile("[^a-zA-Z0-9 ]+")
	q = stripFields.ReplaceAllLiteralString(q, "")
	q = stripPunct.ReplaceAllLiteralString(q, "")
	return q
}

func (res *Results) Match(artIdx int, needle string) bool {
	needle = strings.ToLower(needle)
	art := res.Art(artIdx)
	if strings.Contains(strings.ToLower(art.Headline), needle) {
		return true
	}
	if strings.Contains(strings.ToLower(art.CanonicalURL), needle) {
		return true
	}
	if strings.Contains(strings.ToLower(art.Published), needle) {
		return true
	}
	if strings.Contains(strings.ToLower(art.Pub), needle) {
		return true
	}
	if strings.Contains(strings.ToLower(art.TagsString()), needle) {
		return true
	}
	return false
}

func (res *Results) FindForward(artIdx int, needle string) int {
	for ; artIdx < len(res.arts); artIdx++ {
		if res.Match(artIdx, needle) {
			return artIdx
		}
	}
	return -1
}

func (res *Results) FindReverse(artIdx int, needle string) int {
	for ; artIdx >= 0; artIdx-- {
		if res.Match(artIdx, needle) {
			return artIdx
		}
	}
	return -1
}

func (res *Results) Art(idx int) *store.Article {
	if idx >= 0 && idx < len(res.arts) {
		return res.arts[idx]
	}

	// sometimes get here... seems to be tableview doing one last refresh on
	// old delegates before zapping/recycling them
	// TODO: investigate!
	//	dbug.Printf("bad idx: %d\n", idx)
	return &store.Article{Headline: fmt.Sprintf("<BAD> %d", idx)}
}

func (res *Results) Facet(idx int) *Facet {
	return res.facets[idx]
}

// returns new Results
// order: 1: ascending, 0: descending
func (res *Results) Sort(sortColumn string, sortOrder int) *Results {
	//dbug.Printf("new sorting: %d %d\n", sortColumn, sortOrder)

	sorted := make(store.ArtList, len(res.arts))
	copy(sorted, res.arts)

	var criteria func(a1, a2 *store.Article) bool

	if sortOrder == 0 {
		switch sortColumn {
		case "headline":
			criteria = func(a1, a2 *store.Article) bool { return a1.Headline > a2.Headline }
		case "pub":
			criteria = func(a1, a2 *store.Article) bool { return a1.Pub > a2.Pub }
		case "section":
			criteria = func(a1, a2 *store.Article) bool { return a1.Section > a2.Section }
		case "published":
			criteria = func(a1, a2 *store.Article) bool { return a1.Published > a2.Published }
		case "tags":
			criteria = func(a1, a2 *store.Article) bool { return a1.TagsString() > a2.TagsString() }
		case "byline":
			criteria = func(a1, a2 *store.Article) bool { return a1.Byline > a2.Byline }
		case "url":
			criteria = func(a1, a2 *store.Article) bool { return a1.URL() > a2.URL() }
		case "retweets":
			criteria = func(a1, a2 *store.Article) bool { return a1.Retweets > a2.Retweets }
		case "favourites":
			criteria = func(a1, a2 *store.Article) bool { return a1.Favourites > a2.Favourites }
		case "keywords":
			criteria = func(a1, a2 *store.Article) bool { return a1.KeywordsString() > a2.KeywordsString() }
		case "links":
			criteria = func(a1, a2 *store.Article) bool { return a1.LinksString() > a2.LinksString() }
		}
	} else if sortOrder == 1 {
		switch sortColumn {
		case "headline":
			criteria = func(a1, a2 *store.Article) bool { return a1.Headline < a2.Headline }
		case "pub":
			criteria = func(a1, a2 *store.Article) bool { return a1.Pub < a2.Pub }
		case "section":
			criteria = func(a1, a2 *store.Article) bool { return a1.Section < a2.Section }
		case "published":
			criteria = func(a1, a2 *store.Article) bool { return a1.Published < a2.Published }
		case "tags":
			criteria = func(a1, a2 *store.Article) bool { return a1.TagsString() < a2.TagsString() }
		case "byline":
			criteria = func(a1, a2 *store.Article) bool { return a1.Byline < a2.Byline }
		case "url":
			criteria = func(a1, a2 *store.Article) bool { return a1.URL() < a2.URL() }
		case "retweets":
			criteria = func(a1, a2 *store.Article) bool { return a1.Retweets < a2.Retweets }
		case "favourites":
			criteria = func(a1, a2 *store.Article) bool { return a1.Favourites < a2.Favourites }
		case "keywords":
			criteria = func(a1, a2 *store.Article) bool { return a1.KeywordsString() < a2.KeywordsString() }
		case "links":
			criteria = func(a1, a2 *store.Article) bool { return a1.LinksString() < a2.LinksString() }
		}
	}
	if criteria != nil {
		store.By(criteria).Sort(sorted)
	}

	return &Results{
		Query:    res.Query,
		arts:     sorted,
		Len:      len(sorted),
		facets:   res.facets, // facets don't change
		FacetLen: res.FacetLen,
	}
}
