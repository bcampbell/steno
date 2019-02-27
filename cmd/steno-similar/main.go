package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	//	"github.com/blevesearch/bleve/analysis"
	"github.com/bcampbell/steno/steno/simrep"
	"github.com/bcampbell/steno/steno/store"
	"log"
	"os"
	"time"
)

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: %s [OPTIONS] DBFILE1 DBFILE2
Options:
`, os.Args[0])

	flag.PrintDefaults()
}

func main() {
	var opts simrep.Opts
	flag.Usage = usage

	flag.BoolVar(&opts.Verbose, "v", false, "verbose output")
	flag.IntVar(&opts.NGramSize, "n", 3, "ngram size")
	flag.IntVar(&opts.MinWords, "s", 100, "ignore articles shorter than this this number of words")
	flag.Float64Var(&opts.MatchThreshold, "m", 0.4, "match threshold (0=no matching, 1=all ngrams matched)")
	flag.StringVar(&opts.Lang, "l", "en", "language rules to use for text tokenising  (en,es,ru)")
	flag.Parse()
	if flag.NArg() < 2 {
		fmt.Fprintf(os.Stderr, "ERR: missing args\n")
		usage()
		os.Exit(1)
	}

	if opts.Verbose {
		opts.Dbug = log.New(os.Stderr, "", 0)
	} else {
		opts.Dbug = log.New(ioutil.Discard, "", 0)
	}

	dbFile1 := flag.Arg(0)
	dbFile2 := flag.Arg(1)
	// TODO: KILL KILL KILL!!!
	opts.IgnoreSameID = true //(dbFile1 == dbFile2)

	err := doit(dbFile1, dbFile2, &opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func doit(dbFile1 string, dbFile2 string, opts *simrep.Opts) error {
	dbug := opts.Dbug
	// Open the stores
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		return err
	}
	db1, err := store.New(dbFile1, dbug, "en", loc)
	if err != nil {
		return err
	}
	defer db1.Close()

	db2, err := store.New(dbFile2, dbug, "en", loc)
	if err != nil {
		return err
	}
	defer db2.Close()

	return simrep.Run(db1, db2, opts)
}
