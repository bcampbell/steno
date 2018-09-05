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
	"semprini/steno/steno/store"
	"time"
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

func doOptionsWindow(opts *simrep.Opts, filename string) {
	fenster := ui.NewWindow("steno-similar", 640, 480, true)
	fenster.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})

	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)
	form := ui.NewForm()
	form.SetPadded(true)

	// filename input
	{
		box := ui.NewHorizontalBox()
		fileEntry := ui.NewEntry()
		fileEntry.SetText(filename)
		fileEntry.OnChanged(func(*ui.Entry) {
			filename = fileEntry.Text()
		})

		browseBtn := ui.NewButton("Browse...")
		browseBtn.OnClicked(func(*ui.Button) {
			f := ui.OpenFile(fenster)
			if f != "" {
				filename = f
				fileEntry.SetText(f)
			}
		})
		box.Append(fileEntry, true)
		box.Append(browseBtn, false)
		form.Append(".db to index", box, false)
	}

	// ngram size
	{
		spin := ui.NewSpinbox(1, 100)
		spin.SetValue(opts.NGramSize)
		spin.OnChanged(func(*ui.Spinbox) {
			opts.NGramSize = spin.Value()
		})
		form.Append("ngram size", spin, false)
	}

	// minwords size
	{
		s := ui.NewSpinbox(0, 5000)
		s.SetValue(opts.MinWords)
		s.OnChanged(func(*ui.Spinbox) {
			opts.MinWords = s.Value()
		})
		form.Append("Min Words", s, false)
	}

	// match Threshold %
	{
		s := ui.NewSlider(0, 100)
		s.SetValue(int(100 * opts.MatchThreshold))
		s.OnChanged(func(*ui.Slider) {
			opts.MatchThreshold = float64(s.Value()) / 100.0
		})
		form.Append("Match Threshold %", s, false)
	}

	// language
	{
		langCodes := []string{"en", "ru", "es"}
		langLabels := []string{"English", "Russian", "Spanish"}

		langPicker := ui.NewCombobox()
		sel := -1
		for i, l := range langLabels {
			langPicker.Append(l)
			if opts.Lang == langCodes[i] {
				sel = i
			}
		}
		langPicker.SetSelected(sel)
		langPicker.OnSelected(func(*ui.Combobox) {
			i := langPicker.Selected()
			if i == -1 {
				opts.Lang = ""
			} else {
				opts.Lang = langCodes[i]
			}
		})
		form.Append("Language for indexing", langPicker, false)
	}

	/*
		MinWords       int
		NGramSize      int
		MatchThreshold float64
		IgnoreSameID   bool
		Lang           string
		Dbug store.Logger
	*/

	vbox.Append(form, false)

	indexBtn := ui.NewButton("Index")
	indexBtn.OnClicked(func(*ui.Button) {
		fmt.Printf("<%s>  %v\n", filename, opts)
		//fenster.Destroy()
		//		ui.Quit()
	})
	vbox.Append(indexBtn, false)

	fenster.SetChild(vbox)
	fenster.SetMargined(true)
	fenster.Show()
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
