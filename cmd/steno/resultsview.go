package main

import (
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

type ResultsView struct {
	widgets.QTableView

	_ func() `constructor:"init"`

	model *ResultsModel
}

func (tv *ResultsView) init() {

	tv.SetShowGrid(false)
	tv.SetSelectionBehavior(widgets.QAbstractItemView__SelectRows)
	tv.SetSelectionMode(widgets.QAbstractItemView__ExtendedSelection)
	tv.VerticalHeader().SetVisible(false)

	tv.ResizeColumnsToContents()

	{
		hdr := tv.HorizontalHeader()
		/*
			// cheesy autosize.
			w := tv.Width()
			if w < 600 {
				w = 600
			}

			// Relative weights for each column
			weights := []int{2, 2, 1, 1, 1, 1}
			//n := v.model.columnCount(core.NewQModelIndex())
			total := 0
			for _, weight := range weights {
				total += weight
			}

			for i, weight := range weights {
				hdr.ResizeSection(i, (w*weight)/total)
			}
		*/
		hdr.SetCascadingSectionResizes(true)
		hdr.SetStretchLastSection(true)
		//hdr.SetSectionResizeMode(widgets.QHeaderView__ResizeToContents)
		//hdr.SetSectionResizeMode(widgets.QHeaderView__Stretch)
		// set up sorting (by clicking on column headers)
		hdr.SetSortIndicatorShown(true)
		hdr.ConnectSortIndicatorChanged(func(logicalIndex int, order core.Qt__SortOrder) {
			if tv.model.results == nil {
				return
			}
			field := ""
			switch logicalIndex {
			case 0:
				field = "url"
			case 1:
				field = "headline"
			case 2:
				field = "published"
			case 3:
				field = "pub"
			case 4:
				field = "tags"
			case 5:
				field = "similar"
			default:
				return
			}

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

//TODO: change this to SetResults(results *Results)
// and manage the ResultsModel internally
func (tv *ResultsView) SetResultsModel(model *ResultsModel) {
	tv.model = model
	tv.SetModel(model)
}
