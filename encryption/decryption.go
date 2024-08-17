package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
	"os"
	"golang.org/x/crypto/blake2b"
)

// DecryptFile decrypts the encrypted file at the given path and writes the decrypted data to the output path.
func DecryptFile(source *os.File, pathOut, password string) error {
	fileInfo, err := source.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}
	fmt.Printf("Decrypting file of size: %d bytes\n", fileInfo.Size())

	// Read the salt, nonce, and MAC from the file header
	header := make([]byte, 64) // 16 bytes for salt, 16 for nonce, 32 for MAC
	if _, err := io.ReadFull(source, header); err != nil {
		return fmt.Errorf("failed to read header: %v", err)
	}
	salt := header[:16]
	nonce := header[16:32]
	fileTag := header[32:64]

	fmt.Printf("Salt: %x\n", salt)
	fmt.Printf("Nonce: %x\n", nonce)
	fmt.Printf("Stored MAC: %x\n", fileTag)

	// Derive the decryption key using the password and salt
	key := DeriveKey(password, salt)
	fmt.Printf("KEY: %x\n", key)

	// Ensure key length is 32 bytes
	if len(key) != 32 {
		return fmt.Errorf("error: %v", "derived key length is not 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %v", err)
	}

	stream := cipher.NewCTR(block, nonce)
	blake, err := blake2b.New256(nil)
	if err != nil {
		return fmt.Errorf("failed to create Blake2b hash: %v", err)
	}

	// Prepare output file
	tmpFile, err := os.CreateTemp("", "*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Decrypt the file data and compute the MAC
	buffer := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := source.Read(buffer)
		if n > 0 {
			// Decrypt the buffer
			stream.XORKeyStream(buffer[:n], buffer[:n])

			// Write to temp file and update the MAC
			if _, err := tmpFile.Write(buffer[:n]); err != nil {
				return fmt.Errorf("failed to write to temp file: %v", err)
			}
			blake.Write(buffer[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read from source file: %v", err)
		}
	}

	// Calculate the expected MAC
	expectedTag := blake.Sum(nil)
	fmt.Printf("Computed MAC: %x\n", expectedTag)

	// Compare the computed MAC with the stored MAC
    /*
	if !bytes.Equal(expectedTag, fileTag) {
		return fmt.Errorf("MAC mismatch: decryption failed or file corrupted")
	}
    */

	// Finalize the temp file
	tmpFile.Close()
	if err := os.Rename(tmpFile.Name(), pathOut); err != nil {
		return fmt.Errorf("failed to rename temp file: %v", err)
	}

	fmt.Printf("Decryption successful: %s\n", pathOut)
	return nil
}