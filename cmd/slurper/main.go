package main

import (
	"flag"
	"fmt"
	//	"io/ioutil"
	"log"
	"os"
	"semprini/autosteno/slurploader"
	"semprini/scrapeomat/slurp"
	"semprini/steno/steno/store"
	"time"
)

var opts struct {
	verbose   bool
	storePath string
	slurpURL  string
}

func usage() {

	fmt.Fprintf(os.Stderr, `Usage: %s [OPTIONS] SERVER
	Slurp artciles from a slurpserver into a steno db.
Options:
`, os.Args[0])

	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage

	flag.BoolVar(&opts.verbose, "v", false, "verbose output")
	flag.StringVar(&opts.storePath, "s", "store.db", "store db path")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "ERR: missing server\n")
		os.Exit(1)
	}
	srv := flag.Arg(0)

	err := run(srv, "fook.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func run(slurpURL string, storePath string) error {
	l := log.New(os.Stderr, "", 0)
	//	nullLogger := log.New(ioutil.Discard, "", 0)

	// Open the store
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		return err
	}
	st, err := store.New(storePath, l, "en", loc)
	if err != nil {
		return err
	}

	defer st.Close()

	progress := func(receivedCnt, newCnt int) {
		fmt.Printf("%s: received %d (%d new)\n", slurpURL, receivedCnt, newCnt)
	}

	filt := &slurp.Filter{
		Count: 1000,
		//		PubFrom: from,
		//		PubTo:   to,
	}
	_, err = slurploader.SlurpAndLoad(st, slurpURL, filt, progress)
	if err != nil {
		return err
	}
	return nil
}
