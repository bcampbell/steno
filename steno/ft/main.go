package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"semprini/steno/steno/store"
	"time"
)

var nullLogger = log.New(ioutil.Discard, "", 0)
var dbug = log.New(os.Stderr, "", 0)

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "poop.\n")
		os.Exit(1)
	}

	db, err := store.New(flag.Arg(0), dbug, time.UTC)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	defer db.Close()

	err = dumpTagged(db, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
}
