package store

// TODO: Ditch this in favour of generic Batch mechanism!

// Stasher is a helper for loading articles into a store.
// It batches up stashes, to improve indexing speed.
// TODO:
// - add logging for bad/partly bad articles
// - track articles we already had
type Stasher struct {
	Dest      *Store
	BatchSize int
	arts      []*Article
	// StashedIDs contains IDs of successfully-stashed articles
	StashedIDs ArtList
}

func NewStasher(dest *Store) *Stasher {
	return &Stasher{
		Dest:      dest,
		BatchSize: 200,
	}
}

// Stash submits another article to the store.
// When the article is added (and it might be during
// a subsequent call - remember, articles are batched)
// it's ID field will be set.
func (s *Stasher) Stash(art *Article) error {
	s.arts = append(s.arts, art)
	if len(s.arts) >= s.BatchSize {
		return s.flush()
	}
	return nil
}

func (s *Stasher) flush() error {
	err := s.Dest.Stash(s.arts)
	if err != nil {
		return err
	}
	for _, art := range s.arts {
		s.StashedIDs = append(s.StashedIDs, art.ID)
	}
	s.arts = nil
	return nil
}

// Flushes any batched articles
// Safe to call multiple times
func (s *Stasher) Close() error {
	return s.flush()
}
