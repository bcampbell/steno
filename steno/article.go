package main

import (
	"github.com/bcampbell/htmlutil"
	"golang.org/x/net/html"
	"html/template"
	"strings"
)

type Publication struct {
	Name   string `json:"content"`
	Domain string `json:"domain"`
	Code   string `json:"code"`
}

type Author struct {
	Name string `json:"name"`
	/*
	   RelLink string `json:"rel_link,omitempty"`
	   Email   string `json:"email,omitempty"`
	   Twitter string `json:"twitter,omitempty"`
	*/
}

type Article struct {
	ID           int
	CanonicalURL string `json:"canonical_url"`

	// all known URLs for article (including canonical)
	// TODO: first url should be considered "preferred" if no canonical?
	URLs []string `json:"urls"`

	Headline    string      `json:"headline"`
	Authors     []Author    `json:"authors,omitempty"`
	Content     string      `json:"content"`
	Published   string      `json:"published"`
	Updated     string      `json:"updated"`
	Publication Publication `json:"publication"`

	//Keywords []Keyword
	Section string `json:"section"`

	Pub    string
	Byline string
	Tags   []string `json:"tags"`
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
