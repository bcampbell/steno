package main

import (
	"fmt"
	"gopkg.in/qml.v1"
)

type Control struct {
	Window *qml.Window

	CurrentQuery string
	arts         ArtList
	Len          int
	TotalArts    int

	store *Store
}

func NewControl() (*Control, error) {
	var err error
	engine := qml.NewEngine()
	ctx := engine.Context()

	ctrl := &Control{}
	ctrl.store = DummyStore()
	// populate the initial query
	ctrl.arts, err = ctrl.store.AllArts()
	if err != nil {
		return nil, err
	}
	ctrl.Len = len(ctrl.arts)
	ctrl.TotalArts = ctrl.store.TotalArts()

	// expose us to the qml side
	ctx.SetVar("ctrl", ctrl)

	component, err := engine.LoadFile("fook.qml")
	if err != nil {
		return nil, err
	}

	// instantiate the gui
	ctrl.Window = component.CreateWindow(nil)
	ctrl.Window.Show()

	/*
		obj := window.Root().ObjectByName("query")
		obj.Set("text", "")
		fmt.Printf("%v\n", obj)
	*/

	return ctrl, nil
}

func (ctrl *Control) Art(idx int) *Article {
	return ctrl.arts[idx]
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

func (ctrl *Control) SetDB(fileName string) {
	fmt.Printf("SetDB(%s)\n", fileName)

	ctrl.store.Close()
	newStore, err := NewStore(fileName)
	if err != nil {
		dbug.Printf("SetDB error: %s\n", err)
		ctrl.store = DummyStore()
	} else {
		ctrl.store = newStore
	}

	// populate the initial query
	ctrl.arts, err = ctrl.store.AllArts()
	if err != nil {
		return
	}
	ctrl.Len = len(ctrl.arts)
	ctrl.TotalArts = ctrl.store.TotalArts()

	ctrl.forceArtsRefresh()
}

func (ctrl *Control) AddTag(artIndices []int, tag string) {

	cnt := 0
	for _, artIdx := range artIndices {
		art := ctrl.arts[artIdx]
		if art.AddTag(tag) {
			cnt++
		}
	}
	dbug.Printf("addTag(%s): changed %d articles\n", tag, cnt)

	// TODO: signal changed arts instead of whole list
	ctrl.forceArtsRefresh()
}

func (ctrl *Control) RemoveTag(artIndices []int, tag string) {
	dbug.Printf("removeTag(%s)\n", tag)
	cnt := 0
	for _, artIdx := range artIndices {
		art := ctrl.arts[artIdx]
		if art.RemoveTag(tag) {
			cnt++

		}
	}
	dbug.Printf("removeTag(%s): changed %d articles\n", tag, cnt)
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
}
