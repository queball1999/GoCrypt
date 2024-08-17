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
func EncryptFile(source *os.File, pathOut, password string) error {
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

    // Write the salt and nonce to the header
    if err := writeHeader(tmpFile, salt, nonce); err != nil {
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

    // Calculate and write the MAC tag at the end of the file
    tag := blake.Sum(nil)
    fmt.Printf("BLAKE TAG: %x\n", tag)
    if _, err := tmpFile.Write(tag); err != nil {
        return err
    }

    tmpFile.Close()
    if err := os.Rename(tmpFile.Name(), pathOut); err != nil {
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

// writeHeader writes the encryption header (salt and nonce) to the output file.
func writeHeader(file *os.File, salt, nonce []byte) error {
	// Write the salt to the header
	if _, err := file.Write(salt); err != nil {
		return err
	}

	// Write the nonce to the header
	if _, err := file.Write(nonce); err != nil {
		return err
	}
	return nil
}
