package encryption

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/chacha20poly1305"
)

// DecryptFile decrypts the file at the given path and writes the decrypted data to the output path using ChaCha20-Poly1305.
func DecryptFile(source *os.File, pathOut, password string) error {
	/*
	logFile, err := os.Create("decryption.log")
	if err != nil {
		return fmt.Errorf("failed to create log file: %v", err)
	}
	defer logFile.Close()
	*/

	// Read the nonce from the beginning of the file
	nonce := make([]byte, 24) // 24 bytes nonce
	if _, err := io.ReadFull(source, nonce); err != nil {
		return fmt.Errorf("failed to read nonce: %v", err)
	}
	//fmt.Fprintf(logFile, "Nonce: %x\n", nonce)

	// Read the salt from the file
	salt := make([]byte, 16) // 16 bytes salt
	if _, err := io.ReadFull(source, salt); err != nil {
		return fmt.Errorf("failed to read salt: %v", err)
	}
	//fmt.Fprintf(logFile, "Salt: %x\n", salt)

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
			//fmt.Fprintf(logFile, "Decrypting chunk of size: %d bytes with nonce: %x\n", n, nonce)
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
func LayeredDecryptFile(source *os.File, pathOut, password string, layers int) error {
    var currentSource *os.File = source
    fmt.Printf("Initial source file: %s\n", source.Name())

    for layer := 0; layer < layers; layer++ {
        fmt.Printf("\nStarting layer %d decryption...\n", layer+1)

        // Read the nonce from the beginning of the file
        nonce := make([]byte, 24) // 24 bytes nonce
        if _, err := io.ReadFull(currentSource, nonce); err != nil {
            return fmt.Errorf("failed to read nonce: %v", err)
        }
        fmt.Printf("Nonce for layer %d: %x\n", layer+1, nonce)

        // Read the salt from the file
        salt := make([]byte, 16) // 16 bytes salt
        if _, err := io.ReadFull(currentSource, salt); err != nil {
            return fmt.Errorf("failed to read salt: %v", err)
        }
        fmt.Printf("Salt for layer %d: %x\n", layer+1, salt)

        aead, err := chacha20poly1305.NewX(DeriveKey(password, salt))
        if err != nil {
            return fmt.Errorf("failed to create AEAD: %v", err)
        }

        // Create temp file for decrypted output
        tmpFile, err := os.CreateTemp("", "*.tmp")
        if err != nil {
            return err
        }
        fmt.Printf("Created temp file for layer %d: %s\n", layer+1, tmpFile.Name())
        //defer os.Remove(tmpFile.Name())

        // Buffer setup
        encryptedBuffer := make([]byte, 32*1024+16) // Buffer to hold ciphertext (32KB + 16 bytes MAC)
        plaintextBuffer := make([]byte, 32*1024)    // Buffer for decrypted plaintext

        for {
            n, err := currentSource.Read(encryptedBuffer)
            if n > 0 {
                // Decrypt the buffer chunk
                plaintext, err := aead.Open(plaintextBuffer[:0], nonce, encryptedBuffer[:n], nil)

				// comment out for debugging
				//fmt.Println("Plaintext", plaintext)

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
        fmt.Printf("Closed temp file for layer %d: %s\n", layer+1, tmpFile.Name())

        // Reopen temp file and reset file pointer to the beginning
        if currentSource != source {
            fmt.Printf("Closing previous source file: %s\n", currentSource.Name())
            currentSource.Close()
        }

        currentSource, err = os.Open(tmpFile.Name())
        if err != nil {
            return err
        }
        fmt.Printf("Opened new source file for layer %d: %s\n", layer+1, currentSource.Name())
    }

    // Final source close and rename the decrypted file to output path
    currentSource.Close()
    fmt.Printf("Final source file closed: %s\n", currentSource.Name())

    // Rename the final temp file to the output path
    if err := os.Rename(currentSource.Name(), pathOut); err != nil {
        return err
    }
    fmt.Printf("Renamed final decrypted file to output path: %s\n", pathOut)

    return nil
}