package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"semprini/scrapeomat/slurp"
	"semprini/steno/steno/store"
)

type SlurpSource struct {
	Name string
	Loc  string
}

// article format we expect down the wire from the slurp API
type wireFmtArt struct {
	slurp.Article
	// extra fields from twitcooker
	Extra struct {
		RetweetCount  int `json:"retweet_count,omitempty"`
		FavoriteCount int `json:"favorite_count,omitempty"`
		// resolved links
		Links []string `json:"links,omitempty"`
	} `json:"extra,omitempty"`
}

// article or error, cooked and ready for steno
type Msg struct {
	Article *store.Article
	Error   string
}

func LoadSlurpSources(fileName string) ([]SlurpSource, error) {
	srcs := []SlurpSource{}

	inFile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer inFile.Close()
	in := csv.NewReader(inFile)
	rows, err := in.ReadAll()
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		srcs = append(srcs, SlurpSource{Name: row[0], Loc: row[1]})
	}

	return srcs, nil
}

func Slurp(server SlurpSource, dayFrom, dayTo string) chan Msg {
	out := make(chan Msg)

	go func() {
		defer close(out)
		u := fmt.Sprintf("%s/api/slurp?from=%s&to=%s", server.Loc, dayFrom, dayTo)

		fmt.Println(u)
		resp, err := http.Get(u)
		if err != nil {
			out <- Msg{Error: fmt.Sprintf("HTTP Get failed: %s", err)}
			return
		}
		defer resp.Body.Close()

		dec := json.NewDecoder(resp.Body)
		for {
			// read in a message from the wire
			var msg struct {
				Article *wireFmtArt `json:"article,omitempty"`
				Error   string      `json:"error,omitempty"`
			}
			if err := dec.Decode(&msg); err == io.EOF {
				break
			} else if err != nil {
				out <- Msg{Error: fmt.Sprintf("Decode error: %s", err)}
				return
			}

			cooked := Msg{}

			if msg.Error != "" {
				cooked.Error = msg.Error
			}

			if msg.Article != nil {
				cooked.Article = convertArt(msg.Article)
			}

			out <- cooked
		}
	}()

	return out
}

// convert the wire-format article into our local form
func convertArt(in *wireFmtArt) *store.Article {
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
		Keywords: make([]string, len(in.Keywords)),
		Section:  in.Section,

		Retweets:   in.Extra.RetweetCount,
		Favourites: in.Extra.FavoriteCount,
		Links:      make([]string, len(in.Extra.Links)),
	}

	for i, u := range in.URLs {
		out.URLs[i] = u
	}
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
	for i, l := range in.Extra.Links {
		out.Links[i] = l
	}

	if out.CanonicalURL == "" && len(out.URLs) > 0 {
		out.CanonicalURL = out.URLs[0]
	}

	out.Pub = out.Publication.Code
	out.Byline = out.BylineString()
	// truncate date to day
	/*
		if len(out.Published) > 10 {
			// ugh :-)
			out.Published = msg.Article.Published[0:10]
		}
	*/
	return out
}
