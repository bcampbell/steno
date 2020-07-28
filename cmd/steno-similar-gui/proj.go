package main

import (
	"fmt"
	"github.com/andlabs/ui"
	"github.com/bcampbell/steno/simrep"
	"github.com/bcampbell/steno/simrep/sim"
	"github.com/bcampbell/steno/store"
	"github.com/pkg/browser"
	"io"
	"os"
	"time"
)

// holder for an opened store and it's sim.Index
type Proj struct {
	DB         *store.Store
	DBFilename string
	Index      *sim.Index
	Opts       simrep.Opts
}

// open a store and index it
func NewProject(opts *simrep.Opts, dbFile string, progFn func(int, int)) (*Proj, error) {
	proj := &Proj{}
	proj.DBFilename = dbFile
	proj.Opts = *opts

	// Open the store
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		return nil, err
	}
	proj.DB, err = store.NewWithoutIndex(dbFile, opts.Dbug, "en", loc)
	if err != nil {
		return nil, err
	}

	// build the index
	proj.Index, err = simrep.BuildIndex(proj.DB, opts, progFn)
	return proj, err
}

func singleMatchUI(proj *Proj, fenster *ui.Window) ui.Control {
	var htmlFileEntry *ui.Entry
	var txtEntry *ui.MultilineEntry
	var okBtn *ui.Button

	rethink := func() {
		if len(txtEntry.Text()) > 0 && len(htmlFileEntry.Text()) > 0 {
			okBtn.Enable()
		} else {
			okBtn.Disable()
		}
	}

	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)

	// text entry
	{
		txtEntry = ui.NewMultilineEntry()
		txtEntry.OnChanged(func(*ui.MultilineEntry) { rethink() })
		vbox.Append(ui.NewLabel("text to search for"), false)
		vbox.Append(txtEntry, true)
	}

	// output html report file
	{
		box := ui.NewHorizontalBox()
		htmlFileEntry = ui.NewEntry()
		htmlFileEntry.SetText("")
		htmlFileEntry.OnChanged(func(*ui.Entry) { rethink() })

		browseBtn := ui.NewButton("Browse...")
		browseBtn.OnClicked(func(*ui.Button) {
			f := ui.SaveFile(fenster)
			if f != "" {
				htmlFileEntry.SetText(f)
				rethink()
			}
		})
		box.Append(htmlFileEntry, true)
		box.Append(browseBtn, false)
		vbox.Append(ui.NewLabel("output HTML report"), false)
		vbox.Append(box, false)
	}

	// match Threshold %
	{
		s := ui.NewSlider(0, 100)
		s.SetValue(int(100 * proj.Opts.MatchThreshold))
		s.OnChanged(func(*ui.Slider) {
			proj.Opts.MatchThreshold = float64(s.Value()) / 100.0
			rethink()
		})
		vbox.Append(ui.NewLabel("Match Threshold %"), false)
		vbox.Append(s, false)
	}

	okBtn = ui.NewButton("Perform Search")
	okBtn.OnClicked(func(*ui.Button) {
		// Perform single match
		pw := NewProgressWindow("steno-similar", "")
		txt := txtEntry.Text()
		outputHTMLFilename := htmlFileEntry.Text()
		go func() {
			defer pw.Close()
			err := doSingleMatch(proj, txt, outputHTMLFilename, pw)
			if err != nil {
				ui.QueueMain(func() {
					errMsg := fmt.Sprintf("ERROR: %s", err)
					ui.MsgBoxError(pw.w, "steno-similar error", errMsg)
				})
				return
			}

			// launch web browser
			browser.OpenFile(outputHTMLFilename)
		}()
	})
	vbox.Append(okBtn, false)
	rethink()
	return vbox

}

func bulkMatchUI(proj *Proj, fenster *ui.Window) ui.Control {
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)

	var dbFileEntry *ui.Entry
	var htmlFileEntry *ui.Entry
	var okBtn *ui.Button

	rethink := func() {
		if len(dbFileEntry.Text()) > 0 && len(htmlFileEntry.Text()) > 0 {
			okBtn.Enable()
		} else {
			okBtn.Disable()
		}
	}

	form := ui.NewForm()
	form.SetPadded(true)
	// filename input
	{
		box := ui.NewHorizontalBox()
		dbFileEntry = ui.NewEntry()
		dbFileEntry.SetText("")
		dbFileEntry.OnChanged(func(*ui.Entry) { rethink() })

		browseBtn := ui.NewButton("Browse...")
		browseBtn.OnClicked(func(*ui.Button) {
			f := ui.OpenFile(fenster)
			if f != "" {
				dbFileEntry.SetText(f)
				rethink()
			}
		})
		box.Append(dbFileEntry, true)
		box.Append(browseBtn, false)
		form.Append("DB to match against this index", box, false)
	}

	// output html report file
	{
		box := ui.NewHorizontalBox()
		htmlFileEntry = ui.NewEntry()
		htmlFileEntry.SetText("")
		htmlFileEntry.OnChanged(func(*ui.Entry) { rethink() })

		browseBtn := ui.NewButton("Browse...")
		browseBtn.OnClicked(func(*ui.Button) {
			f := ui.SaveFile(fenster)
			if f != "" {
				htmlFileEntry.SetText(f)
				rethink()
			}
		})
		box.Append(htmlFileEntry, true)
		box.Append(browseBtn, false)
		form.Append("output HTML report", box, false)
	}

	// match Threshold %
	{
		s := ui.NewSlider(0, 100)
		s.SetValue(int(100 * proj.Opts.MatchThreshold))
		s.OnChanged(func(*ui.Slider) {
			proj.Opts.MatchThreshold = float64(s.Value()) / 100.0
			rethink()
		})
		form.Append("Match Threshold %", s, false)
	}

	vbox.Append(form, true)

	okBtn = ui.NewButton("Perform Bulk Match")
	okBtn.OnClicked(func(*ui.Button) {
		// Perform Bulk Match!
		pw := NewProgressWindow("steno-similar", "")
		otherDBFilename := dbFileEntry.Text()
		outputHTMLFilename := htmlFileEntry.Text()
		go func() {
			defer pw.Close()
			err := doBulkMatch(proj, otherDBFilename, outputHTMLFilename, pw)
			if err != nil {
				ui.QueueMain(func() {
					errMsg := fmt.Sprintf("ERROR: %s", err)
					ui.MsgBoxError(pw.w, "steno-similar error", errMsg)
				})
				return
			}

			// launch web browser
			browser.OpenFile(outputHTMLFilename)
		}()
	})
	vbox.Append(okBtn, false)
	rethink()
	return vbox
}

