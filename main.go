package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"
	"strings"
	"sync"

	"GoCrypt/encryption"
	"GoCrypt/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
)

const defaultLayers = 10

// FIXME: - Implement a failsafe to prevent the application from encrypting its own files.
//		  - Associate .enc files with application (done through inno setup)
//		  - Optimize RAM usage (this will be done with chunks but initial testing was not ideal)
//		  - Need to implement the following features of Eddy
//              - Should be capable of generating passphrase if none is provided. Need to ensure the user writes down, prints, or saves to file.
//				- Should use regex to ensure password is safe. If unsafe, give warning to user.
//				- What is keyed BLAKE2b for data authentication?
//		  - We need to change the behavior for the folder context menu and support "Compress and Encrypt", which compresses folder into .zip and encrypts the zip.

func main() {
	// Define flags
	outputDir := flag.String("output", "", "Specify the output directory")
	flag.StringVar(outputDir, "o", "", "Specify the output directory")	// Alias -o for --output

	noUI := flag.Bool("no-ui", false, "Disable the GUI")
	flag.BoolVar(noUI, "n", false, "Disable the GUI")	// Alias -n for --no-ui

	method := flag.String("method", "chacha20poly1305", "Specify the encryption method (chacha20poly1305 or aes)")
	flag.StringVar(method, "m", "chacha20poly1305", "Specify the encryption method (chacha20poly1305 or aes)")	// Alias -m for --method

	// Parse the command-line flags
	flag.Parse()

	// Determine the command (encrypt, decrypt)
	if len(flag.Args()) < 2 {
		fmt.Println("Usage: GoCrypt [encrypt|decrypt] [file1 file2 ...] [flags]")
		return
	}

	command := strings.ToLower(flag.Args()[0])
	files := flag.Args()[1:]

	// FIXME: Need to see if the file actually exists before bring up password prompt. Should display an error if file does not exist.
	// Check if the files exist
	var nonExistentFiles []string
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			nonExistentFiles = append(nonExistentFiles, file)
		}
	}

	if len(nonExistentFiles) > 0 {
		handleFileNotExistError(nonExistentFiles, *noUI)
		return
	}

	// Determine the action based on the command
	switch command {
	case "encrypt", "enc", "e":
		handleEncryption(files, *outputDir, *method, *noUI)
	case "decrypt", "dec", "d":
		handleDecryption(files, *outputDir, *method, *noUI)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Usage: GoCrypt [encrypt|decrypt] [file1 file2 ...] [flags]")
		return
	}
}

func selectFile(application fyne.App) string {
	var filePath string
	// Implement file selection logic
	return filePath
}

func handleFileNotExistError(files []string, noUI bool) {
	errorMessage := fmt.Sprintf("Error: The following files do not exist:\n%s", strings.Join(files, "\n"))
	if noUI {
		fmt.Println(errorMessage)
	} else {
		ui.ShowErrorDialog(errorMessage)
	}
}

func handleEncryption(files []string, outputDir, method string, noUI bool) {
	fmt.Println("Handling Encryption: ", noUI)
	if noUI {
		// Handle encryption without UI
		fmt.Printf("Encrypting %s with method %s...\n", files, method)
		// Implement encryption logic here
		// Here is where we would prompt user for password and convert to key
		
	} else {
		// Handle encryption with UI
		app := app.New()
		ui.ShowPasswordPrompt(app, "encrypt", method, strings.Join(files, "\n"), func(password string, deleteAfter bool) {
			encryptFile(app, files, []byte(password), method, deleteAfter)
		})
	}
}

func handleDecryption(files []string, outputDir, method string, noUI bool) {
	if noUI {
		// Handle decryption without UI
		fmt.Printf("Decrypting %s with method %s...\n", files, method)
		// Implement encryption logic here
		// Here is where we would prompt user for password and convert to key

	} else {
		// Handle decryption with UI
		app := app.New()
		ui.ShowPasswordPrompt(app, "decrypt", method, strings.Join(files, "\n"), func(password string, deleteAfter bool) {
			decryptFile(app, files, []byte(password), method, deleteAfter)
		})
	}
}

func encryptFile(application fyne.App, files []string, key []byte, method string, deleteAfter bool) {
	var wg sync.WaitGroup
	
	for _, filePath := range files {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()

			// Check if the file is already encrypted
			if strings.HasSuffix(filePath, ".enc") {
				fmt.Printf("File %s is already encrypted. Skipping...\n", filePath)
				return
			}

			// Generate a unique salt for each file
			salt, err := encryption.GenerateSalt()
			if err != nil {
				fmt.Printf("Error generating salt for %s: %v\n", filePath, err)
				return
			}
			
			// Derive the encryption key using the password and unique salt
			derivedKey := encryption.DeriveKey(string(key), salt)

			// Ensure key length is 32 bytes
			if len(derivedKey) != 32 {
				fmt.Println("Error: Derived key length is not 32 bytes")
				return
			}

			// Try to read input file
			inputData, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Printf("Error reading input file: %v\n", err)
				return
			}

			// Render UI menu if no-ui is false
			progressBar, win := ui.ShowProgressBar(application, "GoCrypt - Encryption Progress", defaultLayers)

			// Layer the encryption
			for i := 0; i < defaultLayers; i++ {
				inputData, err = encryption.EncryptData(inputData, derivedKey, method)
				if err != nil {
					fmt.Printf("Error during encryption: %v\n", err)
					win.Close()
					return
				}
				progressBar.SetValue(float64(i + 1))
				time.Sleep(10 * time.Millisecond) //needed so the UI can catch up
			}

			outputPath := filePath + ".enc"
			err = os.WriteFile(outputPath, append(salt, inputData...), 0644)
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

		}(filePath)
	}

	wg.Wait()
	fmt.Println("All files have been encrypted.")
}

func decryptFile(application fyne.App, files []string, key []byte, method string, deleteAfter bool) {
	var wg sync.WaitGroup
	
	for _, filePath := range files {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()

			// Check if the file is already encrypted
			if !strings.HasSuffix(filePath, ".enc") {
				fmt.Printf("File %s is not encrypted. Skipping...\n", filePath)
				return
			}

			// Try to read input file
			inputData, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Printf("Error reading input file: %v\n", err)
				return
			}
			
			// Extract the salt from the beginning of the file
			salt := inputData[:16] // Assuming the salt is 16 bytes
			ciphertext := inputData[16:]

			// Derive the encryption key using the password and unique salt
			derivedKey := encryption.DeriveKey(string(key), salt)
			
			// Ensure key length is 32 bytes
			if len(derivedKey) != 32 {
				fmt.Println("Error: Derived key length is not 32 bytes")
				return
			}

			// Render UI menu if no-ui is false
			progressBar, win := ui.ShowProgressBar(application, "GoCrypt - Encryption Progress", defaultLayers)

			// Layer the encryption
			for i := 0; i < defaultLayers; i++ {
				ciphertext, err = encryption.DecryptData(ciphertext, derivedKey, method)
				if err != nil {
					dialog.ShowError(errors.New("decryption failed: wrong password or corrupted data"), win)
					return
				}
				progressBar.SetValue(float64(i + 1))
				time.Sleep(5 * time.Millisecond) //needed so the UI can catch up
			}

			outputPath := filePath[:len(filePath)-4] // Remove .enc extension
			err = os.WriteFile(outputPath, ciphertext, 0644)
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

		}(filePath)
	}

	wg.Wait()
	fmt.Println("All files have been encrypted.")
}

/*
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
*/

/*
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

*/
