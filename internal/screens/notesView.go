package screens

import (
	"encoding/json"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/onepeerlabs/fairpass/internal/utils"
)

type notesListView struct {
	widget.BaseWidget

	view     *widget.Table
	items    []*note
	mainView *mainView
}

func newNotesListView(mainView *mainView) *notesListView {
	items := []*note{}
	list, err := mainView.index.dfsAPI.DocFind(mainView.index.sessionID, utils.PodName, utils.NotesTable, "id>0", 100)
	if err == nil {
		for _, v := range list {
			r := &note{}
			err := json.Unmarshal(v, r)
			if err != nil {
				continue
			}
			items = append(items, r)
		}
	}
	table := widget.NewTable(
		func() (int, int) {
			return len(items) + 1, 3
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("wide content")
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			if id.Row == 0 {
				label := o.(*widget.Label)
				switch id.Col {
				case 0:

					label.SetText(fmt.Sprintf("%d", id.Row))
				case 1:
					label.SetText("Title")
				case 2:
					label.SetText("Info")
				}
				return
			}
			label := o.(*widget.Label)
			label.Wrapping = fyne.TextWrapWord
			switch id.Col {
			case 0:
				label.SetText(fmt.Sprintf("%d", id.Row))
			case 1:
				label.SetText(items[id.Row-1].Title)
			default:
				label.SetText("View")
			}
		},
	)
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 {
			return
		}
		mainView.setContent(mainView.makeAddNoteView(items[id.Row-1]))
	}

	return &notesListView{
		BaseWidget: widget.BaseWidget{},
		items:      items,
		mainView:   mainView,
		view:       table,
	}
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (l *notesListView) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(l.view)
}
