package sim

import (
	"fmt"
	//	"io/ioutil"
	"encoding/gob"
	"github.com/blevesearch/bleve/analysis"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/analysis/lang/es"
	"github.com/blevesearch/bleve/analysis/lang/ru"
	"github.com/blevesearch/bleve/registry"
	"hash"
	"hash/fnv"
	"io"
	"os"
	"strings"
)

type Hash uint64
type DocID uint32
type Lookup map[Hash][]DocID

type Index struct {
	DocSource string
	Lang      string // language used for indexing
	NgramSize int
	cache     *registry.Cache
	analyser  *analysis.Analyzer
	lookup    Lookup

	hasher hash.Hash64
}

func (index *Index) Write(w io.Writer) error {
	enc := gob.NewEncoder(w)
	err := enc.Encode(index.DocSource)
	if err != nil {
		return err
	}
	err = enc.Encode(index.NgramSize)
	if err != nil {
		return err
	}
	err = enc.Encode(index.Lang)
	if err != nil {
		return err
	}
	err = enc.Encode(index.lookup)
	return err
}

func ReadIndex(r io.Reader) (*Index, error) {
	dec := gob.NewDecoder(r)
	var err error
	var docSource, lang string
	var ngramSize int

	err = dec.Decode(&docSource)
	if err != nil {
		return nil, err
	}
	err = dec.Decode(&ngramSize)
	if err != nil {
		return nil, err
	}
	err = dec.Decode(&lang)
	if err != nil {
		return nil, err
	}

	index, err := NewIndex(ngramSize, lang) // TODO: include ngramsize in index!!!!!
	if err != nil {
		return nil, err
	}
	index.DocSource = docSource

	err = dec.Decode(&index.lookup)
	if err != nil {
		return nil, err
	}
	return index, nil
}

func LoadIndex(filename string) (*Index, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadIndex(f)
}

func (index *Index) Save(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return index.Write(f)
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

/*
// helper: remove duplicates
func uniqStrings(src []string) []string {
	out := make([]string, 0, len(src))
	seen := map[string]struct{}{}
	for _, s := range src {
		if _, got := seen[s]; !got {
			lookup[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}
*/

// helper: remove duplicates
func uniqHashes(src []Hash) []Hash {
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

func (index *Index) AddDoc(docID DocID, txt string) {
	//	fmt.Fprintf(os.Stderr, "--- %d ---\n", docID)
	hashes := index.HashDoc(txt)
	hashes = uniqHashes(hashes)
	for _, hash := range hashes {
		index.lookup[hash] = append(index.lookup[hash], docID)
	}
}

func (index *Index) HashDoc(txt string) []Hash {
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

type DocMatch struct {
	ID     DocID
	Factor float64
}

func (index *Index) Match(txt string, threshold float64) []DocMatch {
	hashes := index.HashDoc(txt)
	hashes = uniqHashes(hashes)
	raw := map[DocID]int{}
	for _, h := range hashes {
		for _, id := range index.lookup[h] {
			raw[id]++
		}
	}

	out := []DocMatch{}
	for id, cnt := range raw {
		f := float64(cnt) / float64(len(hashes))
		if f > threshold {
			out = append(out, DocMatch{id, f})
		}
	}
	return out

}
