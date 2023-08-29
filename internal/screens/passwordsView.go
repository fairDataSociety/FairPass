package screens

import (
	"encoding/json"
	"fmt"

	"fyne.io/fyne/v2/layout"

	"fyne.io/fyne/v2/theme"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/fairdatasociety/fairpass/internal/utils"
)

type listView struct {
	widget.BaseWidget

	view     fyne.CanvasObject
	items    []*password
	mainView *mainView
}

var (
	cachedPasswords []*password
	searchTerm      string
)

func newListView(mainView *mainView, forceUpdate bool, expr string) *listView {
	if cachedPasswords == nil {
		forceUpdate = true
	}
	if forceUpdate {
		items := []*password{}
		if expr == "" {
			expr = "domain=>"
		}
		list, err := mainView.index.dfsAPI.DocFind(mainView.index.sessionID, utils.PodName, utils.PasswordsTable, expr, 100)
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
		cachedPasswords = items
	}
	table := widget.NewTable(
		func() (int, int) {
			return len(cachedPasswords) + 1, 7
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
				case 5:
					label.SetText("Edit")
				case 6:
					label.SetText("Delete")
				}
				return
			}
			label := o.(*widget.Label)
			label.Wrapping = fyne.TextWrapWord
			switch id.Col {
			case 0:
				label.SetText(fmt.Sprintf("%d", id.Row))
			case 1:
				label.SetText(cachedPasswords[id.Row-1].Domain)
			case 2:
				label.SetText(cachedPasswords[id.Row-1].Username)
			case 3:
				label.SetText("Copy")
			case 4:
				label.SetText("View")
			case 5:
				label.SetText("Edit")
			default:
				label.SetText("Delete")
			}
		},
	)
	table.SetColumnWidth(1, 250)
	table.SetColumnWidth(2, 200)
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 {
			return
		}
		switch id.Col {
		case 0:
			return
		case 1:
			mainView.index.Window.Clipboard().SetContent(cachedPasswords[id.Row-1].Domain)
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "FairPass",
				Content: "Domain copied to clipboard",
			})
		case 2:
			mainView.index.Window.Clipboard().SetContent(cachedPasswords[id.Row-1].Username)
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "FairPass",
				Content: "Username copied to clipboard",
			})
		case 3:
			password, err := mainView.index.encryptor.DecryptContent(mainView.index.password, cachedPasswords[id.Row-1].Password)
			if err != nil {
				fmt.Println("failed to decrypt password ", err)
			}
			mainView.index.Window.Clipboard().SetContent(password)
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "FairPass",
				Content: "Password copied to clipboard",
			})
		case 4:
			mainView.setContent(mainView.makeAddPasswordView(cachedPasswords[id.Row-1], false))
		case 5:
			mainView.setContent(mainView.makeAddPasswordView(cachedPasswords[id.Row-1], true))
		default:
			dialog.NewConfirm("Delete Password", "Are you sure?", func(choice bool) {
				if choice {
					err := mainView.index.dfsAPI.DocDel(mainView.index.sessionID, utils.PodName, utils.PasswordsTable, cachedPasswords[id.Row-1].ID)
					if err != nil {
						fmt.Println("failed to delete password ", err)
						return
					}
					passwordsView := newListView(mainView, true, "")
					mainView.setContent(passwordsView.view)
					return
				}
				table.UnselectAll()
			}, mainView.index.Window).Show()

		}
	}

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search...")
	searchEntry.OnChanged = func(s string) {
		searchTerm = s
	}
	searchEntry.SetText(searchTerm)
	searchButton := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		expr = ""
		if searchTerm != "" {
			expr = "domain=>" + searchTerm
		}
		newListView(mainView, true, expr)
	})
	topContent := container.NewPadded(searchEntry, container.New(layout.NewHBoxLayout(), layout.NewSpacer(), searchButton))

	return &listView{
		BaseWidget: widget.BaseWidget{},
		items:      cachedPasswords,
		mainView:   mainView,
		view:       container.NewBorder(topContent, nil, nil, nil, table),
	}
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (l *listView) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(l.view)
}
