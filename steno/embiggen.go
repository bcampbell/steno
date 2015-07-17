package main

import (
	//	"bufio"
	"fmt"
	"github.com/PuerkitoBio/purell"
	"net"
	"net/http"
	//	"net/url"
	//	"os"
	"semprini/steno/steno/store"
	//	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func isShortlink(link string) bool {
	/*u, err := url.Parse(link)
	if err != nil {
		return false
	}
	*/
	if len(link) < 30 {
		return true
	}
	return false
}

// consider a link to be a shortlink if:
// a) domain is on our list
// b) entire URL < 30 chars
func embiggenArts(arts store.ArtList, shortLinkDomainsFile string, progress ProgressFunc) (store.ArtList, error) {

	numWorkers := 32
	/*
		// read in list of shortlink domains
		f, err := os.Open(shortLinkDomainsFile)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)

		domains := map[string]struct{}{}

		for scanner.Scan() {
			foo := strings.TrimSpace(scanner.Text())
			if foo == "" || foo[0] == '#' {
				continue
			}
			domains[foo] = struct{}{}
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}
	*/
	// find articles with shortlinks
	totalShortLinks := 0

	affected := store.ArtList{}
	/* XYZZY */
	/*
		for _, art := range arts {
			nShortLinks := 0
			for _, link := range art.Links {
				if isShortlink(link) {
					nShortLinks++
				}
			}

			if nShortLinks > 0 {
				totalShortLinks += nShortLinks
				affected = append(affected, art)
			}
		}
	*/

	dbug.Printf("%s shortlinks to resolve\n", totalShortLinks)

	// create work queue and workers
	wg := new(sync.WaitGroup)
	queue := make(chan *store.Article)
	var numLinksResolved int64 = 0
	var numErrors int64 = 0
	transport := http.Transport{
		Dial: dialTimeout,
	}
	client := &http.Client{Transport: &transport}
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for art := range queue {
				newLinks := make([]string, len(art.Links))
				for i, link := range art.Links {
					newLinks[i] = link
					if isShortlink(link) {
						newLink, err := embiggenLink(client, link)
						if err != nil {
							atomic.AddInt64(&numErrors, 1)
							dbug.Printf("FAIL (%s) %s \n", link, err)
						} else {
							newLinks[i] = newLink
							atomic.AddInt64(&numLinksResolved, 1)
							//dbug.Printf("%s => %s\n", link, newLink)
						}
					}
				}
				art.Links = newLinks
			}
		}()
	}

	// do it!
	/* XYZZY */
	/*
		for _, art := range affected {
			queue <- art
			if progress != nil {
				nOK := atomic.LoadInt64(&numLinksResolved)
				nErr := atomic.LoadInt64(&numErrors)
				progress(int(totalShortLinks), int(nOK+nErr), fmt.Sprintf("Embiggening shortlinks: %d/%d (%d ok, %d failed)", nOK+nErr, totalShortLinks, nOK, nErr))
			}
		}
	*/
	close(queue)
	wg.Wait()
	return affected, nil
}

// can use the http.Client.Timeout field instead for go1.3+
func dialTimeout(network, addr string) (net.Conn, error) {

	return net.DialTimeout(network, addr, 5*time.Second)
}

func embiggenLink(client *http.Client, u string) (string, error) {

	u, err := purell.NormalizeURLString(u, purell.FlagsSafe)
	if err != nil {
		return "", err
	}

	resp, err := client.Head(u)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	code := resp.StatusCode

	// some non-2xx error codes are OK, eg 401 or 403 might indicate we've
	// hit a paywall
	if code >= 200 && code < 300 /* ) || code == 401 || code == 403*/ {
		finalURL := resp.Request.URL.String()
		return finalURL, nil
	} else {
		return "", fmt.Errorf("HTTP code %d", resp.StatusCode)
	}

}
