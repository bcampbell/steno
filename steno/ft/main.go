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
	db, err := store.New("/tmp/eu2.db", dbug, time.UTC)
	if err != nil {
		fmt.Fprintf(os.Stderr, "poop.\n")
		os.Exit(1)
	}

	err = apply(
		db,
		"/home/ben/proj/fastText/fasttext",
		"/tmp/eu1.model.bin",
		0.1)
	if err != nil {
		fmt.Println("ERROR:", err)
	}

}

func OLD2main() {
	err := trainem(
		"/home/ben/proj/fastText/fasttext",
		"../../fasttext/extract/eu1.dump",
		"/tmp/eu1.model",
		func(perc float64) {
			fmt.Printf("Progress: %f\n", perc)
		})
	if err != nil {
		fmt.Println("ERROR:", err)
	}

}
func OLDmain() {
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
