package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"syscall"

	"golang.org/x/term"

	"GoCrypt/encryption"
	"GoCrypt/fileutils"
	"GoCrypt/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

// Main function initializes flags, processes inputs, and handles encryption/decryption based on commands.
func main() {
	// Define and parse command-line flags
	outputDir, noUI, layers := setupFlags()

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
	if nonExistentFiles := checkFilesExist(files); len(nonExistentFiles) > 0 {
		handleFileNotExistError(application, nonExistentFiles, *noUI)
		return
	}

	// Function to detect file types and automatically route user
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

// setupFlags initializes the command-line flags and returns pointers to the flag variables.
func setupFlags() (*string, *bool, *int) {
	outputDir := flag.String("output", "", "Specify the output directory")
	flag.StringVar(outputDir, "o", "", "Specify the output directory (alias: -o)")

	noUI := flag.Bool("no-ui", false, "Disable the GUI")
	flag.BoolVar(noUI, "n", false, "Disable the GUI (alias: -n)")

	layers := flag.Int("layers", 5, "Layers of encryption")
	flag.IntVar(layers, "l", 5, "Layers of encryption (alias: -l)")

	flag.Parse()

	return outputDir, noUI, layers
}

// checkFilesExist verifies whether the provided files exist. Returns a slice of non-existent files.
func checkFilesExist(files []string) []string {
	var nonExistentFiles []string
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			nonExistentFiles = append(nonExistentFiles, file)
		}
	}
	return nonExistentFiles
}

// handleFileNotExistError displays an error message for non-existent files based on the UI mode.
func handleFileNotExistError(application fyne.App, files []string, noUI bool) {
	errorMessage := fmt.Sprintf("the following files do not exist:\n%s", strings.Join(files, "\n"))
	handleError(application, fmt.Errorf("%s", errorMessage), noUI)
}

// handleEncryption manages encryption logic based on whether the UI is enabled or not.
func handleEncryption(application fyne.App, files []string, outputDir string, noUI bool, layers int) {
	if noUI {
		password, err := promptPasswordCLI()
		if err != nil {
			fmt.Printf("Error reading password: %v\n", err)
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
		password, err := promptPasswordCLI()
		if err != nil {
			fmt.Printf("Error reading password: %v\n", err)
			return
		}

		decryptFiles(nil, files, []byte(password), layers, false, noUI)

	} else {
		ui.ShowPasswordPrompt(application, "decrypt", "chacha20poly1305", strings.Join(files, "\n"), func(password string, deleteAfter bool) {
			decryptFiles(application, files, []byte(password), layers, deleteAfter, noUI)
		})
	}
}

// promptPasswordCLI handles secure password input for CLI mode.
func promptPasswordCLI() (string, error) {
	// Read password from the terminal securely
	fmt.Printf("Enter password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Newline after password input
	if err != nil {
		return "", err
	}

	// Ask for password confirmation
	fmt.Printf("Confirm password: ")
	confirmPasswordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}

	password := string(passwordBytes)
	confirmPassword := string(confirmPasswordBytes)

	if password != confirmPassword {
		return "", fmt.Errorf("passwords do not match")
	}

	return password, nil
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
			}
		}(index, filePath)
	}

	wg.Wait()
	if success {
		fmt.Printf("all files encrypted successfully in: %s", time.Since(startTime))
	}
}

// performFileEncryption handles encryption of a single file and reports the status.
func performFileEncryption(index int, filePath string, key []byte, layers int, deleteAfter bool, fileLength int) error {
	startTime := time.Now()
	isDir := false // Track if the file is a directory

	// Skip already encrypted files
	//FIXME: update with IsFileEncrypted function in fileutils
	if strings.HasSuffix(filePath, ".enc") {
		return fmt.Errorf("file %s is already encrypted. Skipping... ", filePath)
	}

	// Check if the file is protected
	if fileutils.IsFileProtected(filePath) {
		return fmt.Errorf("skipping protected file: %s", filePath)
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

	// Optionally delete the original file
	if deleteAfter || isDir {
		if err := fileutils.DeleteFile(filePath); err != nil {
			return fmt.Errorf("error deleting file: %v", err)
		}
	}

	fmt.Printf("File %d / %d encrypted successfully in %s\n", index+1, fileLength, time.Since(startTime))
	return nil
}

// decryptFiles performs the decryption on the provided files using the specified password.
func decryptFiles(application fyne.App, files []string, key []byte, layers int, deleteAfter bool, noUI bool) {
	var wg sync.WaitGroup
	startTime := time.Now()
	success := true

	for _, filePath := range files {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()
			err := performFileDecryption(filePath, key, deleteAfter)
			if (err != nil) {
				success = false
				//handleError(application, err, noUI)
				fmt.Println(err)
			}
		}(filePath)
	}

	wg.Wait()
	if success {
		fmt.Printf("all files decrypted successfully in: %s\n", time.Since(startTime))
	}
}

// performFileDecryption handles decryption of a single file and reports the status.
func performFileDecryption(filePath string, key []byte, deleteAfter bool) error{
	// Skip files that are not encrypted
	if !strings.HasSuffix(filePath, ".enc") {
		return fmt.Errorf("file %s is not encrypted. Skipping... ", filePath)
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

	// Optionally delete the encrypted file
	if deleteAfter {
		if err := fileutils.DeleteFile(filePath); err != nil {
			return fmt.Errorf("error deleting file: %v", err)
		}
	}

	return nil
}

func handleError(application fyne.App, err error, noUI bool) {
    if noUI {
        fmt.Printf("Error: %v\n", err)
    } else {
        ui.ShowErrorDialog(application, err.Error())
    }
}
