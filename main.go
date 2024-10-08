package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"GoCrypt/encryption"
	"GoCrypt/fileutils"
	"GoCrypt/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

var logger *log.Logger

// Main function initializes flags, processes inputs, and handles encryption/decryption based on commands.
func main() {
	// Initialize logger
	logger = fileutils.InitLogger()

	// Define and parse command-line flags
	outputDir, noUI, layers := ui.SetupFlags()

	// Initialize the Fyne app only if necessary
	var application fyne.App
	if !*noUI {
		application = app.New()
	}

	// Check if there are enough command-line arguments
	if len(flag.Args()) < 2 {
		handleError(application, fmt.Errorf("usage: gocrypt [encrypt|decrypt] [file1 file2 ...] [flags]"), *noUI)
		return
	}

	// Validate maximum layers limit
	if *layers > 200 {
		handleError(application, fmt.Errorf("maximum allowed encryption layers is 200"), *noUI)
		return
	}

	// Get the command and files from the arguments
	command := strings.ToLower(flag.Args()[0])
	files := flag.Args()[1:]

	// Check if all files exist
	if nonExistentFiles := fileutils.CheckFilesExist(files); len(nonExistentFiles) > 0 {
		errorMessage := fmt.Sprintf("the following files do not exist:\n%s", strings.Join(files, "\n"))
		handleError(application, fmt.Errorf("%s", errorMessage), *noUI)
		return
	}

	// Function to detect file types
	// WORK IN PROGRESS
	/*
	err := fileutils.CheckFileCommand(files, command)
	if err != nil {
		// Display error either in terminal or UI based on no-ui flag
		handleError(application, err, *noUI)
		return
	}
	*/
	
	// Handle the encryption or decryption command
	switch command {
	case "encrypt", "enc", "e":
		handleEncryption(application, files, *outputDir, *noUI, *layers)
	case "decrypt", "dec", "d":
		handleDecryption(application, files, *outputDir, *noUI, *layers)
	default:
		handleError(application, fmt.Errorf("unknown command: %s\nusage: GoCrypt [encrypt|decrypt] [file1 file2 ...] [flags]", command), *noUI)
	}
}

// handleEncryption manages encryption logic based on whether the UI is enabled or not.
func handleEncryption(application fyne.App, files []string, outputDir string, noUI bool, layers int) {
	if noUI {
		password, err := ui.PromptPasswordCLI()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		encryptFiles(nil, files, []byte(password), layers, false, noUI)

	} else {
		ui.ShowPasswordPrompt(application, "encrypt", "chacha20poly1305", strings.Join(files, "\n"), func(password string, deleteAfter bool) {
			encryptFiles(application, files, []byte(password), layers, deleteAfter, noUI)
		})
	}
}

// handleDecryption manages decryption logic based on whether the UI is enabled or not.
func handleDecryption(application fyne.App, files []string, outputDir string, noUI bool, layers int) {
	if noUI {
		password, err := ui.PromptPasswordCLI()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		decryptFiles(nil, files, []byte(password), layers, false, noUI)

	} else {
		ui.ShowPasswordPrompt(application, "decrypt", "chacha20poly1305", strings.Join(files, "\n"), func(password string, deleteAfter bool) {
			decryptFiles(application, files, []byte(password), layers, deleteAfter, noUI)
		})
	}
}

// encryptFiles performs the encryption on the provided files using the specified password and options.
func encryptFiles(application fyne.App, files []string, key []byte, layers int, deleteAfter bool, noUI bool) {
	var wg sync.WaitGroup
	startTime := time.Now() // Track the time for the entire encryption process
	success := true

	for index, filePath := range files {
		wg.Add(1)
		go func(index int, filePath string) {
			defer wg.Done()
			err := performFileEncryption(index, filePath, key, layers, deleteAfter, len(files))
			if (err != nil) {
				success = false
				//handleError(application, err, noUI)
				fmt.Println(err)
				logger.Println(err)
			}
		}(index, filePath)
	}

	wg.Wait()
	if success {
		fmt.Printf("All files encrypted successfully in: %s", time.Since(startTime))
		logger.Printf("All files encrypted successfully in: %s", time.Since(startTime))
	}
}

