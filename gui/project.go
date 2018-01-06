package gui

import (
	"github.com/bcampbell/ui"
	"semprini/steno/steno/store"
)

type Project struct {
	App   *App
	Store *store.Store
	Views map[*ProjView]struct{}
}

func NewProject(db *store.Store, app *App) (*Project, error) {
	proj := &Project{}
	proj.App = app
	proj.Store = db
	proj.Views = make(map[*ProjView]struct{})

	var err error
	_, err = NewProjView(proj)
	if err != nil {
		return nil, err
	}

	return proj, err
}

func (proj *Project) attachView(v *ProjView) {
	proj.Views[v] = struct{}{}
}

func (proj *Project) detachView(v *ProjView) {
	delete(proj.Views, v)

	if len(proj.Views) == 0 {
		ui.Quit()
	}
}
