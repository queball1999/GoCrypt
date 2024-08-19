package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
	"syscall"

	"GoCrypt/encryption"
	"GoCrypt/ui"
	"GoCrypt/fileutils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

// FIXME: - Associate .enc files with application (done through inno setup)
//		  - Optimize RAM usage (I see consistent 80-90MB usage. Can we condense?)
//		  - Need to modify the user to explicitely state wether they are encrypting files or folders
//		  - Layered encryption should be a while loop that detects layers. maybe we need layer count in header?

func main() {
	// Define flags
	outputDir := flag.String("output", "", "Specify the output directory")
	flag.StringVar(outputDir, "o", "", "Specify the output directory") // Alias -o for --output

	noUI := flag.Bool("no-ui", false, "Disable the GUI")
	flag.BoolVar(noUI, "n", false, "Disable the GUI") // Alias -n for --no-ui

	layers := flag.Int("layers", 5, "Layers of encryption")
	flag.IntVar(layers, "l", 5, "Layers of encryption") // Alias -l for --layers

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
		handleDecryption(application, files, *outputDir, *noUI, *layers)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Usage: GoCrypt [encrypt|decrypt] [file1 file2 ...] [flags]")
		return
	}
}

func handleEncryption(application fyne.App, files []string, outputDir string, noUI bool, layers int) {
	if noUI {
		// Handle encryption without UI
		fmt.Printf("Enter password for encryption: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println() // For a newline after password input
		if err != nil {
			fmt.Printf("Error reading password: %v\n", err)
			return
		}
		password := string(passwordBytes)
		encryptFile(nil, files, []byte(password), layers, false)

	} else {
		// Handle encryption with UI
		ui.ShowPasswordPrompt(application, "encrypt", "chacha20poly1305", strings.Join(files, "\n"), func(password string, deleteAfter bool) {
			encryptFile(application, files, []byte(password), layers, deleteAfter)
		})
	}
}

func handleDecryption(application fyne.App, files []string, outputDir string, noUI bool, layers int) {
	if noUI {
		// Handle decryption without UI
		fmt.Printf("Enter password for decryption: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println() // For a newline after password input
		if err != nil {
			fmt.Printf("Error reading password: %v\n", err)
			return
		}
		password := string(passwordBytes)
		decryptFile(nil, files, []byte(password), layers, false)

	} else {
		// Handle decryption with UI
		ui.ShowPasswordPrompt(application, "decrypt", "chacha20poly1305", strings.Join(files, "\n"), func(password string, deleteAfter bool) {
			decryptFile(application, files, []byte(password), layers, deleteAfter)
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
			isDir := false	// bool to track if driectory

			// Check if the file is already encrypted
			if strings.HasSuffix(filePath, ".enc") {
				fmt.Printf("File %s is already encrypted. Skipping...\n", filePath)
				return
			}

			// Implement failsafe
			if fileutils.IsFileProtected(filePath) {
				fmt.Printf("Skipping protected file or directory: %s\n", filePath)
				return
			}

			// If it's a directory, compress it first
			if fileutils.IsDirectory(filePath) {
				isDir = true
				zipPath := filePath + ".zip"
				err := fileutils.CompressFolder(filePath, zipPath)
				if err != nil {
					fmt.Printf("Error compressing folder: %v\n", err)
					return
				}
				filePath = zipPath
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
			
			err = encryption.EncryptFile(inputFile, outputPath, string(key))
			//err = encryption.LayeredEncryptFile(inputFile, outputPath, string(key), layers)
			if err != nil {
				fmt.Printf("Error encrypting file: %v\n", err)
				return
			}

			if deleteAfter || isDir {
				err := fileutils.DeleteFile(filePath)
				if err != nil {
					fmt.Printf("%v\n", err)
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

func decryptFile(application fyne.App, files []string, key []byte, layers int, deleteAfter bool) {
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

			//err = encryption.LayeredDecryptFile(inputFile, outputPath, string(key), layers)
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
				err := fileutils.DeleteFile(filePath)
				if err != nil {
					fmt.Printf("%v\n", err)
				}
			}

			fmt.Printf("File decrypted successfully to %s\n", outputPath)
		}(filePath)
	}

	wg.Wait()
	fmt.Printf("All files decrypted in: %s\n", time.Since(startTime))
}