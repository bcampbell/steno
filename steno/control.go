package main

// a bit of a mishmash between core functionality and gui.
// TODO: Refector this, along with app.go
// separate core functionality from GUI

import (
	"fmt"
	"gopkg.in/qml.v1"
	"os"
	"strings"
)

type Facet struct {
	Txt string
	Cnt int
}

type SlurpProgress struct {
	TotalCnt int
	GotCnt   int
	NewCnt   int
	InFlight bool
	ErrorMsg string
}

func (p *SlurpProgress) String() string {
	if p.ErrorMsg != "" {
		return p.ErrorMsg
	} else {
		return fmt.Sprintf("Received %d articles (%d new)", p.GotCnt+p.NewCnt, p.NewCnt)
	}
}

type Results struct {
	Query string
	arts  ArtList
	Len   int

	FacetLen int
	facets   []*Facet
}

func NewResults(store *Store, query string) (*Results, error) {

	var arts ArtList
	var err error
	if query == "" {
		arts, err = store.AllArts()
	} else {
		arts, err = store.Search(query)
	}
	if err != nil {
		return nil, err
	}

	res := Results{Query: query}
	res.setArts(arts)
	return &res, nil
}

func (res *Results) setArts(arts ArtList) {
	res.arts = arts
	res.Len = len(res.arts)

	// calc facets
	tab := map[string]int{}
	for _, art := range res.arts {
		tab[art.Pub]++
	}
	res.facets = []*Facet{}
	for txt, cnt := range tab {
		res.facets = append(res.facets, &Facet{txt, cnt})
	}
	res.FacetLen = len(res.facets)
}

func (res *Results) Match(artIdx int, needle string) bool {
	needle = strings.ToLower(needle)
	art := res.Art(artIdx)
	if strings.Contains(strings.ToLower(art.Headline), needle) {
		return true
	}
	if strings.Contains(strings.ToLower(art.CanonicalURL), needle) {
		return true
	}
	if strings.Contains(strings.ToLower(art.Published), needle) {
		return true
	}
	if strings.Contains(strings.ToLower(art.Pub), needle) {
		return true
	}
	if strings.Contains(strings.ToLower(art.TagsString()), needle) {
		return true
	}
	return false
}

func (res *Results) FindForward(artIdx int, needle string) int {
	for ; artIdx < len(res.arts); artIdx++ {
		if res.Match(artIdx, needle) {
			return artIdx
		}
	}
	return -1
}

func (res *Results) FindReverse(artIdx int, needle string) int {
	for ; artIdx >= 0; artIdx-- {
		if res.Match(artIdx, needle) {
			return artIdx
		}
	}
	return -1
}

func (res *Results) Art(idx int) *Article {
	if idx >= 0 && idx < len(res.arts) {
		return res.arts[idx]
	}

	// sometimes get here... seems to be tableview doing one last refresh on
	// old delegates before zapping/recycling them
	// TODO: investigate!
	//	dbug.Printf("bad idx: %d\n", idx)
	return &Article{Headline: fmt.Sprintf("<BAD> %d", idx)}
}

func (res *Results) Facet(idx int) *Facet {
	return res.facets[idx]
}

