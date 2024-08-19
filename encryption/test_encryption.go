package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/blake2b"
)

// EncryptFile encrypts the file at the given path and writes the encrypted data to the output path.
func EncryptFile1(source *os.File, pathOut, password string) error {
	fileInfo, err := source.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}
	fmt.Printf("Encrypting file of size: %d bytes\n", fileInfo.Size())

	// Generate a unique salt for each file
	salt, err := GenerateSalt()
	if err != nil {
		return err
	}
	fmt.Printf("Salt: %x\n", salt)

	// Derive the encryption key using the password and unique salt
	key := DeriveKey(password, salt)
	fmt.Printf("KEY: %x\n", key)

	// Ensure key length is 32 bytes
	if len(key) != 32 {
		fmt.Println("Error: Derived key length is not 32 bytes")
		return nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	nonce := generateNonce(block.BlockSize())
	stream := cipher.NewCTR(block, nonce)
	blake, err := blake2b.New256(nil)
	if err != nil {
		return err
	}
	fmt.Printf("Nonce: %x\n", nonce)

	tmpFile, err := os.CreateTemp("", "*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	// Write the salt and nonce to the header (reserve space for MAC)
	header := make([]byte, 64)
	copy(header[:16], salt)
	copy(header[16:32], nonce)
	if _, err := tmpFile.Write(header); err != nil {
		return err
	}

	buffer := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := source.Read(buffer)
		if n > 0 {
			// Encrypt the buffer
			stream.XORKeyStream(buffer[:n], buffer[:n])

			// Write to temp file and update the MAC
			if _, err := tmpFile.Write(buffer[:n]); err != nil {
				return err
			}
			blake.Write(buffer[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	// Calculate the MAC
	mac := blake.Sum(nil)
	fmt.Printf("BLAKE TAG: %x\n", mac)

	// Write the MAC into the reserved header space
	if _, err := tmpFile.Seek(32, io.SeekStart); err != nil {
		return err
	}
	if _, err := tmpFile.Write(mac); err != nil {
		return err
	}

	tmpFile.Close()
	if err := os.Rename(tmpFile.Name(), pathOut); err != nil {
		return err
	}

	return nil
}

func LayeredEncryptFile1(source *os.File, pathOut, password string, layers int) error {
	fileInfo, err := source.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}
	fmt.Printf("Encrypting file of size: %d bytes\n", fileInfo.Size())

	// Generate the initial salt and key
	salt, err := GenerateSalt()
	if err != nil {
		return err
	}
	fmt.Printf("Initial Salt: %x\n", salt)
	key := DeriveKey(password, salt)

	// Ensure key length is 32 bytes
	if len(key) != 32 {
		fmt.Println("Error: Derived key length is not 32 bytes")
		return nil
	}

	// Temporary files for layers
	var currentSource *os.File = source

	for layer := 0; layer < layers; layer++ {
		fmt.Printf("Starting layer %d encryption...\n", layer+1)

		// Create a temporary file for this layer's output
		tmpFile, err := os.CreateTemp("", "*.tmp")
		if err != nil {
			return err
		}
		defer os.Remove(tmpFile.Name())

		// Generate nonce for this layer
		block, err := aes.NewCipher(key)
		if err != nil {
			return err
		}
		nonce := generateNonce(block.BlockSize())
		stream := cipher.NewCTR(block, nonce)
		blake, err := blake2b.New256(nil)
		if err != nil {
			return err
		}
		fmt.Printf("Layer %d Nonce: %x\n", layer+1, nonce)

		// Write the salt and nonce to the header (reserve space for MAC)
		header := make([]byte, 64)
		copy(header[:16], salt)
		copy(header[16:32], nonce)
		if _, err := tmpFile.Write(header); err != nil {
			return err
		}

		// Process the file in chunks
		buffer := make([]byte, 32*1024) // 32KB buffer
		for {
			n, err := currentSource.Read(buffer)
			if n > 0 {
				// Encrypt the buffer
				stream.XORKeyStream(buffer[:n], buffer[:n])

				// Write to temp file and update the MAC
				if _, err := tmpFile.Write(buffer[:n]); err != nil {
					return err
				}
				blake.Write(buffer[:n])
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
		}

		// Calculate the MAC for this layer
		mac := blake.Sum(nil)
		fmt.Printf("Layer %d MAC: %x\n", layer+1, mac)

		// Write the MAC into the reserved header space
		if _, err := tmpFile.Seek(32, io.SeekStart); err != nil {
			return err
		}
		if _, err := tmpFile.Write(mac); err != nil {
			return err
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

		// Update the salt and key for the next layer
		salt, err = GenerateSalt()
		if err != nil {
			return err
		}
		key = DeriveKey(password, salt)
	}
	
	currentSource.Close()
	// Rename the final temp file to the output path
	if err := os.Rename(currentSource.Name(), pathOut); err != nil {
		return err
	}
	

	return nil
}

// generateNonce generates a nonce for AES-CTR.
func generateNonce(size int) []byte {
	nonce := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err) // In production, handle this properly
	}
	return nonce
}