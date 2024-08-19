package fileutils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

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
	// Get the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return false
	}

	// Check if the file is within any GoCrypt-related directory or is GoCrypt executable
	if strings.Contains(filePath, "GoCrypt") || strings.Contains(filePath, "gocrypt.exe") || strings.HasPrefix(filePath, currentDir) {
		return true
	}
	return false
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