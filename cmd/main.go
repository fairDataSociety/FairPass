package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/fairdatasociety/fairpass/internal/screens"
	"github.com/fairdatasociety/fairpass/internal/utils"
)

func main() {
	a := app.NewWithID(utils.PackageName)
	w := a.NewWindow(utils.AppName)
	w.SetMaster()

	w.Resize(fyne.NewSize(800, 600))

	w.SetContent(screens.Make(a, w))
	w.ShowAndRun()
}
