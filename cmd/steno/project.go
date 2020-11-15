package main

// NOTE: this stuff should be gui-agnostic, and could be moved into non-GUI package.

import (
	"strings"
	"time"

	"github.com/bcampbell/steno/store"
)

// Project bundles together a Store and some Views.
type Project struct {
	Store *store.Store
	Views map[View]struct{}
}

// A View watches a project, and updates itself appropriately
// when the data is modified.
type View interface {
	OnArtsModified(store.ArtList)
	OnArtsAdded(store.ArtList)
	OnArtsDeleted(store.ArtList)
	OnRethink()
}

// OpenProject opens a project based on an existing store.
func OpenProject(dbFilename string) (*Project, error) {
	db, err := store.Open(dbFilename, dbug, "en", time.Local)
	if err != nil {
		return nil, err
	}
	proj := &Project{}
	proj.Store = db
	proj.Views = make(map[View]struct{})

	return proj, nil
}

// CreateProject opens a project based on a new store.
func CreateProject(dbFilename string) (*Project, error) {
	db, err := store.Create(dbFilename, dbug, "en", time.Local)
	if err != nil {
		return nil, err
	}
	proj := &Project{}
	proj.Store = db
	proj.Views = make(map[View]struct{})

	return proj, nil
}

func (proj *Project) attachView(v View) {
	proj.Views[v] = struct{}{}
}

func (proj *Project) detachView(v View) {
	delete(proj.Views, v)
}

func (proj *Project) ArtsAdded(ids store.ArtList) {
	for v, _ := range proj.Views {
		v.OnArtsAdded(ids)
	}
}
func (proj *Project) ArtsDeleted(ids store.ArtList) {
	for v, _ := range proj.Views {
		v.OnArtsDeleted(ids)
	}
}
func (proj *Project) ArtsModified(ids store.ArtList) {
	for v, _ := range proj.Views {
		v.OnArtsModified(ids)
	}
}

// Rethink tells all the views to refresh.
func (proj *Project) Rethink() {
	for v, _ := range proj.Views {
		v.OnRethink()
	}
}

func (proj *Project) DoAddTags(artIDs store.ArtList, tags string) {
	// TODO: plug in a progress dialog...
	tagList := strings.Fields(tags)
	affected, err := proj.Store.AddTags(artIDs, tagList)
	if err != nil {
		dbug.Printf("AddTags(%q): ERROR: %s\n", tagList, err)
	} else {
		dbug.Printf("AddTags(%q): %d affected\n", tagList, len(affected))
	}

	proj.ArtsModified(affected)
}

func (proj *Project) DoRemoveTags(artIDs store.ArtList, tags string) {
	// TODO: plug in a progress dialog...
	tagList := strings.Fields(tags)
	affected, err := proj.Store.RemoveTags(artIDs, tagList)
	if err != nil {
		dbug.Printf("RemoveTags(%q): ERROR: %s\n", tagList, err)
	} else {
		dbug.Printf("RemoveTags(%q): %d affected\n", tagList, len(affected))
	}

	proj.ArtsModified(affected)
}

func (proj *Project) DoDeleteArts(artIDs store.ArtList) {
	// TODO: plug in a progress dialog...
	err := proj.Store.Delete(artIDs, nil)
	if err != nil {
		dbug.Printf("Delete: ERROR: %s\n", err)
	}

	proj.ArtsDeleted(artIDs)
}
