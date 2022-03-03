package screens

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/google/uuid"
	"github.com/onepeerlabs/fairpass/internal/utils"
	generator "github.com/sethvargo/go-password/password"
)

type item struct {
	ID               string            `json:"id,omitempty"`
	Domain           string            `json:"domain,omitempty"`
	Username         string            `json:"username,omitempty"`
	Description      string            `json:"description,omitempty"`
	Password         string            `json:"password,omitempty"`
	GeneratorOptions *generatorOptions `json:"options,omitempty"`
	IsStarred        bool              `json:"starred,omitempty"`
	CreatedAt        int64             `json:"created_at,omitempty"`
	UpdatedAt        int64             `json:"updated_at,omitempty"`
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

	view    *fyne.Container
	sidebar *fyne.Container
	content *fyne.Container

	index *index
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
		w := widget.NewButton("Password", func() {
			main.setContent(main.makeAddItemView(nil))
			modal.Hide()
		})
		c.Add(w)
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
			return container.NewHBox(widget.NewLabel("Identity label"))
		},
		func(id int, obj fyne.CanvasObject) {
			obj.(*fyne.Container).Objects[0].(*widget.Label).SetText(sideBarItems[id])
		},
	)
	list.OnSelected = func(id widget.ListItemID) {
		switch id {
		case 0:
			main.setContent(container.NewMax(widget.NewLabel("Passwords")))
		case 1:
			main.setContent(container.NewMax(widget.NewLabel("Notes")))
		}
	}
	main.sidebar = container.NewPadded(container.NewMax(container.NewBorder(nil, widget.NewLabel(fmt.Sprintf("%s-%s", utils.Version, utils.Commit)), nil, nil, container.NewMax(list))))
	intro := widget.NewLabel("Now store Passwords with FireOS in the Swarm")
	main.content = container.NewBorder(nil, bottomContent, nil, nil, container.NewCenter(container.NewVBox(intro)))
	split := container.NewHSplit(main.sidebar, main.content)
	split.Offset = 0.3
	main.view = container.NewPadded(split)
}

// makeAddItemButton returns the button used to add an item
func (main *mainView) makeAddItemButton() fyne.CanvasObject {
	button := widget.NewButtonWithIcon("Add Item", theme.ContentAddIcon(), func() {
		fmt.Println("Add item")
	})
	button.Importance = widget.HighImportance
	return button
}

// setContent sets the content view with the provided object and refresh
func (main *mainView) setContent(o fyne.CanvasObject) {
	main.content.Objects[0] = o
	main.content.Refresh()
}

func (main *mainView) makeAddItemView(i *item) fyne.CanvasObject {
	if i == nil {
		i = &item{}
	} else {
		// Ask password
		var err error
		i.Password, err = main.index.encryptor.DecryptContent("159263487", i.Password)
		if err != nil {
			fmt.Println("failed to decrypt password")
		}
	}
	websiteEntry := widget.NewEntryWithData(binding.BindString(&i.Domain))
	websiteEntry.SetPlaceHolder("Domain Name")
	websiteEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("please enter a domain name")
		}
		return nil
	}

	usernameEntry := widget.NewEntryWithData(binding.BindString(&i.Username))
	usernameEntry.SetPlaceHolder("Username")
	usernameEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("please enter a username")
		}
		return nil
	}

	// the note field
	noteEntry := widget.NewEntryWithData(binding.BindString(&i.Description))
	noteEntry.SetPlaceHolder("Description")
	noteEntry.MultiLine = true
	noteEntry.Validator = nil

	// center
	passwordEntry := widget.NewPasswordEntry()
	passwordBind := binding.BindString(&i.Password)

	passwordEntry.Bind(passwordBind)
	passwordEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("please enter a password")
		}
		return nil
	}
	passwordEntry.SetPlaceHolder("Password")
	passwordCopyButton := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		main.index.Window.Clipboard().SetContent(passwordEntry.Text)
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Locker",
			Content: "Password copied to clipboard",
		})
	})

	passwordMakeButton := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		d := dialog.NewCustomConfirm("Generate password", "Ok", "Cancel", i.makePasswordDialog(), func(b bool) {
			if b {
				passwordBind.Set(i.Password)
			}
		}, main.index.Window)
		d.Show()
	})
	var passwordObject fyne.CanvasObject
	if i.ID != "" {
		websiteEntry.Disable()
		usernameEntry.Disable()
		passwordEntry.Disable()
		noteEntry.Disable()
		passwordObject = container.NewBorder(nil, nil, nil, container.NewHBox(passwordCopyButton), passwordEntry)
	} else {
		passwordObject = container.NewBorder(nil, nil, nil, container.NewHBox(passwordCopyButton, passwordMakeButton), passwordEntry)
	}

	form := container.NewVBox(
		labelWithStyle("Domain"),
		websiteEntry,
		labelWithStyle("Username"),
		usernameEntry,
		labelWithStyle("Password"),
		passwordObject,
		labelWithStyle("Description"),
		noteEntry,
	)

	cancelBtn := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		main.setContent(container.NewMax(widget.NewLabel("Passwords")))
	})

	var top fyne.CanvasObject
	if i.ID == "" {
		saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
			main.index.progress.Show()
			defer main.index.progress.Hide()
			if i.Domain == "" {
				return
			}
			if i.Username == "" {
				return
			}
			if i.Password == "" {
				return
			}
			i.ID = uuid.New().String()
			// Ask password
			var err error
			i.Password, err = main.index.encryptor.EncryptContent("159263487", i.Password)
			if err != nil {
				fmt.Println("failed to save password")
				return
			}
			data, err := json.Marshal(i)
			if err != nil {
				fmt.Println("failed to save password")
				return
			}
			err = main.index.dfsAPI.DocPut(main.index.sessionID, utils.PodName, utils.PasswordsTable, data)
			if err != nil {
				fmt.Println("failed to save password")
				return
			}
			main.setContent(container.NewMax(widget.NewLabel("Passwords")))
		})
		saveBtn.Importance = widget.HighImportance
		top = container.NewBorder(nil, nil, cancelBtn, saveBtn, widget.NewLabel(""))
	} else {
		top = container.NewBorder(nil, nil, cancelBtn, nil, widget.NewLabel(""))
	}

	return container.NewPadded(container.NewBorder(top, nil, nil, nil, form))
}

