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
	"github.com/fairdatasociety/fairOS-dfs/pkg/contracts"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	dfsUtils "github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/fairdatasociety/fairpass/internal/utils"
	"github.com/fairdatasociety/fairpass/internal/utils/crypto"
	"github.com/sirupsen/logrus"
)

type fairOSConfig struct {
	BeeEndpoint string
	BatchId     string
	RPC         string
}

type userRequest struct {
	Username string
	Password string
	Address  string
	Mnemonic string
}

func (i *index) initLoginView() fyne.CanvasObject {
	// load config
	data, err := ioutil.ReadFile(filepath.Join(i.app.Storage().RootURI().Path(), config))
	if err != nil {
		if os.IsNotExist(err) {
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
	configButton := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		i.setContent(i.initConfigView(true))
	})
	topContent := container.NewPadded(container.New(layout.NewHBoxLayout(), layout.NewSpacer(), configButton))
	if i.dfsAPI == nil {
		logger := logging.New(os.Stdout, logrus.ErrorLevel)
		ensConfig := &contracts.Config{
			ENSRegistryAddress:        utils.EnsRegistryAddress,
			SubdomainRegistrarAddress: utils.SubdomainRegistrarAddress,
			PublicResolverAddress:     utils.PublicResolverAddress,
			ProviderDomain:            utils.ProviderDomain,
			ProviderBackend:           i.config.RPC,
		}
		api, err := dfs.NewDfsAPI(
			"",
			i.config.BeeEndpoint,
			i.config.BatchId,
			false,
			ensConfig,
			logger,
		)
		if err != nil {
			i.view = container.NewBorder(topContent, nil, nil, nil, container.NewCenter(widget.NewLabel(fmt.Sprintf("Failed to connect with bee : %s", err.Error()))))
			return i.view
		}
		i.dfsAPI = api
	}

	t := canvas.NewText(utils.LoginWelcome, theme.ForegroundColor())
	t.TextStyle = fyne.TextStyle{Bold: true}
	t.TextSize = theme.TextSubHeadingSize()
	t.Alignment = fyne.TextAlignCenter

	usernameInput := widget.NewEntry()
	usernameInput.SetPlaceHolder("username")
	passwordInput := widget.NewPasswordEntry()
	passwordInput.SetPlaceHolder("password")
	var username string
	loginBtn := widget.NewButton("Login", func() {
		i.progress = dialog.NewProgressInfinite("", "Login is progress", i) //lint:ignore SA1019 fyne-io/fyne/issues/2782
		i.progress.Show()
		defer i.progress.Hide()
		// Do login
		ui, _, _, err := i.dfsAPI.LoginUserV2(username, passwordInput.Text, "")
		if err != nil {
			fmt.Printf("Login Failed : %s\n", err.Error())
			return
		}
		_, enBytes, err := ui.GetFeed().GetFeedData(dfsUtils.HashString(username), ui.GetAccount().GetAddress(-1))
		if err != nil {
			fmt.Printf("Login Failed : %s\n", err.Error())
			return
		}
		i.encryptor = crypto.New(string(enBytes))
		i.sessionID = ui.GetSessionId()
		i.password = passwordInput.Text
		if !i.dfsAPI.IsPodExist(utils.PodName, i.sessionID) {
			_, err = i.dfsAPI.CreatePod(utils.PodName, i.password, i.sessionID)
			if err != nil {
				fmt.Printf("Create Pod Failed : %s\n", err.Error())
				return
			}
		} else {
			_, err = i.dfsAPI.OpenPod(utils.PodName, i.password, i.sessionID)
			if err != nil {
				fmt.Printf("Open pod Failed : %s\n", err.Error())
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
			fmt.Printf("Failed to create doc table : %s\n", err.Error())
			return
		}
		err = i.dfsAPI.DocOpen(i.sessionID, utils.PodName, utils.PasswordsTable)
		if err != nil {
			fmt.Printf("Failed to open doc table : %s\n", err.Error())
			return
		}
		notesIndexes := make(map[string]collection.IndexType)
		notesIndexes["title"] = collection.StringIndex
		notesIndexes["starred"] = collection.StringIndex
		err = i.dfsAPI.DocCreate(i.sessionID, utils.PodName, utils.NotesTable, notesIndexes, true)
		if err != nil && err != collection.ErrDocumentDBAlreadyPresent {
			fmt.Printf("Failed to create doc table : %s\n", err.Error())
			return
		}
		err = i.dfsAPI.DocOpen(i.sessionID, utils.PodName, utils.NotesTable)
		if err != nil {
			fmt.Printf("Failed to open doc table : %s\n", err.Error())
			return
		}

		main := newMainView(i)
		i.setContent(main.view)
	})

	addButton := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		i.setContent(i.initSignupView(true))
	})
	bottomContent := container.NewPadded(container.New(layout.NewHBoxLayout(), layout.NewSpacer(), addButton))
	return container.NewBorder(topContent, bottomContent, nil, nil, container.NewCenter(container.NewVBox(t, usernameInput, passwordInput, container.NewGridWithColumns(3, layout.NewSpacer(), layout.NewSpacer(), loginBtn))))
}

func (i *index) initSignupView(allowBack bool) fyne.CanvasObject {
	tabs := container.NewAppTabs(
		container.NewTabItem("Signup", i.signupTab(allowBack)),
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

	mnemonicEntry := widget.NewEntryWithData(binding.BindString(&user.Mnemonic))
	mnemonicEntry.SetPlaceHolder("Mnemonic")

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
		labelWithStyle("Mnemonic"),
		mnemonicEntry,
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
		i.progress = dialog.NewProgressInfinite("", "Creating User", i.Window) //lint:ignore SA1019 fyne-io/fyne/issues/2782
		i.progress.Show()
		// Do signup
		dialogTitle := "Caution !!"
		confirm := "Yes, I have kept is safe"
		cb := func(b bool) {
			i.Reload()
		}
		address, mnemonic, _, _, _, err := i.dfsAPI.CreateUserV2(user.Username, user.Password, user.Mnemonic, "")
		if err != nil {
			i.progress.Hide()
			if err != eth.ErrInsufficientBalance {
				fmt.Printf("Failed to create user : %s\n", err.Error())
				return
			}
			dialogTitle = "Insufficient balance !!"
			confirm = "Yes, I will transfer balance to this wallet"
			cb = func(b bool) {}
		}
		i.progress.Hide()
		fmt.Printf("%s\n%s\n", user.Username, address)
		d := dialog.NewCustomConfirm(dialogTitle, confirm, "Cancel", i.displayMnemonic(address, mnemonic), cb, i.Window)
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

	rpcEntry := widget.NewEntryWithData(binding.BindString(&i.config.RPC))
	rpcEntry.SetPlaceHolder("RPC Endpoint")
	rpcEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("please enter rpc endpoint")
		}
		return nil
	}

	form := container.NewVBox(
		labelWithStyle("Bee Endpoint"),
		beeEntry,
		labelWithStyle("Batch ID"),
		stampEntry,
		labelWithStyle("RPC Endpoint"),
		rpcEntry,
	)

	saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
		if i.config.BeeEndpoint == "" {
			return
		}
		if i.config.BatchId == "" {
			return
		}
		if i.config.RPC == "" {
			return
		}

		// Save config
		configBytes, err := json.Marshal(i.config)
		if err != nil {
			fmt.Println("config save failed")
			return
		}
		err = ioutil.WriteFile(filepath.Join(i.app.Storage().RootURI().Path(), config), configBytes, 0700)
		if err != nil {
			fmt.Println("config write failed ", err)
			return
		}
		i.dfsAPI = nil
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
