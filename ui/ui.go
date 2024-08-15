package ui

import (
	"fmt"
    "os"
	"errors"
	"path/filepath"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/container"
)

func ShowPasswordPrompt(application fyne.App, action, method string, filePath string, onPasswordEntered func(password string, deleteAfter bool)) {
	// Get the directory where the executable is located
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		return
	}
	exeDir := filepath.Dir(exePath)

	// Construct the path to the images folder
	imagesPath := filepath.Join(exeDir, "images")

	// Load the icon image
	iconPath := filepath.Join(imagesPath, "GoCrypt.png")
	icon, err := loadIcon(iconPath)
	if err != nil {
		fmt.Println("Error loading icon:", err)
	} else {
		window := application.NewWindow("GoCrypt - Enter Password")
		window.SetIcon(icon)
		window.Resize(fyne.NewSize(450, 250))

		passwordEntry := widget.NewPasswordEntry()

		label := "Enter password to " + action + " your item."
		fileLabel := widget.NewLabelWithStyle("File: "+filePath, fyne.TextAlignLeading, fyne.TextStyle{Italic: true})

		// Create form items
		formItems := []*widget.FormItem{
			widget.NewFormItem("Password", passwordEntry),
		}

		var deleteFileAfterEncrypt *widget.Check
		if action == "encrypt" {
			// Add confirm password field and delete file checkbox if action is encrypt
			confirmPasswordEntry := widget.NewPasswordEntry()
			confirmPasswordEntry.Resize(fyne.NewSize(200, 0))
			formItems = append(formItems, widget.NewFormItem("Confirm", confirmPasswordEntry))

			deleteFileAfterEncrypt = widget.NewCheck("Delete original file after encryption", nil)
			formItems = append(formItems, widget.NewFormItem("", deleteFileAfterEncrypt))

			// Modify the OnSubmit function to check for password match
			form := widget.NewForm(formItems...)
			form.OnSubmit = func() {
				password := passwordEntry.Text
				confirmPassword := confirmPasswordEntry.Text
				if password == "" || confirmPassword == "" {
					// Show error dialog if passwords are empty
					dialog.ShowError(errors.New("passwords cannot be empty"), window)
				} else if password != confirmPassword {
					// Show error dialog if passwords do not match
					dialog.ShowError(errors.New("passwords do not match, please try again"), window)
				} else {
					onPasswordEntered(password, deleteFileAfterEncrypt.Checked)
					window.Close()
					
				}
			}

			form.OnCancel = func() {
				window.Close()
			}

			form.SubmitText = "OK"
			form.CancelText = "Cancel"

			content := container.NewVBox(
				widget.NewLabelWithStyle(label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				fileLabel,
				form,
			)

			paddedContent := container.NewPadded(content)
			window.SetContent(paddedContent)
		} else {
			// For decrypt, simple form with only password
			deleteFileAfterEncrypt = widget.NewCheck("Delete original file after encryption", nil)
			formItems = append(formItems, widget.NewFormItem("", deleteFileAfterEncrypt))
			form := widget.NewForm(formItems...)
			form.OnSubmit = func() {
				password := passwordEntry.Text
				onPasswordEntered(password, deleteFileAfterEncrypt.Checked)
				window.Close()
			}

			form.OnCancel = func() {
				window.Close()
			}

			form.SubmitText = "OK"
			form.CancelText = "Cancel"

			content := container.NewVBox(
				widget.NewLabelWithStyle(label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				fileLabel,
				form,
			)

			paddedContent := container.NewPadded(content)
			window.SetContent(paddedContent)
		}

		window.ShowAndRun()
	}
}

func ShowProgressBar(application fyne.App, title string, max int) (*widget.ProgressBar, fyne.Window) {
    window := application.NewWindow(title)
	window.Resize(fyne.NewSize(450, 215))
	//icon := fyne.NewStaticResource("icon", loadIcon("images/GoCrypt.png"))
	//window.SetIcon(icon)

    progressBar := widget.NewProgressBar()
    progressBar.Max = float64(max)
	progressBar.Resize(fyne.NewSize(200, 10))

	content := container.NewVBox(
		widget.NewLabelWithStyle("Processing...", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), 
		progressBar,
	)

	paddedContent := container.NewPadded(content)
	window.SetContent(paddedContent)
    
    window.Show()
    return progressBar, window
}

// loadIcon is a helper function to load an image from the given path
func loadIcon(path string) (fyne.Resource, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return fyne.NewStaticResource(filepath.Base(path), data), nil
}