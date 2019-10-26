package gui

import (
	"fmt"
	"github.com/andlabs/ui"
	"github.com/bcampbell/steno/steno"
	"github.com/bcampbell/steno/steno/store"
	"os"
	"strings"
	"time"
)

// view onto an open store
type ProjView struct {
	Proj    *Project
	results *steno.Results

	selected map[int]struct{}

	// controls
	c struct {
		window        *ui.Window
		query         *ui.Entry
		table         *ui.Table
		model         *ui.TableModel
		resultSummary *ui.Label
		selSummary    *ui.Label
		showArt       *ui.Button

		tagEntry        *ui.Entry
		addTagButton    *ui.Button
		removeTagButton *ui.Button
	}
}

// View implementation
func (v *ProjView) OnArtsModified(store.ArtList) {
	v.SetQuery(v.results.Query) // re-run current query
}

func (v *ProjView) OnArtsAdded(store.ArtList) {
	v.SetQuery(v.results.Query)
}

func (v *ProjView) OnArtsDeleted(store.ArtList) {
	v.SetQuery(v.results.Query)
}

// TableModelHandler support
// TODO: should be wrapper around Results?

func (v *ProjView) ColumnTypes(m *ui.TableModel) []ui.TableValue {
	return []ui.TableValue{
		ui.TableInt(0),
		ui.TableString(""),
		ui.TableString(""),
		ui.TableString(""),
		ui.TableString(""),
		ui.TableString(""),
	}
}

func (v *ProjView) NumRows(m *ui.TableModel) int {
	return v.results.Len
}

func (v *ProjView) CellValue(m *ui.TableModel, row int, col int) ui.TableValue {
	art := v.results.Art(row)
	switch col {
	case 0:
		_, present := v.selected[row]
		if present {
			return ui.TableInt(1)
		} else {
			return ui.TableInt(0)
		}
	case 1:
		return ui.TableString(art.CanonicalURL)
	case 2:
		return ui.TableString(art.Headline)
	case 3:
		return ui.TableString(art.Published)
	case 4:
		return ui.TableString(art.Pub)
	case 5:
		return ui.TableString(strings.Join(art.Tags, " "))
	}
	panic("unreachable")
}

func (v *ProjView) SetCellValue(m *ui.TableModel, row int, col int, value ui.TableValue) {
	if col == 0 {
		sel := value.(ui.TableInt)
		if sel == ui.TableFalse {
			// remove.
			delete(v.selected, row)
		} else {
			v.selected[row] = struct{}{}
		}
		v.rethinkSelectionSummary()
	}
}

//
func NewProjView(proj *Project) (*ProjView, error) {

	v := &ProjView{}
	v.selected = make(map[int]struct{})
	v.Proj = proj
	var err error
	v.results, err = steno.NewResults(v.Proj.Store, "")
	if err != nil {
		return nil, err
	}

	box := ui.NewVerticalBox()
	box.SetPadded(true)

	// stand-in menu
	box.Append(v.buildToolbar(), false)

	// query entry
	{
		qbox := ui.NewHorizontalBox()
		qbox.SetPadded(true)
		v.c.query = ui.NewSearchEntry()
		/*
			v.c.query.OnChanged(func(e *ui.Entry) {
				fmt.Printf("changed.\n")
				v.SetQuery(v.c.query.Text())
			})
		*/
		button := ui.NewButton("Search")
		button.OnClicked(func(but *ui.Button) { v.SetQuery(v.c.query.Text()) })
		qbox.Append(v.c.query, true)
		qbox.Append(button, false)
		box.Append(qbox, false)
	}

	// result summary
	{
		v.c.resultSummary = ui.NewLabel("")
		hbox := ui.NewHorizontalBox()
		hbox.SetPadded(true)
		hbox.Append(v.c.resultSummary, false)

		box.Append(hbox, false)
	}
	// selection summary & tagging
	{
		v.c.selSummary = ui.NewLabel("")
		v.c.showArt = ui.NewButton("Show")
		hbox := ui.NewHorizontalBox()
		hbox.SetPadded(true)
		hbox.Append(v.c.selSummary, false)
		hbox.Append(v.c.showArt, false)

		// tagging
		v.c.tagEntry = ui.NewEntry()
		v.c.tagEntry.OnChanged(func(e *ui.Entry) { v.rethinkSelectionSummary() })
		v.c.addTagButton = ui.NewButton("Add Tag")
		v.c.addTagButton.OnClicked(func(b *ui.Button) {
			artIDs := make(store.ArtList, 0, len(v.selected))
			for row, _ := range v.selected {
				artIDs = append(artIDs, v.results.Arts[row])
			}
			v.DoAddTags(artIDs, v.c.tagEntry.Text())
		})
		v.c.removeTagButton = ui.NewButton("Remove Tag")
		v.c.removeTagButton.OnClicked(func(b *ui.Button) {
			artIDs := make(store.ArtList, 0, len(v.selected))
			for row, _ := range v.selected {
				artIDs = append(artIDs, v.results.Arts[row])
			}
			v.DoRemoveTags(artIDs, v.c.tagEntry.Text())
		})

		hbox.Append(v.c.tagEntry, false)
		hbox.Append(v.c.addTagButton, false)
		hbox.Append(v.c.removeTagButton, false)

		v.c.showArt.OnClicked(func(b *ui.Button) {
			if len(v.selected) > 0 {
				for row, _ := range v.selected {
					art := v.results.Art(row)
					NewArtView(v.Proj, art)
					break
				}

				// TODO: make db access explict! + proper error handling
			}
		})

		box.Append(hbox, false)
	}

	// set up results table
	{
		v.c.model = ui.NewTableModel(v)
		v.c.table = ui.NewTable(&ui.TableParams{
			Model: v.c.model,
			RowBackgroundColorModelColumn: -1,
		})

		v.c.table.AppendCheckboxColumn("sel", 0, ui.TableModelColumnAlwaysEditable)
		v.c.table.AppendTextColumn("URL", 1, ui.TableModelColumnNeverEditable, nil)
		v.c.table.AppendTextColumn("headline", 2, ui.TableModelColumnNeverEditable, nil)
		v.c.table.AppendTextColumn("published", 3, ui.TableModelColumnNeverEditable, nil)
		v.c.table.AppendTextColumn("pub", 4, ui.TableModelColumnNeverEditable, nil)
		v.c.table.AppendTextColumn("tags", 5, ui.TableModelColumnNeverEditable, nil)
		/*
			v.c.table.OnSelectionChanged(func(t *ui.Table) {
				v.selected = v.c.table.CurrentSelection()
				v.rethinkSelectionSummary()
			})
		*/

		box.Append(v.c.table, true)
	}

	//

	window := ui.NewWindow("Steno", 700, 400, true)
	window.SetMargined(true)
	window.SetChild(box)

	window.OnClosing(func(*ui.Window) bool {
		v.Proj.detachView(v)
		return true
	})

	v.rethinkResultSummary()
	v.rethinkSelectionSummary()

	window.Show()

	v.c.window = window
	/*
		box.Disable()

		pw := ui.NewWindow("Progress", 500, 200, false)
		pw.SetMargined(true)
		prog := ui.NewProgressBar()
		pw.SetChild(prog)

		pw.OnClosing(func(*ui.Window) bool {
			box.Enable()
			return true
		})
		pw.Show()
	*/

	v.Proj.attachView(v)

	return v, err
}