// returns new Results
// TODO: this is a bit brittle - the sortColumn comes directly from the QML side of things...
// Do something better!
func (res *Results) Sort(sortColumn, sortOrder int) *Results {
	// order: 1: ascending, 0: descending
	//dbug.Printf("new sorting: %d %d\n", sortColumn, sortOrder)

	sorted := make(ArtList, len(res.arts))
	copy(sorted, res.arts)

	var criteria func(a1, a2 *Article) bool

	if sortOrder == 0 {
		switch sortColumn {
		case 0:
			criteria = func(a1, a2 *Article) bool { return a1.Headline > a2.Headline }
		case 1:
			criteria = func(a1, a2 *Article) bool { return a1.Pub > a2.Pub }
		case 2:
			criteria = func(a1, a2 *Article) bool { return a1.Section > a2.Section }
		case 3:
			criteria = func(a1, a2 *Article) bool { return a1.Published > a2.Published }
		case 4:
			criteria = func(a1, a2 *Article) bool { return a1.TagsString() > a2.TagsString() }
		case 5:
			criteria = func(a1, a2 *Article) bool { return a1.Byline > a2.Byline }
		case 6:
			criteria = func(a1, a2 *Article) bool { return a1.URL() > a2.URL() }
		}
	} else if sortOrder == 1 {
		switch sortColumn {
		case 0:
			criteria = func(a1, a2 *Article) bool { return a1.Headline < a2.Headline }
		case 1:
			criteria = func(a1, a2 *Article) bool { return a1.Pub < a2.Pub }
		case 2:
			criteria = func(a1, a2 *Article) bool { return a1.Section < a2.Section }
		case 3:
			criteria = func(a1, a2 *Article) bool { return a1.Published < a2.Published }
		case 4:
			criteria = func(a1, a2 *Article) bool { return a1.TagsString() < a2.TagsString() }
		case 5:
			criteria = func(a1, a2 *Article) bool { return a1.Byline < a2.Byline }
		case 6:
			criteria = func(a1, a2 *Article) bool { return a1.URL() < a2.URL() }
		}
	}
	if criteria != nil {
		By(criteria).Sort(sorted)
	}

	return &Results{
		Query:    res.Query,
		arts:     sorted,
		Len:      len(sorted),
		facets:   res.facets, // facets don't change
		FacetLen: res.FacetLen,
	}
}

//
type Control struct {
	App *App

	obj qml.Object

	Results    *Results
	TotalArts  int
	SortColumn int
	SortOrder  int
	store      *Store

	SlurpProgress SlurpProgress
	StatusText    string
	HelpText      string
}

