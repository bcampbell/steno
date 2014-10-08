package main

import (
	"encoding/csv"
	"fmt"
	"github.com/bcampbell/badger"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// isXHR() returns true if the request was made via an XMLHttpRequest (XHR), ie via AJAX
func isXHR(req *http.Request) bool {
	return strings.ToLower(req.Header.Get("Http_X_Requested_With")) == "xmlhttprequest"
}

func handleSearch(w http.ResponseWriter, req *http.Request) {
	limit := 1000
	performed := false
	var arts []*Article
	var err error
	var total int
	queryString := req.FormValue("q")
	if queryString != "" {
		arts, err = search(queryString)
		performed = true
	}

	// calculate publication/tag counts
	pubCnts := map[string]int{}
	tagCnts := map[string]int{}
	for _, art := range arts {
		pubCnts[art.Pub] += 1
		for _, tag := range art.Tags {
			tagCnts[tag] += 1
		}
	}

	total = len(arts)

	if len(arts) > limit {
		arts = arts[0:limit]
	}
	t := tmpls.MustGet("search")

	params := struct {
		Publications []string
		PubCnts      map[string]int
		TagCnts      map[string]int
		Performed    bool
		Arts         []*Article
		Total        int
		Clipped      bool
		Err          error
		Query        string
	}{
		publications,
		pubCnts,
		tagCnts,
		performed,
		arts,
		total,
		len(arts) != total,
		err,
		queryString,
	}
	t.Execute(w, params)
}

func handleOp(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	queryString := req.FormValue("q")
	op := strings.ToLower(req.FormValue("op"))
	ids := req.Form["id"]
	tagList := req.FormValue("tag")

	tags := []string{}
	if tagList != "" {
		for _, tag := range strings.Split(tagList, ",") {
			tag = strings.ToLower(strings.TrimSpace(tag))
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	// build the query
	// (if articles were individually selected, that has precedence over query string)
	var q badger.Query
	if len(ids) > 0 {
		q, err = buildQueryFromIDs(ids)
	} else {
		q, err = buildQuery(queryString)
	}
	if err != nil {
		http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	switch op {
	case "tag":
		if len(tags) > 0 {
			var changed int
			changed, err = addTags(q, tags)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			dbug.Printf("%d articles tagged\n", changed)
		}
		// redirect back to search with same query
		http.Redirect(w, req, fmt.Sprintf("/?q=%s", url.QueryEscape(queryString)), http.StatusSeeOther)
		return
	case "untag":
		if len(tags) > 0 {
			var changed int
			changed, err = removeTags(q, tags)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			dbug.Printf("removed tag from %d articles\n", changed)
		}
		// redirect back to search with same query
		http.Redirect(w, req, fmt.Sprintf("/?q=%s", url.QueryEscape(queryString)), http.StatusSeeOther)
		return
	case "delete":
		zapped := zap(q)
		dbug.Printf("deleted %d articles.\n", zapped)
		// redirect back to search with same query
		http.Redirect(w, req, fmt.Sprintf("/?q=%s", url.QueryEscape(queryString)), http.StatusSeeOther)
		return
	}

	http.Error(w, "unknown op", http.StatusBadRequest)
	//	fmt.Println(queryString)
	//	fmt.Println(op)
	//	fmt.Println(ids)

}

func handleCSVDownload(w http.ResponseWriter, req *http.Request) {
	var arts []*Article
	var err error
	queryString := req.FormValue("q")
	arts, err = search(queryString)

	if err != nil {
		http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	filename := fileNameFromQuery(queryString) + ".csv"

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	out := csv.NewWriter(w)
	err = out.Write([]string{"id", "headline", "published", "pub", "publication.domain", "urls", "tags"})

	if err != nil {
		http.Error(w, "write error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for _, art := range arts {
		rec := []string{
			"/arts/" + art.ID,
			art.Headline,
			art.Published,
			art.Pub,
			art.Publication.Domain,
			strings.Join(art.URLs, "\n"),
			strings.Join(art.Tags, " "),
		}

		err = out.Write(rec)
		if err != nil {
			http.Error(w, "write error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	out.Flush()
}

func handleArticle(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	objID := vars["id"]

	art := getArtByID(objID)
	if art == nil {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	t := tmpls.MustGet("art")
	params := struct {
		Art *Article
	}{
		art,
	}
	t.Execute(w, params)
}

func handleHelp(w http.ResponseWriter, req *http.Request) {
	t := tmpls.MustGet("help")
	params := struct {
		Publications []string
	}{
		publications,
	}
	t.Execute(w, params)
}

func handleBulkTag(w http.ResponseWriter, req *http.Request) {

	scripts, err := loadScripts(path.Join(baseDir, "bulk"))
	if err != nil {
		http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if req.Method == "POST" {
		scriptName := req.FormValue("script")

		script, got := scripts[scriptName]
		if !got {
			http.Error(w, "error: can't find script "+scriptName, http.StatusInternalServerError)
			return
		}

		tagSet := map[string]struct{}{}

		for _, line := range script.lines {
			q, err := buildQuery(line.query)
			if err != nil {
				http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			changed, _ := addTags(q, line.tags)
			dbug.Printf("%s => %d articles tagged %v\n", line.query, changed, line.tags)
			for _, t := range line.tags {
				tagSet[t] = struct{}{}
			}
		}

		// redirect to query showing those tags
		tags := []string{}
		for tag, _ := range tagSet {
			tags = append(tags, tag)
		}
		queryString := "tags:(" + strings.Join(tags, " OR ") + ")"
		http.Redirect(w, req, fmt.Sprintf("/?q=%s", url.QueryEscape(queryString)), http.StatusSeeOther)
		return
	}

	t := tmpls.MustGet("bulktag")
	params := struct {
		Scripts map[string]*script
	}{
		scripts,
	}
	t.Execute(w, params)
}

func buildRouter(baseDir string) http.Handler {

	r := mux.NewRouter()
	r.HandleFunc("/", handleSearch)
	r.HandleFunc("/op", handleOp)
	r.HandleFunc("/csv", handleCSVDownload)
	r.HandleFunc("/arts/{id}", handleArticle)
	r.HandleFunc("/help", handleHelp)
	r.HandleFunc("/bulktag", handleBulkTag)
	r.HandleFunc("/barchart", handleBarChart)
	r.Handle("/{file}", http.FileServer(http.Dir(path.Join(baseDir, "static"))))
	return r
}
