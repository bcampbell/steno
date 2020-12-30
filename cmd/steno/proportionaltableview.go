package main

import (
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

// ProportionalTableView is a QTableView which resizes it's columns when it
// is resized itself. It tries to maintain the relative column widths, and
// to use all the space available.
type ProportionalTableView struct {
	widgets.QTableView

	_ func() `constructor:"init"`
}

func (tv *ProportionalTableView) init() {
	tv.ConnectResizeEvent(tv.resizeEvent)
}

func (tv *ProportionalTableView) resizeEvent(ev *gui.QResizeEvent) {
	numColumns := tv.Model().ColumnCount(core.NewQModelIndex())

	// get current column widths
	weights := make([]float64, numColumns)
	var total float64
	for i := 0; i < numColumns; i++ {
		weights[i] = float64(tv.ColumnWidth(i))
		total += weights[i]
	}
	// normalise
	for i, weight := range weights {
		weights[i] = weight / total
	}

	// apply to new size
	newWidth := float64(ev.Size().Width())
	for i, weight := range weights {
		tv.SetColumnWidth(i, int(newWidth*weight))
	}
}
