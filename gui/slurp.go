package gui

import (
	"fmt"
	"github.com/bcampbell/steno/steno"
	"github.com/bcampbell/ui"
	"strconv"
	"time"
)

//type SlurpDialog struct {

//}

func slurpDialog(sources []steno.SlurpSource, okFn func(steno.SlurpSource, time.Time, int), cancelFn func()) {

	box := ui.NewVerticalBox()
	dayLabel := ui.NewLabel("Start day (YYYY-MM-DD)")
	dayPicker := ui.NewEntry()

	srcLabel := ui.NewLabel("Source")
	srcCombo := ui.NewCombobox()
	for _, s := range sources {
		srcCombo.Append(fmt.Sprintf("%s (%s)", s.Name, s.Loc))
	}

	numDaysLabel := ui.NewLabel("Number of days")
	numDaysEntry := ui.NewEntry()

	box.Append(srcLabel, false)
	box.Append(srcCombo, false)
	box.Append(dayLabel, false)
	box.Append(dayPicker, false)
	box.Append(numDaysLabel, false)
	box.Append(numDaysEntry, false)

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
	var nDays int

	validate := func() {
		ok := true

		// demand a valid date...
		d, err := time.Parse("2006-01-02", dayPicker.Text())
		if err == nil {
			day = d
		} else {
			ok = false
		}

		// ...and valid num of days...
		n, err := strconv.Atoi(numDaysEntry.Text())
		if err == nil && n >= 1 && n <= 14 {
			nDays = n
		} else {
			ok = false
		}

		// ...and a valid slurp source
		if srcCombo.Selected() < 0 {
			ok = false
		}

		if ok {
			okButton.Enable()
		} else {
			okButton.Disable()
		}
	}

	// wire it all up
	okButton.OnClicked(func(b *ui.Button) {
		win.Destroy()
		idx := srcCombo.Selected()
		if idx < 0 {
			cancelFn()
			return
		}
		okFn(sources[idx], day, nDays)
	})
	cancelButton.OnClicked(func(b *ui.Button) {
		win.Destroy()
		cancelFn()
	})
	dayPicker.OnChanged(func(e *ui.Entry) { validate() })
	numDaysEntry.OnChanged(func(e *ui.Entry) { validate() })
	srcCombo.OnSelected(func(c *ui.Combobox) { validate() })

	win.OnClosing(func(w *ui.Window) bool {
		win.Destroy()
		cancelFn()
		return false // already gone
	})

	validate()

	win.Show()
}
