package gui

import (
	"fmt"
	"github.com/andlabs/ui"
	"github.com/bcampbell/steno/steno/store"
)

type ArtView struct {
	Proj *Project
	Art  *store.Article

	// controls we want to track...
	c struct {
		window *ui.Window
	}
}

func NewArtView(proj *Project, art *store.Article) *ArtView {

	v := &ArtView{Proj: proj, Art: art}

	ro := func(txt string) *ui.Entry {
		e := ui.NewEntry()
		e.SetReadOnly(true)
		e.SetText(txt)
		return e
	}

	box := ui.NewVerticalBox()

	box.Append(ui.NewLabel("Headline:"), false)
	box.Append(ro(art.Headline), false)
	box.Append(ui.NewLabel("Pub:"), false)
	box.Append(ro(art.Pub), false)
	box.Append(ui.NewLabel("Published:"), false)
	box.Append(ro(art.Published), false)
	box.Append(ui.NewLabel("Byline:"), false)
	box.Append(ro(art.Byline), false)
	box.Append(ui.NewLabel("CanonicalURL:"), false)
	box.Append(ro(art.CanonicalURL), false)
	box.Append(ui.NewLabel("Content:"), false)

	content := ui.NewMultilineEntry()
	content.SetReadOnly(true)
	content.SetText(art.PlainTextContent())
	box.Append(content, true)

	window := ui.NewWindow(fmt.Sprintf(`Article: "%s"`, art.Headline), 800, 500, true)
	window.SetMargined(true)
	window.SetChild(box)

	window.OnClosing(func(*ui.Window) bool {
		//v.Proj.detachView(v)
		return true
	})
	window.Show()
	v.c.window = window
	return v
}
