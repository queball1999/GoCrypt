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

	// Check if there are enough command-line arguments
	if len(flag.Args()) < 2 {
		fmt.Println("Usage: GoCrypt [encrypt|decrypt] [file1 file2 ...] [flags]")
		return
	}

	// Validate maximum layers limit
	if *layers > 200 {
		fmt.Println("Error: Maximum allowed encryption layers is 200.")
		return
	}

	// Get the command and files from the arguments
	command := strings.ToLower(flag.Args()[0])
	files := flag.Args()[1:]

	// Initialize the Fyne app only if necessary
	var application fyne.App
	if !*noUI {
		application = app.New()
	}

	// Check if all files exist
	if nonExistentFiles := checkFilesExist(files); len(nonExistentFiles) > 0 {
		handleFileNotExistError(application, nonExistentFiles, *noUI)
		return
	}

	// Handle the encryption or decryption command
	switch command {
	case "encrypt", "enc", "e":
		handleEncryption(application, files, *outputDir, *noUI, *layers)
	case "decrypt", "dec", "d":
		handleDecryption(application, files, *outputDir, *noUI, *layers)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Usage: GoCrypt [encrypt|decrypt] [file1 file2 ...] [flags]")
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
	errorMessage := fmt.Sprintf("Error: The following files do not exist:\n%s", strings.Join(files, "\n"))
	if noUI {
		fmt.Println(errorMessage)
	} else {
		ui.ShowErrorDialog(application, errorMessage)
	}
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

		decryptFiles(nil, files, []byte(password), layers, false)

	} else {
		ui.ShowPasswordPrompt(application, "decrypt", "chacha20poly1305", strings.Join(files, "\n"), func(password string, deleteAfter bool) {
			decryptFiles(application, files, []byte(password), layers, deleteAfter)
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

	for index, filePath := range files {
		wg.Add(1)
		go func(index int, filePath string) {
			defer wg.Done()
			performFileEncryption(index, filePath, key, layers, deleteAfter, len(files))
		}(index, filePath)
	}

	wg.Wait()
	fmt.Printf("All files encrypted in: %s\n", time.Since(startTime))
}

// performFileEncryption handles encryption of a single file and reports the status.
func performFileEncryption(index int, filePath string, key []byte, layers int, deleteAfter bool, fileLength int) {
	startTime := time.Now()
	isDir := false // Track if the file is a directory

	// Skip already encrypted files
	if strings.HasSuffix(filePath, ".enc") {
		fmt.Printf("File %s is already encrypted. Skipping...\n", filePath)
		return
	}

	// Check if the file is protected
	if fileutils.IsFileProtected(filePath) {
		fmt.Printf("Skipping protected file: %s\n", filePath)
		return
	}

	// If it's a directory, compress it first
	if fileutils.IsDirectory(filePath) {
		isDir = true
		zipPath := filePath + ".zip"
		if err := fileutils.CompressFolder(filePath, zipPath); err != nil {
			fmt.Printf("Error compressing folder: %v\n", err)
			return
		}
		filePath = zipPath
	}

	// Open the input file for encryption
	inputFile, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening input file: %v\n", err)
		return
	}
	defer inputFile.Close()

	// Perform encryption
	outputPath := filePath + ".enc"
	if err := encryption.LayeredEncryptFile(inputFile, outputPath, string(key), layers); err != nil {
		fmt.Printf("Error encrypting file: %v\n", err)
		return
	}

	// Optionally delete the original file
	if deleteAfter || isDir {
		if err := fileutils.DeleteFile(filePath); err != nil {
			fmt.Printf("Error deleting file: %v\n", err)
		}
	}

	fmt.Printf("File %d / %d encrypted successfully in %s\n", index+1, fileLength, time.Since(startTime))
}

// decryptFiles performs the decryption on the provided files using the specified password.
func decryptFiles(application fyne.App, files []string, key []byte, layers int, deleteAfter bool) {
	var wg sync.WaitGroup
	startTime := time.Now()

	for _, filePath := range files {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()
			performFileDecryption(filePath, key, deleteAfter)
		}(filePath)
	}

	wg.Wait()
	fmt.Printf("All files decrypted in: %s\n", time.Since(startTime))
}

// performFileDecryption handles decryption of a single file and reports the status.
func performFileDecryption(filePath string, key []byte, deleteAfter bool) {
	// Skip files that are not encrypted
	if !strings.HasSuffix(filePath, ".enc") {
		fmt.Printf("File %s is not encrypted. Skipping...\n", filePath)
		return
	}

	// Open the input file for decryption
	inputFile, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening input file: %v\n", err)
		return
	}
	defer inputFile.Close()

	// Perform decryption
	outputPath := strings.TrimSuffix(filePath, ".enc")
	if err := encryption.LayeredDecryptFile(inputFile, outputPath, string(key)); err != nil {
		fmt.Printf("Decryption failed: %v\n", err)
		return
	}

	// Optionally delete the encrypted file
	if deleteAfter {
		if err := fileutils.DeleteFile(filePath); err != nil {
			fmt.Printf("Error deleting file: %v\n", err)
		}
	}

	fmt.Printf("File decrypted successfully to %s\n", outputPath)
}
