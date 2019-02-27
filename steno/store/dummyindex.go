package store

import (
	"fmt"
)

// dummyIndex is a do-nothing implementation of indexer interface
type dummyIndex struct{}

func (idx *dummyIndex) index(...*Article) error {
	return nil
}

func (idx *dummyIndex) zap(...ArtID) error {
	return nil
}
func (idx *dummyIndex) search(string, string) (ArtList, error) {
	return nil, fmt.Errorf("No fulltext index")
}
func (idx *dummyIndex) Close() error {
	return nil
}
