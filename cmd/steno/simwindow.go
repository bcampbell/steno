package main

import (
	"bytes"
	"fmt"
	"github.com/bcampbell/steno/store"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
	"strings"
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

	// a model to hold the list of similar articles
	model *SimListModel

	// controls we want to keep track of
	c struct {
		resultView *widgets.QTableView
		artView    *widgets.QTextBrowser
		otherView  *widgets.QTextBrowser
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

	v.model = NewSimListModel(nil)
	v.SetMinimumSize2(640, 400)

	// Set up menu
	{
		m := v.MenuBar().AddMenu2("&File")
		v.action.close = m.AddAction("Close")
		// set up actions
		{
			v.action.close.ConnectTriggered(func(checked bool) {
				v.Close()
			})

		}
	}

	// top-level widget
	vsplitter := widgets.NewQSplitter2(core.Qt__Horizontal, nil)

	// left pane
	{
		splitter := widgets.NewQSplitter2(core.Qt__Vertical, nil)
		// article view
		v.c.artView = widgets.NewQTextBrowser(nil)
		v.c.artView.SetOpenExternalLinks(true)
		splitter.AddWidget(v.c.artView)

		bottom := widgets.NewQWidget(nil, core.Qt__Widget)
		vbox := widgets.NewQVBoxLayout()
		bottom.SetLayout(vbox)

		label := widgets.NewQLabel2("Similar articles", nil, core.Qt__Widget)
		vbox.AddWidget(label, 0, 0)

		// list of similar articles
		tv := v.initSimListView()
		vbox.AddWidget(tv, 0, 0)
		splitter.AddWidget(bottom)
		vsplitter.AddWidget(splitter)
	}

	// right pane
	{
		v.c.otherView = widgets.NewQTextBrowser(nil)
		v.c.otherView.SetOpenExternalLinks(true)
		vsplitter.AddWidget(v.c.otherView)
	}

	//	vbox := widgets.NewQVBoxLayout()
	//	widget.SetLayout(vbox)
	//	splitter.AddWidget(widget)
	v.SetCentralWidget(vsplitter)

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

	// populate the list of similar articles

	similarIDs := make([]store.ArtID, 0, len(art.Similar))
	scores := make([]float32, 0, len(art.Similar))
	for _, match := range art.Similar {
		similarIDs = append(similarIDs, match.ID)
		scores = append(scores, match.Score)
	}

	// fetch all the similar articles
	arts := make([]*store.Article, 0, len(art.Similar))
	iter := v.proj.Store.IterateArts(similarIDs...)
	for iter.Next() {
		arts = append(arts, iter.Cur())
	}
	err := iter.Err()
	if err != nil {
		// TODO?
		dbug.Printf("couldn't fetch similar articles: %s\n", err)
		return
	}
	v.model.SetArticles(arts, scores)
}

func (v *SimWindow) rethinkWindowTitle() {
	title := "Similar articles"
	/*	if v.proj != nil {
			title += " - " + filepath.Base(v.proj.Store.Filename())
		}
	*/
	v.SetWindowTitle(title)
}

func (v *SimWindow) rethinkActionStates() {
	// always available
	// v.action.close.SetEnabled(true)
}

// Set up the table for displaying the list of similar articles.
// TODO: mostly shared by projectwindow - factor out a results widget instead?
func (v *SimWindow) initSimListView() *widgets.QTableView {

	tv := widgets.NewQTableView(nil)
	tv.SetShowGrid(false)
	tv.SetSelectionBehavior(widgets.QAbstractItemView__SelectRows)
	//	tv.SetSelectionMode(widgets.QAbstractItemView__ExtendedSelection)
	tv.VerticalHeader().SetVisible(false)
	//tv.HorizontalHeader().SetSectionResizeMode(widgets.QHeaderView__Stretch)

	tv.SetModel(v.model)
	tv.ResizeColumnsToContents()

	{
		hdr := tv.HorizontalHeader()
		// set up sorting (by clicking on column headers)
		hdr.SetSortIndicatorShown(true)
		hdr.ConnectSortIndicatorChanged(func(logicalIndex int, order core.Qt__SortOrder) {
			// TODO
		})
	}

	// Callbacks for selecting items in results list
	tv.SelectionModel().ConnectCurrentChanged(func(current *core.QModelIndex, previous *core.QModelIndex) {
		// show article text for most recently-selected article (if any)
		if current.IsValid() {
			otherArt := v.model.arts[current.Row()]

			rawHTML := htmlDiff(v.art, otherArt)
			v.c.otherView.SetHtml(rawHTML)
		}
	})

	tv.ConnectDoubleClicked(func(index *core.QModelIndex) {
		row := index.Row()
		var art *store.Article
		if row >= 0 && row < len(v.model.arts) {
			art = v.model.arts[row]
		}

		if art != nil {
			// double-clicked on an valid article
			w := NewSimWindow(nil, 0)
			w.SetProject(v.proj)
			w.SetArticle(art)
			w.Show()
		}
	})
	return tv
}

// SimListModel implements a QAbstractTableModel around our list of similar articles.
type SimListModel struct {
	core.QAbstractTableModel

	_ func() `constructor:"init"`

	arts   []*store.Article
	scores []float32
}

func (m *SimListModel) init() {
	//	m.modelData = []TableItem{{"john", "doe"}, {"john", "bob"}}
	m.ConnectHeaderData(m.headerData)
	m.ConnectRowCount(m.rowCount)
	m.ConnectColumnCount(m.columnCount)
	m.ConnectData(m.data)
}

// SetArticle installs data in the model - the similar articles + scores.
func (m *SimListModel) SetArticles(arts []*store.Article, scores []float32) {
	m.BeginResetModel()
	m.arts = arts
	m.scores = scores
	m.EndResetModel()
}

func (m *SimListModel) headerData(section int, orientation core.Qt__Orientation, role int) *core.QVariant {
	if role != int(core.Qt__DisplayRole) || orientation == core.Qt__Vertical {
		return m.HeaderDataDefault(section, orientation, role)
	}

	switch section {
	case 0:
		return core.NewQVariant1("Score")
	case 1:
		return core.NewQVariant1("URL")
	case 2:
		return core.NewQVariant1("Headline")
	case 3:
		return core.NewQVariant1("Published")
		/*	case 3:
				return core.NewQVariant1("Pub")
			case 4:
				return core.NewQVariant1("Tags")
			case 5:
				return core.NewQVariant1("Similar")
		*/
	}
	return core.NewQVariant()
}

func (m *SimListModel) columnCount(*core.QModelIndex) int {
	//	fmt.Printf("columnCount()\n")
	return 4
}

func (m *SimListModel) rowCount(*core.QModelIndex) int {
	return len(m.arts)
}

func (m *SimListModel) data(index *core.QModelIndex, role int) *core.QVariant {
	if role != int(core.Qt__DisplayRole) {
		return core.NewQVariant()
	}

	rowNum := index.Row()
	art := m.arts[rowNum]

	//	fmt.Printf("data(): %d %d\n", rowNum, index.Column())
	switch index.Column() {
	case 0:
		return core.NewQVariant1(m.scores[rowNum])
	case 1:
		return core.NewQVariant1(art.CanonicalURL)
	case 2:
		return core.NewQVariant1(art.Headline)
	case 3:
		return core.NewQVariant1(art.Published)
	}
	return core.NewQVariant()
}

func htmlDiff(art1 *store.Article, art2 *store.Article) string {
	header := `<!DOCTYPE html>
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
	footer := `</body>
</html>
`

	txt1 := tidy(art1.PlainTextContent())
	txt2 := tidy(art2.PlainTextContent())

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(txt1, txt2, false)
	diffs = dmp.DiffCleanupSemantic(diffs)

	var buf bytes.Buffer
	w := &buf

	w.WriteString(header)

	fmt.Fprintln(w, `<div class="art1">`)
	fmt.Fprintf(w, "<pre>\n")
	//fmt.Fprintf(w, "Match Factor: %f\n", f)
	fmt.Fprintf(w, "Headline: %s\n", art2.Headline)
	fmt.Fprintf(w, "URL: <a href=\"%s\">%s</a>\n", art2.CanonicalURL, art2.CanonicalURL)
	fmt.Fprintf(w, "Pub: %s\n", art2.Publication.Code)
	fmt.Fprintf(w, "</pre>\n")
	//fmt.Fprintln(w, `<label>show diff</label><input class="showdiffs" type="checkbox" />`)
	fmt.Fprintln(w, `<pre class="diff content">`)
	fmt.Fprintln(w, dmp.DiffPrettyHtml(diffs))
	fmt.Fprintln(w, `</pre>`)
	fmt.Fprintf(w, "</div>\n")

	w.WriteString(footer)

	return buf.String()
}

func tidy(s string) string {
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if len(l) > 0 {
			out = append(out, l)
		}
	}

	return strings.Join(out, "\n")
}
