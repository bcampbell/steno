package main

import "C"

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/bcampbell/steno/script"
	"github.com/therecipe/qt/widgets"
)

var (
	dbug *log.Logger
)

func main() {
	flag.Parse()

	logFilename := filepath.Join(xdg.DataHome, "steno/log.txt")
	f, err := os.Create(logFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	defer f.Close()
	dbug = log.New(f, "", 0)

	//
	script.Log = dbug

	err = Run()
	if err != nil {
		dbug.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}
}

func Run() error {
	dbFilename := ""
	if flag.NArg() > 0 {
		dbFilename = flag.Arg(0)
	}

	qtapp := widgets.NewQApplication(len(os.Args), os.Args)

	win := NewProjWindow(nil, 0)
	if dbFilename != "" {
		var proj *Project
		_, err := os.Stat(dbFilename)
		if err == nil {
			proj, err = OpenProject(dbFilename)
		} else if os.IsNotExist(err) {
			proj, err = CreateProject(dbFilename)
		}
		if err != nil {
			return err
		}
		win.SetProject(proj)
	}

	qtapp.Exec()
	return nil
}
