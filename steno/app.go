package steno

import (
	"fmt"
	"github.com/pkg/browser"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"
)

// TODO: split all this guff up into separate parts
type App struct {
	DataPath    string
	BinPath     string
	PerUserPath string
	ScriptPath  string
	HasCurrent  bool

	scriptCategories []string
	scripts          []*script
	SlurpSources     []SlurpSource

	Wibble      []string
	ErrorMsg    string
	HoveredLink string
}

// TODO: where to?
type Progress struct {
	InFlight     bool
	Title        string
	ExpectedCnt  int // 0=unknown
	CompletedCnt int
	StatusMsg    string
	ErrorMsg     string
}

func (p *Progress) SetError(err error) {
	p.ErrorMsg = err.Error()
	//	qml.Changed(p.ctrl, p)
}

func (p *Progress) SetStatus(msg string) {
	p.StatusMsg = msg
	//	qml.Changed(p.ctrl, p)
}

func (p *Progress) Reset() {
	p.InFlight = false
	p.Title = ""
	p.ExpectedCnt = 0
	p.CompletedCnt = 0
	p.StatusMsg = ""
	p.ErrorMsg = ""
}

var dbug = dbugLog{log: os.Stdout}

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
	binPath, err := kludge.BinPath()
	if err != nil {
		return nil, err
	}

	app := &App{}

	app.DataPath = dataPath
	app.BinPath = binPath
	app.PerUserPath = perUserPath
	app.ScriptPath = filepath.Join(app.PerUserPath, "scripts")

	app.ErrorMsg = "Hello"
	app.RefreshScripts()
	app.initSlurpSources()
	return app, nil
}

func (app *App) GetScript(idx int) *script {
	return app.scripts[idx]
}

func (app *App) SetError(msg string) {
	app.ErrorMsg = msg
	//	qml.Changed(app, &app.ErrorMsg)
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
}

func (app *App) initSlurpSources() {
	srcs, err := LoadSlurpSources(filepath.Join(app.DataPath, "slurp_sources.csv"))
	if err != nil {
		dbug.Printf("ERROR: %s\n", err)
		app.SetError(err.Error())
		return
	}
	app.SlurpSources = srcs
}

func (app *App) GetScriptCategory(idx int) string {
	return app.scriptCategories[idx]
}

func (app *App) GetSlurpSourceName(idx int) string {
	return app.SlurpSources[idx].Name
}

// open link in a web browser
func (app *App) BrowseURL(link string) {

	// some hoop-jumping, mainly to get it working on OSX.
	// We want it to run in the GUI thread, but not immediately.
	/*
		go func() {
			fmt.Printf("Open URL %s\n", link)
			qml.RunMain(func() {
				browser.OpenURL(link)
			})
		}()
	*/
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

func (app *App) GetFasttextExe() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(app.BinPath, "fasttext.exe")
	} else {
		return filepath.Join(app.BinPath, "fasttext")
	}
}
