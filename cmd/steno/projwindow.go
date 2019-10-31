package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	//	"strings"
	//	"time"

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
		addTagButton    *widgets.QToolButton
		removeTagButton *widgets.QToolButton
		deleteButton    *widgets.QToolButton

		//
		artView *widgets.QTextBrowser
	}

	// actions
	action struct {
		open       *widgets.QAction
		close      *widgets.QAction
		newProject *widgets.QAction
		newWindow  *widgets.QAction
		slurp      *widgets.QAction
		importJSON *widgets.QAction
		tagArts    *widgets.QAction
		untagArts  *widgets.QAction
		deleteArts *widgets.QAction
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
		v.rethinkActionStates()
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

	m := v.MenuBar().AddMenu2("&File")
	v.action.open = m.AddAction("Open")
	v.action.newProject = m.AddAction("New")
	m.AddSeparator()
	v.action.newWindow = m.AddAction("New Window")
	m.AddSeparator()
	v.action.close = m.AddAction("Close")

	m = v.MenuBar().AddMenu2("&Tools")
	v.action.slurp = m.AddAction("Slurp...")
	v.action.importJSON = m.AddAction("Import JSON...")
	v.action.tagArts = m.AddAction("Tag")
	v.action.untagArts = m.AddAction("Untag")
	v.action.deleteArts = m.AddAction("Delete")

	widget := widgets.NewQWidget(nil, 0)
	splitter := widgets.NewQSplitter2(core.Qt__Vertical, nil)

	vbox := widgets.NewQVBoxLayout()
	widget.SetLayout(vbox)
	splitter.AddWidget(widget)
	v.SetCentralWidget(splitter)

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
		v.c.tagEntry.ConnectTextEdited(func(tags string) {
			v.rethinkActionStates()
		})
		v.c.addTagButton = widgets.NewQToolButton(nil)
		v.c.addTagButton.SetText("Tag")
		v.c.addTagButton.SetToolButtonStyle(core.Qt__ToolButtonTextOnly)
		v.c.addTagButton.SetDefaultAction(v.action.tagArts)

		v.c.removeTagButton = widgets.NewQToolButton(nil)
		v.c.removeTagButton.SetText("Untag")
		v.c.removeTagButton.SetToolButtonStyle(core.Qt__ToolButtonTextOnly)
		v.c.removeTagButton.SetDefaultAction(v.action.untagArts)

		v.c.deleteButton = widgets.NewQToolButton(nil)
		v.c.deleteButton.SetText("Delete")
		v.c.deleteButton.SetToolButtonStyle(core.Qt__ToolButtonTextOnly)
		v.c.deleteButton.SetDefaultAction(v.action.deleteArts)

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
		v.rethinkActionStates()
	})
	tv.SelectionModel().ConnectCurrentChanged(func(current *core.QModelIndex, previous *core.QModelIndex) {
		// show article text for most recently-selected article (if any)
		if current.IsValid() {
			artIdx := v.model.results.Arts[current.Row()]
			arts, err := v.Proj.Store.Fetch(artIdx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Err: %s\n", err)
				return
			} else {
				art := arts[0]
				v.c.artView.SetHtml(art.FormatContent(""))
				//v.c.artView.SetHtml(art.Content)
			}
		}
	})

	widget.Layout().AddWidget(tv)
	v.c.resultView = tv

	// article view
	v.c.artView = widgets.NewQTextBrowser(nil)
	splitter.AddWidget(v.c.artView)

	// set up actions
	{
		v.action.open.ConnectTriggered(func(checked bool) {
			v.doOpenProject()
		})
		v.action.open.SetShortcuts2(gui.QKeySequence__Open)

		v.action.newWindow.ConnectTriggered(func(checked bool) {
			win := NewProjWindow(nil, 0)
			win.SetProject(v.Proj)
		})

		v.action.slurp.ConnectTriggered(func(checked bool) {
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
		v.action.importJSON.ConnectTriggered(func(checked bool) {
			v.doImportJSON()
		})

		v.action.close.ConnectTriggered(func(checked bool) {
			v.Close()
		})

		v.action.newProject.ConnectTriggered(func(checked bool) {
			v.doNewProject()
		})

		v.action.tagArts.ConnectTriggered(func(checked bool) {
			v.Proj.DoAddTags(v.selectedArts(), v.c.tagEntry.Text())
		})
		v.action.untagArts.ConnectTriggered(func(checked bool) {
			v.Proj.DoRemoveTags(v.selectedArts(), v.c.tagEntry.Text())
		})
		v.action.deleteArts.ConnectTriggered(func(checked bool) {
			v.Proj.DoDeleteArts(v.selectedArts())
		})
	}

	v.rethinkWindowTitle()
	v.rethinkResultSummary()
	v.rethinkSelectionSummary()
	v.rethinkActionStates()

	v.Show()
}

// selectedArts returns a list of the currently-selected article IDs.
func (v *ProjWindow) selectedArts() store.ArtList {
	rowIndices := v.c.resultView.SelectionModel().SelectedRows(0)

	sel := make(store.ArtList, len(rowIndices))
	for i, rowIdx := range rowIndices {
		sel[i] = v.model.results.Arts[rowIdx.Row()]
	}
	return sel
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

	v.rethinkWindowTitle()
	v.rethinkResultSummary()
	v.rethinkSelectionSummary()
	v.rethinkActionStates()
}

func (v *ProjWindow) rethinkWindowTitle() {
	title := "Steno"
	if v.Proj != nil {
		title += " - " + filepath.Base(v.Proj.Store.Filename())
	}
	v.SetWindowTitle(title)
}

func (v *ProjWindow) rethinkResultSummary() {
	var txt string
	if v.model.results != nil {
		txt = fmt.Sprintf("%d matching", len(v.model.results.Arts))
	}
	v.c.resultSummary.SetText(txt)
}

func (v *ProjWindow) rethinkSelectionSummary() {
	selection := v.c.resultView.SelectionModel().SelectedRows(0)
	v.c.selSummary.SetText(fmt.Sprintf("%d selected", len(selection)))
}

func (v *ProjWindow) rethinkActionStates() {
	// update the action states
	selection := v.c.resultView.SelectionModel().SelectedRows(0)
	haveProj := !(v.Proj == nil)
	haveSel := len(selection) > 0
	haveTxt := len(v.c.tagEntry.Text()) > 0

	// always available
	//v.action.open.SetEnabled(true)
	//v.action.newProject.SetEnabled(true)
	// v.action.close.SetEnabled(true)

	v.action.newWindow.SetEnabled(haveProj)
	v.action.slurp.SetEnabled(haveProj)
	v.action.importJSON.SetEnabled(haveProj)
	v.action.tagArts.SetEnabled(haveProj && haveSel && haveTxt)
	v.action.untagArts.SetEnabled(haveProj && haveSel && haveTxt)
	v.action.deleteArts.SetEnabled(haveProj && haveSel)
}

func (v *ProjWindow) doOpenProject() {
	fileDialog := widgets.NewQFileDialog2(v, "Open File...", "", "")
	fileDialog.SetAcceptMode(widgets.QFileDialog__AcceptOpen)
	fileDialog.SetFileMode(widgets.QFileDialog__ExistingFile)
	//	var mimeTypes = []string{"text/html", "text/plain"}
	//	fileDialog.SetMimeTypeFilters(mimeTypes)
	if fileDialog.Exec() != int(widgets.QDialog__Accepted) {
		return
	}
	filename := fileDialog.SelectedFiles()[0]
	proj, err := OpenProject(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewProject failed: %s\n", err)
		// TODO: show error!
		return
	}

	if v.Proj == nil {
		// Use existing window
		v.SetProject(proj)
	} else {
		// open a new window and leave this one in peace
		win := NewProjWindow(nil, 0)
		win.SetProject(proj)
	}

}

func (v *ProjWindow) doNewProject() {
	fileDialog := widgets.NewQFileDialog2(v, "Create new project...", "", "")
	fileDialog.SetAcceptMode(widgets.QFileDialog__AcceptSave)
	fileDialog.SetFileMode(widgets.QFileDialog__AnyFile)
	if fileDialog.Exec() != int(widgets.QDialog__Accepted) {
		return
	}
	filename := fileDialog.SelectedFiles()[0]
	proj, err := CreateProject(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "CreateProject failed: %s\n", err)
		// TODO: show error!
		return
	}

	if v.Proj == nil {
		// Use existing window
		v.SetProject(proj)
	} else {
		// open a new window and leave this one in peace
		win := NewProjWindow(nil, 0)
		win.SetProject(proj)
	}
}

