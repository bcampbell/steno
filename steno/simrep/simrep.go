package simrep

// generate a similarity report between two steno databases...

import (
	"github.com/bcampbell/sim"
	"github.com/bcampbell/steno/steno/store"
	"os"
	"strings"
)

type Opts struct {
	Verbose        bool
	MinWords       int
	NGramSize      int
	MatchThreshold float64
	IgnoreSameID   bool
	Lang           string
	//
	Dbug store.Logger
}

func Run(db1 *store.Store, db2 *store.Store, opts *Opts) error {

	dbug := opts.Dbug

	// index the first store
	dbug.Printf("building similarity index...\n")
	idx, err := BuildIndex(db1, opts, nil)
	if err != nil {
		return err
	}

	dbug.Printf("matching against index...\n")

	matches, err := FindMatches(idx, db2, opts)
	if err != nil {
		return err
	}

	// output report

	dbug.Printf("output report...\n")
	EmitHeader(os.Stdout, opts)
	// for each article...
	for art2ID, m := range matches {
		// fetch the article
		foo, err := db2.Fetch(art2ID)
		if err != nil {
			return err
		}
		art := foo[0]

		// fetch all the articles it matched
		matchArtIDs := make([]store.ArtID, len(m))
		metrics := make([]float64, len(m))
		for i, dm := range m {
			matchArtIDs[i] = store.ArtID(dm.ID)
			metrics[i] = dm.Factor
		}
		matchingArts, err := db1.Fetch(matchArtIDs...)
		if err != nil {
			return err
		}

		EmitMatches(os.Stdout, art, matchingArts, metrics)

	}

	EmitFooter(os.Stdout)
	return nil
}

func tidy(s string) string {
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if len(l) > 0 {
			out = append(out, l)
		}
	}

	return strings.Join(out, "\n")
}

func BuildIndex(db *store.Store, opts *Opts, progFunc func(int, int)) (*sim.Index, error) {
	idx, err := sim.NewIndex(opts.NGramSize, opts.Lang)
	if err != nil {
		return nil, err
	}
	cnt := 0
	tot := db.TotalArts()
	it := db.IterateAllArts()
	for it.Next() {
		art := it.Cur()
		txt := art.PlainTextContent()

		nWords := len(strings.Fields(txt))
		if nWords < opts.MinWords {
			continue
		}
		idx.AddDoc(sim.DocID(art.ID), txt)
		cnt++
		if progFunc != nil {
			progFunc(cnt, tot)
		}
	}
	if it.Err() != nil {
		return nil, it.Err()
	}
	idx.Finalise()
	return idx, nil
}

func FindMatches(idx *sim.Index, db *store.Store, opts *Opts) (map[store.ArtID][]sim.DocMatch, error) {
	matches := map[store.ArtID][]sim.DocMatch{}

	it := db.IterateAllArts()
	for it.Next() {
		art := it.Cur()
		txt := art.PlainTextContent()
		nWords := len(strings.Fields(txt))
		if nWords < opts.MinWords {
			continue
		}
		hits := idx.Match(txt, opts.MatchThreshold)
		if opts.IgnoreSameID {
			tmp := make([]sim.DocMatch, 0, len(hits))
			for _, dm := range hits {
				if store.ArtID(dm.ID) != art.ID {
					tmp = append(tmp, dm)
				}
			}
			hits = tmp
		}

		if len(hits) > 0 {
			matches[art.ID] = hits
		}
	}
	if it.Err() != nil {
		return nil, it.Err()
	}
	return matches, nil
}
