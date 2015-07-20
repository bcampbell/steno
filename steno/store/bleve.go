package store

import (
	"fmt"
	//"github.com/bcampbell/arts/arts"
	"github.com/blevesearch/bleve"
	"strconv"
)

type bleveIndex struct {
	index bleve.Index
	dbug  Logger
}

func newBleveIndex(dbug Logger) (*bleveIndex, error) {
	mapping := bleve.NewIndexMapping()
	index, err := bleve.New("/tmp/example.bleve", mapping)
	if err != nil {
		return nil, err
	}

	idx := &bleveIndex{
		index: index,
		dbug:  dbug}

	/*
		idx.coll.SetWholeWordField("content")
		idx.coll.SetWholeWordField("headline")
		idx.coll.SetWholeWordField("tags")
		idx.coll.SetWholeWordField("byline")
		idx.coll.SetWholeWordField("pub")
		idx.coll.SetWholeWordField("section")
		idx.coll.SetWholeWordField("keywords")
	*/
	return idx, nil
}

func (idx *bleveIndex) add(arts ...*Article) error {
	fmt.Printf("start bleve indexing...\n")
	batch := idx.index.NewBatch()
	for _, art := range arts {
		idStr := strconv.Itoa(int(art.ID))
		batch.Index(idStr, art)
	}
	fmt.Printf("committing...\n")
	idx.index.Batch(batch)
	fmt.Printf("done bleve indexing.\n")
	return nil
}

func (idx *bleveIndex) search(queryString string, order string) (ArtList, error) {
	if queryString == "" {
		return ArtList{}, nil
	}
	q := bleve.NewQueryStringQuery(queryString)
	req := bleve.NewSearchRequest(q)
	results, err := idx.index.Search(req)
	if err != nil {
		return ArtList{}, err
	}

	// TODO: order results

	out := make(ArtList, len(results.Hits))
	for idx, doc := range results.Hits {
		var id int
		id, err := strconv.Atoi(doc.ID)
		if err != nil {
			return ArtList{}, err
		}
		out[idx] = ArtID(id)
	}
	return out, nil
}
