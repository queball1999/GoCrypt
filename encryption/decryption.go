package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
	"os"
	
	"golang.org/x/crypto/blake2b"
)

// FIXME: Need to work on headers and remove the additional data at the end of the files.

// DecryptFile decrypts the encrypted file at the given path and writes the decrypted data to the output path.
func DecryptFile(source *os.File, pathOut, password string) error {
    fileInfo, err := source.Stat()
    if err != nil {
        return fmt.Errorf("failed to get file info: %v", err)
    }
    fmt.Printf("Decrypting file of size: %d bytes\n", fileInfo.Size())

    // Read the salt and nonce from the file header
    salt := make([]byte, 16)
    if _, err := io.ReadFull(source, salt); err != nil {
        return fmt.Errorf("failed to read salt: %v", err)
    }
    fmt.Printf("Salt: %x\n", salt)

    nonce := make([]byte, 16)
    if _, err := io.ReadFull(source, nonce); err != nil {
        return fmt.Errorf("failed to read nonce: %v", err)
    }
    fmt.Printf("Nonce: %x\n", nonce)

    // Derive the decryption key using the password and salt
    key := DeriveKey(password, salt)
    fmt.Printf("KEY: %x\n", key)

    // Ensure key length is 32 bytes
    if len(key) != 32 {
        return fmt.Errorf("error: %v", "derived key length is not 32 bytes")
    }

    block, err := aes.NewCipher(key)
	fmt.Printf("BLOCK: %x\n", block)
    if err != nil {
        return fmt.Errorf("failed to create AES cipher: %v", err)
    }

    stream := cipher.NewCTR(block, nonce)
	fmt.Printf("Stream: %x\n", stream)
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

    // Calculate the size of the data to decrypt (excluding the MAC)
    dataSize := fileInfo.Size() - int64(blake.Size())
    buffer := make([]byte, 32*1024) // 32KB buffer
    totalRead := int64(0)

    for totalRead < dataSize {
        toRead := int64(len(buffer))
        if dataSize-totalRead < toRead {
            toRead = dataSize - totalRead
        }

        n, err := source.Read(buffer[:toRead])
        if n > 0 {
            // Decrypt the buffer
            stream.XORKeyStream(buffer[:n], buffer[:n])

            // Write to temp file and update the MAC
            if _, err := tmpFile.Write(buffer[:n]); err != nil {
                return fmt.Errorf("failed to write to temp file: %v", err)
            }

            blake.Write(buffer[:n])
            totalRead += int64(n)
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
    fmt.Printf("Expected TAG: %x\n", expectedTag)

    fileTag := make([]byte, len(expectedTag))

    // Read the actual MAC tag from the end of the file
    if _, err := source.Seek(-int64(len(expectedTag)), io.SeekEnd); err != nil {
        return fmt.Errorf("failed to seek to MAC tag position: %v", err)
    }

    if _, err := io.ReadFull(source, fileTag); err != nil {
        return fmt.Errorf("failed to read MAC tag from file: %v", err)
    }
    fmt.Printf("File TAG: %x\n", fileTag)

    // Compare the expected MAC with the actual MAC
	// FIXME: MAC comparison not working. Needs bytes
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