package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bcampbell/steno/steno"
	"github.com/bcampbell/steno/steno/store"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

type Project struct {
	//App   *App
	Store *store.Store
	Views map[View]struct{}
}

type View interface {
	OnArtsModified(store.ArtList)
	OnArtsAdded(store.ArtList)
	OnArtsDeleted(store.ArtList)
}

func NewProject(db *store.Store) (*Project, error) {
	proj := &Project{}
	//proj.App = app
	proj.Store = db
	proj.Views = make(map[View]struct{})

	return proj, nil
}

func (proj *Project) attachView(v View) {
	proj.Views[v] = struct{}{}
}

func (proj *Project) detachView(v View) {
	delete(proj.Views, v)

	if len(proj.Views) == 0 {
		// XYZZY QUIT!
		//		ui.Quit()
	}
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

// doSlurp imports articles from a slurp API source.
// It checks urls and doesn't add duplicate articles.
func (proj *Project) doSlurp(src *steno.SlurpSource, dayFrom time.Time, dayTo time.Time) {
	//progress := NewProgressWindow("Slurping...")
	progressDlg := widgets.NewQProgressDialog(nil, core.Qt__Widget)
	progressDlg.SetModal(true)
	progressDlg.SetMinimumDuration(0)
	progressDlg.SetWindowModality(core.Qt__ApplicationModal)
	progressDlg.SetWindowTitle("Slurp from " + src.Name)

	go func() {
		progFn := func(fetchedCnt int, expectedCnt int, newCnt int, msg string) {
			progressDlg.SetRange(0, expectedCnt)
			progressDlg.SetValue(fetchedCnt)

			txt := fmt.Sprintf("%s\nreceived %d/%d articles (%d new)", msg, fetchedCnt, expectedCnt, newCnt)
			progressDlg.SetLabelText(txt)

		}
		dayTo := dayTo.AddDate(0, 0, 1)
		fmt.Printf("slurp %v,%v to %v\n", src, dayFrom, dayTo)
		newArts, err := steno.Slurp(proj.Store, src, dayFrom, dayTo, progFn)
		if err != nil {
			fmt.Printf("slurp ERROR: %s\n", err)
		}
		fmt.Printf("%v %v\n", newArts, err)
		progressDlg.Hide()
		if len(newArts) > 0 {
			proj.ArtsAdded(newArts) // newArts valid even for failed slurp
		}
	}()
}

func (proj *Project) DoAddTags(artIDs store.ArtList, tags string) {
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
	tagList := strings.Fields(tags)
	affected, err := proj.Store.RemoveTags(artIDs, tagList)
	if err != nil {
		dbug.Printf("RemoveTags(%q): ERROR: %s\n", tagList, err)
	} else {
		dbug.Printf("RemoveTags(%q): %d affected\n", tagList, len(affected))
	}

	proj.ArtsModified(affected)
}
