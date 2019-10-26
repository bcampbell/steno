package main

import (
	"github.com/bcampbell/steno/steno/store"
)

type Project struct {
	//App   *App
	Store *store.Store
	Views map[View]struct{}
}

type View interface {
	OnArtsModified(store.ArtList)
	OnArtsAdded(store.ArtList)
	OnArtsDeleted(store.ArtList)
}

func NewProject(db *store.Store) (*Project, error) {
	proj := &Project{}
	//proj.App = app
	proj.Store = db
	proj.Views = make(map[View]struct{})

	return proj, nil
}

func (proj *Project) attachView(v View) {
	proj.Views[v] = struct{}{}
}

func (proj *Project) detachView(v View) {
	delete(proj.Views, v)

	if len(proj.Views) == 0 {
		// XYZZY QUIT!
		//		ui.Quit()
	}
}

func (proj *Project) ArtsAdded(ids store.ArtList) {
	for v, _ := range proj.Views {
		v.OnArtsAdded(ids)
	}
}
func (proj *Project) ArtsDeleted(ids store.ArtList) {
	for v, _ := range proj.Views {
		v.OnArtsDeleted(ids)
	}
}
func (proj *Project) ArtsModified(ids store.ArtList) {
	for v, _ := range proj.Views {
		v.OnArtsModified(ids)
	}
}
