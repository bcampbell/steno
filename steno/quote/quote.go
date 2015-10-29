package quote

import (
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
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

func nextNode(n *html.Node) *html.Node {

	if n.FirstChild != nil {
		return n.FirstChild
	}
	if n.NextSibling != nil {
		return n.NextSibling
	}

	for {
		n = n.Parent
		if n == nil {
			return nil
		}
		if n.NextSibling != nil {
			return n.NextSibling
		}
	}
}

type NodeSpans struct {
	n     *html.Node
	spans []int
}

func HighlightQuotes(n *html.Node) {

	quotes := []NodeSpans{}

	for ; n != nil; n = nextNode(n) {
		if n.Type != html.TextNode {
			continue
		}

		spans := FindQuoted(n.Data)
		if len(spans) > 0 {
			quotes = append(quotes, NodeSpans{n: n, spans: spans})
		}

	}

	for _, q := range quotes {
		HighlightSpans(q.n, q.spans, nil)
	}
}

func buildStringsMatcher(terms []string) *regexp.Regexp {

	parts := make([]string, len(terms))
	for i, term := range terms {
		parts[i] = regexp.QuoteMeta(term)
	}
	pat := "(?i)" + strings.Join(parts, "|")
	return regexp.MustCompile(pat)
}

func HighlightText(n *html.Node, highlightTerms []string) {

	if len(highlightTerms) == 0 {
		return
	}
	pat := buildStringsMatcher(highlightTerms)

	quotes := []NodeSpans{}

	for ; n != nil; n = nextNode(n) {
		if n.Type != html.TextNode {
			continue
		}

		matches := pat.FindAllStringSubmatchIndex(n.Data, -1)
		spans := []int{}
		for _, m := range matches {
			spans = append(spans, m[0], m[1])
		}
		if len(spans) > 0 {
			quotes = append(quotes, NodeSpans{n: n, spans: spans})
		}
	}

	for _, q := range quotes {
		HighlightSpans(q.n, q.spans,
			&html.Node{
				Type:     html.ElementNode,
				DataAtom: atom.Span, //atom.Font,
				Data:     "span",    //"font",
				Attr: []html.Attribute{
					{Key: "class", Val: "hl"},
					//					{Key: "style", Val: "color: #ffdd44;"},
				},
			})
	}
}

func HighlightSpans(orig *html.Node, spans []int, wrapNode *html.Node) {
	if len(spans) == 0 {
		return
	}

	if wrapNode == nil {
		wrapNode = &html.Node{
			Type:     html.ElementNode,
			DataAtom: atom.Span, //atom.Font,
			Data:     "span",    //"font",
			Attr: []html.Attribute{
				{Key: "class", Val: "quote"},
				//{Key: "style", Val: "color: #ff0000; background-color: #ffff00;"},
			},
		}
	}
	newNodes := []*html.Node{}

	pos := 0
	for i := 0; i < len(spans); i += 2 {
		begin, end := spans[i], spans[i+1]
		if pos < begin {
			leading := &html.Node{
				Type: html.TextNode,
				Data: orig.Data[pos:begin],
			}

			pos = begin
			newNodes = append(newNodes, leading)
		}

		var hl = *wrapNode

		frag := orig.Data[begin:end]
		pos = end
		hlContent := &html.Node{Type: html.TextNode, Data: frag}
		hl.AppendChild(hlContent)
		newNodes = append(newNodes, &hl)
	}
	if pos < len(orig.Data) {
		trailing := &html.Node{
			Type: html.TextNode,
			Data: orig.Data[pos:],
		}

		newNodes = append(newNodes, trailing)
	}

	for _, n := range newNodes {
		orig.Parent.InsertBefore(n, orig)
	}
	orig.Parent.RemoveChild(orig)
}

func FindQuoted(s string) []int {
	return lex(s)
}

type stateFn func(*lexer) stateFn

type lexer struct {
	input        string
	pos, prevpos int
	quoteSpans   []int
}

func (l *lexer) next() rune {
	l.prevpos = l.pos
	if l.eof() {
		return 0
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += w
	return r
}

func (l *lexer) eof() bool {
	return l.pos >= len(l.input)
}

func (l *lexer) backup() {
	l.pos = l.prevpos
}

func (l *lexer) peek() rune {
	if l.eof() {
		return '\000'
	}
	r := l.next()
	l.backup()
	return r
}

func lex(in string) []int {
	l := &lexer{input: in, quoteSpans: []int{}}
	for state := lexDefault; state != nil; {
		state = state(l)
	}
	return l.quoteSpans
}

func lexDefault(l *lexer) stateFn {
	var prev rune
	for {
		if l.eof() {
			return nil
		}

		r := l.next()

		if strings.ContainsRune(`"“«„`, r) {
			return lexDoubleQuoted
		}
		if strings.ContainsRune(`‘‹‚`, r) {
			return lexSingleQuoted
		}
		// special case for apostrophe/single quote ambiguity
		if r == '\'' && !(unicode.IsLetter(prev) || unicode.IsDigit(prev)) {
			return lexSingleQuoted
		}

		prev = r

	}
}

func lexDoubleQuoted(l *lexer) stateFn {
	start := l.pos
	var end int
	for {
		if l.eof() {
			end = l.pos
			break
		}
		r := l.next()
		if strings.ContainsRune(`"”»“`, r) {
			end = l.prevpos
			break
		}
	}

	l.quoteSpans = append(l.quoteSpans, start, end)
	return lexDefault
}

func lexSingleQuoted(l *lexer) stateFn {
	start := l.pos
	var end int
	for {
		if l.eof() {
			end = l.pos
			break
		}
		r := l.next()
		if strings.ContainsRune(`'’›‘`, r) {

			// make sure things like "don't" don't close the quote prematurely
			end = l.prevpos
			r2 := l.peek()
			if !unicode.IsLetter(r2) && !unicode.IsDigit(r2) {
				break
			}
		}

	}

	l.quoteSpans = append(l.quoteSpans, start, end)
	return lexDefault
}
