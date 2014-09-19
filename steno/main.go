package main

import (
	//	"encoding/gob"
	"flag"
	"fmt"
	"github.com/bcampbell/badger"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"sync"
	"time"
)

func usage() {
	fmt.Fprintf(os.Stderr, "This tool provides a web-based interface to an article database\n")
	fmt.Fprintf(os.Stderr, "usage:\n")
	flag.PrintDefaults()
}

var coll *badger.Collection

var tmpls *TemplateMgr
var publications []string

func main() {
	//	gob.Register(Article{})

	flag.Usage = usage
	var databaseFile = flag.String("database", "eurobot.db", "database filename")
	var port = flag.Int("port", 8080, "port to run on")
	var launchBrowser = flag.Bool("launch", true, "launch web browser")
	flag.Parse()

	baseDir := "."
	if os.Getenv("EUROBOT") != "" {
		baseDir = path.Join(os.Getenv("EUROBOT"), "browse")
	}

	if *databaseFile == "" {
		fmt.Fprintf(os.Stderr, "missing: -database <filename>\n")
		os.Exit(1)
	}

	var err error
	coll, err = loadDB(*databaseFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading db: %s\n", err)
		os.Exit(1)
	}
	coll.EnableAutosave(*databaseFile)

	templateSources := map[string][]string{
		"search":   {"base.html", "search.html"},
		"art":      {"base.html", "art.html"},
		"help":     {"base.html", "help.html"},
		"bulktag":  {"base.html", "bulktag.html"},
		"barchart": {"base.html", "barchart.html"},
	}

	tmpls, err = NewTemplateMgr(path.Join(baseDir, "templates"), templateSources)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	tmpls.Monitor(true)

	// create database

	fmt.Println("fetching list of publications")
	publications, err = getPublications()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding publications: %s\n", err)
		os.Exit(1)
	}

	r := buildRouter(baseDir)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = http.ListenAndServe(fmt.Sprintf(":%d", *port), r)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}()

	fmt.Printf("running at http://localhost:%d\n", *port)
	if *launchBrowser {
		time.Sleep(100 * time.Millisecond)
		serverURL := fmt.Sprintf("http://localhost:%d", *port)
		launch(serverURL)
	}

	wg.Wait()
}

func loadDB(fileName string) (*badger.Collection, error) {

	fmt.Printf("Loading DB from %s\n", fileName)
	infile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer infile.Close()
	db, err := badger.Read(infile, &Article{})
	if err != nil {
		return nil, err
	}
	fmt.Printf("Loaded %d articles\n", db.Count())
	return db, nil
}

func launch(url string) {

	fmt.Printf("Launch %s\n", url)

	var params []string
	switch runtime.GOOS {
	case "windows":
		params = []string{"cmd", "/c", "start"}
	case "darwin":
		params = []string{"open"}
	default:
		params = []string{"xdg-open"}
	}
	params = append(params, url)
	fmt.Println(params)
	cmd := exec.Command(params[0], params...)
	cmd.Start()
}
