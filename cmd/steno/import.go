package main

import (
	"encoding/json"
	"github.com/bcampbell/scrapeomat/slurp"
	"github.com/bcampbell/steno/store"
	"io"
)

// ToStoreArt converts a wire-format article into the Store form.
func ToStoreArt(in *slurp.Article) *store.Article {
	out := &store.Article{
		CanonicalURL: in.CanonicalURL,
		URLs:         make([]string, len(in.URLs)),
		Headline:     in.Headline,
		Authors:      make([]store.Author, len(in.Authors)),
		Content:      in.Content,
		Published:    in.Published,
		Updated:      in.Updated,
		Publication: store.Publication{
			Code:   in.Publication.Code,
			Name:   in.Publication.Name,
			Domain: in.Publication.Domain,
		},
		Keywords:   make([]string, len(in.Keywords)),
		Section:    in.Section,
		Tags:       make([]string, len(in.Tags)),
		Retweets:   in.Extra.RetweetCount,
		Favourites: in.Extra.FavoriteCount,
		Links:      make([]string, len(in.Extra.Links)),
	}

	copy(out.URLs, in.URLs)
	copy(out.Tags, in.Tags)
	copy(out.Links, in.Extra.Links)

	for i, a := range in.Authors {
		out.Authors[i] = store.Author{
			Name:    a.Name,
			RelLink: a.RelLink,
			Email:   a.Email,
			Twitter: a.Twitter,
		}
	}
	for i, k := range in.Keywords {
		out.Keywords[i] = k.Name
	}

	if out.CanonicalURL == "" && len(out.URLs) > 0 {
		out.CanonicalURL = out.URLs[0]
	}

	out.Pub = out.Publication.Code
	out.Byline = out.BylineString()
	return out
}

// ImportFromJSON imports articles from a JSON object stream.
// It returns a list of IDs of articles which were added to the store.
func ImportFromJSON(db *store.Store, in io.Reader) (store.ArtList, error) {
	dec := json.NewDecoder(in)

	stasher := store.NewStasher(db)
	defer func() {
		stasher.Close()
	}()
	for dec.More() {
		wireFmtArt := &slurp.Article{}
		err := dec.Decode(wireFmtArt)
		if err != nil {
			return stasher.StashedIDs, err
		}

		art := ToStoreArt(wireFmtArt)

		got, err := db.FindArt(art.URLs)
		if err != nil {
			return stasher.StashedIDs, err
		}
		if got > 0 {
			// Skip this article. Already in DB.
			continue
		}

		err = stasher.Stash(art)
		if err != nil {
			return stasher.StashedIDs, err
		}
	}
	return stasher.StashedIDs, nil
}
