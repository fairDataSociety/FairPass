package screens

import (
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"github.com/onepeerlabs/fairpass/internal/utils"
	"github.com/onepeerlabs/fairpass/internal/utils/crypto"
)

const (
	config = ".fairpass.conf"
)

type index struct {
	fyne.Window

	app      fyne.App
	view     *fyne.Container
	progress dialog.Dialog

	config    *fairOSConfig
	dfsAPI    *dfs.DfsAPI
	dataDir   string
	sessionID string
	encryptor *crypto.Encryptor
}

func Make(a fyne.App, w fyne.Window) fyne.CanvasObject {
	installationLocation := a.Storage().RootURI().Path()
	dataDir := filepath.Join(installationLocation, utils.DataDirName)
	i := &index{
		Window:  w,
		app:     a,
		dataDir: dataDir,
		config:  &fairOSConfig{},
	}
	i.view = container.NewMax(i.initLoginView())
	return i.view
}

func (i *index) Reload() {
	i.view.Objects[0] = i.initLoginView()
}