func (v *ProjWindow) doImportJSON() {
	fileDialog := widgets.NewQFileDialog2(v, "Import JSON data...", "", "")
	fileDialog.SetAcceptMode(widgets.QFileDialog__AcceptOpen)
	fileDialog.SetFileMode(widgets.QFileDialog__ExistingFile)
	//	var mimeTypes = []string{"text/html", "text/plain"}
	//	fileDialog.SetMimeTypeFilters(mimeTypes)
	if fileDialog.Exec() != int(widgets.QDialog__Accepted) {
		return
	}
	filename := fileDialog.SelectedFiles()[0]

	inFile, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed: %s\n", err)
		// TODO: show error!
		return
	}
	defer inFile.Close()

	dec := json.NewDecoder(inFile)

	stasher := store.NewStasher(v.Proj.Store)
	defer func() {
		stasher.Close()
		v.Proj.ArtsAdded(stasher.StashedIDs)
	}()
	ids := store.ArtList{}
	for dec.More() {
		art := store.Article{}
		err = dec.Decode(&art)
		if err != nil {
			fmt.Fprintf(os.Stderr, "decode failed: %s\n", err)
			// TODO: show error!
			return
		}

		err = stasher.Stash(&art)
		if err != nil {
			fmt.Fprintf(os.Stderr, "decode failed: %s\n", err)
			// TODO: show error!
			return
		}
		// Stash() sets the article ID field
		ids = append(ids, art.ID)
	}

}
