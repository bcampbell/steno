package quote

import (
	"testing"
)

func TestFindQuoted(t *testing.T) {

	cases := []struct {
		txt    string
		expect []string
	}{
		{`blah blah "hello there" blah blah`, []string{"hello there"}},
		{`"one" and "two"`, []string{"one", "two"}},
		{`"one" and "two"`, []string{"one", "two"}},
		{`"fancy stuff"`, []string{"fancy stuff"}},
		{`some “fancy stuff”`, []string{"fancy stuff"}},
		{`‘fancy stuff’`, []string{"fancy stuff"}},
		{`«fancy stuff» in foreignish`, []string{"fancy stuff"}},
		{`some more ‹fancy stuff›`, []string{"fancy stuff"}},
		{`„fancy stuff“`, []string{"fancy stuff"}},
		{`‚fancy stuff‘`, []string{"fancy stuff"}},
		{`and now: "an unclosed quote`, []string{"an unclosed quote"}},
		{`we don't want anything to match here. We really don't.`, []string{}},
		{`‘This decision doesn't mean you need to stop eating any red and processed meat.`,
			[]string{"This decision doesn't mean you need to stop eating any red and processed meat."}},
		{`This one 'has lot's of ambiguous crap'. But 'that won't screw us up'. So there.`,
			[]string{"has lot's of ambiguous crap", "that won't screw us up"}},
		{`and 'here we have a sudden end`, []string{"here we have a sudden end"}},
	}

	for _, dat := range cases {
		qs := FindQuoted(dat.txt)
		ok := true

		got := []string{}
		for i := 0; i < len(qs); i += 2 {
			got = append(got, dat.txt[qs[i]:qs[i+1]])
		}

		if len(got) != len(dat.expect) {
			ok = false
		} else {
			for i, _ := range got {
				if got[i] != dat.expect[i] {
					ok = false
					break
				}
			}
		}

		if !ok {
			t.Errorf("findQuotes(%q) got %q, expected %q", dat.txt, got, dat.expect)
		}
	}
}
