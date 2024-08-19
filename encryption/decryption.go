package encryption

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/chacha20poly1305"
)

// DecryptFile decrypts the file at the given path and writes the decrypted data to the output path using ChaCha20-Poly1305.
func DecryptFile(source *os.File, pathOut, password string) error {
	logFile, err := os.Create("decryption.log")
	if err != nil {
		return fmt.Errorf("failed to create log file: %v", err)
	}
	defer logFile.Close()

	// Read the nonce from the beginning of the file
	nonce := make([]byte, 24) // 24 bytes nonce
	if _, err := io.ReadFull(source, nonce); err != nil {
		return fmt.Errorf("failed to read nonce: %v", err)
	}
	fmt.Printf("Nonce: %x\n", nonce)

	// Read the salt from the file
	salt := make([]byte, 16) // 16 bytes salt
	if _, err := io.ReadFull(source, salt); err != nil {
		return fmt.Errorf("failed to read salt: %v", err)
	}
	fmt.Printf("Salt: %x\n", salt)

	// Derive the decryption key using the password and the salt
	key := DeriveKey(password, salt)
	fmt.Printf("Key: %x\n", key)

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return fmt.Errorf("failed to create AEAD: %v", err)
	}

	tmpFile, err := os.CreateTemp("", "*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	// Adjust the buffer size to account for the MAC overhead
	encryptedBuffer := make([]byte, 32*1024+16) // Buffer to hold ciphertext
	plaintextBuffer := make([]byte, 32*1024)    // Buffer for decrypted plaintext

	for {
		n, err := source.Read(encryptedBuffer)
		if n > 0 {
			// Decrypt the buffer chunk
			plaintext, err := aead.Open(plaintextBuffer[:0], nonce, encryptedBuffer[:n], nil)
			if err != nil {
				return fmt.Errorf("decryption failed: %v", err)
			}
			if _, err := tmpFile.Write(plaintext); err != nil {
				return fmt.Errorf("failed to write decrypted data: %v", err)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	tmpFile.Close()
	if err := os.Rename(tmpFile.Name(), pathOut); err != nil {
		return err
	}

	return nil
}

// LayeredDecryptFile decrypts the file with multiple layers using ChaCha20-Poly1305.
// FIXME: Need to fix layered issues. Reaching EOF after first layer.
func LayeredDecryptFile(source *os.File, pathOut, password string, layers int) error {
	// Temporary files for layers
	var currentSource *os.File = source

	for layer := 0; layer < layers; layer++ {
		fmt.Printf("Starting layer %d decryption...\n", layer+1)

        // Read the nonce from the beginning of the file
		nonce := make([]byte, 24) // 24 bytes nonce
		if _, err := io.ReadFull(currentSource, nonce); err != nil {
			return fmt.Errorf("failed to read nonce: %v", err)
		}
		fmt.Printf("Nonce: %x\n", nonce)

		// Read the salt from the file
		salt := make([]byte, 16) // 16 bytes salt
		if _, err := io.ReadFull(currentSource, salt); err != nil {
			return fmt.Errorf("failed to read salt: %v", err)
		}
		fmt.Printf("Salt: %x\n", salt)

		// Derive the decryption key using the password and the salt
		key := DeriveKey(password, salt)
		fmt.Printf("Key: %x\n", key)

		aead, err := chacha20poly1305.NewX(key)
		if err != nil {
			return fmt.Errorf("failed to create AEAD: %v", err)
		}

		// Create a temporary file for this layer's output
		tmpFile, err := os.CreateTemp("", "*.tmp")
		if err != nil {
			return err
		}
		defer os.Remove(tmpFile.Name())

		// Adjust the buffer size to account for the MAC overhead
		encryptedBuffer := make([]byte, 32*1024+16) // Buffer to hold ciphertext
		plaintextBuffer := make([]byte, 32*1024)    // Buffer for decrypted plaintext
		
		for {
			n, err := currentSource.Read(encryptedBuffer)
			if n > 0 {
				// Decrypt the buffer chunk
				plaintext, err := aead.Open(plaintextBuffer[:0], nonce, encryptedBuffer[:n], nil)
				if err != nil {
					return fmt.Errorf("layer %d decryption failed: %v", layer+1, err)
				}
				if _, err := tmpFile.Write(plaintext); err != nil {
					return fmt.Errorf("failed to write decrypted data: %v", err)
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
		}

		tmpFile.Close()

		// Close the previous source file and set the current source to the new temp file
		if currentSource != source {
			currentSource.Close()
		}
		currentSource, err = os.Open(tmpFile.Name())
		if err != nil {
			return err
		}
	}

	currentSource.Close()
	// Rename the final temp file to the output path
	if err := os.Rename(currentSource.Name(), pathOut); err != nil {
		return err
	}

	return nil
}