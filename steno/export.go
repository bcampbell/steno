package main

import (
	"encoding/csv"
	//	"fmt"
	"io"
	"semprini/steno/steno/store"
	"strconv"
)

func getArtField(art *store.Article, field string) string {
	switch field {
	case "canonical_url":
		return art.CanonicalURL
		//TODO: URLs
	case "headline":
		return art.Headline
		//TODO: Authors
	case "content":
		return art.Content
	case "published":
		return art.Published
	// TODO: Updated, Publication, Keywords
	case "keywords":
		return art.KeywordsString()
	case "section":
		return art.Section
	case "tags":
		return art.TagsString()
	case "retweets":
		return strconv.Itoa(art.Retweets)
	case "favourites":
		return strconv.Itoa(art.Favourites)
	case "links":
		return art.LinksString()

		// assorted fudge feilds
	case "url":
		return art.URL()
	case "pub":
		return art.Pub
	case "byline":
		return art.Byline
	}

	return "?????"
}

func exportCSV(arts store.ArtList, out io.Writer) error {

	fields := []string{"headline", "published", "tags", "byline", "url", "retweets", "favourites", "keywords", "links"}

	w := csv.NewWriter(out)

	// header
	err := w.Write(fields)
	if err != nil {
		return err
	}

	// rows
	for _, art := range arts {
		row := make([]string, len(fields))

		for i, fld := range fields {
			row[i] = getArtField(art, fld)
		}
		err := w.Write(row)
		if err != nil {
			return err
		}
	}
	w.Flush()

	return nil
}