// performFileEncryption handles encryption of a single file and reports the status.
func performFileEncryption(index int, filePath string, key []byte, layers int, deleteAfter bool, fileLength int) error {
	startTime := time.Now()
	isDir := false // Track if the file is a directory

	// Skip already encrypted files
	//FIXME: update with IsFileEncrypted function in fileutils
	if strings.HasSuffix(filePath, ".enc") {
		logger.Printf("file %s is already encrypted. Skipping... ", filePath)
		return nil
	}

	// Check if the file is protected
	if fileutils.IsFileProtected(filePath) {
		logger.Printf("skipping protected file: %s", filePath)
		return nil
	}

	// If it's a directory, compress it first
	if fileutils.IsDirectory(filePath) {
		isDir = true
		zipPath := filePath + ".zip"
		if err := fileutils.CompressFolder(filePath, zipPath); err != nil {
			return fmt.Errorf("error compressing folder: %v", err)
		}
		filePath = zipPath
	}

	// Open the input file for encryption
	inputFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening input file: %v", err)
	}
	defer inputFile.Close()

	// Perform encryption
	outputPath := filePath + ".enc"
	if err := encryption.LayeredEncryptFile(inputFile, outputPath, string(key), layers); err != nil {
		return fmt.Errorf("error encrypting file: %v", err)
	}
	
	inputFile.Close()	// Ensure file is closed

	// Optionally delete the original file
	if deleteAfter || isDir {
		logger.Printf("Deleting the following item during encryption: %v", filePath)
		if err := fileutils.DeleteFile(filePath); err != nil {
			return fmt.Errorf("error deleting file: %v", err)
		}
	}
	
	logger.Printf("File %d / %d encrypted successfully in %s\n", index+1, fileLength, time.Since(startTime))
	return nil
}

// decryptFiles performs the decryption on the provided files using the specified password.
func decryptFiles(application fyne.App, files []string, key []byte, layers int, deleteAfter bool, noUI bool) {
	var wg sync.WaitGroup
	startTime := time.Now()
	success := true

	for index, filePath := range files {
		wg.Add(1)
		go func(index int, filePath string) {
			defer wg.Done()
			err := performFileDecryption(index, filePath, key, deleteAfter, len(files))
			if (err != nil) {
				success = false
				//handleError(application, err, noUI)
				fmt.Println(err)
				logger.Println(err)
			}
		}(index, filePath)
	}

	wg.Wait()
	if success {
		fmt.Printf("All files decrypted successfully in: %s", time.Since(startTime))
		logger.Printf("All files decrypted successfully in: %s", time.Since(startTime))
	}
}

// performFileDecryption handles decryption of a single file and reports the status.
func performFileDecryption(index int, filePath string, key []byte, deleteAfter bool, fileLength int) error{
	startTime := time.Now()
	
	// Skip files that are not encrypted
	// FIXME: Replace with IsFileEncrypted method
	if !strings.HasSuffix(filePath, ".enc") {
		logger.Printf("file %s is not encrypted. Skipping... ", filePath)
		return nil
	}

	// Open the input file for decryption
	inputFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening input file: %v", err)
	}
	defer inputFile.Close()

	// Perform decryption
	outputPath := strings.TrimSuffix(filePath, ".enc")
	err = encryption.LayeredDecryptFile(inputFile, outputPath, string(key))
	if err != nil {
		// Check if it's an incorrect password error
		if strings.Contains(err.Error(), "cipher: message authentication failed") {
			return fmt.Errorf("decryption failed: incorrect password")
		} else {
			return fmt.Errorf("decryption failed: %v", err)
		}
	}

	inputFile.Close()	// Ensure file is closed	

	// Optionally delete the encrypted file
	if deleteAfter {
		if err := fileutils.DeleteFile(filePath); err != nil {
			logger.Printf("Deleting the following item during decryption: %v", filePath)
			return fmt.Errorf("error deleting file: %v", err)
		}
	}
	logger.Printf("File %d / %d decrypted successfully in %s\n", index+1, fileLength, time.Since(startTime))
	return nil
}

func handleError(application fyne.App, err error, noUI bool) {
	// print error to log file regardless
	logger.Printf("Error: %v\n", err)

    if noUI {
        fmt.Printf("Error: %v\n", err)
    } else {
        ui.ShowErrorDialog(application, err.Error())
    }
}
