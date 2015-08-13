package store

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/qs"
	"strconv"
)

// article data massasged for bleve indexing
type bleveArt struct {
	ID       string   `json:"id"`
	Urls     []string `json:"urls"`
	Headline string   `json:"headline"`
	Content  string   `json:"content"`
	// TODO: this needs to be a time.Time!
	Published string   `json:"published"`
	Keywords  []string `json:"keywords"`
	Section   string   `json:"section"`
	Tags      []string `json:"tags"`
	// note: bleve only indexes float64 numeric fields
	Retweets   float64  `json:"retweets"`
	Favourites float64  `json:"favourites"`
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
	bleve.Config.DefaultKVStore = "goleveldb"

	indexMapping := bleve.NewIndexMapping()

	artMapping := bleve.NewDocumentMapping()

	// english text - stemming etc...
	textFld := bleve.NewTextFieldMapping()
	textFld.Analyzer = "en"
	textFld.Store = false

	// simple field - split by whitespace and lowercase
	simpleFld := bleve.NewTextFieldMapping()
	simpleFld.Analyzer = "simple"
	simpleFld.Store = false

	numFld := bleve.NewNumericFieldMapping()
	numFld.Store = false

	dateFld := bleve.NewDateTimeFieldMapping()
	dateFld.Store = false

	artMapping.AddFieldMappingsAt("urls", textFld)
	artMapping.AddFieldMappingsAt("headline", textFld)
	artMapping.AddFieldMappingsAt("content", textFld)
	artMapping.AddFieldMappingsAt("published", dateFld)
	artMapping.AddFieldMappingsAt("keywords", simpleFld)
	artMapping.AddFieldMappingsAt("section", simpleFld)
	artMapping.AddFieldMappingsAt("tags", simpleFld)
	artMapping.AddFieldMappingsAt("retweets", numFld)
	artMapping.AddFieldMappingsAt("favourites", numFld)
	artMapping.AddFieldMappingsAt("links", textFld)
	artMapping.AddFieldMappingsAt("pub", simpleFld)
	artMapping.AddFieldMappingsAt("byline", textFld)

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
	bleve.Config.DefaultKVStore = "goleveldb"

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
			Retweets:   float64(src.Retweets),
			Favourites: float64(src.Favourites),
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

	//	q := bleve.NewQueryStringQuery(queryString)
	//	q.MustMatch = true

	q, err := qs.Parse(queryString)
	if err != nil {
		return ArtList{}, err
	}

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
