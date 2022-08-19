package main

import (
	"encoding/json"
	//"fmt"
	"github.com/bcampbell/scrapeomat/slurp"
	"github.com/bcampbell/steno/store"
	"io"
	//	"os"
	//	"time"
)

// ToSlurpArt converts a store article into the wire-format slurp article,
// with all the json encoding tags and whatnot.
func ToSlurpArt(in *store.Article) *slurp.Article {
	out := &slurp.Article{
		// NOTE: we discard any ID from the store article (that's DB-specific)
		CanonicalURL: in.CanonicalURL,
		URLs:         make([]string, len(in.URLs)),
		Headline:     in.Headline,
		Authors:      make([]slurp.Author, len(in.Authors)),
		Content:      in.Content,
		Published:    in.Published,
		Updated:      in.Updated,
		Publication: slurp.Publication{
			Code:   in.Publication.Code,
			Name:   in.Publication.Name,
			Domain: in.Publication.Domain,
		},
		Keywords: make([]slurp.Keyword, len(in.Keywords)),
		Section:  in.Section,
	}

	out.Extra.RetweetCount = in.Retweets
	out.Extra.FavoriteCount = in.Favourites
	out.Extra.Links = make([]string, len(in.Links))
	for i, link := range in.Links {
		out.Extra.Links[i] = link
	}

	for i, u := range in.URLs {
		out.URLs[i] = u
	}
	for i, a := range in.Authors {
		out.Authors[i] = slurp.Author{
			Name:    a.Name,
			RelLink: a.RelLink,
			Email:   a.Email,
			Twitter: a.Twitter,
		}
	}
	for i, k := range in.Keywords {
		out.Keywords[i] = slurp.Keyword{Name: k, URL: ""}
	}

	if out.CanonicalURL == "" && len(out.URLs) > 0 {
		out.CanonicalURL = out.URLs[0]
	}

	// truncate date to day
	/*
		if len(out.Published) > 10 {
			// ugh :-)
			out.Published = msg.Article.Published[0:10]
		}
	*/

	// TODO: TAGS!!!!
	return out
}

// ExportToJSON writes out the specified articles into a JSON file.
// TODO: streamed objects or collection?
// TODO: should move this function into the store? Or a new package which uses
// store and slurp.Article?
func ExportToJSON(db *store.Store, artIDs store.ArtList, out io.Writer) error {
	var err error
	enc := json.NewEncoder(out)
	iter := db.IterateArts(artIDs...)
	for iter.Next() {
		art := ToSlurpArt(iter.Cur())
		err = enc.Encode(art)
		if err != nil {
			return err
		}
	}

	// any errors while reading?
	err = iter.Err()
	if err != nil {
		return err
	}
	return nil
}
