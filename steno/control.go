package main

// a bit of a mishmash between core functionality and gui.
// TODO: Refector this, along with app.go
// separate core functionality from GUI

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"golang.org/x/net/html"
	"gopkg.in/qml.v1"
	"os"
	"path/filepath"
	"semprini/steno/steno/quote"
	"semprini/steno/steno/store"
	"strings"
	"time"
)

type Facet struct {
	Txt string
	Cnt int
}

type Progress struct {
	InFlight     bool
	Title        string
	ExpectedCnt  int // 0=unknown
	CompletedCnt int
	StatusMsg    string
	ErrorMsg     string
	ctrl         *Control
}

func (p *Progress) SetError(err error) {
	p.ErrorMsg = err.Error()
	qml.Changed(p.ctrl, p)
}

func (p *Progress) SetStatus(msg string) {
	p.StatusMsg = msg
	qml.Changed(p.ctrl, p)
}

func (p *Progress) Reset() {
	p.InFlight = false
	p.Title = ""
	p.ExpectedCnt = 0
	p.CompletedCnt = 0
	p.StatusMsg = ""
	p.ErrorMsg = ""
}

//
type Control struct {
	App *App

	obj qml.Object

	Results    *Results
	TotalArts  int
	SortColumn int
	SortOrder  int
	store      *store.Store

	ViewMode string // "tweet" or "article"

	DateFmt string // format for date display in results (time.Time format)

	Progress   Progress
	StatusText string
	HelpText   string
}

func NewControl(app *App, storePath string, gui qml.Object) (*Control, error) {
	var err error

	ctrl := &Control{}
	ctrl.App = app

	newStore, err := store.New(storePath, dbug)
	if err != nil {
		return nil, err
	}
	ctrl.store = newStore

	ctrl.Results, err = NewResults(ctrl.store, "")
	if err != nil {
		return nil, err
	}

	ctrl.ViewMode = "article"
	ctrl.SortColumn = 3 // Published
	ctrl.SortOrder = 0

	ctrl.DateFmt = "Mon 02/01/06"

	// expose us to the qml side
	app.ctx.SetVar("ctrl", ctrl)

	// instantiate the gui
	w := app.Window.Root().ObjectByName("mainSpace")
	ctrl.obj = gui.Create(nil)
	ctrl.obj.Set("parent", w)

	//
	ctrl.Progress.ctrl = ctrl

	/*
		obj := window.Root().ObjectByName("query")
		obj.Set("text", "")
		fmt.Printf("%v\n", obj)
	*/

	return ctrl, nil
}

func (ctrl *Control) Close() {
	//dbug.Printf("Close db\n")
	ctrl.obj.Destroy()
	ctrl.store.Close()
	//ctrl.Window.Hide()
}

func (ctrl *Control) SetViewMode(mode string) {
	ctrl.ViewMode = mode
	qml.Changed(ctrl, &ctrl.ViewMode)
}

func (ctrl *Control) SetDateFmt(fmt string) {
	ctrl.DateFmt = fmt
	qml.Changed(ctrl, &ctrl.DateFmt)
	ctrl.forceArtsRefresh()
}

func (ctrl *Control) ApplySorting(sortColumn string, sortOrder int) {
	ctrl.Results = ctrl.Results.Sort(sortColumn, sortOrder)
	qml.Changed(ctrl, &ctrl.Results)
}

// TODO: provide a function to validate query...

// version exposed to gui - only acts if query is different
func (ctrl *Control) SetQuery(q string) {
	if q == ctrl.Results.Query {
		return
	}
	ctrl.setQuery(q)
	ctrl.TotalArts = ctrl.store.TotalArts()
}

// internal version
func (ctrl *Control) setQuery(q string) {
	res, err := NewResults(ctrl.store, q)
	if err != nil {
		//TODO: show error
		e := fmt.Sprintf("Search error: %s", err)
		dbug.Println(e)
		ctrl.App.SetError(e)
		return
	}

	ctrl.Results = res
	qml.Changed(ctrl, &ctrl.Results)
}

func (ctrl *Control) DeleteArticles(artIndices []int) {

	arts := store.ArtList{}
	for _, artIdx := range artIndices {
		arts = append(arts, ctrl.Results.arts[artIdx])
	}
	err := ctrl.store.Delete(arts)
	if err != nil {
		dbug.Printf("ERROR: delete failed: %s\n", err)
		return
	}
	dbug.Printf("%d articles deleted\n", len(arts))

	// rerun the current query
	ctrl.setQuery(ctrl.Results.Query)
}

func (ctrl *Control) AddTags(artIndices []int, tags string) {

	tagList := strings.Fields(tags)

	arts := store.ArtList{}
	for _, artIdx := range artIndices {
		arts = append(arts, ctrl.Results.arts[artIdx])
	}
	affected, err := ctrl.store.AddTags(arts, tagList)
	if err != nil {
		dbug.Printf("AddTags(%q): ERROR: %s\n", tagList, err)
	} else {
		dbug.Printf("AddTags(%q): %d affected\n", tagList, len(affected))
	}

	// rerun the current query
	ctrl.setQuery(ctrl.Results.Query)
}

