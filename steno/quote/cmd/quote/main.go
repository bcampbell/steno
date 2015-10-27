package main

import (
	"fmt"
	//	"github.com/bcampbell/htmlutil"
	"golang.org/x/net/html"
	"os"
	"semprini/steno/steno/quote"
	"strconv"
)

func short(s string, i int) string {
	runes := []rune(s)
	if len(runes) > i {
		return string(runes[:i]) + "..."
	}
	return s
}

func describe(n *html.Node) string {
	switch n.Type {
	case html.TextNode:
		return fmt.Sprintf("%s", strconv.Quote(short(n.Data, 60)))
	case html.ElementNode:
		return fmt.Sprintf("<%s>", n.DataAtom)
	case html.CommentNode:
		return fmt.Sprintf("<!-- %s -->", strconv.Quote(short(n.Data, 60)))
	case html.DocumentNode:
		return "DOC"
	case html.DoctypeNode:
		return "DOCTYPE"
	case html.ErrorNode:
		return "ERROR"
	default:
		return "???"
	}
}

func depth(n *html.Node) int {
	if n.Parent == nil {
		return 0
	} else {
		return depth(n.Parent) + 1
	}
}

func main() {

	infile, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}

	doc, err := html.Parse(infile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	defer infile.Close()

	quote.HighlightQuotes(doc)

	err = html.Render(os.Stdout, doc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}

	//	fmt.Println(htmlutil.RenderNode(doc))
}
