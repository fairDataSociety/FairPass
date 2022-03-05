package screens

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/fairdatasociety/fairpass/internal/utils"
)

type password struct {
	ID               string            `json:"id,omitempty"`
	Domain           string            `json:"domain,omitempty"`
	Username         string            `json:"username,omitempty"`
	Description      string            `json:"description,omitempty"`
	Password         string            `json:"password,omitempty"`
	GeneratorOptions *generatorOptions `json:"options,omitempty"`
	IsStarred        string            `json:"starred,omitempty"`
	CreatedAt        int64             `json:"created_at,omitempty"`
	UpdatedAt        int64             `json:"updated_at,omitempty"`
}

type note struct {
	ID        string `json:"id,omitempty"`
	Title     string `json:"title,omitempty"`
	Message   string `json:"message,omitempty"`
	IsStarred string `json:"starred,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
}

type generatorOptions struct {
	Length       int
	NumDigits    int
	NumSymbols   int
	NoUpper      bool
	AllowSymbols bool
	AllowDigits  bool
	AllowRepeat  bool
}

type mainView struct {
	widget.BaseWidget

	view     *fyne.Container
	sidebar  *fyne.Container
	content  *fyne.Container
	listItem *widget.List
	index    *index
}

func newMainView(i *index) *mainView {
	main := &mainView{
		index: i,
	}
	main.ExtendBaseWidget(main)

	main.initMainView()
	return main
}

func (main *mainView) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(main.view)
}

func (main *mainView) initMainView() {
	addButton := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		var modal *widget.PopUp
		c := container.NewVBox()
		c.Add(widget.NewButton("Password", func() {
			main.setContent(main.makeAddPasswordView(nil))
			modal.Hide()
		}))
		c.Add(widget.NewButton("Notes", func() {
			main.setContent(main.makeAddNoteView(nil))
			modal.Hide()
		}))
		c.Add(widget.NewLabel(""))
		cancelButton := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
			modal.Hide()
		})
		c.Add(cancelButton)

		modal = widget.NewModalPopUp(c, main.index.Window.Canvas())
		modal.Show()
	})
	bottomContent := container.NewPadded(container.New(layout.NewHBoxLayout(), layout.NewSpacer(), addButton))
	sideBarItems := []string{"Passwords", "Notes"}
	list := widget.NewList(
		func() int {
			return len(sideBarItems)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewLabel("label"))
		},
		func(id int, obj fyne.CanvasObject) {
			obj.(*fyne.Container).Objects[0].(*widget.Label).SetText(sideBarItems[id])
		},
	)
	list.OnSelected = func(id widget.ListItemID) {
		switch id {
		case 0:
			passwordsView := newListView(main)
			main.setContent(passwordsView.view)
		case 1:
			notesView := newNotesListView(main)
			main.setContent(notesView.view)
		}
	}
	main.listItem = list
	main.sidebar = container.NewPadded(container.NewMax(container.NewBorder(nil, widget.NewLabel(fmt.Sprintf("%s-%s", utils.Version, utils.Commit)), nil, nil, container.NewMax(list))))
	intro := widget.NewLabel("Now store Passwords with FirePass in the Swarm")
	main.content = container.NewBorder(nil, bottomContent, nil, nil, container.NewCenter(container.NewVBox(intro)))
	split := container.NewHSplit(main.sidebar, main.content)
	split.Offset = 0.3
	main.view = container.NewPadded(split)
}

func (main *mainView) setContent(o fyne.CanvasObject) {
	main.content.Objects[0] = o
	main.content.Refresh()
}
