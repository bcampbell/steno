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
type Lookup map[Hash][]DocID

//
type Index struct {
	DocSource string
	Lang      string // language used for indexing
	NgramSize int
	cache     *registry.Cache
	analyser  *analysis.Analyzer
	lookup    Lookup

	hasher hash.Hash64
}

// group src words into ngrams of length n.
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

// NewIndex creates a new Index.
func NewIndex(ngramSize int, lang string) (*Index, error) {
	var err error
	index := &Index{}
	index.NgramSize = ngramSize
	index.Lang = lang
	index.cache = registry.NewCache()
	config := make(map[string]interface{})
	index.lookup = make(Lookup)
	index.hasher = fnv.New64a()
	switch lang {
	case "en":
		index.analyser, err = en.AnalyzerConstructor(config, index.cache)
	case "ru":
		index.analyser, err = ru.AnalyzerConstructor(config, index.cache)
	case "es":
		index.analyser, err = es.AnalyzerConstructor(config, index.cache)
	default:
		return nil, fmt.Errorf("Unsupported language %s", lang)
	}
	if err != nil {
		return nil, err
	}
	return index, nil
}

// AddDoc indexes a document and adds it to the index.
func (index *Index) AddDoc(docID DocID, txt string) {
	//	fmt.Fprintf(os.Stderr, "--- %d ---\n", docID)
	hashes := index.HashString(txt)
	hashes = UniqHashes(hashes)
	for _, hash := range hashes {
		index.lookup[hash] = append(index.lookup[hash], docID)
	}
}

// HashString tokenises a string and returns a list of hashed ngrams.
func (index *Index) HashString(txt string) []Hash {
	toks := index.analyser.Analyze([]byte(txt))
	strs := make([]string, len(toks))
	for i, tok := range toks {
		strs[i] = string(tok.Term)
	}

	frags := ngrams(index.NgramSize, strs)
	hashes := make([]Hash, len(frags))
	for i, frag := range frags {
		index.hasher.Reset()
		index.hasher.Write([]byte(frag))
		hashes[i] = Hash(index.hasher.Sum64())
		//		fmt.Printf("'%s' (0x%x)\n", frag, hashes[i])
	}
	return hashes
}

func (index *Index) Finalise() {
	// TODO: dedupe hash lists?
}

func (index *Index) Dump() {
	fmt.Println(index.DocSource)
	for hash, docIDs := range index.lookup {
		fmt.Printf("%d 0x%x\n", len(docIDs), hash)
	}
}
