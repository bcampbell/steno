package gui

import (
	"fmt"
	"github.com/bcampbell/steno/steno"
	"github.com/bcampbell/steno/steno/store"
	"github.com/bcampbell/ui"
	"time"
)

type FOO struct{}

func (f *FOO) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

var dbug = &FOO{}

type App struct {
	App  *steno.App
	proj *Project
	//	Projects []*Proj
}

func NewApp() (*App, error) {
	var err error
	app := &App{}
	app.App, err = steno.NewApp()
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
		if dbFilename == "" {
			app.startupWindow()
		} else {
			app.openProject(dbFilename)
		}
	})
	return err
}

func (app *App) startupWindow() {
	window := ui.NewWindow("Steno startup", 400, 400, false)
	window.SetMargined(true)

	box := ui.NewVerticalBox()
	box.SetPadded(true)

	newButton := ui.NewButton("New (blank) Project")
	openButton := ui.NewButton("Open existing project")

	box.Append(newButton, false)
	box.Append(openButton, false)

	newButton.OnClicked(func(b *ui.Button) {
		f := ui.SaveFile(window)
		if f != "" {
			err := app.openProject(f)
			if err != nil {
				ui.MsgBox(window, "Poop", err.Error())
			} else {
				window.Destroy()
			}
		}
	})
	openButton.OnClicked(func(b *ui.Button) {
		f := ui.OpenFile(window)
		if f != "" {
			err := app.openProject(f)
			if err != nil {
				ui.MsgBox(window, "Poop", err.Error())
			} else {
				window.Destroy()
			}
		}
	})

	window.OnClosing(func(w *ui.Window) bool {
		ui.Quit()
		return true
	})

	window.SetChild(box)
	window.Show()
}
