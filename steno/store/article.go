package store

import (
	//	"fmt"
	"github.com/bcampbell/htmlutil"
	"golang.org/x/net/html"
	"html/template"
	"regexp"
	"strings"
)

// dehydrated article (enough for list in gui):
//  headline, pub, section, published, tags, byline, url

type Publication struct {
	// Code is a short, unique name (eg "mirror")
	Code string
	// Name is the 'pretty' name (eg "The Daily Mirror")
	Name   string
	Domain string
}

type Author struct {
	Name    string
	RelLink string
	Email   string
	Twitter string
}

/*
type Keyword struct {
	Name string
	URL  string
}
*/

//
type Article struct {
	ID           int // id in local sqlite db
	CanonicalURL string
	URLs         []string
	Headline     string
	Authors      []Author
	// Content contains HTML, sanitised using a subset of tags
	Content string

	Published   string
	Updated     string
	Publication Publication
	// Hack for now: store keywords as strings to work around badger shortcomings
	// TODO: Restore full structure
	//	Keywords    []Keyword
	Keywords []string
	Section  string

	Tags []string

	// a couple of tweet-specific bits
	RetweetCount  int
	FavoriteCount int
	// resolved links
	Links []string

	// fudge fields
	Pub    string
	Byline string
}

func (art *Article) Day() string {
	l := len(art.Published)
	switch {
	case l == 10:
		return art.Published
	case l < 10:
		return ""
	case l > 10:
		return art.Published[0:10]
	}
	return ""
}

func (art *Article) TextContent() template.HTML {
	txt := art.PlainTextContent()
	txt = strings.Replace(txt, "\n", "<br/>\n", -1)
	return template.HTML(txt)

}

func (art *Article) PlainTextContent() string {
	r := strings.NewReader(art.Content)
	root, err := html.Parse(r)
	if err != nil {
		return ""
	}

	return htmlutil.RenderNode(root)
}

// TODO: this should be out in the gui layer
func (art *Article) FormatContent(highlightTerms string) string {
	breakPat := regexp.MustCompile(`[\n]{2,}`)
	foo := strings.Fields(highlightTerms)
	txt := art.PlainTextContent()

	for _, term := range foo {
		termPat := regexp.MustCompile("(?i)" + regexp.QuoteMeta(term))
		txt = termPat.ReplaceAllStringFunc(txt, func(t string) string {
			return `<b>` + t + "</b>"
		})
	}
	txt = breakPat.ReplaceAllLiteralString(txt, "<br/><br/>")

	return txt
}

/*
func (art *Article) Hoo() *string {
	return &art.Headline
}
*/

func (art *Article) URL() string {
	if art.CanonicalURL != "" {
		return art.CanonicalURL
	}

	if len(art.URLs) > 0 {
		return art.URLs[0]
	}
	return ""
}

func (art *Article) TagsString() string {
	return strings.Join(art.Tags, " ")
}

func (art *Article) BylineString() string {
	names := make([]string, len(art.Authors))
	for i, a := range art.Authors {
		names[i] = a.Name
	}
	return strings.Join(names, ", ")
}

func (art *Article) AddTag(tag string) bool {
	for _, t := range art.Tags {
		if t == tag {
			return false // already got it
		}
	}
	art.Tags = append(art.Tags, tag)
	return true // changed
}

func (art *Article) RemoveTag(tag string) bool {
	dirtied := false
	newTags := []string{}
	for _, t := range art.Tags {
		if t == tag {
			dirtied = true
			continue
		}
		newTags = append(newTags, t)
	}

	if dirtied {
		art.Tags = newTags
	}
	return dirtied
}
