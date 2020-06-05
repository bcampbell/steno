package main

import (
	"github.com/bcampbell/steno/script"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

type ScriptDialog struct {
	widgets.QDialog

	_ func() `constructor:"init"`

	//	_ func() `slot:"revert"`
	//	_ func() `slot:"submit"`

	tree *widgets.QTreeWidget
}

func (d *ScriptDialog) init() {
	//	d.ConnectSubmit(d.submit)

	d.tree = widgets.NewQTreeWidget(nil)
	layout := widgets.NewQVBoxLayout()
	layout.AddWidget(widgets.NewQLabel2("Source:", nil, 0), 0, 0)
	layout.AddWidget(d.tree, 0, 0)

	butts := d.createButtons()
	layout.AddWidget(butts, 0, 0)
	d.SetLayout(layout)

	/*
		d.dateFrom.ConnectSelectionChanged(func() {
			d.dateTo.SetMinimumDate(d.dateFrom.SelectedDate())
		})
		d.dateTo.ConnectSelectionChanged(func() {
			d.dateFrom.SetMaximumDate(d.dateTo.SelectedDate())
		})
	*/
	d.SetWindowTitle("Run script")
}

func (d *ScriptDialog) SetScripts(scripts []*script.Script) {

	parents := map[string]*widgets.QTreeWidgetItem{}
	for idx, script := range scripts {
		item := widgets.NewQTreeWidgetItem2([]string{script.Name}, 0) //widgets.QTreeWidgetItem__Type)
		// Store the script index in the data field.
		item.SetData(0, int(core.Qt__UserRole), core.NewQVariant5(idx))
		if script.Category == "" {
			d.tree.AddTopLevelItem(item)
		} else {

			cat, got := parents[script.Category]
			if !got {
				cat = widgets.NewQTreeWidgetItem2([]string{script.Category}, 0) //widgets.QTreeWidgetItem__Type)
				parents[script.Category] = cat
				d.tree.AddTopLevelItem(cat)
			}
			cat.AddChild(item)
		}
	}
}

// SelectedIndex returns the index of the selected script, if any.
// The second return value will be false if none selected.
// The index is into the scripts array originally passed in via SetScripts().
func (d *ScriptDialog) SelectedIndex() (int, bool) {
	sel := d.tree.SelectedItems()
	if len(sel) > 0 {
		var got bool
		idx := sel[0].Data(0, int(core.Qt__UserRole)).ToInt(&got)
		if got {
			return idx, true
		}
	}
	return 0, false
}

func (d *ScriptDialog) createButtons() *widgets.QDialogButtonBox {
	closeButton := widgets.NewQPushButton2("&Close", nil)
	submitButton := widgets.NewQPushButton2("&Submit", nil)

	closeButton.SetDefault(true)

	closeButton.ConnectClicked(func(bool) { d.Close() })
	submitButton.ConnectClicked(func(bool) { d.Accept() })

	buttonBox := widgets.NewQDialogButtonBox(nil)
	buttonBox.AddButton(submitButton, widgets.QDialogButtonBox__ResetRole)
	buttonBox.AddButton(closeButton, widgets.QDialogButtonBox__RejectRole)

	return buttonBox
}
