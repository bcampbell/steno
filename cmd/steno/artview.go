package main

import (
	"github.com/bcampbell/steno/steno/store"
	"html/template"
	"strings"
)

func HTMLArt(art *store.Article) (string, error) {

	const tpl = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title></title>
	</head>
	<body>
		<h2>{{.Art.Headline}}</h2><br/>
		URL: <a href="{{.Art.URL}}">{{.Art.URL}}</a><br/>
		Publication: {{.Art.Publication.Name}} {{.Art.Publication.Code}} {{.Art.Publication.Domain}}<br/>
		Published: {{.Art.Published}}<br/>
		<hr/>
		<div>
		{{.Content}}
		</div>
		<hr/>
		Authors: {{range .Art.Authors}}{{.Name}} | {{end}}<br/>
		Section: {{.Art.Section}}<br/>
		Keywords: {{range .Art.Keywords}}{{.}} | {{end}}<br/>
		Published: {{.Art.Published}}<br/>
		Updated: {{.Art.Updated}}<br/>
		Tags: {{range .Art.Tags}}{{.}} | {{end}}<br/>
	</body>
</html>`

	var t = template.Must(template.New("artview").Parse(tpl))

	data := struct {
		Art     *store.Article
		Content template.HTML
	}{
		Art:     art,
		Content: template.HTML(art.FormatContent("")),
	}
	var buf strings.Builder
	err := t.Execute(&buf, data)
	if err != nil {
		return "!!!", err
	}
	return buf.String(), nil
}
