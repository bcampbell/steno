package gui

import (
	"github.com/bcampbell/ui"
	"time"
)

//type SlurpDialog struct {

//}

func slurpDialog(okFn func(string, time.Time, int), cancelFn func()) {

	box := ui.NewVerticalBox()
	dayLabel := ui.NewLabel("Start day (YYYY-MM-DD)")
	dayPicker := ui.NewEntry()

	numDaysLabel := ui.NewLabel("Number of days")
	numDaysSlider := ui.NewSlider(1, 14)

	box.Append(dayLabel, false)
	box.Append(dayPicker, false)
	box.Append(numDaysLabel, false)
	box.Append(numDaysSlider, false)

	box.Append(ui.NewHorizontalSeparator(), false)

	buttonsBox := ui.NewHorizontalBox()
	okButton := ui.NewButton("Slurp")
	cancelButton := ui.NewButton("Cancel")
	buttonsBox.Append(okButton, false)
	buttonsBox.Append(cancelButton, false)

	box.Append(buttonsBox, false)
	//	numDaysSlider.Value()

	win := ui.NewWindow("Slurp", 600, 300, false)
	win.SetMargined(true)
	win.SetChild(box)

	var day time.Time

	validate := func() {
		d, err := time.Parse("2006-01-02", dayPicker.Text())
		if err != nil {
			okButton.Disable()
			return
		}

		day = d
		okButton.Enable()
	}

	// wire it all up
	okButton.OnClicked(func(b *ui.Button) {
		win.Destroy()
		okFn("", day, numDaysSlider.Value())
	})
	cancelButton.OnClicked(func(b *ui.Button) {
		win.Destroy()
		cancelFn()
	})
	dayPicker.OnChanged(func(e *ui.Entry) {
		validate()
	})

	win.OnClosing(func(w *ui.Window) bool {
		win.Destroy()
		cancelFn()
		return false // already gone
	})

	validate()

	win.Show()
}
