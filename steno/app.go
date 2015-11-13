package main

import (
	//"fmt"
	"fmt"
	"gopkg.in/qml.v1"
	"gopkg.in/xlab/clipboard.v2"
	"io/ioutil"
	"path/filepath"
	"semprini/steno/steno/kludge"
	"sort"
	"time"
	//	"strings"
)

type App struct {
	Window        *qml.Window
	HelpText      string
	DataPath      string
	Clipboard     *clipboard.Clipboard
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

	engine := qml.NewEngine()
	ctx := engine.Context()
	app := &App{}
	app.ctx = ctx
	app.DataPath = dataPath
	app.Clipboard = clipboard.New(engine)

	// all the qml/js/html stuff is in the ui dir
	uiPath := filepath.Join(app.DataPath, "ui")
	buf, err := ioutil.ReadFile(filepath.Join(uiPath, "help.html"))
	if err != nil {
		return nil, err
	}
	app.HelpText = string(buf)
	//	ctrl.HelpText = strings.Replace(ctrl.HelpText, "\n", "<br/>\n", -1)

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
	scripts, err := loadScripts(filepath.Join(app.DataPath, "scripts"))
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
