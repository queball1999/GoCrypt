package encryption

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/chacha20poly1305"
)

// EncryptFile encrypts the file at the given path and writes the encrypted data to the output path using ChaCha20-Poly1305.
func EncryptFile(source *os.File, pathOut, password string) error {
	/*
	logFile, err := os.Create("encryption.log")
	if err != nil {
		return fmt.Errorf("failed to create log file: %v", err)
	}
	defer logFile.Close()
	*/

	// Generate a unique salt for each file
	salt, err := GenerateSalt()
	if err != nil {
		return err
	}
	//fmt.Fprintf(logFile, "Salt: %x\n", salt)

	aead, err := chacha20poly1305.NewX(DeriveKey(password, salt))
	if err != nil {
		return fmt.Errorf("failed to create AEAD: %v", err)
	}

	// Generate a nonce
	nonce := make([]byte, 24) // 24 bytes nonce
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %v", err)
	}
	//fmt.Fprintf(logFile, "Nonce: %x\n", nonce)

	// Create temp file
	tmpFile, err := os.CreateTemp("", "*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	// Write the nonce and salt to the output file
	if _, err := tmpFile.Write(nonce); err != nil {
		return fmt.Errorf("failed to write nonce: %v", err)
	}
	if _, err := tmpFile.Write(salt); err != nil {
		return fmt.Errorf("failed to write salt: %v", err)
	}

	// Adjust the buffer size to account for the MAC overhead
	buffer := make([]byte, 32*1024)                 // 32KB buffer for reading plaintext
	encryptedBuffer := make([]byte, 32*1024+16) // Buffer to hold ciphertext (plaintext + 16 bytes MAC)

	for {
		n, err := source.Read(buffer)
		if n > 0 {
			//fmt.Fprintf(logFile, "Decrypting chunk of size: %d bytes with nonce: %x\n", n, nonce)
			// Encrypt the buffer chunk
			ciphertext := aead.Seal(encryptedBuffer[:0], nonce, buffer[:n], nil)
			if _, err := tmpFile.Write(ciphertext); err != nil {
				return fmt.Errorf("failed to write encrypted data: %v", err)
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

// LayeredEncryptFile encrypts the file with multiple layers using ChaCha20-Poly1305.
// FIXME: Need a way to auto-detect layer count.
func LayeredEncryptFile(source *os.File, pathOut, password string, layers int) error {
	if layers <= 0 {
		return fmt.Errorf("invalid number of layers: %d", layers)
	}

	var currentSource *os.File = source

	for layer := 0; layer < layers; layer++ {
		fmt.Printf("Starting layer %d encryption...\n", layer+1)

		// Generate a unique salt for each layer
		salt, err := GenerateSalt()
		if err != nil {
			return err
		}
		fmt.Printf("Salt: %x\n", salt)

		aead, err := chacha20poly1305.NewX(DeriveKey(password, salt))
		if err != nil {
			return fmt.Errorf("failed to create AEAD: %v", err)
		}

		// Generate a nonce
		nonce := make([]byte, 24) // 24 bytes nonce
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return fmt.Errorf("failed to generate nonce: %v", err)
		}
		fmt.Printf("Nonce: %x\n", nonce)

		// Create temp file for current layer
		tmpFile, err := os.CreateTemp("", "*.tmp")
		if err != nil {
			return err
		}
		defer os.Remove(tmpFile.Name())

		// Write the headers to the temp file (layer, nonce, and salt)
		layerHeader := []byte{byte(layer + 1)} // Convert the current layer to a 1-byte value
		if _, err := tmpFile.Write(layerHeader); err != nil {
			return fmt.Errorf("failed to write layer header: %v", err)
		}
		if _, err := tmpFile.Write(nonce); err != nil {
			return fmt.Errorf("failed to write nonce: %v", err)
		}
		if _, err := tmpFile.Write(salt); err != nil {
			return fmt.Errorf("failed to write salt: %v", err)
		}

		// Adjust the buffer size to account for the MAC overhead
		buffer := make([]byte, 32*1024)             // 32KB buffer for reading plaintext
		encryptedBuffer := make([]byte, 32*1024+16) // Buffer to hold ciphertext (plaintext + 16 bytes MAC)

		for {
			n, err := currentSource.Read(buffer)
			if n > 0 {
				// Encrypt the buffer chunk
				ciphertext := aead.Seal(encryptedBuffer[:0], nonce, buffer[:n], nil)
				if _, err := tmpFile.Write(ciphertext); err != nil {
					return fmt.Errorf("failed to write encrypted data: %v", err)
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