func (ctrl *Control) RemoveTags(artIndices []int, tags string) {
	tagList := strings.Fields(tags)
	arts := store.ArtList{}
	for _, artIdx := range artIndices {
		arts = append(arts, ctrl.Results.arts[artIdx])
	}
	affected, err := ctrl.store.RemoveTags(arts, tagList)
	if err != nil {
		dbug.Printf("RemoveTags(%q): ERROR: %s\n", tagList, err)
	} else {
		dbug.Printf("RemoveTags(%q): %d affected\n", tagList, len(affected))
	}

	// rerun the current query
	newResults, err := NewResults(ctrl.store, ctrl.Results.Query)
	if err != nil {
		dbug.Printf("Rerun query: ERROR: %s\n", err)
		return
	}
	ctrl.Results = newResults
	qml.Changed(ctrl, &ctrl.Results)
}

func (ctrl *Control) forceArtsRefresh() {
	// fudge to force tableview to rethink itself:
	// create a new Results, with the same data
	r := *ctrl.Results
	ctrl.Results = &r
	qml.Changed(ctrl, &ctrl.Results)
}

func (ctrl *Control) ExportOveralls(outFile string) {
	/* XYZZY */
	/*
		out, err := os.Create(outFile)

		if err != nil {
			// TODO: error on gui...
			dbug.Printf("ERROR: %s", err)
			return
		}
		err = exportOverallsCSV(ctrl.Results.arts, out)
		if err != nil {
			// TODO: error on gui...
			dbug.Printf("ERROR exporting overalls: %s", err)
			return
		}
		dbug.Printf("Wrote to %s\n", outFile)
	*/
}

func (ctrl *Control) ExportCSV(outFile string) {

	out, err := os.Create(outFile)

	if err != nil {
		// TODO: error on gui...
		dbug.Printf("ERROR: %s", err)
		return
	}
	err = exportCSV(ctrl.Results.arts, out)
	if err != nil {
		// TODO: error on gui...
		dbug.Printf("ERROR exporting overalls: %s", err)
		return
	}
	dbug.Printf("Wrote to %s\n", outFile)
}

func (ctrl *Control) Slurp(slurpSourceName string, dayFrom, dayTo string) {
	prog := &ctrl.Progress
	prog.Reset()

	// look up the server by name
	var server *SlurpSource
	for _, src := range ctrl.App.SlurpSources {
		if src.Name == slurpSourceName {
			server = &src
			break
		}
	}
	if server == nil {
		prog.SetError(fmt.Errorf("unknown server '%s'", slurpSourceName))
		return
	}

	//
	const shortForm = "2006-01-02"
	timeFrom, err := time.ParseInLocation(shortForm, dayFrom, time.Local)
	if err != nil {
		prog.SetError(fmt.Errorf("ERROR: bad dayFrom: %s (%s)\n", dayFrom, err))
		return
	}
	timeTo, err := time.ParseInLocation(shortForm, dayTo, time.Local)
	if err != nil {
		prog.SetError(fmt.Errorf("ERROR: bad dayTo: %s (%s)\n", dayTo, err))
		return
	}
	// HACK: want one day's worth
	timeTo = timeTo.AddDate(0, 0, 1)

	// -----

	go func() {
		startTime := time.Now()
		defer func() {
			prog.InFlight = false
			qml.Changed(ctrl, prog)
		}()

		prog.InFlight = true
		prog.Title = "Slurping"
		prog.StatusMsg = "Slurping..."
		qml.Changed(ctrl, prog)

		err := Slurp(ctrl.store, server, timeFrom, timeTo, prog)
		if err != nil {
			prog.SetError(err)
		}

		// re-run the current query
		r2, err := NewResults(ctrl.store, ctrl.Results.Query)
		if err != nil {
			dbug.Printf("ERROR failed to refresh query: %s\n", err)
			return
		}
		ctrl.Results = r2
		qml.Changed(ctrl, &ctrl.Results)

		elapsed := time.Since(startTime)
		dbug.Printf("Slurp finished (took %s)\n", elapsed)
	}()
}

func (ctrl *Control) RunScript(scriptIdx int) {
	s := ctrl.App.GetScript(scriptIdx)

	// run as goroutine to avoid freezing gui
	go func() {
		prog := &ctrl.Progress
		prog.Reset()

		defer func() {
			prog.InFlight = false
			qml.Changed(ctrl, &ctrl.Progress)
		}()

		prog.Title = fmt.Sprintf("running %s...", s.Name)
		ctrl.Progress.InFlight = true
		qml.Changed(ctrl, prog)
		err := s.Run(ctrl.store, func(expected int, completed int, status string) {
			prog.StatusMsg = status
			prog.ExpectedCnt = expected
			prog.CompletedCnt = completed
			qml.Changed(ctrl, prog)
		})
		if err != nil {
			dbug.Printf("ERROR running script %s: %s\n", s.Name, err)
			prog.ErrorMsg = err.Error()
			qml.Changed(ctrl, prog)
			ctrl.App.SetError(err.Error())
		}
		// rerun the current query
		ctrl.setQuery(ctrl.Results.Query)
	}()
}

