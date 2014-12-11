package main

import (
	"fmt"
	"gopkg.in/qml.v1"
	//	"strings"
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
	res := Results{Query: query}

	var err error
	if query == "" {
		res.arts, err = store.AllArts()
	} else {
		res.arts, err = store.Search(query)
	}
	if err != nil {
		return nil, err
	}

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

	return &res, nil
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
func (res *Results) Sort(sortColumn, sortOrder int) *Results {
	// order: 1: ascending, 0: descending
	dbug.Printf("new sorting: %d %d\n", sortColumn, sortOrder)

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
			criteria = func(a1, a2 *Article) bool { return a1.Published > a2.Published }
		}
	} else if sortOrder == 1 {
		switch sortColumn {
		case 0:
			criteria = func(a1, a2 *Article) bool { return a1.Headline < a2.Headline }
		case 1:
			criteria = func(a1, a2 *Article) bool { return a1.Pub < a2.Pub }
		case 2:
			criteria = func(a1, a2 *Article) bool { return a1.Published < a2.Published }
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
	dbug.Printf("Close db\n")
	ctrl.obj.Destroy()
	ctrl.store.Close()
	//ctrl.Window.Hide()
}

func (ctrl *Control) ApplySorting(sortColumn, sortOrder int) {
	ctrl.Results = ctrl.Results.Sort(sortColumn, sortOrder)
	qml.Changed(ctrl, &ctrl.Results)
}

// TODO: provide a function to validate query...

func (ctrl *Control) SetQuery(q string) {
	if q == ctrl.Results.Query {
		return
	}

	res, err := NewResults(ctrl.store, q)
	if err != nil {
		//TODO: show error
		dbug.Printf("Search error: %s", err)
		return
	}

	fmt.Printf("SetQuery(%s)\n", q)
	ctrl.TotalArts = ctrl.store.TotalArts()
	ctrl.Results = res
	qml.Changed(ctrl, &ctrl.Results)
	fmt.Printf("END SetQuery(%s)\n", q)

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

	// TODO: signal changed arts instead of whole list
	ctrl.forceArtsRefresh()
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

	// TODO: signal changed arts instead of whole list
	ctrl.forceArtsRefresh()
}

func (ctrl *Control) forceArtsRefresh() {
	// horrible fudge to force tableview to rethink itself,
	// until go-qml lets you use a proper type as model.

	// create a new Results, with the same data
	r := *ctrl.Results
	ctrl.Results = &r
	qml.Changed(ctrl, &ctrl.Results)
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
				uhoh := fmt.Sprintf("Stash failed: %s", msg.Error)
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
