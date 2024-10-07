package encryption

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptAndDecryptFile(t *testing.T) {
	// Get the current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Prepare a test file with sample data in the current directory
	testFileName := "test_input.txt"
	testFilePath := filepath.Join(workingDir, testFileName) // Use full path
	key := []byte("testpassword")
	originalData := []byte("This is a test file for encryption.")

	err = os.WriteFile(testFilePath, originalData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test input file: %v", err)
	}

	// Compute hash of the original file
	originalHash, err := hashFile(testFilePath)
	if err != nil {
		t.Fatalf("Failed to hash the original file: %v", err)
	}

	inputFile, err := os.Open(testFilePath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer inputFile.Close()

	// Create the encrypted file path
	outputPath := testFilePath + ".enc"

	// Run the encryption function
	err = LayeredEncryptFile(inputFile, outputPath, string(key), 3) // 3 layers of encryption as an example
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Check if the encrypted file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Encrypted file does not exist: %v", err)
	}

	encryptedFile, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("Failed to open encrypted test file: %v", err)
	}
	defer encryptedFile.Close()

	// Create the decrypted file path
	decryptedFilePath := testFilePath + ".dec"

	// Decrypt the encrypted file
	err = LayeredDecryptFile(encryptedFile, decryptedFilePath, string(key))
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	// Compute hash of the decrypted file
	decryptedHash, err := hashFile(decryptedFilePath)
	if err != nil {
		t.Fatalf("Failed to hash the decrypted file: %v", err)
	}

	// Compare the hashes of the original and decrypted files
	if originalHash != decryptedHash {
		t.Errorf("Hash mismatch!\nOriginal Hash: %s\nDecrypted Hash: %s", originalHash, decryptedHash)
	}

	// Cleanup the test files
	os.Remove(outputPath)
	os.Remove(decryptedFilePath)
	os.Remove(testFilePath)
}

// hashFile computes the SHA-256 hash of a file and returns it as a hex string
func hashFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}