// open link in a web browser
func (ctrl *Control) OpenLink(link string) {
	openURL(link)
}

func (ctrl *Control) EmbiggenShortlinks() {

	shortlinkDomainsFile := filepath.Join(ctrl.App.DataPath, "shortlink_domains.txt")

	allArts, err := ctrl.store.AllArts()
	if err != nil {
		dbug.Println(err.Error())
		ctrl.App.SetError(err.Error())
		return
	}

	prog := &ctrl.Progress
	prog.InFlight = true
	qml.Changed(ctrl, prog)
	// run as goroutine to prevent gui freezing
	go func() {
		defer func() {
			prog.InFlight = false
			qml.Changed(ctrl, prog)
		}()
		affected, err := embiggenArts(allArts, shortlinkDomainsFile, func(expected int, completed int, status string) {
			prog.ExpectedCnt = expected
			prog.CompletedCnt = completed
			prog.StatusMsg = status
			qml.Changed(ctrl, prog)
		})
		if err != nil {
			prog.ErrorMsg = err.Error()
			qml.Changed(ctrl, prog)
			dbug.Println(err.Error())
			ctrl.App.SetError(err.Error())
			return
		}
		dbug.Printf("Committing changes to db...\n")
		err = ctrl.store.UpdateLinks(affected)
		if err != nil {
			prog.ErrorMsg = err.Error()
			qml.Changed(ctrl, prog)
			dbug.Println(err.Error())
			ctrl.App.SetError(err.Error())
			return
		}
		dbug.Printf("finished embiggening\n")
		ctrl.forceArtsRefresh()
	}()
}

func (ctrl *Control) TagRetweets() {
	/* XYZZY */
	/*
		allArts, err := ctrl.store.AllArts()
		if err != nil {
			dbug.Println(err.Error())
			ctrl.App.SetError(err.Error())
			return
		}

		rts := store.ArtList{}
		for _, art := range allArts {
			if strings.Index(art.Content, "RT ") == 0 {
				rts = append(rts, art)
			}
		}

		if len(rts) > 0 {
			tagList := []string{"rt"}
			affected, err := ctrl.store.AddTags(rts, tagList)
			if err != nil {
				dbug.Printf("AddTags(%q): ERROR: %s\n", tagList, err)
			} else {
				dbug.Printf("AddTags(%q): %d affected\n", tagList, len(affected))
			}

			// rerun the current query
			ctrl.setQuery(ctrl.Results.Query)
		}
	*/
}

func (ctrl *Control) CopyCells(artIndices []int, colName string) {

	bits := []string{}

	for _, artIdx := range artIndices {
		foo := ctrl.Results.Art(artIdx).FieldString(colName)
		bits = append(bits, foo)
	}

	val := strings.Join(bits, "\n")

	err := ctrl.App.Clipboard.WriteAll(val)
	if err != nil {
		dbug.Printf("Copy failed: %s\n", err)
	}
}

func (ctrl *Control) CopyRows(artIndices []int) {
	fieldNames := []string{"headline", "published", "tags", "byline", "url", "retweets", "favourites", "keywords", "links"}
	var out bytes.Buffer
	w := csv.NewWriter(&out)
	w.Comma = '\t'
	for _, artIdx := range artIndices {
		fieldVals := []string{}
		art := ctrl.Results.Art(artIdx)
		for _, fldName := range fieldNames {
			fieldVals = append(fieldVals, art.FieldString(fldName))
		}
		err := w.Write(fieldVals)
		if err != nil {
			dbug.Printf("Copy failed (csv write failed): %s\n", err)
			return
		}
	}

	w.Flush()

	err := ctrl.App.Clipboard.WriteAll(out.String())
	if err != nil {
		dbug.Printf("Copy failed: %s\n", err)
	}
}

func (ctrl *Control) CopyArtSummaries(artIndices []int) {
	var out bytes.Buffer
	for _, artIdx := range artIndices {
		art := ctrl.Results.Art(artIdx)

		var pretty string
		t, err := time.ParseInLocation(time.RFC3339, art.Published, time.Local)
		if err == nil {
			pretty = t.Format("2-Jan-2006")
		} else {
			pretty = art.Published
		}
		fmt.Fprintf(&out, "%s %s %s\n", pretty, art.Headline, art.URL())
	}

	err := ctrl.App.Clipboard.WriteAll(out.String())
	if err != nil {
		dbug.Printf("Copy failed: %s\n", err)
	}
}

func (ctrl *Control) RenderContent(art *store.Article) string {
	r := strings.NewReader(art.Content)
	root, err := html.Parse(r)
	if err != nil {
		return ""
	}
	quote.HighlightQuotes(root)

	var buf bytes.Buffer
	err = html.Render(&buf, root)
	if err != nil {
		return ""
	}

	return buf.String()
}
