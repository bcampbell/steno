package sim

type DocMatch struct {
	ID    DocID
	Score float64
}

/*
// Match returns a list of document IDs matching txt, within the given threshold.
// The threshold is the proportion of matching ngrams (0..1), where 1 is
// perfect match.
func (index *Index) Match(txt string, threshold float64) []DocMatch {
	hashes := index.HashString(txt)
	hashes = UniqHashes(hashes)
	return index.Match(hashes, threshold)
	raw := map[DocID]int{}
	for _, h := range hashes {
		for _, id := range index.lookup[h] {
			raw[id]++
		}
	}

	out := []DocMatch{}
	for id, cnt := range raw {
		f := float64(cnt) / float64(len(hashes))
		if f > threshold {
			out = append(out, DocMatch{id, f})
		}
	}
	return out
}
*/

// Match finds documents containing the given hashes.
// Matches below the given threshold factor are ignored.
func (index *Index) Match(hashes []Hash, threshold float64) []DocMatch {
	raw := map[DocID]int{}
	for _, h := range hashes {
		// find all the docs containing this hash
		for _, id := range index.Hashes[h] {
			raw[id]++
		}
	}

	out := []DocMatch{}
	for id, cnt := range raw {
		// percentage of matched hashes
		score := float64(cnt) / float64(len(hashes))
		if score >= threshold {
			out = append(out, DocMatch{id, score})
		}
	}
	return out
}
