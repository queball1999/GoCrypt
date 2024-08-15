// go: generate goversioninfo -icon = gocrypt.ico
package main

import (
	"flag"
	"fmt"
	"os"
	"time"
	"errors"

	"GoCrypt/encryption"
	"GoCrypt/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
)

const defaultLayers = 5

// FIXME: - Implement a failsafe to prevent the application from encrypting its own files.
//        - The application should automatically detect file types. If the file has a .enc extension, it should only attempt decryption. For any other file type, suggest encryption.

func main() {
	// Define command-line flags
	inputFile := flag.String("input", "", "Input file to be encrypted or decrypted")
	action := flag.String("action", "encrypt", "Action to perform: encrypt or decrypt")
	method := flag.String("method", "chacha20poly1305", "Encryption method: chacha20poly1305, aes")
	flag.Parse()

	application := app.New()

	// Check if an input file is provided via the flag
	var filePath string
	if *inputFile != "" {
		filePath = *inputFile
	} else {
		// Show file selection dialog if no input file is provided
		filePath = selectFile(application)
		if filePath == "" {
			return
		}
	}

	ui.ShowPasswordPrompt(application, *action, *method, filePath, func(password string, deleteAfter bool) {
		key := encryption.DeriveKey(password)

		if *action == "encrypt" {
			encryptFile(application, filePath, key, *method, deleteAfter)
		} else {
			decryptFile(application, filePath, key, *method, deleteAfter)
		}
	})
}

func selectFile(application fyne.App) string {
	var filePath string
	// Implement file selection logic
	return filePath
}

func encryptFile(application fyne.App, filePath string, key []byte, method string, deleteAfter bool) {
	inputData, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		return
	}

	progressBar, win := ui.ShowProgressBar(application, "GoCrypt - Encryption Progress", defaultLayers)

	for i := 0; i < defaultLayers; i++ {
		inputData, err = encryption.EncryptData(inputData, key, method)
		if err != nil {
			fmt.Printf("Error during encryption: %v\n", err)
			win.Close()
			return
		}
		progressBar.SetValue(float64(i + 1))
		time.Sleep(10 * time.Millisecond) //needed so the UI can catch up
	}
	
	outputPath := filePath + ".enc"
	err = os.WriteFile(outputPath, inputData, 0644)
	if err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		return
	}

	if deleteAfter {
		err := os.Remove(filePath)
		if err != nil {
			fmt.Printf("Failed to delete the original file: %v\n", err)
		} else {
			fmt.Printf("Original file %s deleted successfully\n", filePath)
		}
	}

	fmt.Printf("File encrypted successfully to %s\n", outputPath)
	win.Close()
}

func decryptFile(application fyne.App, filePath string, key []byte, method string, deleteAfter bool) {
	inputData, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		return
	}

	progressBar, win := ui.ShowProgressBar(application, "GoCrypt - Decryption Progress", defaultLayers)

	for i := 0; i < defaultLayers; i++ {
		inputData, err = encryption.DecryptData(inputData, key, method)
		if err != nil {
			dialog.ShowError(errors.New("decryption failed: wrong password or corrupted data"), win)
			return
		}
		progressBar.SetValue(float64(i + 1))
		time.Sleep(10 * time.Millisecond) //needed so the UI can catch up
	}
	
	outputPath := filePath[:len(filePath)-4] // Remove .enc extension
	err = os.WriteFile(outputPath, inputData, 0644)
	if err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		return
	}

	if deleteAfter {
		err := os.Remove(filePath)
		if err != nil {
			fmt.Printf("Failed to delete the original file: %v\n", err)
		} else {
			fmt.Printf("Original file %s deleted successfully\n", filePath)
		}
	}

	fmt.Printf("File decrypted successfully to %s\n", outputPath)
	win.Close()
}
