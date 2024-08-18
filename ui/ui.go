package ui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var (
	iconCache fyne.Resource
	iconOnce  sync.Once
)

func loadIcon() (fyne.Resource, error) {
	var err error
	iconOnce.Do(func() {
		exePath, err := os.Executable()
		if err != nil {
			return
		}
		exeDir := filepath.Dir(exePath)
		iconPath := filepath.Join(exeDir, "images", "GoCrypt.png")

		data, err := os.ReadFile(iconPath)
		if err != nil {
			return
		}
		iconCache = fyne.NewStaticResource(filepath.Base(iconPath), data)
	})
	return iconCache, err
}

func ShowPasswordPrompt(application fyne.App, action, method, filePath string, onPasswordEntered func(password string, deleteAfter bool)) {
	icon, err := loadIcon()
	if err != nil {
		fmt.Println(err)
		return
	}

	window := application.NewWindow("GoCrypt - Enter Password")
	window.SetIcon(icon)
	window.Resize(fyne.NewSize(450, 250))
	window.CenterOnScreen()

	passwordEntry := widget.NewPasswordEntry()
	fileLabel := widget.NewLabelWithStyle("File(s): "+filePath, fyne.TextAlignLeading, fyne.TextStyle{Italic: true})

	formItems := []*widget.FormItem{widget.NewFormItem("Password", passwordEntry)}

	var deleteFileAfterEncrypt *widget.Check

	if action == "encrypt" {
		confirmPasswordEntry := widget.NewPasswordEntry()
		formItems = append(formItems, widget.NewFormItem("Confirm", confirmPasswordEntry))
	}

	deleteFileAfterEncrypt = widget.NewCheck("Delete original file after encryption", nil)
	formItems = append(formItems, widget.NewFormItem("", deleteFileAfterEncrypt))
	
	form := widget.NewForm(formItems...)
	form.OnSubmit = func() {
		password := passwordEntry.Text
		if action == "encrypt" && password == "" {
			dialog.ShowError(errors.New("password cannot be empty"), window)
			return
		}
		if action == "encrypt" && password != formItems[1].Widget.(*widget.Entry).Text {
			dialog.ShowError(errors.New("passwords do not match, please try again"), window)
			return
		}
		onPasswordEntered(password, deleteFileAfterEncrypt.Checked)
		window.Close()
	}

	form.OnCancel = func() { window.Close() }
	form.SubmitText = "OK"
	form.CancelText = "Cancel"

	content := container.NewVBox(
		widget.NewLabelWithStyle("Enter password to "+action+" your item.", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		fileLabel,
		form,
	)
	window.SetContent(container.NewPadded(content))
	window.ShowAndRun()
}

func ShowErrorDialog(application fyne.App, message string) {
	icon, err := loadIcon()
	if err != nil {
		fmt.Println(err)
		return
	}

	window := application.NewWindow("Error")
	window.Resize(fyne.NewSize(450, 215))
	window.CenterOnScreen()
	window.SetIcon(icon)

	infoDialog := dialog.NewInformation("Error", message, window)
	infoDialog.SetOnClosed(func() { window.Close() })
	infoDialog.Show()
	window.ShowAndRun()
}
