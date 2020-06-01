package steno

import (
	"encoding/csv"
	"fmt"
	"github.com/bcampbell/scrapeomat/slurp"
	"github.com/bcampbell/steno/store"
	"io"
	"os"
	"time"
)

type SlurpSource struct {
	Name string
	Loc  string
}

/*
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
*/

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

// returns IDs of sucessfully-added articles
func Slurp(db *store.Store, server *SlurpSource, timeFrom, timeTo time.Time, progressFn func(fetchedCnt int, expectedCnt int, newCnt int, msg string)) (store.ArtList, error) {
	newlySlurped := store.ArtList{}

	slurper := slurp.NewSlurper(server.Loc)

	filt := &slurp.Filter{
		PubFrom: timeFrom,
		PubTo:   timeTo,
	}

	progressFn(0, 0, 0, "Fetching count...")
	totalCnt, err := slurper.FetchCount(filt)
	if err != nil {
		return newlySlurped, err
	}

	progressFn(0, totalCnt, 0, "Slurping...")
	stream := slurper.Slurp2(filt)
	defer stream.Close()

	batchSize := 200

	newCnt := 0
	receivedCnt := 0
	done := false
	for {
		// all done?
		if done {
			break
		}
		// read a batch of articles in from the wire...
		arts := []*store.Article{}
		for i := 0; i < batchSize; i++ {
			wireArt, err := stream.Next()
			if err != nil {
				if err == io.EOF {
					done = true
					break
				} else {
					// uhoh.
					return newlySlurped, err
				}
			}
			art := convertArt(wireArt)
			arts = append(arts, art)
			receivedCnt += 1
		}

		// check which articles are new
		newArts := []*store.Article{}
		for _, art := range arts {
			got, err := db.FindArt(art.URLs)
			if err != nil {
				return newlySlurped, fmt.Errorf("FindArt() failed: %s", err)
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
				return newlySlurped, fmt.Errorf("Stash failed: %s", err)
			}
			// Stash will have assigned article IDs
			for _, a := range newArts {
				newlySlurped = append(newlySlurped, a.ID)
			}

		}
		//dbug.Printf("stashed %s as %d\n", art.Headline, art.ID)
		// TODO: not right, but hey
		newCnt += len(newArts)
		//		progressFn(fmt.Sprintf("Received %d (%d new)", receivedCnt, newCnt))
		progressFn(receivedCnt, totalCnt, newCnt, "Slurping...")
	}
	return newlySlurped, nil
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
