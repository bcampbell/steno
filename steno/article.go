package main

import (
	"code.google.com/p/go.net/html"
	"github.com/bcampbell/arts/arts"
	"github.com/bcampbell/htmlutil"
	"html/template"
	"strings"
)

type Article struct {
	arts.Article `json:",inline"`
	ID           string   `json:"id"`
	Pub          string   `json:"pub"`
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
