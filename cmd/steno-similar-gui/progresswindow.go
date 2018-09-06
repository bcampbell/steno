package main

import (
	"fmt"
	"github.com/andlabs/ui"
	"semprini/steno/steno/simrep"
)

//

type ProgressWindow struct {
	w        *ui.Window
	bar      *ui.ProgressBar
	OnCancel func()
}

func NewProgressWindow(title, msg string) *ProgressWindow {
	prog := &ProgressWindow{}

	prog.w = ui.NewWindow(title, 640, 480, true)
	prog.w.OnClosing(func(*ui.Window) bool {
		if prog.OnCancel {
			prog.OnCancel()
		}
		return false
	})

	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)

	msgLabel := ui.NewLabel(msg)
	vbox.Append(msgLabel, false)

	prog.bar = ui.NewProgressBar()
	vbox.Append(prog.bar, false)

	cancel := ui.NewButton("Cancel")
	cancel.OnClicked(func(*ui.Button) {
		if prog.OnCancel {
			prog.OnCancel()
		}
	})

	prog.w.SetChild(vbox)
	prog.w.SetMargined(true)
	prog.w.Show()
	return prog
}

func (prog *ProgressWindow) Close() {
	ui.QueueMain(prog.w.Destroy)
}

func (prog *ProgressWindow) SetProgress(int i) {
	ui.QueueMain(func() {
		prog.bar.SetValue(i)
	})
}
