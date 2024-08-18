package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"GoCrypt/encryption"
	"GoCrypt/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

// FIXME: - Implement a failsafe to prevent the application from encrypting its own files.
//		  - Associate .enc files with application (done through inno setup)
//		  - Optimize RAM usage (I see consistent 80-90MB usage. Can we condense?)
//		  - We need to change the behavior for the folder context menu and support "Compress and Encrypt", which compresses folder into .zip and encrypts the zip.

func main() {
	// Define flags
	outputDir := flag.String("output", "", "Specify the output directory")
	flag.StringVar(outputDir, "o", "", "Specify the output directory") // Alias -o for --output

	noUI := flag.Bool("no-ui", false, "Disable the GUI")
	flag.BoolVar(noUI, "n", false, "Disable the GUI") // Alias -n for --no-ui

	layers := flag.Int("layers", 1, "Layers of encryption")
	flag.IntVar(layers, "l", 1, "Layers of encryption") // Alias -l for --layers

	// Parse the command-line flags
	flag.Parse()

	// Determine the command (encrypt, decrypt)
	if len(flag.Args()) < 2 {
		fmt.Println("Usage: GoCrypt [encrypt|decrypt] [file1 file2 ...] [flags]")
		return
	}

	// Separate commands and files into corresponding variables
	command := strings.ToLower(flag.Args()[0])
	files := flag.Args()[1:]

	// Initialize the app only if necessary
	var application fyne.App
	if !*noUI {
		application = app.New()
	}

	// Check if the files exist
	var nonExistentFiles []string
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			nonExistentFiles = append(nonExistentFiles, file)
		}
	}

	if len(nonExistentFiles) > 0 {
		handleFileNotExistError(application, nonExistentFiles, *noUI)
		return
	}

	// Determine the action based on the command
	switch command {
	case "encrypt", "enc", "e":
		handleEncryption(application, files, *outputDir, *noUI, *layers)
	case "decrypt", "dec", "d":
		handleDecryption(application, files, *outputDir, *noUI)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Usage: GoCrypt [encrypt|decrypt] [file1 file2 ...] [flags]")
		return
	}
}

func handleEncryption(application fyne.App, files []string, outputDir string, noUI bool, layers int) {
	if noUI {
		// Handle encryption without UI
		fmt.Printf("Encrypting %s with ChaCha20-Poly1305...\n", files)
		// Implement encryption logic here
		for _, file := range files {
			fmt.Printf("Enter password for encryption: ")
			var password string
			fmt.Scanln(&password)
			encryptFile(nil, []string{file}, []byte(password), layers, false)
		}
	} else {
		// Handle encryption with UI
		ui.ShowPasswordPrompt(application, "encrypt", "chacha20poly1305", strings.Join(files, "\n"), func(password string, deleteAfter bool) {
			encryptFile(application, files, []byte(password), layers, deleteAfter)
		})
	}
}

func handleDecryption(application fyne.App, files []string, outputDir string, noUI bool) {
	if noUI {
		// Handle decryption without UI
		fmt.Printf("Decrypting %s with ChaCha20-Poly1305...\n", files)
		// Implement decryption logic here
		for _, file := range files {
			fmt.Printf("Enter password for decryption: ")
			var password string
			fmt.Scanln(&password)
			decryptFile(nil, []string{file}, []byte(password), false)
		}
	} else {
		// Handle decryption with UI
		ui.ShowPasswordPrompt(application, "decrypt", "chacha20poly1305", strings.Join(files, "\n"), func(password string, deleteAfter bool) {
			decryptFile(application, files, []byte(password), deleteAfter)
		})
	}
}

func handleFileNotExistError(application fyne.App, files []string, noUI bool) {
	errorMessage := fmt.Sprintf("Error: The following files do not exist:\n%s", strings.Join(files, "\n"))
	if noUI {
		fmt.Println(errorMessage)
	} else {
		ui.ShowErrorDialog(application, errorMessage)
	}
}

func encryptFile(application fyne.App, files []string, key []byte, layers int, deleteAfter bool) {
	var wg sync.WaitGroup
	startTime := time.Now() // Track the time for the entire encryption process

	for index, filePath := range files {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()
			fileDuration := time.Now()

			// Check if the file is already encrypted
			if strings.HasSuffix(filePath, ".enc") {
				fmt.Printf("File %s is already encrypted. Skipping...\n", filePath)
				return
			}

			inputFile, err := os.Open(filePath)
			if err != nil {
				fmt.Printf("Error opening input file: %v\n", err)
				return
			}
			defer inputFile.Close()

			// Render UI menu if no-ui is false
			//progressBar, _ := ui.ShowProgressBar(application, "GoCrypt - Encryption Progress", defaultLayers)
			//progressBar.SetValue(10)

			outputPath := filePath + ".enc"
			// FIXME: Handle layer input flag and passing to layered encryption.
			err = encryption.EncryptFile(inputFile, outputPath, string(key))
			//err = encryption.LayeredEncryptFile(inputFile, outputPath, string(key), layers)
			if err != nil {
				fmt.Printf("Error encrypting file: %v\n", err)
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

			
			//progressBar.SetValue(100)
			//win.Close()
			fmt.Printf("File %d / %d Encrypted successfully in %s\n", index+1, len(files), time.Since(fileDuration))

		}(filePath)
	}

	wg.Wait()
	fmt.Printf("All files encrypted in: %s\n", time.Since(startTime))
}

func decryptFile(application fyne.App, files []string, key []byte, deleteAfter bool) {
	var wg sync.WaitGroup
	startTime := time.Now() // Track the time for the entire decryption process

	for _, filePath := range files {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()

			// Check if the file is already encrypted
			if !strings.HasSuffix(filePath, ".enc") {
				fmt.Printf("File %s is not encrypted. Skipping...\n", filePath)
				return
			}

			inputFile, err := os.Open(filePath)
			if err != nil {
				fmt.Printf("Error opening input file: %v\n", err)
				return
			}
			defer inputFile.Close()

			outputPath := filePath[:len(filePath)-4] // Remove .enc extension
			err = encryption.DecryptFile(inputFile, outputPath, string(key))
			if err != nil {
				fmt.Printf("Decryption failed: %v\n", err)
				/*
				if application != nil {
					//dialog.ShowError(errors.New("decryption failed: wrong password or corrupted data"), nil)
					fmt.Printf("Decryption failed: %v\n", err)
				} else {
					fmt.Printf("Decryption failed: %v\n", err)
				}
				*/
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
		}(filePath)
	}

	wg.Wait()
	fmt.Printf("All files decrypted in: %s\n", time.Since(startTime))
}
