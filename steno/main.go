package main

// #cgo pkg-config: Qt5Core Qt5Widgets Qt5Quick
import "C"

// previous gubbins needed to force the app to compile with cgo, and to make sure
// the linker gets the right -L path passed in

import (
	"flag"
	"fmt"
	"github.com/limetext/qml-go"
	"os"
	"path/filepath"
	"semprini/steno/steno/kludge"
)

func usage() {
	fmt.Fprintf(os.Stderr, "This tool provides a web-based interface to an article database\n")
	fmt.Fprintf(os.Stderr, "usage:\n")
	flag.PrintDefaults()
}

var dbug *dbugLog

func main() {
	if err := qml.Run(run); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var err error
	flag.Parse()
	perUserDir, err := kludge.PerUserPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	dbug = NewDbugLog(filepath.Join(perUserDir, "log.txt"))
	defer dbug.Close()

	// GUI startup
	app, err := NewApp()
	if err != nil {
		dbug.Printf("Error starting App: %s\n", err)
		os.Exit(1)
	}

	if flag.NArg() > 0 {
		dbFilename := flag.Arg(0)
		app.OpenProject(dbFilename)
	}

	app.Window.Wait()
	return nil
}
