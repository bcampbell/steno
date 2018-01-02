package steno

import (
	"encoding/csv"
	"fmt"
	"os"
	"semprini/scrapeomat/slurp"
	"semprini/steno/steno/store"
	"time"
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

func Slurp(db *store.Store, server *SlurpSource, timeFrom, timeTo time.Time, progress *Progress) error {

	slurper := slurp.NewSlurper(server.Loc)

	filt := &slurp.Filter{
		PubFrom: timeFrom,
		PubTo:   timeTo,
	}

	stream, cancel := slurper.Slurp(filt)

	batchSize := 200

	newCnt := 0
	receivedCnt := 0
	for {
		// read a batch of articles in from the wire...
		arts := []*store.Article{}
		for i := 0; i < batchSize; i++ {
			msg, ok := <-stream

			if !ok {
				break
			}

			// handle errors
			if msg.Error != "" {
				//cancel <- struct{}{} // TODO: this isn't enough.
				return fmt.Errorf("Slurp error from server: %s", msg.Error)
			}
			if msg.Article == nil {
				dbug.Printf("Slurp WARN: missing article\n")
				continue
			}

			art := convertArt(msg.Article)
			arts = append(arts, art)
			receivedCnt += 1
		}

		// empty batch? all done?
		if len(arts) == 0 {
			break
		}

		// check which articles are new
		newArts := []*store.Article{}
		for _, art := range arts {
			got, err := db.FindArt(art.URLs)
			if err != nil {
				cancel <- struct{}{} // TODO: this isn't enough.
				return fmt.Errorf("FindArt() failed: %s", err)
			}
			if got > 0 {
				// already got it.
				continue
			}
			newArts = append(newArts, art)
		}

		// stash the new articles
		if len(newArts) > 0 {
			err := db.Stash(newArts)
			if err != nil {
				cancel <- struct{}{} // TODO: this isn't enough.
				return fmt.Errorf("Stash failed: %s", err)
			}
		}
		//dbug.Printf("stashed %s as %d\n", art.Headline, art.ID)
		// TODO: not right, but hey
		newCnt += len(newArts)
		progress.SetStatus(fmt.Sprintf("Received %d (%d new)", receivedCnt, newCnt))
	}
	return nil
}

// convert the wire-format article into our local form
func convertArt(in *slurp.Article) *store.Article {
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
