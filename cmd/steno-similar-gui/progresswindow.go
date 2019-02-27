package main

import (
	//	"fmt"
	"github.com/andlabs/ui"
	//	"github.com/bcampbell/steno/steno/simrep"
)

//

type ProgressWindow struct {
	w        *ui.Window
	bar      *ui.ProgressBar
	msgLabel *ui.Label
	OnCancel func()
}

func NewProgressWindow(title, msg string) *ProgressWindow {
	prog := &ProgressWindow{}

	prog.w = ui.NewWindow(title, 640, 480, true)
	prog.w.OnClosing(func(*ui.Window) bool {
		if prog.OnCancel != nil {
			prog.OnCancel()
		}
		return false
	})

	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)

	prog.msgLabel = ui.NewLabel(msg)
	vbox.Append(prog.msgLabel, false)

	prog.bar = ui.NewProgressBar()
	vbox.Append(prog.bar, false)

	cancel := ui.NewButton("Cancel")
	cancel.OnClicked(func(*ui.Button) {
		if prog.OnCancel != nil {
			prog.OnCancel()
		}
	})

	prog.w.SetChild(vbox)
	prog.w.SetMargined(true)
	prog.w.Show()
	return prog
}

// safe to call from any goroutine
func (prog *ProgressWindow) Close() {
	ui.QueueMain(prog.w.Destroy)
}

// set progress 0..100 (or -1 for indefinite)
// safe to call from any goroutine
func (prog *ProgressWindow) SetProgress(i int) {
	ui.QueueMain(func() {
		prog.bar.SetValue(i)
	})
}

func (prog *ProgressWindow) SetMessage(msg string) {
	ui.QueueMain(func() {
		prog.msgLabel.SetText(msg)
	})

}
