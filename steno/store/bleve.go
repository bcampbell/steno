package store

import (
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
	idx  bleve.Index
	dbug Logger
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
		idx:  index,
		dbug: dbug}

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
		idx:  index,
		dbug: dbug}

	return idx, nil
}

func (idx *bleveIndex) index(srcArts ...*Article) error {
	idx.dbug.Printf("bleve: indexing %d articles\n", len(srcArts))
	batch := idx.idx.NewBatch()
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
	idx.dbug.Printf("bleve: committing...\n")
	idx.idx.Batch(batch)
	idx.dbug.Printf("bleve: done indexing\n")
	return nil
}

func (idx *bleveIndex) search(queryString string, order string) (ArtList, error) {
	if queryString == "" {
		return ArtList{}, nil
	}
	q := bleve.NewQueryStringQuery(queryString)
	// TODO: improve upon kludgy max size
	req := bleve.NewSearchRequestOptions(q, 1000000, 0, false)
	results, err := idx.idx.Search(req)
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

func (idx *bleveIndex) zap(theDoomed ...ArtID) error {
	idx.dbug.Printf("bleve: delete %d articles\n", len(theDoomed))
	batch := idx.idx.NewBatch()
	for _, id := range theDoomed {
		artIDStr := strconv.Itoa(int(id))
		batch.Delete(artIDStr)
	}
	idx.dbug.Printf("bleve: committing...\n")
	idx.idx.Batch(batch)
	idx.dbug.Printf("bleve: done deleting\n")
	return nil
}
