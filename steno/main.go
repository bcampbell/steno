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

var dbug *dbugLog

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

func openURL(url string) {

	dbug.Printf("open %s\n", url)

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
	cmd := exec.Command(params[0], params[1:]...)
	err := cmd.Start()
	if err != nil {

		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}
