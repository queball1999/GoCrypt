package encryption

import (
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
