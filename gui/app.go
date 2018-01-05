package gui

import (
	"fmt"
	"github.com/bcampbell/ui"
	"semprini/steno/steno"
	"semprini/steno/steno/store"
	"time"
)

type FOO struct{}

func (f *FOO) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

var dbug = &FOO{}

type App struct {
	app *steno.App

	proj *Project
	//	Projects []*Proj
}

func NewApp() (*App, error) {
	var err error
	app := &App{}
	app.app, err = steno.NewApp()
	return app, err
}

func (app *App) openProject(storePath string) error {
	db, err := store.New(storePath, dbug, "en", time.Local)
	if err != nil {
		return err
	}
	proj, err := NewProject(db, app)
	_ = proj
	return err
}

func (app *App) Run(dbFilename string) error {
	err := ui.Main(func() {
		app.openProject(dbFilename)
	})
	return err
}
