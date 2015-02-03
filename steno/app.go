package main

import (
	//"fmt"
	"gopkg.in/qml.v1"
	"io/ioutil"
	"path"
	"semprini/steno/steno/kludge"
	//	"strings"
)

type App struct {
	Window        *qml.Window
	HelpText      string
	dataPath      string
	projComponent qml.Object
	ctx           *qml.Context
	project       *Control
	HasCurrent    bool

	Scripts    []*script
	ScriptsLen int

	ErrorMsg string
}

func NewApp() (*App, error) {
	var err error

	//    dataPath := "/Users/ben/semprini/steno/steno.app/Contents/Resources"
	dataPath, err := kludge.DataPath()
	if err != nil {
		return nil, err
	}
	dbug.Printf("Data path: %s\n", dataPath)

	engine := qml.NewEngine()
	ctx := engine.Context()
	app := &App{}
	app.ctx = ctx
	app.dataPath = dataPath
	buf, err := ioutil.ReadFile(path.Join(app.dataPath, "help.html"))
	if err != nil {
		return nil, err
	}
	app.HelpText = string(buf)
	//	ctrl.HelpText = strings.Replace(ctrl.HelpText, "\n", "<br/>\n", -1)

	app.ErrorMsg = "Hello"

	// expose us to the qml side
	ctx.SetVar("app", app)

	component, err := engine.LoadFile(path.Join(app.dataPath, "main.qml"))
	if err != nil {
		return nil, err
	}

	proj, err := engine.LoadFile(path.Join(app.dataPath, "project.qml"))
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

	return app, nil
}

func (app *App) Current() *Control {
	return app.project
}

func (app *App) GetScript(idx int) *script {
	return app.Scripts[idx]
}

func (app *App) SetError(msg string) {
	app.ErrorMsg = msg
	qml.Changed(app, &app.ErrorMsg)
}

func (app *App) OpenProject(storePath string) {
	dbug.Printf("open %s\n", storePath)

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
	dbug.Printf("new %s\n", storePath)

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
	scripts, err := loadScripts(path.Join(app.dataPath, "scripts"))
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
	app.Scripts = scripts
	app.ScriptsLen = len(scripts)
	qml.Changed(app, &app.ScriptsLen)
}

func (app *App) CloseProject() {
	dbug.Printf("close\n")
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