func doBulkMatch(proj *Proj, otherDBName string, outputHTMLFilename string, pw *ProgressWindow) error {

	pw.SetMessage(fmt.Sprintf("opening %s", otherDBName))
	pw.SetProgress(-1)

	// Open the other store
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		return err
	}

	otherDB, err := store.New(otherDBName, proj.Opts.Dbug, "en", loc)
	if err != nil {
		return err
	}

	pw.SetMessage("matching...")
	pw.SetProgress(-1)

	matches, err := simrep.FindMatches(proj.Index, otherDB, &proj.Opts)
	if err != nil {
		return err
	}
	pw.SetMessage("Generating report...")
	pw.SetProgress(0)
	outFile, err := os.Create(outputHTMLFilename)
	if err != nil {
		return err
	}
	defer outFile.Close()
	err = generateReport(proj, otherDB, matches, outFile, pw)
	if err != nil {
		return err
	}

	return nil
}

func doSingleMatch(proj *Proj, txt string, outputHTMLFilename string, pw *ProgressWindow) error {

	pw.SetMessage("matching...")
	pw.SetProgress(-1)

	hits := proj.Index.Match(txt, proj.Opts.MatchThreshold)

	pw.SetMessage("Generating report...")
	pw.SetProgress(-1)
	outFile, err := os.Create(outputHTMLFilename)
	if err != nil {
		return err
	}
	defer outFile.Close()
	//matches := map[store.ArtID][]simDocMatch{}
	//matches[0] = hits

	matchArtIDs := make([]store.ArtID, len(hits))
	metrics := make([]float64, len(hits))
	for i, dm := range hits {
		matchArtIDs[i] = store.ArtID(dm.ID)
		metrics[i] = dm.Factor
	}
	matchingArts, err := proj.DB.Fetch(matchArtIDs...)
	if err != nil {
		return err
	}

	simrep.EmitHeader(outFile, &proj.Opts)
	// fake article for report
	art := &store.Article{}
	art.Content = txt
	simrep.EmitMatches(outFile, art, matchingArts, metrics)
	simrep.EmitFooter(outFile)

	return nil
}

func generateReport(proj *Proj, otherDB *store.Store, matches map[store.ArtID][]sim.DocMatch,
	out io.Writer, pw *ProgressWindow) error {

	// output report
	db1 := proj.DB
	db2 := otherDB

	//dbug.Printf("output report...\n")
	simrep.EmitHeader(out, &proj.Opts)
	// for each article...
	cnt := 0
	for art2ID, m := range matches {
		// fetch the article
		foo, err := db2.Fetch(art2ID)
		if err != nil {
			return err
		}
		art := foo[0]

		// fetch all the articles it matched
		matchArtIDs := make([]store.ArtID, len(m))
		metrics := make([]float64, len(m))
		for i, dm := range m {
			matchArtIDs[i] = store.ArtID(dm.ID)
			metrics[i] = dm.Factor
		}
		matchingArts, err := db1.Fetch(matchArtIDs...)
		if err != nil {
			return err
		}

		simrep.EmitMatches(out, art, matchingArts, metrics)

		cnt++
		pw.SetProgress((cnt * 100) / len(matches))
	}

	simrep.EmitFooter(out)
	return nil
}

func projectWindow(proj *Proj) {
	winTitle := fmt.Sprintf("steno-similar: %s", proj.DBFilename)
	fenster := ui.NewWindow(winTitle, 640, 480, true)
	fenster.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true // destroy the window
	})
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)

	{
		lMsg := fmt.Sprintf("Currently indexed: %s\n", proj.DBFilename)
		l := ui.NewLabel(lMsg)
		vbox.Append(l, false)
	}

	tabs := ui.NewTab()
	tabs.Append("Single", singleMatchUI(proj, fenster))
	tabs.Append("Bulk Match", bulkMatchUI(proj, fenster))
	tabs.SetMargined(0, true)
	tabs.SetMargined(1, true)
	vbox.Append(tabs, true)

	fenster.SetChild(vbox)
	fenster.SetMargined(true)
	fenster.Show()
}
