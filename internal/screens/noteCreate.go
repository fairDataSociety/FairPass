package screens

import (
	"encoding/json"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/google/uuid"
	"github.com/onepeerlabs/fairpass/internal/utils"
)

func (main *mainView) makeAddNoteView(i *note) fyne.CanvasObject {
	if i == nil {
		i = &note{}
	}
	titleEntry := widget.NewEntryWithData(binding.BindString(&i.Title))
	titleEntry.SetPlaceHolder("Title")
	titleEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("please enter a title")
		}
		return nil
	}

	noteEntry := widget.NewEntryWithData(binding.BindString(&i.Message))
	noteEntry.SetPlaceHolder("Message")
	noteEntry.MultiLine = true
	noteEntry.Validator = nil

	if i.Message != "" {
		var err error
		i.Message, err = main.index.encryptor.DecryptContent(main.index.password, i.Message)
		if err != nil {
			fmt.Println("failed to decrypt note ", err)
		}
	}

	if i.ID != "" {
		titleEntry.Disable()
		noteEntry.Disable()
	}

	form := container.NewVBox(
		labelWithStyle("Title"),
		titleEntry,
		labelWithStyle("Message"),
		noteEntry,
	)

	cancelBtn := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		notesView := newNotesListView(main)
		main.setContent(notesView.view)
	})

	var top fyne.CanvasObject
	if i.ID == "" {
		saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
			main.index.progress = dialog.NewProgressInfinite("", "Saving Note", main.index)
			main.index.progress.Show()
			defer main.index.progress.Hide()
			if i.Title == "" {
				return
			}
			if i.Message == "" {
				return
			}

			i.ID = uuid.New().String()

			var err error
			i.Message, err = main.index.encryptor.EncryptContent(main.index.password, i.Message)
			if err != nil {
				fmt.Println("failed to save note :", err.Error())
				return
			}
			i.CreatedAt = time.Now().Unix()
			i.UpdatedAt = time.Now().Unix()
			i.IsStarred = "false"
			data, err := json.Marshal(i)
			if err != nil {
				fmt.Println("failed to save note :", err.Error())
				return
			}
			err = main.index.dfsAPI.DocPut(main.index.sessionID, utils.PodName, utils.NotesTable, data)
			if err != nil {
				fmt.Println("failed to save note :", err.Error())
				return
			}
			notesView := newNotesListView(main)
			main.setContent(notesView.view)
		})
		saveBtn.Importance = widget.HighImportance
		top = container.NewBorder(nil, nil, cancelBtn, saveBtn, widget.NewLabel(""))
	} else {
		top = container.NewBorder(nil, nil, cancelBtn, nil, widget.NewLabel(""))
	}

	return container.NewPadded(container.NewBorder(top, nil, nil, nil, form))
}
