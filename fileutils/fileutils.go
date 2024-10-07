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