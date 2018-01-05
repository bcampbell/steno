package gui

import (
	"fmt"
	"github.com/bcampbell/ui"
	"os"
	"semprini/steno/steno"
	"semprini/steno/steno/store"
)

type Project struct {
	App   *App
	Store *store.Store
	Views map[*ProjView]struct{}
}

func NewProject(db *store.Store, app *App) (*Project, error) {
	proj := &Project{}
	proj.App = app
	proj.Store = db
	proj.Views = make(map[*ProjView]struct{})

	var err error
	_, err = NewProjView(proj)
	_, err = NewProjView(proj)
	_, err = NewProjView(proj)
	if err != nil {
		return nil, err
	}
	return proj, err
}

func (proj *Project) attachView(v *ProjView) {
	proj.Views[v] = struct{}{}
}

func (proj *Project) detachView(v *ProjView) {
	delete(proj.Views, v)

	if len(proj.Views) == 0 {
		ui.Quit()
	}
}

// view onto an open store
type ProjView struct {
	Proj    *Project
	results *steno.Results

	// controls
	c struct {
		query *ui.Entry
		table *ui.Table
		model *ui.TableModel
	}
}

// TableModelHandler support
// TODO: should be wrapper around Results?

func (v *ProjView) NumColumns(m *ui.TableModel) int {
	return 4
}

func (v *ProjView) ColumnType(m *ui.TableModel, col int) ui.TableModelColumnType {
	return ui.StringColumn
}

func (v *ProjView) NumRows(m *ui.TableModel) int {
	return v.results.Len
}

func (v *ProjView) CellValue(m *ui.TableModel, row int, col int) interface{} {
	art := v.results.Art(row)
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

func (v *ProjView) SetCellValue(m *ui.TableModel, row int, col int, value interface{}) {
}

//
func NewProjView(proj *Project) (*ProjView, error) {

	v := &ProjView{}
	v.Proj = proj
	var err error
	v.results, err = steno.NewResults(v.Proj.Store, "")
	if err != nil {
		return nil, err
	}

	qbox := ui.NewHorizontalBox()
	v.c.query = ui.NewEntry()
	button := ui.NewButton("Search")
	button.OnClicked(func(but *ui.Button) { v.SetQuery(v.c.query.Text()) })
	qbox.Append(v.c.query, true)
	qbox.Append(button, false)

	//
	v.c.model = ui.NewTableModel(v)
	v.c.table = ui.NewTable(v.c.model, ui.TableStyleMultiSelect)
	v.c.table.AppendTextColumn("URL", 0)
	v.c.table.AppendTextColumn("Headline", 1)
	v.c.table.AppendTextColumn("Published", 2)
	v.c.table.AppendTextColumn("Pub", 3)

	box := ui.NewVerticalBox()
	box.Append(qbox, false)
	box.Append(ui.NewLabel("Results"), false)
	box.Append(v.c.table, true)

	window := ui.NewWindow("Steno", 700, 400, true)
	window.SetMargined(true)
	window.SetChild(box)

	window.OnClosing(func(*ui.Window) bool {
		v.Proj.detachView(v)
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

	v.Proj.attachView(v)

	return v, err
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
	fmt.Printf("%d hits\n", res.Len)
}
