package sim

/*
// routines to load/save index

import (
	"encoding/gob"
	"io"
	"os"
)

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

	index, err := NewIndex(ngramSize, lang)
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
*/
