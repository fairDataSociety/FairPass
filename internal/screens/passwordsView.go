package screens

import (
	"encoding/json"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/onepeerlabs/fairpass/internal/utils"
)

type listView struct {
	widget.BaseWidget

	view     *widget.Table
	items    []*password
	mainView *mainView
}

func newListView(mainView *mainView) *listView {
	items := []*password{}
	list, err := mainView.index.dfsAPI.DocFind(mainView.index.sessionID, utils.PodName, utils.PasswordsTable, "id>0", 100)
	if err == nil {
		for _, v := range list {
			r := &password{}
			err := json.Unmarshal(v, r)
			if err != nil {
				continue
			}
			items = append(items, r)
		}
	}
	table := widget.NewTable(
		func() (int, int) {
			return len(items) + 1, 5
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
					label.SetText("Domain")
				case 2:
					label.SetText("Username")
				case 3:
					label.SetText("Password")
				case 4:
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
				label.SetText(items[id.Row-1].Domain)
			case 2:
				label.SetText(items[id.Row-1].Username)
			case 3:
				label.SetText("Copy")
			default:
				label.SetText("View")
			}
		},
	)
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 {
			return
		}
		switch id.Col {
		case 0:
			return
		case 1:
			mainView.index.Window.Clipboard().SetContent(items[id.Row-1].Domain)
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "FairPass",
				Content: "Domain copied to clipboard",
			})
		case 2:
			mainView.index.Window.Clipboard().SetContent(items[id.Row-1].Username)
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "FairPass",
				Content: "Username copied to clipboard",
			})
		case 3:
			password, err := mainView.index.encryptor.DecryptContent(mainView.index.password, items[id.Row-1].Password)
			if err != nil {
				fmt.Println("failed to decrypt password ", err)
			}
			mainView.index.Window.Clipboard().SetContent(password)
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "FairPass",
				Content: "Password copied to clipboard",
			})
		default:
			mainView.setContent(mainView.makeAddPasswordView(items[id.Row-1]))
		}
	}

	return &listView{
		BaseWidget: widget.BaseWidget{},
		items:      items,
		mainView:   mainView,
		view:       table,
	}
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (l *listView) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(l.view)
}
