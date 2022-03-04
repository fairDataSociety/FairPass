package screens

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	dfsUtils "github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/onepeerlabs/fairpass/internal/utils"
	"github.com/onepeerlabs/fairpass/internal/utils/crypto"
	"github.com/sirupsen/logrus"
)

type fairOSConfig struct {
	BeeEndpoint string
	BatchId     string
}

type userRequest struct {
	Username string
	Password string
	Address  string
}

func (i *index) initLoginView() fyne.CanvasObject {
	// load config
	data, err := ioutil.ReadFile(filepath.Join(i.dataDir, config))
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(i.dataDir, 0700)
			if err != nil {
				i.view = container.NewPadded(container.NewCenter(widget.NewLabel(fmt.Sprintf("Failed to load config : %s", err.Error()))))
				return i.view
			}
			// return show config page, dont allow block
			return i.initConfigView(false)
		}
		i.view = container.NewPadded(container.NewCenter(widget.NewLabel(fmt.Sprintf("Failed to load config : %s", err.Error()))))
		return i.view
	}
	err = json.Unmarshal(data, i.config)
	if err != nil {
		i.view = container.NewPadded(container.NewCenter(widget.NewLabel(fmt.Sprintf("Failed to load config : %s", err.Error()))))
		return i.view
	}
	if i.dfsAPI == nil {
		logger := logging.New(os.Stdout, logrus.ErrorLevel)
		api, err := dfs.NewDfsAPI(
			i.dataDir,
			i.config.BeeEndpoint,
			"",
			i.config.BatchId,
			logger,
		)
		if err != nil {
			i.view = container.NewPadded(container.NewCenter(widget.NewLabel(fmt.Sprintf("Failed to connect with bee : %s", err.Error()))))
			return i.view
		}
		i.dfsAPI = api
	}

	t := canvas.NewText(utils.LoginWelcome, theme.ForegroundColor())
	t.TextStyle = fyne.TextStyle{Bold: true}
	t.TextSize = theme.TextSubHeadingSize()
	t.Alignment = fyne.TextAlignCenter

	var users []string
	// get users
	files, err := ioutil.ReadDir(filepath.Join(i.dataDir, "user"))
	if err != nil {
		users = []string{}
	} else {
		for _, f := range files {
			users = append(users, f.Name())
		}
	}

	if len(users) == 0 {
		return i.initSignupView(false)
	}

	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("Password")
	password.Hide()
	var username string
	loginBtn := widget.NewButton("Login", func() {
		i.progress = dialog.NewProgressInfinite("", "Login is progress", i)
		i.progress.Show()
		defer i.progress.Hide()
		// Do login
		ui, err := i.dfsAPI.LoginUser(username, password.Text, "")
		if err != nil {
			fmt.Printf("Login Failed : %s", err.Error())
			return
		}
		_, enBytes, err := ui.GetFeed().GetFeedData(dfsUtils.HashString(username), ui.GetAccount().GetAddress(-1))
		if err != nil {
			fmt.Printf("Login Failed : %s", err.Error())
			return
		}
		i.encryptor = crypto.New(string(enBytes))
		i.sessionID = ui.GetSessionId()
		i.password = password.Text
		if !i.dfsAPI.IsPodExist(utils.PodName, i.sessionID) {
			_, err = i.dfsAPI.CreatePod(utils.PodName, i.password, i.sessionID)
			if err != nil {
				fmt.Printf("Create Pod Failed : %s", err.Error())
				return
			}
		} else {
			_, err = i.dfsAPI.OpenPod(utils.PodName, i.password, i.sessionID)
			if err != nil {
				fmt.Printf("Open pod Failed : %s", err.Error())
				return
			}
		}
		passwordIndexes := make(map[string]collection.IndexType)
		passwordIndexes["domain"] = collection.StringIndex
		passwordIndexes["username"] = collection.StringIndex
		passwordIndexes["password"] = collection.StringIndex
		passwordIndexes["starred"] = collection.StringIndex
		err = i.dfsAPI.DocCreate(i.sessionID, utils.PodName, utils.PasswordsTable, passwordIndexes, true)
		if err != nil && err != collection.ErrDocumentDBAlreadyPresent {
			fmt.Printf("Failed to create doc table : %s", err.Error())
			return
		}
		err = i.dfsAPI.DocOpen(i.sessionID, utils.PodName, utils.PasswordsTable)
		if err != nil {
			fmt.Printf("Failed to open doc table : %s", err.Error())
			return
		}
		notesIndexes := make(map[string]collection.IndexType)
		notesIndexes["title"] = collection.StringIndex
		notesIndexes["starred"] = collection.StringIndex
		err = i.dfsAPI.DocCreate(i.sessionID, utils.PodName, utils.NotesTable, notesIndexes, true)
		if err != nil && err != collection.ErrDocumentDBAlreadyPresent {
			fmt.Printf("Failed to create doc table : %s", err.Error())
			return
		}
		err = i.dfsAPI.DocOpen(i.sessionID, utils.PodName, utils.NotesTable)
		if err != nil {
			fmt.Printf("Failed to open doc table : %s", err.Error())
			return
		}

		main := newMainView(i)
		i.setContent(main.view)
	})
	loginBtn.Hide()
	combo := widget.NewSelect(users, func(value string) {
		username = value
		password.Show()
		loginBtn.Show()
	})
	addButton := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		i.setContent(i.initSignupView(true))
	})
	bottomContent := container.NewPadded(container.New(layout.NewHBoxLayout(), layout.NewSpacer(), addButton))
	configButton := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		i.setContent(i.initConfigView(true))
	})
	topContent := container.NewPadded(container.New(layout.NewHBoxLayout(), layout.NewSpacer(), configButton))
	return container.NewBorder(topContent, bottomContent, nil, nil, container.NewCenter(container.NewVBox(t, combo, password, loginBtn)))
}

