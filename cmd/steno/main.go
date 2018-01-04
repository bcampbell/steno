package main

import "C"

import (
	"flag"
	"fmt"
	"os"
	"semprini/steno/gui"
)

type FOO struct{}

func (f *FOO) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

var dbug = &FOO{}

func main() {
	var err error
	flag.Parse()

	app, err := gui.NewApp()
	if err != nil {
		dbug.Printf("Error starting App: %s\n", err)
		os.Exit(1)
	}

	dbFilename := ""
	if flag.NArg() > 0 {
		dbFilename = flag.Arg(0)
	}

	err = app.Run(dbFilename)

	if err != nil {
		dbug.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}
}
