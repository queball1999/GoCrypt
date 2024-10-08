package encryption

import (
    "io"
	"os"
    "fmt"
    "crypto/rand"
    "crypto/sha256"
    "golang.org/x/crypto/pbkdf2"
)

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 16) // 16 bytes is a common salt length
	_, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %v", err)
	}
	return salt, nil
}

func DeriveKey(password string, salt []byte) []byte {
    return pbkdf2.Key([]byte(password), salt, 4096, 32, sha256.New)
}

func incrementNonce(nonce []byte) {
    for i := len(nonce) - 1; i >= 0; i-- {
        nonce[i]++
        if nonce[i] != 0 {
            break
        }
    }
}

// Helper function to copy file contents
func copyFile(src, dst string) error {
    sourceFile, err := os.Open(src)
    if err != nil {
        return err
    }
    defer sourceFile.Close()

    destinationFile, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer destinationFile.Close()

    _, err = io.Copy(destinationFile, sourceFile)
    return err
}