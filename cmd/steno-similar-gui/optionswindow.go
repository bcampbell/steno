package main

import (
	"fmt"
	"github.com/andlabs/ui"
	"semprini/steno/steno/simrep"
)

// show a window to pick db file and options and kick off the indexing...
func doOptionsWindow(opts *simrep.Opts, filename string) {
	fenster := ui.NewWindow("steno-similar", 640, 480, true)
	fenster.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})

	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)
	form := ui.NewForm()
	form.SetPadded(true)

	// filename input
	{
		box := ui.NewHorizontalBox()
		fileEntry := ui.NewEntry()
		fileEntry.SetText(filename)
		fileEntry.OnChanged(func(*ui.Entry) {
			filename = fileEntry.Text()
		})

		browseBtn := ui.NewButton("Browse...")
		browseBtn.OnClicked(func(*ui.Button) {
			f := ui.OpenFile(fenster)
			if f != "" {
				filename = f
				fileEntry.SetText(f)
			}
		})
		box.Append(fileEntry, true)
		box.Append(browseBtn, false)
		form.Append(".db to index", box, false)
	}

	// ngram size
	{
		spin := ui.NewSpinbox(1, 100)
		spin.SetValue(opts.NGramSize)
		spin.OnChanged(func(*ui.Spinbox) {
			opts.NGramSize = spin.Value()
		})
		form.Append("ngram size", spin, false)
	}

	// minwords size
	{
		s := ui.NewSpinbox(0, 5000)
		s.SetValue(opts.MinWords)
		s.OnChanged(func(*ui.Spinbox) {
			opts.MinWords = s.Value()
		})
		form.Append("Min Words", s, false)
	}

	// match Threshold %
	{
		s := ui.NewSlider(0, 100)
		s.SetValue(int(100 * opts.MatchThreshold))
		s.OnChanged(func(*ui.Slider) {
			opts.MatchThreshold = float64(s.Value()) / 100.0
		})
		form.Append("Match Threshold %", s, false)
	}

	// language
	{
		langCodes := []string{"en", "ru", "es"}
		langLabels := []string{"English", "Russian", "Spanish"}

		langPicker := ui.NewCombobox()
		sel := -1
		for i, l := range langLabels {
			langPicker.Append(l)
			if opts.Lang == langCodes[i] {
				sel = i
			}
		}
		langPicker.SetSelected(sel)
		langPicker.OnSelected(func(*ui.Combobox) {
			i := langPicker.Selected()
			if i == -1 {
				opts.Lang = ""
			} else {
				opts.Lang = langCodes[i]
			}
		})
		form.Append("Language for indexing", langPicker, false)
	}

	/*
		MinWords       int
		NGramSize      int
		MatchThreshold float64
		IgnoreSameID   bool
		Lang           string
		Dbug store.Logger
	*/

	vbox.Append(form, false)

	indexBtn := ui.NewButton("Index")
	indexBtn.OnClicked(func(*ui.Button) {
		fmt.Printf("<%s>  %v\n", filename, opts)
		//fenster.Destroy()
		//		ui.Quit()
	})
	vbox.Append(indexBtn, false)

	fenster.SetChild(vbox)
	fenster.SetMargined(true)
	fenster.Show()
}