func (v *ProjView) buildToolbar() *ui.Box {
	toolbar := ui.NewHorizontalBox()

	slurpButton := ui.NewButton("Slurp...")
	slurpButton.OnClicked(func(b *ui.Button) { v.SlurpTool() })
	toolbar.Append(slurpButton, false)

	newViewButton := ui.NewButton("New window...")
	newViewButton.OnClicked(func(b *ui.Button) {
		NewProjView(v.Proj)
	})
	toolbar.Append(newViewButton, false)

	return toolbar
}

// run the user through the slurping process
func (v *ProjView) SlurpTool() {
	// TODO: window disable doesn't work
	v.c.window.Disable()
	slurpDialog(
		v.Proj.App.App.SlurpSources,
		func(src steno.SlurpSource, day time.Time, nDays int) {

			progress := NewProgressWindow("Slurping...")
			go func() {
				progFn := func(msg string) {
					fmt.Printf("progress: %s\n", msg)
					ui.QueueMain(func() { progress.SetStatus(msg) })
				}
				dayTo := day.AddDate(0, 0, nDays)
				fmt.Printf("slurp %v,%v to %v,%d\n", src, day, dayTo, nDays)
				newArts, err := steno.Slurp(v.Proj.Store, &src, day, dayTo, progFn)
				if err != nil {
					fmt.Printf("slurp ERROR: %s\n", err)
				}
				ui.QueueMain(func() {
					progress.Close()
					v.c.window.Enable()
					if len(newArts) > 0 {
						v.Proj.ArtsAdded(newArts) // newArts valid even for failed slurp
					}
				})
			}()
		},
		func() {
			fmt.Printf("no slurp\n")
			v.c.window.Enable()
		})
}

func (v *ProjView) rethinkResultSummary() {
	v.c.resultSummary.SetText(fmt.Sprintf("%d matching", v.results.Len))
}

func (v *ProjView) rethinkSelectionSummary() {
	v.c.selSummary.SetText(fmt.Sprintf("%d selected", len(v.selected)))

	tagtxt := v.c.tagEntry.Text()
	if tagtxt != "" && len(v.selected) > 0 {
		v.c.addTagButton.Enable()
		v.c.removeTagButton.Enable()
	} else {
		v.c.addTagButton.Disable()
		v.c.removeTagButton.Disable()
	}

	if len(v.selected) == 1 {
		v.c.showArt.Enable()
	} else {
		v.c.showArt.Disable()
	}

}

func (v *ProjView) SetQuery(q string) {
	res, err := steno.NewResults(v.Proj.Store, q)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %s", err)
		//TODO: show error
		//e := fmt.Sprintf("Search error: %s", err)
		//dbug.Println(e)
		//ctrl.App.SetError(e)
		return
	}

	// cheesy-as-hell
	for i := v.results.Len - 1; i >= 0; i-- {
		v.c.model.RowDeleted(i)
	}

	v.results = res
	for i := 0; i < v.results.Len; i++ {
		v.c.model.RowInserted(i)
	}
	v.rethinkResultSummary()
	fmt.Printf("%d hits\n", res.Len)
}

func (v *ProjView) DoAddTags(artIDs store.ArtList, tags string) {
	tagList := strings.Fields(tags)
	affected, err := v.Proj.Store.AddTags(artIDs, tagList)
	if err != nil {
		dbug.Printf("AddTags(%q): ERROR: %s\n", tagList, err)
	} else {
		dbug.Printf("AddTags(%q): %d affected\n", tagList, len(affected))
	}

	v.Proj.ArtsModified(affected)
}

func (v *ProjView) DoRemoveTags(artIDs store.ArtList, tags string) {
	tagList := strings.Fields(tags)
	affected, err := v.Proj.Store.RemoveTags(artIDs, tagList)
	if err != nil {
		dbug.Printf("RemoveTags(%q): ERROR: %s\n", tagList, err)
	} else {
		dbug.Printf("RemoveTags(%q): %d affected\n", tagList, len(affected))
	}

	v.Proj.ArtsModified(affected)
}