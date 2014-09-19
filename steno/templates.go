package main

import (
	"fmt"
	"html/template"
	"os"
	"path"
	"sync"
	"time"
)

type entry struct {
	tmpl    *template.Template
	built   time.Time // when last compiled
	sources []string  // source files
}

func newEntry(tmplDir string, srcFiles []string) (*entry, error) {
	e := entry{
		sources: make([]string, len(srcFiles)),
	}
	for i, srcFile := range srcFiles {
		e.sources[i] = path.Join(tmplDir, srcFile)
	}
	err := e.build()
	if err != nil {
		return nil, err
	}
	return &e, err
}

// returns true if any of the source files have changed since
// template was last compiled
func (e *entry) isStale() (bool, error) {
	for _, srcFile := range e.sources {
		fi, err := os.Stat(srcFile)
		if err != nil {
			return false, err
		}
		if fi.ModTime().After(e.built) {
			return true, nil
		}
	}
	return false, nil
}

func (e *entry) build() error {
	e.built = time.Now()
	t, err := template.ParseFiles(e.sources...)
	if err != nil {
		return err
	}
	e.tmpl = t
	return nil
}

type TemplateMgr struct {
	dir      string
	compiled map[string]*entry
	mut      sync.Mutex
	monitor  bool
}

// NewTemplateMgr compiles a set of templates from their source files
//
// eg:
//    NewTemplateMgr( "templates",
//        map[string][]string{
//          "front": { "base.html", "frontpage.html"},
//          "search": { "base.html", "search.html"},
//          })
//
//
func NewTemplateMgr(templateDir string, sources map[string][]string) (*TemplateMgr, error) {
	out := TemplateMgr{
		dir:      templateDir,
		compiled: map[string]*entry{},
	}

	for name, srcFiles := range sources {
		e, err := newEntry(templateDir, srcFiles)
		if err != nil {
			return nil, err
		}
		out.compiled[name] = e
	}

	return &out, nil
}

// Monitor enables or disables monitoring. When monitoring is enabled,
// the templates will be rebuilt automatically if the source files change.
func (tmpls *TemplateMgr) Monitor(enable bool) { tmpls.monitor = enable }

// Fetch a template
func (tmpls *TemplateMgr) Get(name string) (*template.Template, error) {
	tmpls.mut.Lock()
	defer tmpls.mut.Unlock()
	if entry, ok := tmpls.compiled[name]; ok {
		if tmpls.monitor {
			stale, err := entry.isStale()
			if err != nil {
				return nil, fmt.Errorf("failed checking template '%s': %s", name, err)
			}
			if stale {
				err := entry.build()
				if err != nil {
					return nil, fmt.Errorf("failed to rebuild template '%s': %s", name, err)
				}
			}
		}
		return entry.tmpl, nil
	}
	return nil, fmt.Errorf("unknown template '%s'", name)
}

// Fetch a template, or panic if an error occurs
func (tmpls *TemplateMgr) MustGet(name string) *template.Template {

	t, err := tmpls.Get(name)
	if err != nil {
		panic(err)
	}
	return t
}
