package screens

import (
	"context"
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
		// testnet config
		//ensConfig, _ := contracts.TestnetConfig(contracts.Sepolia)
		ensConfig, _ := contracts.PlayConfig()
		ensConfig.ProviderBackend = i.config.RPC
		api, err := dfs.NewDfsAPI(
			context.TODO(),
			i.config.BeeEndpoint,
			i.config.BatchId,
			ensConfig,
			nil,
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

	usernameInput.SetText("check")
	passwordInput.SetText("passwordpassword")

	loginBtn := widget.NewButton("Login", func() {
		i.progress = dialog.NewProgressInfinite("", "Login is progress", i) //lint:ignore SA1019 fyne-io/fyne/issues/2782
		i.progress.Show()
		defer i.progress.Hide()
		// Do login
		if usernameInput.Text == "" {
			dialog.NewError(utils.ErrBlankUsername, i.Window).Show()
			return
		}
		if passwordInput.Text == "" {
			dialog.NewError(utils.ErrBlankPassword, i.Window).Show()
			return
		}
		lr, err := i.dfsAPI.LoginUserV2(usernameInput.Text, passwordInput.Text, "")
		if err != nil {
			dialog.NewError(fmt.Errorf("login Failed : %s", err.Error()), i.Window).Show()
			return
		}

		ui := lr.UserInfo

		i.encryptor = &crypto.Encryptor{}
		i.sessionID = ui.GetSessionId()
		i.password = passwordInput.Text
		if !i.dfsAPI.IsPodExist(utils.PodName, i.sessionID) {
			_, err = i.dfsAPI.CreatePod(utils.PodName, i.sessionID)
			if err != nil {
				dialog.NewError(fmt.Errorf("create Pod Failed : %s", err.Error()), i.Window).Show()
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
			dialog.NewError(fmt.Errorf("failed to create doc table : %s", err.Error()), i.Window).Show()
			return
		}
		err = i.dfsAPI.DocOpen(i.sessionID, utils.PodName, utils.PasswordsTable)
		if err != nil {
			dialog.NewError(fmt.Errorf("failed to open doc table : %s", err.Error()), i.Window).Show()
			return
		}

		notesIndexes := make(map[string]collection.IndexType)
		notesIndexes["title"] = collection.StringIndex
		notesIndexes["starred"] = collection.StringIndex
		err = i.dfsAPI.DocCreate(i.sessionID, utils.PodName, utils.NotesTable, notesIndexes, true)
		if err != nil && err != collection.ErrDocumentDBAlreadyPresent {
			dialog.NewError(fmt.Errorf("failed to create doc table : %s", err.Error()), i.Window).Show()
			return
		}
		err = i.dfsAPI.DocOpen(i.sessionID, utils.PodName, utils.NotesTable)
		if err != nil {
			dialog.NewError(fmt.Errorf("failed to open doc table : %s", err.Error()), i.Window).Show()
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

	confirmPassword := ""
	confirmPasswordEntry := widget.NewPasswordEntry()
	confirmPasswordBind := binding.BindString(&confirmPassword)

	confirmPasswordEntry.Bind(confirmPasswordBind)
	confirmPasswordEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("please enter a password")
		}
		return nil
	}
	confirmPasswordEntry.SetPlaceHolder("Confirm Password")

	form := container.NewVBox(
		labelWithStyle("Username"),
		usernameEntry,
		labelWithStyle("Mnemonic"),
		mnemonicEntry,
		labelWithStyle("Password"),
		passwordEntry,
		labelWithStyle("Confirm Password"),
		confirmPasswordEntry,
	)

	saveBtn := widget.NewButtonWithIcon("Signup", theme.DocumentSaveIcon(), func() {
		if user.Username == "" {
			dialog.NewError(utils.ErrBlankUsername, i.Window).Show()
			return
		}
		if user.Password == "" {
			dialog.NewError(utils.ErrBlankPassword, i.Window).Show()
			return
		}
		if confirmPassword == "" {
			dialog.NewError(utils.ErrBlankConfirmPassword, i.Window).Show()
			return
		}
		if user.Password != confirmPassword {
			dialog.NewError(utils.ErrPasswordMismatch, i.Window).Show()
			return
		}
		i.progress = dialog.NewProgressInfinite("", "Creating User", i.Window) //lint:ignore SA1019 fyne-io/fyne/issues/2782
		i.progress.Show()
		// Do signup
		dialogTitle := "Caution !!"
		confirm := "I have kept it safe"
		headerText := "Please keep the following safe. If these are lost you cannot recover your data."
		cb := func(b bool) {
			i.Reload()
		}
		rr, err := i.dfsAPI.CreateUserV2(user.Username, user.Password, user.Mnemonic, "")
		if err != nil {
			i.progress.Hide()
			if err != eth.ErrInsufficientBalance {
				dialog.NewError(err, i.Window).Show()
				return
			}
			dialogTitle = "Insufficient balance !!"
			headerText = "Wallet needs to be funded before using. Fund the following wallet and try again with the following mnemonic. " + headerText
			cb = func(b bool) {}
		}
		i.progress.Hide()
		d := dialog.NewCustomConfirm(dialogTitle, confirm, "Cancel", i.displayMnemonic(rr.Address, rr.Mnemonic, headerText), cb, i.Window)
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

func (i *index) displayMnemonic(address, mnemonic, headerText string) fyne.CanvasObject {
	header := labelWithStyle(headerText)
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
			return utils.ErrBlankBee
		}
		return nil
	}

	stampEntry := widget.NewEntryWithData(binding.BindString(&i.config.BatchId))
	stampEntry.SetPlaceHolder("Batch ID")
	stampEntry.Validator = func(s string) error {
		if s == "" {
			return utils.ErrBlankBatchId
		}
		return nil
	}

	rpcEntry := widget.NewEntryWithData(binding.BindString(&i.config.RPC))
	rpcEntry.SetPlaceHolder("RPC Endpoint")
	rpcEntry.Validator = func(s string) error {
		if s == "" {
			return utils.ErrBlankRPC
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
			dialog.NewError(utils.ErrBlankBee, i.Window).Show()
			return
		}
		if i.config.BatchId == "" {
			dialog.NewError(utils.ErrBlankBatchId, i.Window).Show()
			return
		}
		if i.config.RPC == "" {
			dialog.NewError(utils.ErrBlankRPC, i.Window).Show()
			return
		}

		// Save config
		configBytes, err := json.Marshal(i.config)
		if err != nil {
			dialog.NewError(fmt.Errorf("config write failed : %s", err.Error()), i.Window).Show()
			return
		}
		err = ioutil.WriteFile(filepath.Join(i.app.Storage().RootURI().Path(), config), configBytes, 0700)
		if err != nil {
			dialog.NewError(fmt.Errorf("config write failed : %s", err.Error()), i.Window).Show()
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
