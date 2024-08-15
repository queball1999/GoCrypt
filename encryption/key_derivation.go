package encryption

import (
    "crypto/sha256"
    "golang.org/x/crypto/pbkdf2"
)

const salt = "fixedsalt"

func DeriveKey(password string) []byte {
    return pbkdf2.Key([]byte(password), []byte(salt), 4096, 32, sha256.New)
}
