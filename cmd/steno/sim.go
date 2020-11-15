package main

import (
	"github.com/bcampbell/steno/sim"
	"github.com/bcampbell/steno/store"
	"strings"
)

const minWords = 10

func BuildSimilarity(store *store.Store) (*sim.Index, error) {
	idx, err := sim.NewIndex(2, "en" /*TODO: store.Lang */)
	if err != nil {
		return nil, err
	}
	cnt := 0
	//tot := store.TotalArts()
	it := store.IterateAllArts()
	for it.Next() {
		art := it.Cur()
		txt := art.PlainTextContent()

		nWords := len(strings.Fields(txt))
		if nWords < minWords {
			continue
		}
		idx.AddDoc(sim.DocID(art.ID), txt)
		cnt++

		//		if progFunc != nil {
		//			progFunc(cnt, tot)
		//		}
	}
	if it.Err() != nil {
		return nil, it.Err()
	}
	idx.Finalise()
	return idx, nil
}
