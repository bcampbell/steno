package sim

type DocMatch struct {
	ID     DocID
	Factor float64
}

// Match returns a list of document IDs matching txt, within the given threshold.
// The threshold is the proportion of matching ngrams (0..1), where 1 is
// perfect match.
func (index *Index) Match(txt string, threshold float64) []DocMatch {
	hashes := index.HashString(txt)
	hashes = UniqHashes(hashes)
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
