package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"semprini/steno/steno/store"
)

type SlurpSource struct {
	Name string
	Loc  string
}

type Msg struct {
	Article *store.Article `json:"article,omitempty"`
	Error   string         `json:"error,omitempty"`
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

		resp, err := http.Get(u)
		if err != nil {
			out <- Msg{Error: fmt.Sprintf("HTTP Get failed: %s", err)}
			return
		}
		defer resp.Body.Close()

		dec := json.NewDecoder(resp.Body)
		for {
			var msg Msg
			if err := dec.Decode(&msg); err == io.EOF {
				break
			} else if err != nil {
				out <- Msg{Error: fmt.Sprintf("Decode error: %s", err)}
				return
			}

			// massage the data slightly...
			if msg.Article != nil {
				if msg.Article.CanonicalURL == "" && len(msg.Article.URLs) > 0 {
					msg.Article.CanonicalURL = msg.Article.URLs[0]
				}

				msg.Article.Pub = msg.Article.Publication.Code
				msg.Article.Byline = msg.Article.BylineString()
				// truncate date to day
				if len(msg.Article.Published) > 10 {
					// ugh :-)
					msg.Article.Published = msg.Article.Published[0:10]
				}
			}

			out <- msg
		}
	}()

	return out
}
