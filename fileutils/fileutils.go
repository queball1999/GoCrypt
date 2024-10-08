package fileutils

import (
	"archive/zip"
	"fmt"
	"log"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// initLogger initializes the logger
func InitLogger() *log.Logger {
    // Get the user's home directory
    homeDir, err := os.UserHomeDir()
    if err != nil {
        fmt.Printf("Failed to get user home directory: %v\n", err)
        return nil
    }

    // Determine the path to the Documents folder based on the OS
    var documentsDir string
    switch runtime.GOOS {
    case "windows":
        documentsDir = filepath.Join(homeDir, "Documents")
    case "darwin", "linux":
        documentsDir = filepath.Join(homeDir, "Documents") // macOS and Linux typically use ~/Documents
    default:
        fmt.Printf("Unsupported OS: %s\n", runtime.GOOS)
        return nil
    }

    // Create a GoCrypt folder in the Documents directory
    goCryptDir := filepath.Join(documentsDir, "GoCrypt")
    logFilePath := filepath.Join(goCryptDir, "app.log")

    // Ensure the GoCrypt log directory exists
    if err := os.MkdirAll(goCryptDir, 0755); err != nil {
        fmt.Printf("Failed to create log directory: %v\n", err)
        return nil
    }

    // Open the log file in append mode, create it if it doesn't exist
    logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Printf("Failed to open log file: %v\n", err)
        return nil
    }

    // Create a logger that writes to both the log file and stdout, with custom timestamp format
    return log.New(logFile, "GoCrypt: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
}

// CheckFilesExist verifies whether the provided files exist. Returns a slice of non-existent files.
func CheckFilesExist(files []string) []string {
	var nonExistentFiles []string
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			nonExistentFiles = append(nonExistentFiles, file)
		}
	}
	return nonExistentFiles
}

// DeleteFile deletes the specified file.
func DeleteFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("failed to delete the original file: %v", err)
	} 
	fmt.Printf("Original file %s deleted successfully\n", filePath)
	return nil
}

// IsFileProtected checks if the file should be skipped from encryption (e.g., GoCrypt files).
func IsFileProtected(filePath string) bool {
	// Normalize the file path for consistent comparison
	filePath = strings.ToLower(filePath)

	// Get the current working directory
	/*
	currentDir, err := os.Getwd()
	fmt.Println("Workign Directory: ", currentDir)
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return true
	}
	*/

	// Check if the file path is the GoCrypt executable itself
	if strings.Contains(filePath, "gocrypt.exe") {
		fmt.Printf("Protected path: %s is the GoCrypt executable. This file cannot be encrypted or decrypted.\n", filePath)
		return true
	}

	// Check if the file is in the GoCrypt installation directory (C:\Program Files\GoCrypt)
	if strings.HasPrefix(filePath, `c:\program files\gocrypt`) {
		fmt.Printf("Protected path: %s is within the GoCrypt installation directory. Files in this directory are protected.\n", filePath)
		return true
	}

	// Optionally, block common system directories (e.g., C:\Windows)
	if strings.Contains(filePath, `c:\windows`) {
		fmt.Printf("Protected path: %s is a system directory (Windows). This file cannot be encrypted or decrypted.\n", filePath)
		return true
	}

	// Allow everything else
	return false
}

// Check if the file is encrypted by GoCrypt based on the header format
// WORK IN PROGRESS
func IsFileEncrypted(filePath string) (bool, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return false, fmt.Errorf("could not open file: %v", err)
    }
    defer file.Close()

    // We need at least 41 bytes (1 byte for layer, 24 for nonce, 16 for salt)
    header := make([]byte, 41)
    n, err := file.Read(header)
    if err != nil && err != io.EOF {
        return false, fmt.Errorf("error reading file: %v", err)
    }
    if n < 41 {
        // File is too small to contain an encryption header
        return false, nil
    }

    // Check the layer value (first byte)
    layer := header[0]
	fmt.Println("layer:", layer)
    if layer < 1 || layer > 200 {
        // Invalid layer, not an encrypted file
        return false, nil
    }

    // Optional: you could add checks to validate the nonce and salt further if necessary.
    nonce := header[1:25] // 24 bytes nonce
    salt := header[25:41] // 16 bytes salt
	fmt.Println("salt:", salt)
	fmt.Println("nonce:", nonce)

    // Further validation on nonce and salt can be added here if desired.
    if len(nonce) != 24 || len(salt) != 16 {
        return false, fmt.Errorf("invalid nonce or salt size")
    }

    // File has a valid GoCrypt header
    return true, nil
}

func CheckFileCommand(files []string, action string) error {
	var encounteredEncrypted, encounteredNonEncrypted bool

	for _, filePath := range files {
		// Check if the current path is a directory
        info, err := os.Stat(filePath)
        if err != nil {
            return fmt.Errorf("could not access %s: %v", filePath, err)
        }
        
        // If it's a directory, skip the check
        if info.IsDir() {
            fmt.Printf("Skipping directory: %s\n", filePath)
            continue
        }

		// Check if the file is encrypted
		encrypted, err := IsFileEncrypted(filePath)
		fmt.Println("Is file encrypted?", encrypted, err)
		if err != nil {
			return err
		}

		// Track whether we've encountered encrypted and non-encrypted files
        if encrypted {
            encounteredEncrypted = true
        } else {
            encounteredNonEncrypted = true
        }

		if action == "decrypt" && !encrypted {
			return fmt.Errorf("file %s is not encrypted. please select an encrypted file", filePath)
			encounteredNonEncrypted = true
		}
		if action == "encrypt" && encrypted {
			encounteredEncrypted = true
			return fmt.Errorf("file %s is already encrypted. please select a non-encrypted file", filePath)
		}
	}	

	// If there is a mix of encrypted and non-encrypted files, return an error
    if encounteredEncrypted && encounteredNonEncrypted {
        return fmt.Errorf("mismatch of file types: both encrypted and non-encrypted files detected. please select files of the same type")
    }

	return nil
}

// IsDirectory checks if the given path is a directory.
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// CompressFolder compresses a folder into a .zip file.
func CompressFolder(folderPath, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %v", err)
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	err = filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relativePath := strings.TrimPrefix(path, folderPath)
		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		writer, err := archive.Create(relativePath)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		return err
	})

	return err
}