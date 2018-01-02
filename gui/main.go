package main

import "C"

import (
	"flag"
	"fmt"
	"github.com/bcampbell/ui"
	"os"
	"semprini/steno/steno"
)

type FOO struct{}

func (f *FOO) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

var dbug = &FOO{}

func main() {
	var err error
	flag.Parse()

	app, err := steno.NewApp()
	if err != nil {
		dbug.Printf("Error starting App: %s\n", err)
		os.Exit(1)
	}
	_ = app
	dbFilename := ""
	if flag.NArg() > 0 {
		dbFilename = flag.Arg(0)
	} else {
		dbug.Printf("Missing db file\n")
		os.Exit(1)

	}

	err = ui.Main(func() {
		_, _ = NewProj(dbFilename)
	})
	if err != nil {
		dbug.Printf("Error starting GUI: %s\n", err)
		os.Exit(1)
	}
}

func guiMain(initialDB string) {
}
