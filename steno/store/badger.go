package store

import (
	//	"sort"
	//"github.com/bcampbell/arts/arts"
	"github.com/bcampbell/badger"
	"github.com/bcampbell/badger/query"
)

type badgerIndex struct {
	coll *badger.Collection
}

func newBadgerIndex() *badgerIndex {
	idx := &badgerIndex{}
	idx.coll = badger.NewCollection(&Article{})
	idx.coll.SetWholeWordField("content")
	idx.coll.SetWholeWordField("headline")
	idx.coll.SetWholeWordField("tags")
	idx.coll.SetWholeWordField("byline")
	idx.coll.SetWholeWordField("pub")
	idx.coll.SetWholeWordField("section")
	idx.coll.SetWholeWordField("keywords")
	return idx
}

func (idx *badgerIndex) add(arts ...*Article) error {
	for _, art := range arts {
		idx.coll.Put(art)
	}
	return nil
}

func (idx *badgerIndex) search(queryString string, order string) (ArtList, error) {
	q, err := query.Parse(queryString, idx.coll.ValidFields(), defaultField)
	if err != nil {
		return nil, err
	}
	// TODO: fix badger so it's not so silly!
	if q == nil {
		return ArtList{}, nil
	}

	var arts []*Article
	idx.coll.Find(q, &arts)

	publishedDesc := func(a1, a2 *Article) bool {
		return a1.Published > a2.Published
	}
	By(publishedDesc).Sort(arts)

	out := make(ArtList, len(arts))
	for idx, art := range arts {
		out[idx] = art.ID
	}
	return out, nil
}
