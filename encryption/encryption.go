package encryption

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "golang.org/x/crypto/chacha20poly1305"
    "io"
)

func EncryptData(plaintext []byte, key []byte, method string) ([]byte, error) {
    switch method {
    case "aes":
        return aesEncrypt(plaintext, key)
    default:
        return chacha20poly1305Encrypt(plaintext, key)
    }
}

func aesEncrypt(plaintext []byte, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }

    ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
    return ciphertext, nil
}

func chacha20poly1305Encrypt(plaintext []byte, key []byte) ([]byte, error) {
    aead, err := chacha20poly1305.NewX(key)
    if err != nil {
        return nil, err
    }

    nonce := make([]byte, aead.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }

    ciphertext := aead.Seal(nonce, nonce, plaintext, nil)
    return ciphertext, nil
}
