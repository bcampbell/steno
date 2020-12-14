package sim

import (
	"fmt"
	//	"io/ioutil"
	"github.com/blevesearch/bleve/analysis"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/analysis/lang/es"
	"github.com/blevesearch/bleve/analysis/lang/ru"
	"github.com/blevesearch/bleve/registry"
	"hash"
	"hash/fnv"
	"strings"
)

type Hash uint64
type DocID uint32

type Index struct {
	// Docs holds a lists of hashes for each document.
	Docs map[DocID][]Hash
	// Hashes holds lists of documents for each hash.
	Hashes map[Hash][]DocID
}

//
type Indexer struct {
	Lang      string // language used for indexering
	NgramSize int
	cache     *registry.Cache
	analyser  *analysis.Analyzer
	hasher    hash.Hash64
}

// ngrams groups source words into ngrams of length n.
// Joins the words and returns each ngram as a single string.
// eg
// ngram(2, []string{"hello", "there", "bob"})
// => {"hello there", "there bob"}
func ngrams(n int, src []string) []string {
	cnt := (len(src) - n) + 1
	if cnt <= 0 {
		return []string{strings.Join(src, " ")}
	}
	out := make([]string, cnt)
	for i := 0; i < cnt; i++ {
		out[i] = strings.Join(src[i:i+n], " ")
	}
	return out
}

// UniqHashes removes duplicates from a list of hashes.
// Ordering is otherwise preserved.
func UniqHashes(src []Hash) []Hash {
	out := make([]Hash, 0, len(src))
	seen := map[Hash]struct{}{}
	for _, v := range src {
		if _, got := seen[v]; !got {
			seen[v] = struct{}{}
			out = append(out, v)
		}
	}
	return out
}

// NewIndex creates a new Index
func NewIndex() *Index {
	return &Index{Docs: make(map[DocID][]Hash), Hashes: make(map[Hash][]DocID)}
}

// NewIndexer creates a new Indexer.
func NewIndexer(ngramSize int, lang string) (*Indexer, error) {
	var err error
	indexer := &Indexer{}
	indexer.NgramSize = ngramSize
	indexer.Lang = lang
	indexer.cache = registry.NewCache()
	config := make(map[string]interface{})
	indexer.hasher = fnv.New64a()
	switch lang {
	case "en":
		indexer.analyser, err = en.AnalyzerConstructor(config, indexer.cache)
	case "ru":
		indexer.analyser, err = ru.AnalyzerConstructor(config, indexer.cache)
	case "es":
		indexer.analyser, err = es.AnalyzerConstructor(config, indexer.cache)
	default:
		return nil, fmt.Errorf("Unsupported language %s", lang)
	}
	if err != nil {
		return nil, err
	}

	return indexer, nil
}

// IndexDoc indexes a document and adds it to the target index.
// Assumes doc does not already exist in index!
func (indexer *Indexer) IndexDoc(targ *Index, docID DocID, txt string) {
	//	fmt.Fprintf(os.Stderr, "--- %d ---\n", docID)
	hashes := indexer.HashString(txt)
	hashes = UniqHashes(hashes)
	for _, hash := range hashes {
		targ.Hashes[hash] = append(targ.Hashes[hash], docID)
	}
	targ.Docs[docID] = hashes
}

// HashString tokenises a string and returns a list of hashed ngrams.
func (indexer *Indexer) HashString(txt string) []Hash {
	toks := indexer.analyser.Analyze([]byte(txt))
	strs := make([]string, len(toks))
	for i, tok := range toks {
		strs[i] = string(tok.Term)
	}

	frags := ngrams(indexer.NgramSize, strs)
	hashes := make([]Hash, len(frags))
	for i, frag := range frags {
		indexer.hasher.Reset()
		indexer.hasher.Write([]byte(frag))
		hashes[i] = Hash(indexer.hasher.Sum64())
		//		fmt.Printf("'%s' (0x%x)\n", frag, hashes[i])
	}
	return hashes
}
