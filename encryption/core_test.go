package encryption

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// This function tests encrypting and decrypting a file. It also verifies the output hash matches the original file.
// EncryptTestFile encrypts a file with the provided password and returns the path to the encrypted file
func EncryptTestFile(inputFilePath, password string, layers int) (string, error) {
	// Open the input file
	inputFile, err := os.Open(inputFilePath)
	if err != nil {
		return "", err
	}
	defer inputFile.Close()

	// Create the encrypted file path
	encryptedFilePath := inputFilePath + ".enc"

	// Run the encryption function
	err = LayeredEncryptFile(inputFile, encryptedFilePath, password, layers)
	if err != nil {
		return "", err
	}

	// Return the path to the encrypted file
	return encryptedFilePath, nil
}

// DecryptTestFile decrypts an encrypted file with the provided password and returns the path to the decrypted file
func DecryptTestFile(encryptedFilePath, password string) (string, error) {
	// Open the encrypted file
	encryptedFile, err := os.Open(encryptedFilePath)
	if err != nil {
		return "", err
	}
	defer encryptedFile.Close()

	// Create the decrypted file path
	decryptedFilePath := encryptedFilePath[:len(encryptedFilePath)-4] + ".dec" // Remove ".enc" and add ".dec"

	// Run the decryption function
	err = LayeredDecryptFile(encryptedFile, decryptedFilePath, password)
	if err != nil {
		return "", err
	}

	// Return the path to the decrypted file
	return decryptedFilePath, nil
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

// TestEncryptAndDecryptFile tests both the encryption and decryption process
func TestEncryptAndDecryptFile(t *testing.T) {
	// Get the current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Prepare a test file with sample data in the current directory
	testFileName := "test_input.txt"
	testFilePath := filepath.Join(workingDir, testFileName) // Use full path
	key := "testpassword"
	originalData := []byte("This is a test file for encryption.")

	// Write the original data to the test file
	err = os.WriteFile(testFilePath, originalData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test input file: %v", err)
	}

	// Compute the hash of the original file
	originalHash, err := hashFile(testFilePath)
	if err != nil {
		t.Fatalf("Failed to hash the original file: %v", err)
	}

	// Encrypt the file
	encryptedFilePath, err := EncryptTestFile(testFilePath, key, 3) // 3 layers of encryption as an example
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Check if the encrypted file exists
	if _, err := os.Stat(encryptedFilePath); os.IsNotExist(err) {
		t.Fatalf("Encrypted file does not exist: %v", err)
	}

	// Decrypt the file
	decryptedFilePath, err := DecryptTestFile(encryptedFilePath, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	// Compute the hash of the decrypted file
	decryptedHash, err := hashFile(decryptedFilePath)
	if err != nil {
		t.Fatalf("Failed to hash the decrypted file: %v", err)
	}

	// Compare the hashes of the original and decrypted files
	if originalHash != decryptedHash {
		t.Errorf("Hash mismatch!\nOriginal Hash: %s\nDecrypted Hash: %s", originalHash, decryptedHash)
	}

	// Cleanup the test files
	os.Remove(encryptedFilePath)
	os.Remove(decryptedFilePath)
	os.Remove(testFilePath)
}

// TestEncryptAndDecryptFileError tests decryption with a wrong password
func TestEncryptAndDecryptFileError(t *testing.T) {
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	testFileName := "test_input.txt"
	testFilePath := filepath.Join(workingDir, testFileName)
	key := "testpassword"
	originalData := []byte("This is a test file for encryption.")

	// Write the original data to the test file
	err = os.WriteFile(testFilePath, originalData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test input file: %v", err)
	}

	// Encrypt the file using the helper function
	encryptedFilePath, err := EncryptTestFile(testFilePath, key, 3)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}
	defer os.Remove(encryptedFilePath) // Cleanup

	// Try to decrypt the file with the wrong password
	_, err = DecryptTestFile(encryptedFilePath, "wrongpassword")
	if err == nil {
		t.Fatalf("Expected decryption to fail with wrong password")
	}

	// Cleanup
	os.Remove(testFilePath)
}

// TestEncryptFile tests that encryption produces a file
func TestEncryptFile(t *testing.T) {
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	testFileName := "test_input.txt"
	testFilePath := filepath.Join(workingDir, testFileName)
	key := "testpassword"
	originalData := []byte("Test data")

	// Write the original data to the test file
	err = os.WriteFile(testFilePath, originalData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test input file: %v", err)
	}

	// Encrypt the file using the helper function
	encryptedFilePath, err := EncryptTestFile(testFilePath, key, 3)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}
	defer os.Remove(encryptedFilePath) // Cleanup

	// Check if the encrypted file exists
	if _, err := os.Stat(encryptedFilePath); os.IsNotExist(err) {
		t.Fatalf("Encrypted file does not exist: %v", err)
	}

	// Cleanup
	os.Remove(testFilePath)
}

// TestEncryptFileError tests encryption failure for invalid input
func TestEncryptFileError(t *testing.T) {
	invalidFilePath := "invalid_file_path.txt"
	key := "testpassword"

	// Try to encrypt a non-existent file
	_, err := EncryptTestFile(invalidFilePath, key, 3)
	if err == nil {
		t.Fatalf("Expected encryption to fail with invalid input")
	}
}

// TestDecryptFile tests the decryption of an encrypted file
func TestDecryptFile(t *testing.T) {
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	testFileName := "test_input.txt"
	testFilePath := filepath.Join(workingDir, testFileName)
	key := "testpassword"
	originalData := []byte("This is a test file for encryption.")

	// Write the original data to the test file
	err = os.WriteFile(testFilePath, originalData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test input file: %v", err)
	}

	// Encrypt the file
	encryptedFilePath, err := EncryptTestFile(testFilePath, key, 3)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}
	defer os.Remove(encryptedFilePath) // Cleanup

	// Decrypt the file using the helper function
	decryptedFilePath, err := DecryptTestFile(encryptedFilePath, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}
	defer os.Remove(decryptedFilePath) // Cleanup

	// Check if the decrypted file exists
	if _, err := os.Stat(decryptedFilePath); os.IsNotExist(err) {
		t.Fatalf("Decrypted file does not exist: %v", err)
	}

	// Cleanup
	os.Remove(testFilePath)
}

// TestDecryptFileError tests decryption failure for invalid input or incorrect password
func TestDecryptFileError(t *testing.T) {
	invalidFilePath := "invalid_file_path.enc"
	key := "testpassword"

	// Try to decrypt a non-existent file
	_, err := DecryptTestFile(invalidFilePath, key)
	if err == nil {
		t.Fatalf("Expected decryption to fail with invalid input")
	}
}