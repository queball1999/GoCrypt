package encryption

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/chacha20poly1305"
)

// DecryptFile decrypts the file at the given path and writes the decrypted data to the output path using ChaCha20-Poly1305.
// THIS FUNCTION IS DEPRECIATED BUT STILL FUNCTIONAL
func DecryptFile(source *os.File, pathOut, password string) error {
	// Read the nonce from the beginning of the file
	nonce := make([]byte, 24) // 24 bytes nonce
	if _, err := io.ReadFull(source, nonce); err != nil {
		return fmt.Errorf("failed to read nonce: %v", err)
	}

	// Read the salt from the file
	salt := make([]byte, 16) // 16 bytes salt
	if _, err := io.ReadFull(source, salt); err != nil {
		return fmt.Errorf("failed to read salt: %v", err)
	}

	aead, err := chacha20poly1305.NewX(DeriveKey(password, salt))
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
// This functin automatically detects the layer count in the header.
func LayeredDecryptFile(source *os.File, pathOut, password string) error {
	var currentSource *os.File = source

	// Read the layer header before entering the loop
	layerHeader := make([]byte, 1)
	if _, err := io.ReadFull(currentSource, layerHeader); err != nil {
		return fmt.Errorf("failed to read initial layer header: %v", err)
	}
	totalLayers := int(layerHeader[0])
	//fmt.Printf("Total layers to decrypt: %d\n", totalLayers)

	for layer := 0; layer < totalLayers; layer++ {
		//fmt.Printf("Starting layer %d decryption...\n", layer+1)

		// Skip the first byte (layer header) after the first loop
		if layer > 0 {
			if _, err := io.ReadFull(currentSource, layerHeader); err != nil {
				return fmt.Errorf("failed to skip layer header: %v", err)
			}
		}

		// Read the nonce from the beginning of the file
		nonce := make([]byte, 24) // 24 bytes nonce
		if _, err := io.ReadFull(currentSource, nonce); err != nil {
			return fmt.Errorf("failed to read nonce: %v", err)
		}

		// Read the salt from the file
		salt := make([]byte, 16) // 16 bytes salt
		if _, err := io.ReadFull(currentSource, salt); err != nil {
			return fmt.Errorf("failed to read salt: %v", err)
		}

		aead, err := chacha20poly1305.NewX(DeriveKey(password, salt))
		if err != nil {
			return fmt.Errorf("failed to create AEAD: %v", err)
		}

		// Create temp file for decrypted output
		tmpFile, err := os.CreateTemp("", "*.tmp")
		if err != nil {
			return err
		}

		// Buffer setup
		encryptedBuffer := make([]byte, 32*1024+16) // Buffer to hold ciphertext (32KB + 16 bytes MAC)
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

		// Explicitly close and flush temp file
		tmpFile.Close()

		// Reopen temp file and reset file pointer to the beginning
		if currentSource != source {
			currentSource.Close()
		}

		currentSource, err = os.Open(tmpFile.Name())
		if err != nil {
			return err
		}
	}

	currentSource.Close()

	// Copy the contents of the temp file to the output file
	err := copyFile(currentSource.Name(), pathOut)
	if err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	// Remove the temp file after a successful copy
	err = os.Remove(currentSource.Name())
	if err != nil {
		return fmt.Errorf("failed to remove temp file: %v", err)
	}

	return nil
}
