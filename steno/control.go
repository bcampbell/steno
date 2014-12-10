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
	GotCnt   int
	NewCnt   int
	InFlight bool
}

func (p *SlurpProgress) String() string {
	return fmt.Sprintf("Received %d articles (%d new)", p.GotCnt+p.NewCnt, p.NewCnt)
}

type Control struct {
	App *App

	obj qml.Object

	CurrentQuery string
	arts         ArtList
	Len          int
	TotalArts    int
	SortColumn   int
	SortOrder    int
	FacetLen     int
	facets       []*Facet
	store        *Store

	SlurpProgress SlurpProgress
	StatusText    string
	HelpText      string
}

func NewControl(app *App, storePath string, gui qml.Object) (*Control, error) {
	var err error

	ctrl := &Control{}
	ctrl.App = app

	err = ctrl.SetDB(storePath)
	if err != nil {
		return nil, err
	}

	// expose us to the qml side
	app.ctx.SetVar("ctrl", ctrl)
	w := app.Window.Root().ObjectByName("mainSpace")
	// instantiate the gui

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

func (ctrl *Control) Art(idx int) *Article {
	return ctrl.arts[idx]
}
func (ctrl *Control) Facet(idx int) *Facet {
	return ctrl.facets[idx]
}

func (ctrl *Control) ApplySorting(sortColumn, sortOrder int) {
	// order: 1: ascending, 0: descending
	dbug.Printf("new sorting: %d %d\n", sortColumn, sortOrder)

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
		By(criteria).Sort(ctrl.arts)
	}
	ctrl.forceArtsRefresh()
}

// TODO: provide a function to validate query...

func (ctrl *Control) SetQuery(q string) {
	if ctrl.CurrentQuery == q {
		return
	}
	ctrl.CurrentQuery = q

	fmt.Printf("SetQuery(%s)\n", q)
	var err error
	if q == "" {
		ctrl.arts, err = ctrl.store.AllArts()
	} else {
		ctrl.arts, err = ctrl.store.Search(q)
	}
	if err != nil {
		// TODO: display error...
		dbug.Printf("Error in allArts(): %s\n", err)
		//os.Exit(1)
	}
	ctrl.Len = len(ctrl.arts)
	ctrl.TotalArts = ctrl.store.TotalArts()

	ctrl.updateFacets()
	ctrl.forceArtsRefresh()
}

/*
func (ctrl *Control) OLDLoadDB(fileName string) {
	fmt.Printf("loadDB(%s)\n", fileName)
	var err error

	coll = badger.NewCollection(&Article{})
	coll, err = loadDB(fileName)
	if err != nil {
		dbug.Printf("loadDB error: %s\n", err)
	}
	// populate the initial query
	ctrl.arts, err = allArts()
	if err != nil {
		dbug.Printf("Query error: %s\n", err)
	}
	ctrl.Len = len(ctrl.arts)
	ctrl.TotalArts = coll.Count()
	ctrl.forceArtsRefresh()

	dbug.Printf("Save to sqlite!\n")
	err = debadger(ctrl.arts, "fancy.db")
	if err != nil {
		dbug.Printf("debadger error: %s\n", err)
	}
	dbug.Printf("Load complete\n")
}
*/

func (ctrl *Control) SetDB(fileName string) error {
	fmt.Printf("SetDB(%s)\n", fileName)

	//	ctrl.store.Close()
	newStore, err := NewStore(fileName)
	if err != nil {
		dbug.Printf("SetDB error: %s\n", err)

		return err
	}
	ctrl.store = newStore

	// populate the initial query
	ctrl.arts, err = ctrl.store.AllArts()
	if err != nil {
		return err
	}
	ctrl.Len = len(ctrl.arts)
	ctrl.TotalArts = ctrl.store.TotalArts()
	ctrl.updateFacets()
	//	ctrl.forceArtsRefresh()
	return nil
}

func (ctrl *Control) updateFacets() {
	tab := map[string]int{}
	for _, art := range ctrl.arts {
		tab[art.Pub]++
	}
	ctrl.facets = []*Facet{}
	for txt, cnt := range tab {
		ctrl.facets = append(ctrl.facets, &Facet{txt, cnt})
	}
	ctrl.FacetLen = len(ctrl.facets)
}

func (ctrl *Control) AddTag(artIndices []int, tag string) {

	arts := ArtList{}
	for _, artIdx := range artIndices {
		arts = append(arts, ctrl.arts[artIdx])
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
		arts = append(arts, ctrl.arts[artIdx])
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
	foo := ctrl.Len
	ctrl.Len = 0
	qml.Changed(ctrl, &ctrl.Len)
	ctrl.Len = foo
	qml.Changed(ctrl, &ctrl.Len)

	foo = ctrl.FacetLen
	ctrl.FacetLen = 0
	qml.Changed(ctrl, &ctrl.FacetLen)
	ctrl.FacetLen = foo
	qml.Changed(ctrl, &ctrl.FacetLen)
}

func (ctrl *Control) Slurp(dayFrom, dayTo string) {

	ctrl.SlurpProgress = SlurpProgress{}

	defer func() {
		ctrl.SlurpProgress.InFlight = false
		qml.Changed(ctrl, &ctrl.SlurpProgress.InFlight)
	}()

	ctrl.SlurpProgress.InFlight = true
	qml.Changed(ctrl, &ctrl.SlurpProgress.InFlight)
	incoming := Slurp(dayFrom, dayTo)

	for msg := range incoming {
		/*
			ctrl.StatusText = fmt.Sprintf("Slurping - receieved %d articles (%d new)",
				ctrl.SlurpProgress.GotCnt+ctrl.SlurpProgress.NewCnt,
				ctrl.SlurpProgress.NewCnt)
			qml.Changed(&ctrl, &ctrl.StatusText)
		*/
		ctrl.App.SetError(ctrl.SlurpProgress.String())
		if msg.Error != "" {
			dbug.Printf("Slurp error from server: %s\n", msg.Error)
			ctrl.App.SetError(fmt.Sprintf("Slurp error from server: %s\n", msg.Error))
			return
		}
		if msg.Article == nil {
			dbug.Printf("Slurp WARN: missing article\n")
			continue
		}

		art := msg.Article
		got, err := ctrl.store.FindArt(art.URLs)
		if err != nil {
			// TODO: display error
			dbug.Printf("ERROR FindArt() failed: %s\n", err)
			return
		}
		if got > 0 {
			ctrl.SlurpProgress.GotCnt++
			qml.Changed(ctrl, &ctrl.SlurpProgress.GotCnt)
			continue
		}
		err = ctrl.store.Stash(art)
		if err != nil {
			dbug.Printf("ERROR: Stash failed: %s\n", err)
			ctrl.App.SetError(fmt.Sprintf("Stash failed: %s\n", msg.Error))
			return
		}
		//dbug.Printf("stashed %s as %d\n", art.Headline, art.ID)
		ctrl.SlurpProgress.NewCnt++
		qml.Changed(ctrl, &ctrl.SlurpProgress.NewCnt)
	}

	dbug.Printf("Slurp finished.\n")
	ctrl.App.SetError("")
	//dbug.Printf("slurped %d (%d new)\n", gotCnt+newCnt, newCnt)
	ctrl.forceArtsRefresh()
}
