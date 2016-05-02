package store

import (
	"fmt"
	"github.com/bcampbell/qs"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzers/custom_analyzer"
	"github.com/blevesearch/bleve/analysis/analyzers/simple_analyzer"
	"github.com/blevesearch/bleve/analysis/char_filters/zero_width_non_joiner"
	"github.com/blevesearch/bleve/analysis/token_filters/lower_case_filter"
	"github.com/blevesearch/bleve/analysis/tokenizers/regexp_tokenizer"
	"github.com/blevesearch/bleve/index/store/goleveldb"
	"regexp"
	"strconv"
	"time"
)

// article data massasged for bleve indexing
type bleveArt struct {
	ID        string    `json:"id"`
	Urls      []string  `json:"urls"`
	Headline  string    `json:"headline"`
	Content   string    `json:"content"`
	Published time.Time `json:"published"`
	Keywords  []string  `json:"keywords"`
	Section   string    `json:"section"`
	Tags      []string  `json:"tags"`
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
	loc  *time.Location
}

func newBleveIndex(dbug Logger, idxName string, loc *time.Location) (*bleveIndex, error) {
	bleve.Config.DefaultKVStore = goleveldb.Name

	indexMapping := bleve.NewIndexMapping()
	// need to do this for sensible handling of default fields ("_all" uses this)
	indexMapping.DefaultAnalyzer = "en"

	// add a custom tokenizer and analyzer for handling urls
	var err error
	err = indexMapping.AddCustomTokenizer("url_parts",
		map[string]interface{}{
			"regexp": `(\p{L}+)|([\d]+)|[\S]`,
			"type":   regexp_tokenizer.Name,
		})
	if err != nil {
		return nil, err
	}

	err = indexMapping.AddCustomAnalyzer("url",
		map[string]interface{}{
			"type": custom_analyzer.Name,
			"char_filters": []interface{}{
				zero_width_non_joiner.Name,
			},
			"tokenizer": `url_parts`,
			"token_filters": []interface{}{
				lower_case_filter.Name,
			},
		})
	if err != nil {
		return nil, err
	}

	artMapping := bleve.NewDocumentMapping()

	// english text - stemming, remove stopwords etc...
	textFld := bleve.NewTextFieldMapping()
	textFld.Analyzer = "en"
	textFld.Store = false

	// urls - split into words or one-char tokens
	// http://www.example.com/wibble/foo-bar-wibble.html
	// => [ http : / / www . example . com / wibble / foo - bar - wibble . html ]
	urlFld := bleve.NewTextFieldMapping()
	urlFld.Analyzer = "url"
	urlFld.Store = false

	// simple field - split by whitespace and lowercase, no stemming or stopwords
	simpleFld := bleve.NewTextFieldMapping()
	simpleFld.Analyzer = simple_analyzer.Name
	simpleFld.Store = false

	numFld := bleve.NewNumericFieldMapping()
	numFld.Store = false

	dateFld := bleve.NewDateTimeFieldMapping()
	dateFld.Store = false

	artMapping.AddFieldMappingsAt("urls", urlFld)
	artMapping.AddFieldMappingsAt("headline", textFld)
	artMapping.AddFieldMappingsAt("content", textFld)
	artMapping.AddFieldMappingsAt("published", dateFld)
	artMapping.AddFieldMappingsAt("keywords", simpleFld)
	artMapping.AddFieldMappingsAt("section", simpleFld)
	artMapping.AddFieldMappingsAt("tags", simpleFld)
	artMapping.AddFieldMappingsAt("retweets", numFld)
	artMapping.AddFieldMappingsAt("favourites", numFld)
	artMapping.AddFieldMappingsAt("links", urlFld)
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
		dbug: dbug,
		loc:  loc}

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

func openBleveIndex(dbug Logger, idxName string, loc *time.Location) (*bleveIndex, error) {
	bleve.Config.DefaultKVStore = goleveldb.Name

	index, err := bleve.Open(idxName)
	if err != nil {
		return nil, err
	}

	idx := &bleveIndex{
		idx:  index,
		dbug: dbug,
		loc:  loc,
	}

	return idx, nil
}

var varPat = regexp.MustCompile(`[$][{][_A-Z]+[}]`)

func (idx *bleveIndex) expandQuery(q string) string {
	now := time.Now().In(idx.loc)

	const layout = "2006-01-02"

	return varPat.ReplaceAllStringFunc(q, func(s string) string {
		switch s {
		case "${TODAY}":
			return fmt.Sprintf(`[%s TO %s]`, now.Format(layout), now.Format(layout))
		case "${PAST_WEEK}":
			return fmt.Sprintf(`[%s TO %s]`, now.AddDate(0, 0, -6).Format(layout), now.Format(layout))
		case "${PAST_MONTH}":
			return fmt.Sprintf(`[%s TO %s]`, now.AddDate(0, -1, 0).Format(layout), now.Format(layout))
		case "${PAST_YEAR}":
			return fmt.Sprintf(`[%s TO %s]`, now.AddDate(-1, 0, 0).Format(layout), now.Format(layout))
		}
		return q
	})
}

func (idx *bleveIndex) index(srcArts ...*Article) error {
	idx.dbug.Printf("bleve: indexing %d articles\n", len(srcArts))
	batch := idx.idx.NewBatch()

	for _, src := range srcArts {
		artID := strconv.Itoa(int(src.ID))

		pubTime, err := time.ParseInLocation(time.RFC3339, src.Published, idx.loc)
		if err != nil {
			idx.dbug.Printf("WARN: art %d: bad time '%s'\n", src.ID, src.Published)
			pubTime = time.Time{}
		}

		art := bleveArt{
			Urls:       src.URLs,
			Headline:   src.Headline,
			Content:    src.PlainTextContent(),
			Published:  pubTime,
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
	//idx.dbug.Printf("bleve: committing...\n")
	idx.idx.Batch(batch)
	//idx.dbug.Printf("bleve: done indexing\n")
	return nil
}

func (idx *bleveIndex) search(queryString string, order string) (ArtList, error) {

	if queryString == "" {
		return ArtList{}, nil
	}

	queryString = idx.expandQuery(queryString)
	fmt.Println(queryString)

	parser := &qs.Parser{}
	parser.DefaultOp = qs.AND
	parser.Loc = idx.loc

	q, err := parser.Parse(queryString)
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

	batchSize := 200
	start := 0
	for start < len(theDoomed) {
		end := start + batchSize
		if end > len(theDoomed) {
			end = len(theDoomed)
		}
		idx.dbug.Printf("bleve: delete %d articles (%d...%d)\n", end-start, start, end)
		batch := idx.idx.NewBatch()
		for _, id := range theDoomed[start:end] {
			artIDStr := strconv.Itoa(int(id))
			batch.Delete(artIDStr)
		}
		idx.dbug.Printf("bleve: committing...\n")
		idx.idx.Batch(batch)
		start = end
	}
	idx.dbug.Printf("bleve: done deleting\n")
	return nil
}
