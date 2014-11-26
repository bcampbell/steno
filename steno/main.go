package main

// #cgo pkg-config: Qt5Core Qt5Widgets Qt5Quick
import "C"

// previous gubbins needed to force the app to compile with cgo, and to make sure
// the linker gets the right -L path passed in

import (
	//	"encoding/gob"
	"flag"
	"fmt"
	//"github.com/bcampbell/arts/arts"
	"github.com/bcampbell/badger"
	"gopkg.in/qml.v1"
	"os"
	"os/exec"
	//	"path"
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

func run() error {
	flag.Parse()
	dbug = NewDbugLog()
	defer dbug.Close()

	var err error

	coll = badger.NewCollection(&Article{})
	// create database
	/*
		coll, err = loadDB(databaseFile)
		if err != nil {
			dbug.Printf("Error loading db: %s\n", err)
			os.Exit(1)
		}
		coll.EnableAutosave(databaseFile)
	*/

	/*
		dummyArts := []*Article{
			&Article{Article: arts.Article{Headline: "Dummy article 1", CanonicalURL: "http://example.com/art1", Published: "2014-01-01"}, Pub: "dailyfilth"},
		}

		for _, dummyArt := range dummyArts {
			coll.Put(dummyArt)
		}
	*/

	dbug.Printf("fetching list of publications\n")
	publications, err = getPublications()
	if err != nil {
		dbug.Printf("Error finding publications: %s\n", err)
		os.Exit(1)
	}

	// GUI startup
	ctrl, err := NewControl()
	if err != nil {
		dbug.Printf("Error starting GUI: %s\n", err)
		os.Exit(1)
	}
	if flag.NArg() > 0 {
		dbFilename := flag.Arg(0)
		ctrl.LoadDB(dbFilename)
	}
	/*
		for i := 0; i < 1000000; i++ {
			d := &Doc{Headline: fmt.Sprintf("Headline %d", i), Author: fmt.Sprintf("fred%d bloggs", i)}
			docs.list = append(docs.list, d)
		}
		docs.Len = len(docs.list)
		qml.Changed(docs, &docs.Len)
	*/

	ctrl.Window.Wait()
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
