package store

import (
	"fmt"
	//"github.com/bcampbell/arts/arts"
	"github.com/blevesearch/bleve"
	"strconv"
)

// article data massasged for bleve indexing
type bleveArt struct {
	ID         string   `json:"id"`
	Urls       []string `json:"urls"`
	Headline   string   `json:"headline"`
	Content    string   `json:"content"`
	Published  string   `json:"published"`
	Keywords   []string `json:"keywords"`
	Section    string   `json:"section"`
	Tags       []string `json:"tags"`
	Retweets   int      `json:"retweets"`
	Favourites int      `json:"favourites"`
	Links      []string `json:"links"`

	// fudge fields
	Pub    string `json:"pub"`
	Byline string `json:"byline"`
}

type bleveIndex struct {
	index bleve.Index
	dbug  Logger
}

func newBleveIndex(dbug Logger, idxName string) (*bleveIndex, error) {
	indexMapping := bleve.NewIndexMapping()

	artMapping := bleve.NewDocumentMapping()

	textFieldMappings := map[string]*bleve.FieldMapping{
		"urls":       bleve.NewTextFieldMapping(),
		"Headline":   bleve.NewTextFieldMapping(),
		"content":    bleve.NewTextFieldMapping(),
		"published":  bleve.NewTextFieldMapping(),
		"keywords":   bleve.NewTextFieldMapping(),
		"section":    bleve.NewTextFieldMapping(),
		"tags":       bleve.NewTextFieldMapping(),
		"retweets":   bleve.NewTextFieldMapping(),
		"favourites": bleve.NewTextFieldMapping(),
		"links":      bleve.NewTextFieldMapping(),
		"pub":        bleve.NewTextFieldMapping(),
		"byline":     bleve.NewTextFieldMapping(),
	}

	for name, fm := range textFieldMappings {
		fm.Analyzer = "en"
		fm.Store = false
		artMapping.AddFieldMappingsAt(name, fm)
	}

	indexMapping.DefaultType = "article" //artMapping
	indexMapping.AddDocumentMapping("article", artMapping)

	index, err := bleve.New(idxName, indexMapping)
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

func openBleveIndex(dbug Logger, idxName string) (*bleveIndex, error) {
	index, err := bleve.Open(idxName)
	if err != nil {
		return nil, err
	}

	idx := &bleveIndex{
		index: index,
		dbug:  dbug}

	return idx, nil
}

func (idx *bleveIndex) add(srcArts ...*Article) error {
	fmt.Printf("start bleve indexing...\n")
	batch := idx.index.NewBatch()
	for _, src := range srcArts {
		artID := strconv.Itoa(int(src.ID))
		art := bleveArt{
			Urls:       src.URLs,
			Headline:   src.Headline,
			Content:    src.PlainTextContent(),
			Published:  src.Published,
			Keywords:   src.Keywords,
			Section:    src.Section,
			Tags:       src.Tags,
			Retweets:   src.Retweets,
			Favourites: src.Favourites,
			Links:      src.Links,
			Pub:        src.Pub,
			Byline:     src.BylineString(),
		}

		batch.Index(artID, art)
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
	// TODO: improve upon kludgy max size
	req := bleve.NewSearchRequestOptions(q, 1000000, 0, false)
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
