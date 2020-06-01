package main

import (
	"time"

	"github.com/bcampbell/steno/steno"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

type SlurpDialog struct {
	widgets.QDialog

	_ func() `constructor:"init"`

	//	_ func() `slot:"revert"`
	//	_ func() `slot:"submit"`

	source   *widgets.QComboBox
	dateFrom *widgets.QCalendarWidget
	dateTo   *widgets.QCalendarWidget
}

func (d *SlurpDialog) init() {
	//	d.ConnectSubmit(d.submit)

	d.source = widgets.NewQComboBox(nil)
	d.dateFrom = widgets.NewQCalendarWidget(nil)
	d.dateTo = widgets.NewQCalendarWidget(nil)
	layout := widgets.NewQVBoxLayout()
	layout.AddWidget(widgets.NewQLabel2("Source:", nil, 0), 0, 0)
	layout.AddWidget(d.source, 0, 0)
	layout.AddWidget(widgets.NewQLabel2("From:", nil, 0), 0, 0)
	layout.AddWidget(d.dateFrom, 0, 0)
	layout.AddWidget(widgets.NewQLabel2("To:", nil, 0), 0, 0)
	layout.AddWidget(d.dateTo, 0, 0)

	butts := d.createButtons()
	layout.AddWidget(butts, 0, 0)
	d.SetLayout(layout)

	d.dateFrom.ConnectSelectionChanged(func() {
		d.dateTo.SetMinimumDate(d.dateFrom.SelectedDate())
	})
	d.dateTo.ConnectSelectionChanged(func() {
		d.dateFrom.SetMaximumDate(d.dateTo.SelectedDate())
	})

	d.SetWindowTitle("Wibble")
}

func (d *SlurpDialog) SetSources(srcs []steno.SlurpSource) {
	for _, src := range srcs {
		d.source.AddItem(src.Name, core.NewQVariant())
	}
}

func (d *SlurpDialog) SourceIndex() int {
	// TODO: should get variant and unpack
	return d.source.CurrentIndex()
}

func (d *SlurpDialog) createButtons() *widgets.QDialogButtonBox {
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

func (d *SlurpDialog) DateRange() (time.Time, time.Time) {
	f := d.dateFrom.SelectedDate()
	t := d.dateTo.SelectedDate()

	fromDay := time.Date(f.Year(), time.Month(f.Month()), f.Day(), 0, 0, 0, 0, time.UTC)
	toDay := time.Date(t.Year(), time.Month(t.Month()), t.Day(), 0, 0, 0, 0, time.UTC)
	return fromDay, toDay
}

/*
func (d *SlurpDialog) initWith() {

	inputWidgetBox := d.createInputWidgets()
	buttonBox := d.createButtons()

	layout := widgets.NewQVBoxLayout()
	layout.AddWidget(inputWidgetBox, 0, 0)
	layout.AddWidget(buttonBox, 0, 0)
	d.SetLayout(layout)

	d.SetWindowTitle("Add Album")
}

func (d *Dialog) submit() {

	artist := d.artistEditor.Text()
	title := d.titleEditor.Text()

	if artist == "" || title == "" {
		widgets.QMessageBox_Information(d, "Add Album",
			`Please provide both the name of the artist
and the title of the album.`, 0, 0)
	} else {
		d.Accept()
	}
}
*/
