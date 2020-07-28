package simrep

import (
	"fmt"
	"github.com/bcampbell/steno/store"
	"github.com/sergi/go-diff/diffmatchpatch"
	"io"
)

func EmitHeader(w io.Writer, opts *Opts) {

	raw := `<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">

<style>
body { /* max-width: 80rem; margin: auto; */ }
.grp { border-bottom: 1px solid black; margin-top: 2rem; }
.art1 { margin-left: 2rem; margin-bottom: 2rem; border-top: 1px solid black; }
.diff { display: none; }
.showdiffs:checked + .diff { display:block; }
.content { background: #ffe; padding: 1rem; max-width: 65rem; white-space: pre-wrap; }
</style>
</head>
<body>
`
	io.WriteString(w, raw)

	fmt.Fprintf(w, "<pre>ngramSize=%d, minWords=%d, matchThreshold=%.02f</pre>\n",
		opts.NGramSize, opts.MinWords, opts.MatchThreshold)
}

func EmitFooter(w io.Writer) {
	raw := `</body>
</html>
`
	io.WriteString(w, raw)
}

func EmitMatches(w io.Writer, art *store.Article, matching []*store.Article, metrics []float64) {

	//
	txt2 := tidy(art.PlainTextContent())
	fmt.Fprintln(w, `<div class="grp">`)
	fmt.Fprintf(w, `<pre>`)
	fmt.Fprintf(w, "Headline: %s\n", art.Headline)
	fmt.Fprintf(w, "URL: <a href=\"%s\">%s</a>\n", art.CanonicalURL, art.CanonicalURL)
	fmt.Fprintf(w, "Pub: %s\n", art.Publication.Code)
	fmt.Fprintf(w, "Content:\n")
	fmt.Fprintln(w, `<div class="content">`)
	fmt.Fprintf(w, "%s\n", txt2)
	fmt.Fprintln(w, `</div>`)
	fmt.Fprintf(w, `</pre>`)
	for i, art1 := range matching {
		f := metrics[i]

		// get the matching article
		txt1 := tidy(art1.PlainTextContent())
		fmt.Fprintln(w, `<div class="art1">`)
		fmt.Fprintf(w, "<pre>\n")
		fmt.Fprintf(w, "Match Factor: %f\n", f)
		fmt.Fprintf(w, "Headline: %s\n", art1.Headline)
		fmt.Fprintf(w, "URL: <a href=\"%s\">%s</a>\n", art1.CanonicalURL, art1.CanonicalURL)
		fmt.Fprintf(w, "Pub: %s\n", art1.Publication.Code)
		fmt.Fprintf(w, "</pre>\n")
		fmt.Fprintln(w, `<label>show diff</label><input class="showdiffs" type="checkbox" />`)
		fmt.Fprintln(w, `<pre class="diff content">`)
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(txt1, txt2, false)
		diffs = dmp.DiffCleanupSemantic(diffs)
		fmt.Fprintln(w, dmp.DiffPrettyHtml(diffs))
		fmt.Fprintln(w, `</pre>`)
		fmt.Fprintf(w, "</div>\n")
	}
	fmt.Fprintf(w, "</div> <!-- end .grp -->\n")
}
