package main

import (
	"fmt"
	"os"
	//	"strings"
	"time"

	"github.com/bcampbell/steno/steno"
	"github.com/bcampbell/steno/steno/store"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

// view onto an open store
type ProjWindow struct {
	widgets.QMainWindow

	_ func() `constructor:"init"`

	// The project we're attached to (nil = none)
	Proj *Project

	// currently-shown query results stored here.
	// (each window has it's own results)
	model *ResultsModel

	// controls we want to keep track of
	c struct {
		query         *widgets.QLineEdit
		resultView    *widgets.QTableView
		resultSummary *widgets.QLabel
		//
		selSummary *widgets.QLabel

		//
		tagEntry        *widgets.QLineEdit
		addTagButton    *widgets.QPushButton
		removeTagButton *widgets.QPushButton
		deleteButton    *widgets.QPushButton
	}

	// actions
	actions struct {
		open       *widgets.QAction
		close      *widgets.QAction
		newProject *widgets.QAction
		newWindow  *widgets.QAction
		slurp      *widgets.QAction
	}
}

// View implementation
func (v *ProjWindow) OnArtsModified(store.ArtList) {
	v.rerun()
}

func (v *ProjWindow) OnArtsAdded(store.ArtList) {
	v.rerun()
}

func (v *ProjWindow) OnArtsDeleted(store.ArtList) {
	v.rerun()
}

func (v *ProjWindow) rerun() {
	results, err := steno.NewResults(v.Proj.Store, v.c.query.Text())
	if err == nil {
		v.model.setResults(results)
		v.rethinkResultSummary()
		v.rethinkSelectionSummary()
	} else {
		// TODO XYZZY - show bad query message
		fmt.Printf("Error: %s\n", err)
	}
}

func (v *ProjWindow) init() {

	//v.ConnectDestroyed(func(obj *core.QObject) {
	// doesn't get here...
	//})
	v.ConnectCloseEvent(func(event *gui.QCloseEvent) {
		v.SetProject(nil)
		fmt.Printf("byebye!\n")
		event.Accept()
	})

	v.model = NewResultsModel(nil)

	v.SetMinimumSize2(640, 400)
	v.SetWindowTitle("Steno")

	m := v.MenuBar().AddMenu2("&File")
	v.actions.open = m.AddAction("Open")
	v.actions.newProject = m.AddAction("New")
	m.AddSeparator()
	v.actions.newWindow = m.AddAction("New Window")
	m.AddSeparator()
	v.actions.close = m.AddAction("Close")

	m = v.MenuBar().AddMenu2("&Tools")
	v.actions.slurp = m.AddAction("Slurp...")

	widget := widgets.NewQWidget(nil, 0)
	vbox := widgets.NewQVBoxLayout()
	widget.SetLayout(vbox)
	v.SetCentralWidget(widget)

	{

		// query entry:
		query := widgets.NewQLineEdit(nil)
		query.SetClearButtonEnabled(true)
		//	query.addAction(":/resources/search.ico", QLineEdit::LeadingPosition);
		query.SetPlaceholderText("Search...")
		v.c.query = query
		query.ConnectEditingFinished(func() {
			if v.Proj == nil {
				// TODO: complain!
				return
			}
			fmt.Printf("new query: %s\n", query.Text())
			v.rerun()
		})

		resultSummary := widgets.NewQLabel(nil, 0)
		resultSummary.SetText("")
		v.c.resultSummary = resultSummary

		group1 := widgets.NewQHBoxLayout()
		group1.AddWidget(query, 3, 0)
		group1.AddWidget(resultSummary, 1, 0)
		group1.SetContentsMargins(0, 0, 0, 0)

		vbox.AddLayout(group1, 0)
	}

	// selection info and operations (tag, delete etc)
	{
		v.c.selSummary = widgets.NewQLabel(nil, 0)
		v.c.selSummary.SetText("")

		//
		v.c.tagEntry = widgets.NewQLineEdit(nil)
		v.c.tagEntry.SetPlaceholderText("enter tag")
		v.c.addTagButton = widgets.NewQPushButton2("tag", nil)
		v.c.removeTagButton = widgets.NewQPushButton2("untag", nil)
		v.c.deleteButton = widgets.NewQPushButton2("delete", nil)

		group1 := widgets.NewQHBoxLayout()
		group1.AddWidget(v.c.selSummary, 0, 0)
		group1.AddWidget(v.c.tagEntry, 0, 0)
		group1.AddWidget(v.c.addTagButton, 0, 0)
		group1.AddWidget(v.c.removeTagButton, 0, 0)
		group1.AddWidget(v.c.deleteButton, 0, 0)
		group1.AddStretch(0)
		group1.SetContentsMargins(0, 0, 0, 0)
		vbox.AddLayout(group1, 0)

	}

	tv := widgets.NewQTableView(nil)
	tv.SetShowGrid(false)
	tv.SetSelectionBehavior(widgets.QAbstractItemView__SelectRows)
	tv.SetSelectionMode(widgets.QAbstractItemView__ExtendedSelection)
	tv.VerticalHeader().SetVisible(false)
	//tv.HorizontalHeader().SetSectionResizeMode(widgets.QHeaderView__Stretch)

	tv.SetModel(v.model)
	tv.ResizeColumnsToContents()

	// cheesy autosize.
	{
		w := tv.Width()
		if w < 600 {
			w = 600
		}
		hdr := tv.HorizontalHeader()
		hdr.ResizeSection(0, w/3) // url
		hdr.ResizeSection(1, w/3) // headline
		hdr.ResizeSection(2, w/6) // published
		hdr.ResizeSection(3, w/6) // pub
		hdr.ResizeSection(4, w/6) // tags
	}
	tv.SelectionModel().ConnectSelectionChanged(func(selected *core.QItemSelection, deselected *core.QItemSelection) {
		v.rethinkSelectionSummary()
	})

	widget.Layout().AddWidget(tv)
	v.c.resultView = tv

	// set up actions
	{
		v.actions.open.ConnectTriggered(func(checked bool) {
			fileDialog := widgets.NewQFileDialog2(v, "Open File...", "", "")
			fileDialog.SetAcceptMode(widgets.QFileDialog__AcceptOpen)
			fileDialog.SetFileMode(widgets.QFileDialog__ExistingFile)
			//	var mimeTypes = []string{"text/html", "text/plain"}
			//	fileDialog.SetMimeTypeFilters(mimeTypes)
			if fileDialog.Exec() != int(widgets.QDialog__Accepted) {
				return
			}
			filename := fileDialog.SelectedFiles()[0]
			// TODO: move into project
			db, err := store.New(filename, dbug, "en", time.Local)
			if err != nil {
				fmt.Fprintf(os.Stderr, "store.New failed: %s\n", err)
				// TODO: show error!
				return
			}
			proj, err := NewProject(db)
			if err != nil {
				fmt.Fprintf(os.Stderr, "NewProject failed: %s\n", err)
				// TODO: show error!
				return
			}
			v.SetProject(proj)
		})
		v.actions.open.SetShortcuts2(gui.QKeySequence__Open)

		v.actions.newWindow.ConnectTriggered(func(checked bool) {
			win := NewProjWindow(nil, 0)
			win.SetProject(v.Proj)
		})

		v.actions.slurp.ConnectTriggered(func(checked bool) {
			srcs, err := steno.LoadSlurpSources("slurp_sources.csv")
			if err != nil {
				// TODO: show error!!!
				return
			}
			dlg := NewSlurpDialog(nil, 0)
			dlg.SetSources(srcs)
			if dlg.Exec() != int(widgets.QDialog__Accepted) {
				return
			}
			from, to := dlg.DateRange()
			sel := dlg.SourceIndex()
			if sel < 0 || sel >= len(srcs) {
				// TODO: show error?
			}
			v.Proj.doSlurp(&srcs[sel], from, to)
		})

		v.actions.close.ConnectTriggered(func(checked bool) {
			v.Close()
		})
	}

	v.rethinkResultSummary()
	v.rethinkSelectionSummary()

	v.Show()
}

// proj can be nil
func (v *ProjWindow) SetProject(proj *Project) {

	if v.Proj != nil {
		v.model.setResults(nil)
		v.Proj.detachView(v)
	}
	v.Proj = proj
	if v.Proj != nil {
		fmt.Printf("Setting new project...\n")
		v.Proj.attachView(v)

		// Show default query
		results, err := steno.NewResults(proj.Store, "")
		if err != nil {
			// TODO: report error!
			return
		}
		v.model.setResults(results)
	}

	v.rethinkResultSummary()
	v.rethinkSelectionSummary()
}

func (v *ProjWindow) rethinkResultSummary() {
	var txt string
	if v.model.results != nil {
		txt = fmt.Sprintf("%d matching", len(v.model.results.Arts))
	}
	v.c.resultSummary.SetText(txt)
}

/*
func (v *ProjWindow) buildToolbar() *ui.Box {
	toolbar := ui.NewHorizontalBox()

	slurpButton := ui.NewButton("Slurp...")
	slurpButton.OnClicked(func(b *ui.Button) { v.SlurpTool() })
	toolbar.Append(slurpButton, false)

	newViewButton := ui.NewButton("New window...")
	newViewButton.OnClicked(func(b *ui.Button) {
		NewProjWindow(v.Proj)
	})
	toolbar.Append(newViewButton, false)

	return toolbar
}

*/

func (v *ProjWindow) rethinkSelectionSummary() {
	sel := v.c.resultView.SelectionModel().SelectedRows(0)
	v.c.selSummary.SetText(fmt.Sprintf("%d selected", len(sel)))
	tagtxt := v.c.tagEntry.Text()
	if tagtxt != "" && len(sel) > 0 {
		v.c.addTagButton.SetEnabled(true)
		v.c.removeTagButton.SetEnabled(true)
	} else {
		v.c.addTagButton.SetEnabled(false)
		v.c.removeTagButton.SetEnabled(false)
	}

}
