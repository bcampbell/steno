package main

import (
	"github.com/bcampbell/htmlutil"
	"golang.org/x/net/html"
	"html/template"
	"strings"
)

type Article struct {
	ID           int
	CanonicalURL string
	// all known URLs for article (including canonical)
	// TODO: first url should be considered "preferred" if no canonical?
	URLs     []string
	Headline string
	//	Authors     []Author
	Content   string
	Published string
	Updated   string
	//	Publication Publication
	//Keywords []Keyword

	Pub  string
	Tags []string
}

func (art *Article) TextContent() template.HTML {

	r := strings.NewReader(art.Content)
	root, err := html.Parse(r)
	if err != nil {
		return ""
	}

	txt := htmlutil.RenderNode(root)

	txt = strings.Replace(txt, "\n", "<br/>\n", -1)
	return template.HTML(txt)

}

func (art *Article) Hoo() *string {
	return &art.Headline
}

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
