package main

import (
	"github.com/bcampbell/steno/sim"
	"github.com/bcampbell/steno/store"
	"strings"
)

const minWords = 10

// BuildSimilarity populates the 'similar' table in the store.
func BuildSimilarity(db *store.Store, progressFn func(int, int, string)) error {

	// part one - build up index of all documents in the store

	idx := sim.NewIndex()

	indexer, err := sim.NewIndexer(2, "en" /*TODO: store.Lang */)
	if err != nil {
		return err
	}
	cnt := 0
	tot := db.TotalArts()
	it := db.IterateAllArts()
	for it.Next() {
		art := it.Cur()
		txt := art.PlainTextContent()

		nWords := len(strings.Fields(txt))
		if nWords < minWords {
			continue
		}
		indexer.IndexDoc(idx, sim.DocID(art.ID), txt)
		cnt++

		//		if progFunc != nil {
		//			progFunc(cnt, tot)
		//		}
		if progressFn != nil {
			progressFn(cnt, tot, "indexing")
		}
	}
	if it.Err() != nil {
		return it.Err()
	}

	// part two - compare document hashes and load matches into DB

	batch, err := db.BeginBatch()
	err = batch.ClearSimilar()
	if err != nil {
		batch.Rollback()
		return err
	}
	cnt = 0
	for docID, hashes := range idx.Docs {
		artID := store.ArtID(docID)
		matches := idx.Match(hashes, 0.80)

		// need to filter out self-matching!
		filteredMatches := make([]sim.DocMatch, 0, len(matches))
		for _, match := range matches {
			if match.ID != docID {
				filteredMatches = append(filteredMatches, match)
			}
		}

		err = batch.AddSimilar(artID, filteredMatches)
		if err != nil {
			batch.Rollback()
			return err
		}
		if progressFn != nil {
			progressFn(cnt, tot, "matching")
		}
		cnt++
	}

	err = batch.Commit()
	if err != nil {
		return err
	}
	return nil
}