func NewControl(app *App, storePath string, gui qml.Object) (*Control, error) {
	var err error

	ctrl := &Control{}
	ctrl.App = app

	newStore, err := NewStore(storePath)
	if err != nil {
		return nil, err
	}
	ctrl.store = newStore

	ctrl.Results, err = NewResults(ctrl.store, "")
	if err != nil {
		return nil, err
	}

	// TODO: this is brittle, magic numbers from QML side
	ctrl.SortColumn = 3 // Published
	ctrl.SortOrder = 0

	// expose us to the qml side
	app.ctx.SetVar("ctrl", ctrl)

	// instantiate the gui
	w := app.Window.Root().ObjectByName("mainSpace")
	ctrl.obj = gui.Create(nil)
	ctrl.obj.Set("parent", w)

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

func (ctrl *Control) ApplySorting(sortColumn, sortOrder int) {
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

	arts := ArtList{}
	for _, artIdx := range artIndices {
		arts = append(arts, ctrl.Results.arts[artIdx])
	}
	err := ctrl.store.Delete(arts)
	if err != nil {
		dbug.Printf("ERROR: delete failed: %s\n", err)
		return
	}
	//	dbug.Printf("%d articles deleted\n", len(arts))

	// rerun the current query
	ctrl.setQuery(ctrl.Results.Query)
}
func (ctrl *Control) AddTag(artIndices []int, tag string) {

	arts := ArtList{}
	for _, artIdx := range artIndices {
		arts = append(arts, ctrl.Results.arts[artIdx])
	}
	affected, err := ctrl.store.AddTag(arts, tag)
	if err != nil {
		dbug.Printf("AddTag(%s): ERROR: %s\n", tag, err)
	} else {
		dbug.Printf("AddTag(%s): %d affected\n", tag, len(affected))
	}

	// rerun the current query
	ctrl.setQuery(ctrl.Results.Query)
}

func (ctrl *Control) RemoveTag(artIndices []int, tag string) {
	arts := ArtList{}
	for _, artIdx := range artIndices {
		arts = append(arts, ctrl.Results.arts[artIdx])
	}
	affected, err := ctrl.store.RemoveTag(arts, tag)
	if err != nil {
		dbug.Printf("RemoveTag(%s): ERROR: %s\n", tag, err)
	} else {
		dbug.Printf("RemoveTag(%s): %d affected\n", tag, len(affected))
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
}

func (ctrl *Control) Slurp(dayFrom, dayTo string) {

	go func() {

		ctrl.SlurpProgress = SlurpProgress{}
		prog := &ctrl.SlurpProgress

		defer func() {
			prog.InFlight = false
			qml.Changed(ctrl, &ctrl.SlurpProgress)
		}()

		ctrl.SlurpProgress.InFlight = true
		qml.Changed(ctrl, &ctrl.SlurpProgress)
		incoming := Slurp(dayFrom, dayTo)

		dbug.Printf("Slurping...\n")
		for msg := range incoming {
			if ctrl.SlurpProgress.TotalCnt%10 == 0 {
				dbug.Printf("%s\n", ctrl.SlurpProgress.String())
			}
			/*
				ctrl.StatusText = fmt.Sprintf("Slurping - receieved %d articles (%d new)",
					ctrl.SlurpProgress.GotCnt+ctrl.SlurpProgress.NewCnt,
					ctrl.SlurpProgress.NewCnt)
				qml.Changed(&ctrl, &ctrl.StatusText)
			*/
			if msg.Error != "" {
				uhoh := fmt.Sprintf("Slurp error from server: %s", msg.Error)
				ctrl.SlurpProgress.ErrorMsg = uhoh
				qml.Changed(ctrl, &ctrl.SlurpProgress)
				dbug.Printf("%s\n", uhoh)
				return
			}
			if msg.Article == nil {
				dbug.Printf("Slurp WARN: missing article\n")
				continue
			}

			art := msg.Article
			got, err := ctrl.store.FindArt(art.URLs)
			if err != nil {
				uhoh := fmt.Sprintf("FindArt() failed: %s", err)
				ctrl.SlurpProgress.ErrorMsg = uhoh
				qml.Changed(ctrl, &ctrl.SlurpProgress)
				dbug.Printf("%s\n", uhoh)
				return
			}
			if got > 0 {
				ctrl.SlurpProgress.GotCnt++
				ctrl.SlurpProgress.TotalCnt++
				qml.Changed(ctrl, &ctrl.SlurpProgress)
				continue
			}
			err = ctrl.store.Stash(art)
			if err != nil {
				uhoh := fmt.Sprintf("Stash failed: %s", err)
				ctrl.SlurpProgress.ErrorMsg = uhoh
				qml.Changed(ctrl, &ctrl.SlurpProgress)
				dbug.Printf("%s\n", uhoh)
				return
			}
			//dbug.Printf("stashed %s as %d\n", art.Headline, art.ID)
			ctrl.SlurpProgress.NewCnt++
			ctrl.SlurpProgress.TotalCnt++
			qml.Changed(ctrl, &ctrl.SlurpProgress)
		}

		dbug.Printf("Slurp finished.\n")
		ctrl.App.SetError("")
		//dbug.Printf("slurped %d (%d new)\n", gotCnt+newCnt, newCnt)

		// re-run the current query
		r2, err := NewResults(ctrl.store, ctrl.Results.Query)
		if err != nil {
			dbug.Printf("ERROR failed to refresh query: %s\n", err)
			return
		}
		ctrl.Results = r2
		qml.Changed(ctrl, &ctrl.Results)
	}()
}

func (ctrl *Control) RunScript(scriptIdx int) {
	s := ctrl.App.Scripts[scriptIdx]
	err := s.Run(ctrl.store)
	if err != nil {
		dbug.Printf("ERROR running script %s: %s\n", s.Name, err)
		ctrl.App.SetError(err.Error())
	}
	// rerun the current query
	ctrl.setQuery(ctrl.Results.Query)
}

func (ctrl *Control) Train() {
	Train(ctrl.Results.arts)
}

func (ctrl *Control) Classify() {
	Classify(ctrl.Results.arts, ctrl.store)
}
