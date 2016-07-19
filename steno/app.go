package main

import (
	"fmt"
	"github.com/limetext/qml-go"
	"github.com/pkg/browser"
	"os"
	"path/filepath"
	"semprini/steno/steno/kludge"
	"sort"
	"time"
)

type App struct {
	Window      *qml.Window
	DataPath    string
	PerUserPath string
	ScriptPath  string
	//Clipboard     *clipboard.Clipboard
	projComponent qml.Object
	ctx           *qml.Context
	project       *Control
	HasCurrent    bool

	scriptCategories    []string
	ScriptCategoriesLen int

	scripts    []*script
	ScriptsLen int

	SlurpSources    []SlurpSource
	SlurpSourcesLen int

	Wibble   []string
	ErrorMsg string
}

// return seconds offset as [+-]HH:MM
func formatZone(offset int) string {
	sign := '+'
	if offset < 0 {
		offset = -offset
		sign = '-'
	}
	return fmt.Sprintf("%c%02d:%02d", sign, offset/3600, (offset%3600)/60)
}

func NewApp() (*App, error) {
	var err error

	// show the current timezone
	zone, offset := time.Now().Zone()
	dbug.Printf("current timezone: %s %s\n", zone, formatZone(offset))

	dataPath, err := kludge.DataPath()
	if err != nil {
		return nil, err
	}
	perUserPath, err := kludge.PerUserPath()
	if err != nil {
		return nil, err
	}

	engine := qml.NewEngine()
	ctx := engine.Context()
	app := &App{}
	app.ctx = ctx
	app.DataPath = dataPath
	app.PerUserPath = perUserPath
	app.ScriptPath = filepath.Join(app.PerUserPath, "scripts")
	//app.Clipboard = clipboard.New(engine)

	// all the qml/js/html stuff is in the ui dir
	uiPath := filepath.Join(app.DataPath, "ui")

	app.ErrorMsg = "Hello"

	// expose us to the qml side
	ctx.SetVar("app", app)

	component, err := engine.LoadFile(filepath.Join(uiPath, "main.qml"))
	if err != nil {
		return nil, err
	}
	proj, err := engine.LoadFile(filepath.Join(uiPath, "project.qml"))
	app.projComponent = proj
	if err != nil {
		return nil, err
	}

	// instantiate the gui
	app.Window = component.CreateWindow(nil)
	app.Window.Show()

	/*
		obj := window.Root().ObjectByName("query")
		obj.Set("text", "")
		fmt.Printf("%v\n", obj)
	*/

	app.RefreshScripts()

	app.initSlurpSources()

	return app, nil
}

func (app *App) Current() *Control {
	return app.project
}

func (app *App) GetScript(idx int) *script {
	return app.scripts[idx]
}

func (app *App) SetError(msg string) {
	app.ErrorMsg = msg
	qml.Changed(app, &app.ErrorMsg)
}

func (app *App) OpenProject(storePath string) {
	//dbug.Printf("open %s\n", storePath)

	proj, err := NewControl(app, storePath, app.projComponent)
	if err != nil {
		dbug.Printf("ERROR: %s\n", err)
		return
	}
	app.project = proj
	app.HasCurrent = true
	qml.Changed(app, &app.HasCurrent)
}

func (app *App) NewProject(storePath string) {
	//dbug.Printf("new %s\n", storePath)

	proj, err := NewControl(app, storePath, app.projComponent)
	if err != nil {
		dbug.Printf("ERROR: %s\n", err)
		return
	}
	app.project = proj
	app.HasCurrent = true
	qml.Changed(app, &app.HasCurrent)
}

func (app *App) RefreshScripts() {

	var scripts []*script
	var err error

	// create scripts dir if it doesn't exist
	// (does nothing if already created)
	err = os.MkdirAll(app.ScriptPath, 0755)
	if err == nil {
		scripts, err = loadScripts(app.ScriptPath)
	}

	if err != nil {
		dbug.Printf("ERROR: %s\n", err)
		app.SetError(err.Error())
		return
	}
	/*
		for _, s := range scripts {
			fmt.Printf("%s - %s\n", s.Name, s.Desc)
			for _, l := range s.lines {
				fmt.Println(l)
			}
		}
	*/

	app.scripts = scripts
	app.scriptCategories = []string{}
	cats := map[string]struct{}{}

	for _, s := range scripts {
		cats[s.Category] = struct{}{}
	}

	for cat, _ := range cats {
		app.scriptCategories = append(app.scriptCategories, cat)
	}
	sort.Strings(app.scriptCategories)

	app.ScriptsLen = len(app.scripts)
	qml.Changed(app, &app.ScriptsLen)

	app.ScriptCategoriesLen = len(app.scriptCategories)
	qml.Changed(app, &app.ScriptCategoriesLen)
}

func (app *App) initSlurpSources() {
	srcs, err := LoadSlurpSources(filepath.Join(app.DataPath, "slurp_sources.csv"))
	if err != nil {
		dbug.Printf("ERROR: %s\n", err)
		app.SetError(err.Error())
		return
	}
	app.SlurpSources = srcs
	app.SlurpSourcesLen = len(srcs)
	qml.Changed(app, &app.SlurpSourcesLen)
}

func (app *App) GetScriptCategory(idx int) string {
	return app.scriptCategories[idx]
}

func (app *App) GetSlurpSourceName(idx int) string {
	return app.SlurpSources[idx].Name
}

func (app *App) CloseProject() {
	//dbug.Printf("close\n")
	if app.project != nil {
		app.project.Close()
		app.project = nil
		app.HasCurrent = false
		qml.Changed(app, &app.HasCurrent)
	}
}
func (app *App) Quit() {
	app.CloseProject()
	app.Window.Hide()
}

// open link in a web browser
func (app *App) BrowseURL(link string) {

	// some hoop-jumping, mainly to get it working on OSX.
	// We want it to run in the GUI thread, but not immediately.
	go func() {
		fmt.Printf("Open URL %s\n", link)
		qml.RunMain(func() {
			browser.OpenURL(link)
		})
	}()
}

// open in a web browser
func (app *App) OpenManual() {
	helpFile := filepath.Join(app.DataPath, "doc", "steno.html")
	go browser.OpenFile(helpFile)
}

// open a directory in a file browser
func (app *App) OpenFileBrowser(dir string) {
	// an abuse of browse.OpenURL(), but should be fine for now...
	app.BrowseURL(dir)
}
