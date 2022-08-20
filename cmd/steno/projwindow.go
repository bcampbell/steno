package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"
	"github.com/bcampbell/steno/script"
	"github.com/bcampbell/steno/steno"
	"github.com/bcampbell/steno/store"
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
	// TODO: just keep *Results here. Hide ResultsModel inside resultsview.go
	model *ResultsModel

	// controls we want to keep track of
	c struct {
		query         *widgets.QLineEdit
		resultView    *ResultsView
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
		open          *widgets.QAction
		close         *widgets.QAction
		newProject    *widgets.QAction
		newWindow     *widgets.QAction
		slurp         *widgets.QAction
		runScript     *widgets.QAction
		importJSON    *widgets.QAction
		exportJSON    *widgets.QAction
		exportArts    *widgets.QAction
		tagArts       *widgets.QAction
		untagArts     *widgets.QAction
		deleteArts    *widgets.QAction
		runSimilarity *widgets.QAction
	}
}

// View implementation
func (v *ProjWindow) OnArtsModified(modified store.ArtList) {
	v.model.artsChanged(modified)
}

func (v *ProjWindow) OnArtsAdded(store.ArtList) {
	v.rerun()
}

func (v *ProjWindow) OnArtsDeleted(store.ArtList) {
	v.rerun()
}

func (v *ProjWindow) OnRethink() {
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
		dbug.Printf("bad query: %s\n", err)
	}
}

