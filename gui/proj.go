package gui

import (
	"fmt"
	"github.com/bcampbell/ui"
	"os"
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

type FOO struct{}

func (f *FOO) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

var dbug = &FOO{}

// TableModelHandler support
// TODO: should be wrapper around Results?

func (proj *Proj) NumColumns(m *ui.TableModel) int {
	return 4
}

func (proj *Proj) ColumnType(m *ui.TableModel, col int) ui.TableModelColumnType {
	return ui.StringColumn
}

func (proj *Proj) NumRows(m *ui.TableModel) int {
	return proj.results.Len
}

func (proj *Proj) CellValue(m *ui.TableModel, row int, col int) interface{} {
	art := proj.results.Art(row)
	switch col {
	case 0:
		return art.CanonicalURL
	case 1:
		return art.Headline
	case 2:
		return art.Published
	case 3:
		return art.Pub
	}
	return ""
}

func (proj *Proj) SetCellValue(m *ui.TableModel, row int, col int, value interface{}) {
}

//
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

	qbox := ui.NewHorizontalBox()
	proj.c.query = ui.NewEntry()
	button := ui.NewButton("Search")
	button.OnClicked(func(but *ui.Button) { proj.SetQuery(proj.c.query.Text()) })
	qbox.Append(proj.c.query, true)
	qbox.Append(button, false)

	//
	proj.c.model = ui.NewTableModel(proj)
	proj.c.table = ui.NewTable(proj.c.model, ui.TableStyleMultiSelect)
	proj.c.table.AppendTextColumn("URL", 0)
	proj.c.table.AppendTextColumn("Headline", 1)
	proj.c.table.AppendTextColumn("Published", 2)
	proj.c.table.AppendTextColumn("Pub", 3)

	box := ui.NewVerticalBox()
	box.Append(qbox, false)
	box.Append(ui.NewLabel("Results"), false)
	box.Append(proj.c.table, true)

	window := ui.NewWindow("Steno", 700, 400, true)
	window.SetMargined(true)
	window.SetChild(box)

	window.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	window.Show()
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
	return proj, err
}

func (proj *Proj) SetQuery(q string) {
	res, err := steno.NewResults(proj.store, q)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %s", err)
		//TODO: show error
		//e := fmt.Sprintf("Search error: %s", err)
		//dbug.Println(e)
		//ctrl.App.SetError(e)
		return
	}

	// cheesy-as-hell
	for i := proj.results.Len - 1; i >= 0; i-- {
		proj.c.model.RowDeleted(i)
	}

	proj.results = res
	for i := 0; i < proj.results.Len; i++ {
		proj.c.model.RowInserted(i)
	}
	fmt.Printf("%d hits\n", res.Len)
}