func (i *index) initSignupView(allowBack bool) fyne.CanvasObject {
	tabs := container.NewAppTabs(
		container.NewTabItem("Signup", i.signupTab(allowBack)),
		container.NewTabItem("Import user", i.importTab(allowBack)),
	)
	tabs.SetTabLocation(container.TabLocationTop)
	return tabs
}

func (i *index) signupTab(allowBack bool) fyne.CanvasObject {
	user := &userRequest{}
	usernameEntry := widget.NewEntryWithData(binding.BindString(&user.Username))
	usernameEntry.SetPlaceHolder("Username")
	usernameEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("please enter a username")
		}
		return nil
	}

	passwordEntry := widget.NewPasswordEntry()
	passwordBind := binding.BindString(&user.Password)

	passwordEntry.Bind(passwordBind)
	passwordEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("please enter a password")
		}
		return nil
	}
	passwordEntry.SetPlaceHolder("Password")

	form := container.NewVBox(
		labelWithStyle("Username"),
		usernameEntry,
		labelWithStyle("Password"),
		passwordEntry,
	)

	saveBtn := widget.NewButtonWithIcon("Signup", theme.DocumentSaveIcon(), func() {
		if user.Username == "" {
			return
		}
		if user.Password == "" {
			return
		}
		i.progress = dialog.NewProgressInfinite("", "Creating User", i.Window)
		i.progress.Show()
		// Do signup
		address, mnemonic, _, err := i.dfsAPI.CreateUser(user.Username, user.Password, "", "")
		if err != nil {
			fmt.Printf("Failed to create user : %s", err.Error())
			i.progress.Hide()
			return
		}
		i.progress.Hide()
		fmt.Printf("User %s got created \n%s\n%s\n", user.Username, address, mnemonic)
		d := dialog.NewCustomConfirm("Caution !!", "Yes, I have kept is safe", "Cancel", i.displayMnemonic(address, mnemonic), func(b bool) {
			i.Reload()
		}, i.Window)
		d.Show()
	})
	saveBtn.Importance = widget.HighImportance
	cancelBtn := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		i.Reload()
	})
	if !allowBack {
		cancelBtn.Hide()
	}
	bottom := container.NewBorder(nil, nil, cancelBtn, saveBtn, widget.NewLabel(""))

	return container.NewPadded(container.NewBorder(nil, bottom, nil, nil, form))
}