func (v *ProjWindow) init() {

	//v.ConnectDestroyed(func(obj *core.QObject) {
	// doesn't get here...
	//})
	v.ConnectCloseEvent(func(event *gui.QCloseEvent) {
		v.SetProject(nil)
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
	v.action.runScript = m.AddAction("Run script...")
	v.action.importJSON = m.AddAction("Import JSON...")
	v.action.exportJSON = m.AddAction("Export JSON...")
	v.action.exportArts = m.AddAction("Export CSV...")
	v.action.tagArts = m.AddAction("Tag")
	v.action.untagArts = m.AddAction("Untag")
	v.action.deleteArts = m.AddAction("Delete")
	v.action.runSimilarity = m.AddAction("Run similarity...")

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
			dbug.Printf("new query: %s\n", query.Text())
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

	// set up the results tableview
	tv := v.initResultsView()
	widget.Layout().AddWidget(tv)
	v.c.resultView = tv

	// article view
	v.c.artView = widgets.NewQTextBrowser(nil)
	v.c.artView.SetOpenExternalLinks(true)
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
			sourcesFile, err := xdg.ConfigFile("steno/slurp_sources.csv")
			if err != nil {
				widgets.QMessageBox_Warning(nil, "Error finding slurp_sources.csv", fmt.Sprintf("Error finding slurp sources:\n%s", err.Error()), widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
				return
			}

			srcs, err := steno.LoadSlurpSources(sourcesFile)
			if err != nil {
				widgets.QMessageBox_Warning(nil, "Error", fmt.Sprintf("Error loading slurp sources:\n%s", err.Error()), widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
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
			v.doSlurp(&srcs[sel], from, to)
		})
		v.action.runScript.ConnectTriggered(func(checked bool) {
			v.doRunScript()
		})
		v.action.runSimilarity.ConnectTriggered(func(checked bool) {
			/*
				button := widgets.QMessageBox_Question(nil, "Build similarity data",
					"This can take a few minutes. Are you sure you want to continue?", widgets.QMessageBox__Ok|widgets.QMessageBox__Cancel, widgets.QMessageBox__Ok)
				if button != widgets.QMessageBox__Ok {
					return // cancelled.
				}
			*/
			var ok bool
			label := `Matches below this threshold will be are discarded.

0.0 = no match, 1.0 = contained entirely

NOTE: this might take a few minutes`

			threshold := widgets.QInputDialog_GetDouble(v, "Enter match threshold", label, 0.75, 0.0, 1.0, 2, &ok, 0)
			if !ok {
				return // cancelled
			}

			progressDlg := widgets.NewQProgressDialog(nil, core.Qt__Widget)
			progressDlg.SetModal(true)
			progressDlg.SetMinimumDuration(0)
			progressDlg.SetWindowModality(core.Qt__ApplicationModal)
			progressDlg.SetWindowTitle("Building similarity data...")
			progFn := func(currCnt int, expectedCnt int, msg string) bool {
				progressDlg.SetRange(0, expectedCnt)
				progressDlg.SetValue(currCnt)

				txt := fmt.Sprintf("%s %d/%d", msg, currCnt, expectedCnt)
				progressDlg.SetLabelText(txt)
				return progressDlg.WasCanceled()
			}
			err := BuildSimilarity(v.Proj.Store, progFn, threshold)
			progressDlg.Hide()
			if err != nil {
				fmt.Printf("POOP: %s\n", err)
			} else {
				fmt.Printf("Done.\n")
			}
		})
		v.action.importJSON.ConnectTriggered(func(checked bool) {
			v.doImportJSON()
		})
		v.action.exportJSON.ConnectTriggered(func(checked bool) {
			v.doExportJSON()
		})
		v.action.exportArts.ConnectTriggered(func(checked bool) {
			v.doExportArts()
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

// doSlurp imports articles from a slurp API source.
// It displays a progress dialog.
// It checks urls and doesn't add duplicate articles.
func (v *ProjWindow) doSlurp(src *steno.SlurpSource, dayFrom time.Time, dayTo time.Time) {
	//progress := NewProgressWindow("Slurping...")
	progressDlg := widgets.NewQProgressDialog(nil, core.Qt__Widget)
	progressDlg.SetModal(true)
	progressDlg.SetMinimumDuration(0)
	progressDlg.SetWindowModality(core.Qt__ApplicationModal)
	progressDlg.SetWindowTitle("Slurp from " + src.Name)

	go func() {
		progFn := func(fetchedCnt int, expectedCnt int, newCnt int, msg string) {
			progressDlg.SetRange(0, expectedCnt)
			progressDlg.SetValue(fetchedCnt)

			txt := fmt.Sprintf("%s\nreceived %d/%d articles (%d new)", msg, fetchedCnt, expectedCnt, newCnt)
			progressDlg.SetLabelText(txt)

		}
		dayTo := dayTo.AddDate(0, 0, 1)
		dbug.Printf("slurp %v,%v to %v\n", src, dayFrom, dayTo)
		newArts, err := steno.Slurp(v.Proj.Store, src, dayFrom, dayTo, progFn)
		if err != nil {
			dbug.Printf("slurp ERROR: %s\n", err)
		}
		dbug.Printf("%v %v\n", newArts, err)
		progressDlg.Hide()
		if len(newArts) > 0 {
			v.Proj.ArtsAdded(newArts) // newArts valid even for failed slurp
		}
	}()
}

func (v *ProjWindow) doRunScript() {

	scriptsDir := filepath.Join(xdg.DataHome, "steno/scripts")
	scripts, err := script.LoadScripts(scriptsDir)
	if err != nil {
		widgets.QMessageBox_Warning(nil, "Error finding scripts", fmt.Sprintf("Error finding scripts:\n%s", err.Error()), widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		return
	}

	dlg := NewScriptDialog(nil, 0)
	dlg.SetScripts(scripts)
	if dlg.Exec() != int(widgets.QDialog__Accepted) {
		return
	}

	//	for _, s := range scripts {
	//		fmt.Println(s.Name, s.Category)
	//	}

	idx, got := dlg.SelectedIndex()
	if !got {
		return
	}

	// All set to run!
	script := scripts[idx]
	dbug.Printf("running %d: '%s'\n", idx, script.Name)
	progressDlg := widgets.NewQProgressDialog(nil, core.Qt__Widget)
	progressDlg.SetModal(true)
	progressDlg.SetMinimumDuration(0)
	progressDlg.SetWindowModality(core.Qt__ApplicationModal)
	progressDlg.SetWindowTitle("Running script " + script.Name)

	go func() {
		progFn := func(expected int, completed int, msg string) {
			progressDlg.SetRange(0, expected)
			progressDlg.SetValue(completed)
			progressDlg.SetLabelText(msg)
		}

		err := script.Run(v.Proj.Store, progFn)

		if err != nil {
			dbug.Printf("script ERROR: %s\n", err)
			// TODO: show error in GUI!
		}
		progressDlg.Hide()
		v.Proj.Rethink() // make sure all the views refresh (including this one!)
	}()
}

// set up the tableview for displaying the list of articles.
func (v *ProjWindow) initResultsView() *ResultsView {

	tv := NewResultsView(nil)
	tv.SetResultsModel(v.model)

	// Callbacks for selecting items in results list
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
				rawHTML, _ := HTMLArt(art)
				v.c.artView.SetHtml(rawHTML)
			}
		}
	})

	// Context menu for results list
	tv.SetContextMenuPolicy(core.Qt__CustomContextMenu) // Qt::CustomContextMenu
	tv.ConnectCustomContextMenuRequested(func(pt *core.QPoint) {
		row := tv.RowAt(pt.Y())
		//col := tv.ColumnAt(pt.X())

		var focusArt *store.Article
		if row >= 0 && row < len(v.model.results.Arts) {
			// clicked on an valid article
			focusArt = v.model.results.Art(row)
		}

		if focusArt != nil {
			menu := widgets.NewQMenu(tv)
			action := menu.AddAction("Open " + focusArt.CanonicalURL)
			action.ConnectTriggered(func(checked bool) {
				u := core.NewQUrl3(focusArt.CanonicalURL, core.QUrl__TolerantMode)
				gui.QDesktopServices_OpenUrl(u)
			})
			menu.Popup(tv.MapToGlobal(pt), nil)
		}
	})

	tv.ConnectDoubleClicked(func(index *core.QModelIndex) {
		row := index.Row()
		var art *store.Article
		if row >= 0 && row < len(v.model.results.Arts) {
			art = v.model.results.Art(row)
		}
		if art != nil {
			// double-clicked on an valid article
			w := NewSimWindow(nil, 0)
			w.SetProject(v.Proj)
			w.SetArticle(art)
			w.Show()
		}
	})

	return tv
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

// proj can be nil to detach project from window
func (v *ProjWindow) SetProject(proj *Project) {

	if v.Proj != nil {
		v.model.setResults(nil)
		v.Proj.detachView(v)
	}
	v.Proj = proj
	if v.Proj != nil {
		dbug.Printf("Setting new project...\n")
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
	v.action.exportArts.SetEnabled(haveProj && haveSel)
	v.action.tagArts.SetEnabled(haveProj && haveSel && haveTxt)
	v.action.untagArts.SetEnabled(haveProj && haveSel && haveTxt)
	v.action.deleteArts.SetEnabled(haveProj && haveSel)
	v.action.runSimilarity.SetEnabled(haveProj)
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
		v.reportError("Import JSON", "", err)
		return
	}
	defer inFile.Close()

	imported, err := ImportFromJSON(v.Proj.Store, inFile)
	// Even if there was an error, some might have been imported.
	if len(imported) > 0 {
		v.Proj.ArtsAdded(imported)
	}
	if err != nil {
		v.reportError("Import JSON", "", err)
		return
	}
}

func (v *ProjWindow) doExportJSON() {
	fileDialog := widgets.NewQFileDialog2(v, "Export to JSON...", "", "")
	fileDialog.SetAcceptMode(widgets.QFileDialog__AcceptSave)
	//	fileDialog.SetFileMode(widgets.QFileDialog__ExistingFile)
	//	var mimeTypes = []string{"text/html", "text/plain"}
	//	fileDialog.SetMimeTypeFilters(mimeTypes)
	if fileDialog.Exec() != int(widgets.QDialog__Accepted) {
		return
	}
	filename := fileDialog.SelectedFiles()[0]

	// export selected arts to JSON file.
	outFile, err := os.Create(filename)
	if err != nil {
		v.reportError("Export", "", err)
		return
	}
	defer outFile.Close()

	err = ExportToJSON(v.Proj.Store, v.selectedArts(), outFile)
	if err != nil {
		v.reportError("Export to JSON", "", err)
	}
}

// reportError() shows an error message to the user via a dialog box, and
// logs it to whatever logging setup we've got set up.
func (v *ProjWindow) reportError(title string, message string, err error) {
	if title == "" {
		title = "Error"
	}
	if message == "" {
		message = err.Error()
	}

	widgets.QMessageBox_Warning(nil, title, message, widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
	dbug.Printf("%s\n", message)
}

func (v *ProjWindow) doExportArts() {
	fileDialog := widgets.NewQFileDialog2(v, "Export to CSV...", "", "")
	fileDialog.SetAcceptMode(widgets.QFileDialog__AcceptSave)
	//	fileDialog.SetFileMode(widgets.QFileDialog__ExistingFile)
	//	var mimeTypes = []string{"text/html", "text/plain"}
	//	fileDialog.SetMimeTypeFilters(mimeTypes)
	if fileDialog.Exec() != int(widgets.QDialog__Accepted) {
		return
	}
	filename := fileDialog.SelectedFiles()[0]

	// export selected arts to text file.
	outFile, err := os.Create(filename)
	if err != nil {
		v.reportError("Export", "", err)
		return
	}
	defer outFile.Close()

	err = v.Proj.ExportToCSV(v.selectedArts(), outFile)
	if err != nil {
		v.reportError("Export to CSV", "", err)
	}
}
