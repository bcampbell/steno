package main

import (
	"github.com/bcampbell/ui"
	"semprini/steno/steno"
	"semprini/steno/steno/store"
	"time"
)

type Proj struct {
	store   *store.Store
	results *steno.Results

	// controls
	c struct {
		query *ui.Entry
		table *ui.Table
		model *ui.TableModel
	}
}

func NewProj(storePath string) (*Proj, error) {

	proj := &Proj{}

	newStore, err := store.New(storePath, dbug, "en", time.Local)
	if err != nil {
		return nil, err
	}
	proj.store = newStore

	proj.results, err = steno.NewResults(proj.store, "")
	if err != nil {
		return nil, err
	}

	input := ui.NewEntry()
	button := ui.NewButton("Greet")
	greeting := ui.NewLabel("")
	box := ui.NewVerticalBox()
	box.Append(ui.NewLabel("Enter your name:"), false)
	box.Append(input, false)
	box.Append(button, false)
	box.Append(greeting, false)

	window := ui.NewWindow("Steno", 700, 400, false)
	window.SetMargined(true)
	window.SetChild(box)
	button.OnClicked(func(*ui.Button) {
		greeting.SetText("Hello, " + input.Text() + "!")
	})
	window.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	window.Show()

	return proj, err
}