func (i *index) displayMnemonic(address, mnemonic string) fyne.CanvasObject {
	header := labelWithStyle("Please keep the following safe. If these are lost you cannot recover your data")
	addressLabel := widget.NewLabel(address)
	mnemonicLabel := widget.NewLabel(mnemonic)

	header.Wrapping = fyne.TextWrapWord
	mnemonicLabel.Wrapping = fyne.TextWrapWord
	addressLabel.Wrapping = fyne.TextWrapWord

	optionsForm := widget.NewForm()
	copyToClipAddr := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		i.Window.Clipboard().SetContent(address)
	})
	optionsForm.Append(
		"Address",
		container.NewBorder(nil, nil, nil, copyToClipAddr, addressLabel),
	)
	copyToClip := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		i.Window.Clipboard().SetContent(mnemonic)
	})
	optionsForm.Append(
		"Mnemonic",
		container.NewBorder(nil, nil, nil, copyToClip, mnemonicLabel),
	)

	return container.NewMax(container.NewBorder(header, optionsForm, nil, nil))
}

func (i *index) importTab(allowBack bool) fyne.CanvasObject {
	user := &userRequest{}
	usernameEntry := widget.NewEntryWithData(binding.BindString(&user.Username))
	usernameEntry.SetPlaceHolder("Username")
	usernameEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("please enter a username")
		}
		return nil
	}

	passwordEntry := widget.NewPasswordEntry()
	passwordBind := binding.BindString(&user.Password)

	passwordEntry.Bind(passwordBind)
	passwordEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("please enter a password")
		}
		return nil
	}
	passwordEntry.SetPlaceHolder("Password")

	addressEntry := widget.NewEntryWithData(binding.BindString(&user.Address))
	addressEntry.SetPlaceHolder("Address")
	addressEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("please enter a address")
		}
		return nil
	}

	form := container.NewVBox(
		labelWithStyle("Username"),
		usernameEntry,
		labelWithStyle("Password"),
		passwordEntry,
		labelWithStyle("Address"),
		addressEntry,
	)

	saveBtn := widget.NewButtonWithIcon("Import", theme.DocumentSaveIcon(), func() {
		if user.Username == "" {
			return
		}
		if user.Password == "" {
			return
		}
		if user.Address == "" {
			return
		}
		i.progress = dialog.NewProgressInfinite("", "Importing user", i)
		i.progress.Show()
		defer i.progress.Hide()
		// TODO import
		i.Reload()
	})
	saveBtn.Importance = widget.HighImportance
	cancelBtn := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		i.Reload()
	})
	if !allowBack {
		cancelBtn.Hide()
	}
	bottom := container.NewBorder(nil, nil, cancelBtn, saveBtn, widget.NewLabel(""))

	return container.NewPadded(container.NewBorder(nil, bottom, nil, nil, form))
}

func (i *index) initConfigView(allowBack bool) fyne.CanvasObject {
	beeEntry := widget.NewEntryWithData(binding.BindString(&i.config.BeeEndpoint))
	beeEntry.SetPlaceHolder("Bee Endpoint")
	beeEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("please enter bee endpoint")
		}
		return nil
	}

	stampEntry := widget.NewEntryWithData(binding.BindString(&i.config.BatchId))
	stampEntry.SetPlaceHolder("Batch ID")
	stampEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("please enter batch ID")
		}
		return nil
	}

	form := container.NewVBox(
		labelWithStyle("Bee Endpoint"),
		beeEntry,
		labelWithStyle("Batch ID"),
		stampEntry,
	)

	saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
		if i.config.BeeEndpoint == "" {
			return
		}
		if i.config.BatchId == "" {
			return
		}

		// Save config
		configBytes, err := json.Marshal(i.config)
		if err != nil {
			fmt.Println("config save failed")
			return
		}
		err = ioutil.WriteFile(filepath.Join(i.dataDir, config), configBytes, 0700)
		if err != nil {
			fmt.Println("config write failed ", err)
			return
		}
		i.setContent(i.initLoginView())
	})
	saveBtn.Importance = widget.HighImportance
	cancelBtn := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		i.setContent(i.initLoginView())
	})
	if !allowBack {
		cancelBtn.Hide()
	}
	bottom := container.NewBorder(nil, nil, cancelBtn, saveBtn, widget.NewLabel(""))

	return container.NewPadded(container.NewBorder(nil, bottom, nil, nil, form))
}

func labelWithStyle(label string) *widget.Label {
	return widget.NewLabelWithStyle(label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
}

// setContent sets the content view with the provided object and refresh
func (i *index) setContent(o fyne.CanvasObject) {
	i.view.Objects = []fyne.CanvasObject{o}
	i.view.Refresh()
}
