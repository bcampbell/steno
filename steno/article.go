package main

import (
	"github.com/bcampbell/arts/arts"
	"github.com/bcampbell/htmlutil"
	"golang.org/x/net/html"
	"html/template"
	"strings"
)

type Article struct {
	arts.Article `json:",inline"`
	ID           string   `json:"id"`
	Pub          string   `json:"pub"`
	KW           []string `json:"kw"`
	Tags         []string `json:"tags"`
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