func (i *item) makePasswordDialog() fyne.CanvasObject {
	g := &generatorOptions{
		Length:       16,
		NumDigits:    0,
		NumSymbols:   0,
		NoUpper:      false,
		AllowRepeat:  true,
		AllowSymbols: true,
		AllowDigits:  true,
	}
	passwordBind := binding.BindString(&i.Password)
	passwordEntry := widget.NewEntryWithData(passwordBind)
	passwordEntry.Validator = nil
	refreshButton := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		err := makePassword(g, passwordBind)
		if err != nil {
			return
		}
	})

	lengthBind := binding.BindInt(&g.Length)
	if g.Length == 0 {
		lengthBind.Set(16)
	}

	lengthEntry := widget.NewEntryWithData(binding.IntToString(lengthBind))
	lengthEntry.Disabled()
	lengthEntry.Validator = nil
	lengthEntry.OnChanged = func(s string) {
		if s == "" {
			return
		}
		l, err := strconv.Atoi(s)
		if err != nil {
			log.Println(err)
			return
		}
		if l < 8 || l > 128 {
			log.Printf("password lenght must be between %d and %d, got %d", 8, 128, l)
			return
		}
		lengthBind.Set(l)
		err = makePassword(g, passwordBind)
		if err != nil {
			return
		}
	}

	lengthSlider := widget.NewSlider(float64(8), float64(128))
	lengthSlider.OnChanged = func(f float64) {
		lengthBind.Set(int(f))
		err := makePassword(g, passwordBind)
		if err != nil {
			return
		}
	}

	digitsButton := widget.NewCheck("0-9", func(isChecked bool) {
		g.AllowDigits = isChecked
		err := makePassword(g, passwordBind)
		if err != nil {
			return
		}
	})
	digitsButton.SetChecked(true)

	symbolsButton := widget.NewCheck("!%$", func(isChecked bool) {
		g.AllowSymbols = isChecked
		err := makePassword(g, passwordBind)
		if err != nil {
			return
		}
	})
	symbolsButton.SetChecked(true)
	uppercaseButton := widget.NewCheck("A-Z", func(isChecked bool) {
		g.NoUpper = !isChecked
		err := makePassword(g, passwordBind)
		if err != nil {
			return
		}
	})
	uppercaseButton.SetChecked(true)

	optionsForm := widget.NewForm()
	optionsForm.Append(
		"Password",
		container.NewBorder(nil, nil, nil, refreshButton, passwordEntry),
	)

	optionsForm.Append(
		"Length",
		container.NewBorder(nil, nil, nil, lengthEntry, lengthSlider),
	)

	optionsForm.Append(
		"",
		container.NewGridWithColumns(3, digitsButton, symbolsButton, uppercaseButton),
	)

	return container.NewMax(optionsForm)
}

func makePassword(g *generatorOptions, externalString binding.ExternalString) error {
	if g.AllowSymbols {
		g.NumSymbols = g.Length * 25 / 100
	} else {
		g.NumSymbols = 0
	}
	if g.AllowDigits {
		g.NumDigits = g.Length * 25 / 100

	} else {
		g.NumDigits = 0
	}

	password, err := generator.Generate(g.Length, g.NumDigits, g.NumSymbols, g.NoUpper, g.AllowRepeat)
	if err != nil {
		return err
	}
	return externalString.Set(password)
}
