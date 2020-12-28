package main

import (
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

type ResultsView struct {
	widgets.QTableView

	_ func() `constructor:"init"`

	model *ResultsModel
}

// TODO: use this to drive ResultsModel too
var resultColumns = []struct {
	name   string
	weight int
}{
	{"url", 10},
	{"headline", 10},
	{"published", 3},
	{"pub", 3},
	{"tags", 3},
	{"similar", 1},
}

func (tv *ResultsView) init() {

	// override member functions
	tv.ConnectSizeHintForColumn(tv.sizeHintForColumn)
	tv.ConnectResizeEvent(tv.resizeEvent)

	tv.SetShowGrid(false)
	tv.SetSelectionBehavior(widgets.QAbstractItemView__SelectRows)
	tv.SetSelectionMode(widgets.QAbstractItemView__ExtendedSelection)
	tv.VerticalHeader().SetVisible(false)

	tv.ResizeColumnsToContents()

	{
		hdr := tv.HorizontalHeader()
		hdr.SetCascadingSectionResizes(true)
		hdr.SetStretchLastSection(true)
		// set up sorting (by clicking on column headers)
		hdr.SetSortIndicatorShown(true)
		hdr.ConnectSortIndicatorChanged(func(logicalIndex int, order core.Qt__SortOrder) {
			if tv.model.results == nil {
				return
			}
			if logicalIndex < 0 || logicalIndex > len(resultColumns) {
				return
			}
			field := resultColumns[logicalIndex].name

			var dir int
			if order == core.Qt__DescendingOrder {
				dir = -1
			} else {
				dir = 1
			}

			newResults := tv.model.results.Sort(field, dir)
			tv.model.setResults(newResults)

			// TODO: need to sort the original query using the current gui setting...

			//			fmt.Printf("Bing. %d\n", logicalIndex)
		})
	}
}

// sizeHintForColumn provides hints used by ResizeColumnsToContents()
func (tv *ResultsView) sizeHintForColumn(col int) int {
	if col < 0 || col > len(resultColumns) {
		return -1
	}
	var total int = 0
	for _, def := range resultColumns {
		total = total + def.weight
	}
	weight := resultColumns[col].weight
	totalw := tv.Size().Width()
	w := (weight * totalw) / total

	return w
}

//TODO: change this to SetResults(results *Results)
// and manage the ResultsModel internally
func (tv *ResultsView) SetResultsModel(model *ResultsModel) {
	tv.model = model
	tv.SetModel(model)
}

func (tv *ResultsView) resizeEvent(ev *gui.QResizeEvent) {
	// TODO: preserve existing column width ratios!
	tv.ResizeColumnsToContents()
}
