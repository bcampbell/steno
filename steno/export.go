package steno

import (
	"encoding/csv"
	//	"fmt"
	"github.com/bcampbell/steno/steno/store"
	"io"
)

func exportCSV(arts store.ArtList, out io.Writer) error {

	fields := []string{"headline", "published", "tags", "byline", "url", "retweets", "favourites", "keywords", "links"}

	w := csv.NewWriter(out)

	// header
	err := w.Write(fields)
	if err != nil {
		return err
	}

	// rows
	/* XYZZY */
	/*
		for _, art := range arts {
			row := make([]string, len(fields))

			for i, fld := range fields {
				row[i] = art.FieldString(fld)
			}
			err := w.Write(row)
			if err != nil {
				return err
			}
		}
	*/
	w.Flush()

	return nil
}
