package ft

import (
	"fmt"
	"github.com/bcampbell/htmlutil"
	"github.com/bcampbell/steno/store"
	"golang.org/x/net/html"
	"io"
	"regexp"
	"strings"
)

func PlainText(rawhtml string) string {
	r := strings.NewReader(rawhtml)
	root, err := html.Parse(r)
	if err != nil {
		return ""
	}

	return htmlutil.RenderNode(root)
}

var wordPat = regexp.MustCompile(`(\w+)|[\S]`)

func tokenise(s string) []string {
	s = strings.ToLower(s)
	return wordPat.FindAllString(s, -1)
}

func dumpTagged(db *store.Store, out io.Writer) error {
	it := db.IterateTaggedArts()

	for it.Next() {
		art := it.Cur()
		dumpArt(art, out)
	}
	return it.Err()
}

func dumpArt(art *store.Article, out io.Writer) {
	headline := strings.Join(tokenise(art.Headline), " ")

	content := PlainText(art.Content)
	content = strings.Join(tokenise(content), " ")

	labels := make([]string, 0, len(art.Tags))
	for _, tag := range art.Tags {
		labels = append(labels, "__label__"+tag)
	}

	fmt.Fprintf(out, "%s %s\n", strings.Join(labels, " "), headline+" | "+content)
}
