package main

import (
	"fmt"
	"github.com/bcampbell/steno/sim"
	"github.com/bcampbell/steno/store"
	"strings"
)

const minWords = 10

// BuildSimilarity populates the 'similar' table in the store.
func BuildSimilarity(db *store.Store, progressFn func(int, int, string) bool, threshold float64) error {

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
			cancel := progressFn(cnt, tot, "Building Index")
			if cancel {
				return fmt.Errorf("Canceled")
			}
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
		matches := idx.Match(hashes, threshold)

		// need to convert structs and filter out self-matches
		filteredMatches := make([]store.Match, 0, len(matches))
		for _, match := range matches {
			if match.ID == docID {
				continue
			}
			converted := store.Match{ID: store.ArtID(match.ID), Score: float32(match.Score)}
			filteredMatches = append(filteredMatches, converted)
		}

		err = batch.AddSimilar(artID, filteredMatches)
		if err != nil {
			batch.Rollback()
			return err
		}
		if progressFn != nil {
			cancel := progressFn(cnt, tot, "Calculating matches")
			if cancel {
				batch.Rollback()
				return fmt.Errorf("Canceled")
			}
		}
		cnt++
	}

	err = batch.Commit()
	if err != nil {
		return err
	}
	return nil
}
