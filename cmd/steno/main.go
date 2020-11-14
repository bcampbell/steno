package main

import "C"

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/bcampbell/steno/script"
	"github.com/bcampbell/steno/steno"
	"github.com/therecipe/qt/widgets"
)

var (
	dbug steno.Logger
)

func main() {
	flag.Parse()

	logFilename := filepath.Join(xdg.DataHome, "steno/log.txt")
	logger, err := steno.NewLog(logFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	//
	dbug = logger
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
