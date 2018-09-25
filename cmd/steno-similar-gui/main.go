package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	//	"github.com/blevesearch/bleve/analysis"
	"github.com/andlabs/ui"
	_ "github.com/andlabs/ui/winmanifest"
	"log"
	"os"
	"semprini/steno/steno/simrep"
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

	// hmm. Appears we're now writing javascript...
	doOptionsWindow(&opts, infile, openProject, func() {
		// Cancelled
		fmt.Printf("Cancelled.\n")
		ui.Quit()
	})

	ui.OnShouldQuit(func() bool {
		fmt.Printf("OnShouldQuit\n")
		//	fenster.Destroy()
		// TODO: what?
		return true
	})
}

func openProject(opts *simrep.Opts, filename string) {

	title := fmt.Sprintf("opening %s", filename)
	pw := NewProgressWindow(title, "indexing...")
	go func() {
		defer pw.Close()
		proj, err := NewProject(opts, filename, func(n int, tot int) {
			if tot == 0 {
				pw.SetProgress(-1)
			} else {
				pw.SetProgress((n * 100) / tot)
			}
		})

		if err != nil {
			ui.QueueMain(func() {
				errMsg := fmt.Sprintf("ERROR: %s", err)
				ui.MsgBoxError(pw.w, "steno-similar error", errMsg)
				ui.Quit()
			})
			return
		}

		ui.QueueMain(func() {
			projectWindow(proj)
		})
	}()

}

func main() {
	ui.Main(gogogo)
}
