package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	//	"github.com/blevesearch/bleve/analysis"
	"github.com/andlabs/ui"
	"log"
	"os"
	//"semprini/sim"
	"semprini/steno/steno/simrep"
	//	"semprini/steno/steno/store"
	//	"time"
)

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: %s [OPTIONS] [DBFILE1]
Options:
`, os.Args[0])

	flag.PrintDefaults()
}

func gogogo() {
	var opts simrep.Opts
	flag.Usage = usage

	flag.BoolVar(&opts.Verbose, "v", false, "verbose output")
	flag.IntVar(&opts.NGramSize, "n", 3, "ngram size")
	flag.IntVar(&opts.MinWords, "s", 100, "ignore articles shorter than this this number of words")
	flag.Float64Var(&opts.MatchThreshold, "m", 0.4, "match threshold (0=no matching, 1=all ngrams matched)")
	flag.StringVar(&opts.Lang, "l", "en", "language rules to use for text tokenising  (en,es,ru)")
	flag.Parse()
	if opts.Verbose {
		opts.Dbug = log.New(os.Stderr, "", 0)
	} else {
		opts.Dbug = log.New(ioutil.Discard, "", 0)
	}

	infile := ""
	if flag.NArg() > 0 {
		infile = flag.Arg(0)
	}

	doOptionsWindow(&opts, infile)

	ui.OnShouldQuit(func() bool {
		fmt.Printf("OnShouldQuit\n")
		//	fenster.Destroy()
		// TODO: what?
		return true
	})
}

func main() {
	ui.Main(gogogo)
}

/*
func buildIndex(dbFile1 string, opts *simrep.Opts) (sim.Index, error) {

	// Open the store
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		return err
	}
	db1, err := store.New(dbFile1, dbug, "en", loc)
	if err != nil {
		return err
	}

	prog := NewProgressWindow("Indexing", "Indexing...")

	dbug := opts.Dbug

	go func() {

		defer db1.Close()

	}()

	return simrep.Run(db1, db2, opts)
}
*/
