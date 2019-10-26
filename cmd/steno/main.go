package main

import "C"

import (
	"flag"
	"fmt"
	"os"
	"time"

	//	"github.com/bcampbell/steno/steno"
	"github.com/bcampbell/steno/steno/store"
	"github.com/therecipe/qt/widgets"
)

type FOO struct{}

func (f *FOO) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

var dbug = &FOO{}

func main() {
	flag.Parse()
	err := Run()
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
		db, err := store.New(dbFilename, dbug, "en", time.Local)
		if err != nil {
			return err
		}
		proj, err := NewProject(db)
		if err != nil {
			return err
		}
		win.SetProject(proj)
	}

	qtapp.Exec()
	return nil
}
