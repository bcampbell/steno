package main

import (
	//	"encoding/gob"
	"flag"
	"fmt"
	"github.com/bcampbell/badger"
	"gopkg.in/qml.v1"
	"os"
	"os/exec"
	"path"
	"runtime"
)

func usage() {
	fmt.Fprintf(os.Stderr, "This tool provides a web-based interface to an article database\n")
	fmt.Fprintf(os.Stderr, "usage:\n")
	flag.PrintDefaults()
}

var coll *badger.Collection

var tmpls *TemplateMgr
var publications []string
var dbug *dbugLog
var baseDir string

func main() {
	if err := qml.Run(run); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

type Control struct {
	arts ArtList
	Len  int
}

func (ctrl *Control) Art(idx int) *Article {
	return ctrl.arts[idx]
}
func (ctrl *Control) Poop(idx int) string {
	return ctrl.arts[idx].Headline
}

func run() error {
	dbug = NewDbugLog()
	defer dbug.Close()

	var databaseFile string
	if flag.NArg() > 0 {
		databaseFile = flag.Arg(0)
	} else {
		databaseFile = path.Join(baseDir, "scotref.db")
	}

	var err error
	coll, err = loadDB(databaseFile)
	if err != nil {
		dbug.Printf("Error loading db: %s\n", err)
		os.Exit(1)
	}
	coll.EnableAutosave(databaseFile)

	// create database

	dbug.Printf("fetching list of publications\n")
	publications, err = getPublications()
	if err != nil {
		dbug.Printf("Error finding publications: %s\n", err)
		os.Exit(1)
	}

	// GUI startup

	ctrl := &Control{}

	engine := qml.NewEngine()
	ctx := engine.Context()
	ctx.SetVar("ctrl", ctrl)

	component, err := engine.LoadFile("fook.qml")
	if err != nil {
		return err
	}
	/*
		for i := 0; i < 1000000; i++ {
			d := &Doc{Headline: fmt.Sprintf("Headline %d", i), Author: fmt.Sprintf("fred%d bloggs", i)}
			docs.list = append(docs.list, d)
		}
		docs.Len = len(docs.list)
		qml.Changed(docs, &docs.Len)
	*/

	// set the query
	ctrl.arts, err = allArts()
	if err != nil {
		dbug.Printf("Error in allArts(): %s\n", err)
		os.Exit(1)
	}
	ctrl.Len = len(ctrl.arts)
	fmt.Printf("%d\n", ctrl.Len)
	qml.Changed(ctrl, &ctrl.Len)

	window := component.CreateWindow(nil)
	window.Show()
	window.Wait()
	return nil
}

func loadDB(fileName string) (*badger.Collection, error) {

	dbug.Printf("Loading DB from %s\n", fileName)
	infile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer infile.Close()
	db, err := badger.Read(infile, &Article{})
	if err != nil {
		return nil, err
	}
	dbug.Printf("Loaded %d articles\n", db.Count())
	return db, nil
}

func openURL(url string) {

	dbug.Printf("Launching web browser...\n")

	var params []string
	switch runtime.GOOS {
	case "windows":
		params = []string{"cmd", "/c", "start"}
	case "darwin":
		params = []string{"open"}
	default:
		params = []string{"xdg-open"}
	}
	params = append(params, url)
	cmd := exec.Command(params[0], params...)
	cmd.Start()
}
