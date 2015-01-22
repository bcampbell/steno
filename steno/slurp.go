package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Msg struct {
	Article *Article `json:"article,omitempty"`
	Error   string   `json:"error,omitempty"`
}

func Slurp(dayFrom, dayTo string) chan Msg {
	out := make(chan Msg)

	go func() {
		defer close(out)
		/*
			for i := 0; i < 10; i++ {
				n := fmt.Sprintf("%d", i)
				art := &Article{
					CanonicalURL: "http://wibble.com/art" + n,
					URLs:         []string{"http://wibble.com/art" + n},
					Headline:     "Article " + n,
					Content:      "blah blah blah",
					Published:    "2014-01-01",
					Updated:      "",
					Pub:          "foo",
				}
				out <- Msg{Article: art}
			}

			out <- Msg{Error: "POOOPY"}
		*/
		server := os.Getenv("STENO_SLURP_ADDR")
		if server == "" {
			server = "foo.scumways.com"
		}
		u := fmt.Sprintf("http://%s/all?from=%s&to=%s", server, dayFrom, dayTo)

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
