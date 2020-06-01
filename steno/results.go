package steno

import (
	"fmt"
	"github.com/bcampbell/steno/store"
	"regexp"
	"strings"
)

type Facet struct {
	Txt string
	Cnt int
}

type Results struct {
	Query string
	Arts  store.ArtList
	Len   int

	FacetLen int
	facets   []*Facet

	db *store.Store
	// cheesy-ass cache. Cacheing should probably be done inside store instead...
	hydrated map[store.ArtID]*store.Article
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

	res := Results{
		Query:    query,
		db:       db,
		hydrated: map[store.ArtID]*store.Article{},
	}
	res.setArts(arts)
	return &res, nil
}

func (res *Results) setArts(arts store.ArtList) {
	res.Arts = arts
	res.Len = len(res.Arts)

	/* XYZZY */
	/*
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
	*/
	//	arts.DumpAverages()

}

// return terms from current query, for highlighting
// eg if query is `chaos headline:"climate change"`, this fn
// should return "chaos","climate change". or something.
// TODO: should really fall out of query parsing...
func (res *Results) HighlightTerms() string {
	// ultrashonky hack for now.
	q := res.Query
	stripBool := regexp.MustCompile(`OR|AND`)
	stripTags := regexp.MustCompile(`tags:\s*[a-zA-Z]+`)
	stripFields := regexp.MustCompile("[a-zA-Z]+:")
	stripPunct := regexp.MustCompile("[^a-zA-Z0-9 ]+")
	q = stripBool.ReplaceAllLiteralString(q, "")
	q = stripTags.ReplaceAllLiteralString(q, "")
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
	for ; artIdx < len(res.Arts); artIdx++ {
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

// TODO: make db access explict! + proper error handling
func (res *Results) Art(idx int) *store.Article {
	if idx < 0 || idx >= len(res.Arts) {
		// sometimes get here... seems to be tableview doing one last refresh on
		// old delegates before zapping/recycling them
		// TODO: investigate!
		//	dbug.Printf("bad idx: %d\n", idx)
		return &store.Article{Headline: fmt.Sprintf("<BAD> %d", idx)}
	}

	artID := res.Arts[idx]
	art, got := res.hydrated[artID]
	if got {
		return art
	}
	// not in cache - fetch it!

	//dbug.Printf("fetch art %d\n", artID)
	fetchedArts, err := res.db.Fetch(artID)
	if err != nil {
		return &store.Article{Headline: fmt.Sprintf("<BAD> %d", idx)}
	}

	if len(fetchedArts) == 0 {
		return &store.Article{Headline: fmt.Sprintf("<Missing article %d>", artID)}
	}
	art = fetchedArts[0]
	// cache it
	res.hydrated[artID] = art
	return art
}

func (res *Results) Facet(idx int) *Facet {
	return res.facets[idx]
}

func (res *Results) Sort(sortColumn string, sortOrder int) *Results {
	ord := store.Ascending
	if sortOrder != 0 {
		ord = store.Descending
	}

	sorted, err := res.db.Sort(res.Arts, sortColumn, ord)
	if err != nil {
		//TODO: log error to dbug?
		return res
	}

	return &Results{
		Query:    res.Query,
		Arts:     sorted,
		Len:      len(sorted),
		facets:   res.facets, // facets don't change
		FacetLen: res.FacetLen,
		db:       res.db,
		hydrated: res.hydrated,
	}
}

// returns new Results
// order: 1: ascending, 0: descending
/*
func (res *Results) Sort(sortColumn string, sortOrder int) *Results {

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
*/
