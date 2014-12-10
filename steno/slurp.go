package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

		//		u := fmt.Sprintf("http://localhost:12345/all?from=%s&to=%s", dayFrom, dayTo)
		u := fmt.Sprintf("http://foo.scumways.com/all?from=%s&to=%s", dayFrom, dayTo)

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
				msg.Article.Pub = msg.Article.Publication.Code
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
