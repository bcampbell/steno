package main

import (
	//"fmt"
	"strings"

	"github.com/bcampbell/steno/steno"
	"github.com/therecipe/qt/core"
	//"github.com/therecipe/qt/widgets"
)

// ResultsModel implements a QAbstractTableModel around our Results struct
type ResultsModel struct {
	core.QAbstractTableModel

	_ func() `constructor:"init"`

	// can be nil
	results *steno.Results
}

func (m *ResultsModel) init() {
	//	m.modelData = []TableItem{{"john", "doe"}, {"john", "bob"}}
	m.ConnectHeaderData(m.headerData)
	m.ConnectRowCount(m.rowCount)
	m.ConnectColumnCount(m.columnCount)
	m.ConnectData(m.data)
}

// Install a new Results (nil is OK)
func (m *ResultsModel) setResults(r *steno.Results) {
	m.BeginResetModel()
	m.results = r
	m.EndResetModel()
}

func (m *ResultsModel) headerData(section int, orientation core.Qt__Orientation, role int) *core.QVariant {
	if role != int(core.Qt__DisplayRole) || orientation == core.Qt__Vertical {
		return m.HeaderDataDefault(section, orientation, role)
	}

	switch section {
	case 0:
		return core.NewQVariant1("URL")
	case 1:
		return core.NewQVariant1("Headline")
	case 2:
		return core.NewQVariant1("Published")
	case 3:
		return core.NewQVariant1("Pub")
	case 4:
		return core.NewQVariant1("Tags")
	case 5:
		return core.NewQVariant1("Similar")
	}
	return core.NewQVariant()
}

func (m *ResultsModel) columnCount(*core.QModelIndex) int {
	//	fmt.Printf("columnCount()\n")
	return 6
}

func (m *ResultsModel) rowCount(*core.QModelIndex) int {
	if m.results == nil {
		return 0
	}
	//	fmt.Printf("rowCount(): %d\n", len(m.results.Arts))
	return len(m.results.Arts)
}

func (m *ResultsModel) data(index *core.QModelIndex, role int) *core.QVariant {
	if m.results == nil {
		return core.NewQVariant()
	}
	if role != int(core.Qt__DisplayRole) {
		return core.NewQVariant()
	}

	rowNum := index.Row()
	art := m.results.Art(rowNum)

	//	fmt.Printf("data(): %d %d\n", rowNum, index.Column())
	switch index.Column() {
	case 0:
		return core.NewQVariant1(art.CanonicalURL)
	case 1:
		return core.NewQVariant1(art.Headline)
	case 2:
		return core.NewQVariant1(art.Published)
	case 3:
		return core.NewQVariant1(art.Pub)
	case 4:
		return core.NewQVariant1(strings.Join(art.Tags, ","))
	case 5:
		return core.NewQVariant1(len(art.Similar))
	}
	return core.NewQVariant()
}
