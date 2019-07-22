package gui

import (
	"github.com/andlabs/ui"
)

type ProgressWindow struct {
	window *ui.Window
	status *ui.Label
	//	bar    *ui.ProgressBar
}

func NewProgressWindow(title string) *ProgressWindow {
	p := &ProgressWindow{}

	p.status = ui.NewLabel("messge goes here...")

	box := ui.NewVerticalBox()
	box.Append(p.status, true)

	w := ui.NewWindow(title, 500, 200, false)
	p.window = w
	w.SetChild(box)
	w.SetMargined(true)
	w.Show()
	return p
}

func (p *ProgressWindow) Close() {
	p.window.Destroy()
}

func (p *ProgressWindow) SetStatus(msg string) {
	p.status.SetText(msg)
}
