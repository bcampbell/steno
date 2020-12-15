package main

import (
	"github.com/bcampbell/steno/store"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

// SimWindow is a window which displays an article and lets the user
// browse/diff similar articles
type SimWindow struct {
	widgets.QMainWindow

	_ func() `constructor:"init"`

	// The project we're attached to
	proj *Project

	// art is holds the focused article. (nil = none)
	art *store.Article

	// controls we want to keep track of
	c struct {
		//
		//selSummary *widgets.QLabel
		//
		artView *widgets.QTextBrowser
	}

	// actions
	action struct {
		close *widgets.QAction
	}
}

// View implementation
func (v *SimWindow) OnArtsModified(store.ArtList) {
}

func (v *SimWindow) OnArtsAdded(store.ArtList) {
}

func (v *SimWindow) OnArtsDeleted(store.ArtList) {
}

func (v *SimWindow) OnRethink() {
}

func (v *SimWindow) init() {

	//v.ConnectDestroyed(func(obj *core.QObject) {
	// doesn't get here...
	//})
	v.ConnectCloseEvent(func(event *gui.QCloseEvent) {
		if v.proj != nil {
			v.proj.detachView(v)
			v.proj = nil
		}
		event.Accept()
	})

	v.SetMinimumSize2(640, 400)

	// Set up menu
	m := v.MenuBar().AddMenu2("&File")
	v.action.close = m.AddAction("Close")

	// article view
	v.c.artView = widgets.NewQTextBrowser(nil)
	v.c.artView.SetOpenExternalLinks(true)
	v.SetCentralWidget(v.c.artView)
	// set up actions
	{
		v.action.close.ConnectTriggered(func(checked bool) {
			v.Close()
		})

	}

	v.Show()
}

func (v *SimWindow) SetProject(proj *Project) {
	v.proj = proj
	v.proj.attachView(v)
}

func (v *SimWindow) SetArticle(art *store.Article) {
	v.art = art
	rawHTML, _ := HTMLArt(art)
	v.c.artView.SetHtml(rawHTML)
}

func (v *SimWindow) rethinkWindowTitle() {
	title := "Steno"
	/*	if v.proj != nil {
			title += " - " + filepath.Base(v.Proj.Store.Filename())
		}
	*/
	v.SetWindowTitle(title)
}

func (v *SimWindow) rethinkActionStates() {
	// always available
	// v.action.close.SetEnabled(true)
}
