package encryption

import (
    "crypto/aes"
    "crypto/cipher"
    "fmt"
    "golang.org/x/crypto/chacha20poly1305"
)

func DecryptData(ciphertext []byte, key []byte, method string) ([]byte, error) {
    switch method {
    case "aes":
        return aesDecrypt(ciphertext, key)
    default:
        return chacha20poly1305Decrypt(ciphertext, key)
    }
}

func aesDecrypt(ciphertext []byte, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonceSize := gcm.NonceSize()
    if len(ciphertext) < nonceSize {
        return nil, fmt.Errorf("ciphertext too short")
    }

    nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return nil, err
    }

    return plaintext, nil
}

func chacha20poly1305Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
    aead, err := chacha20poly1305.NewX(key)
    if err != nil {
        return nil, err
    }

    nonceSize := aead.NonceSize()
    if len(ciphertext) < nonceSize {
        return nil, fmt.Errorf("ciphertext too short")
    }

    nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
    plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return nil, err
    }

    return plaintext, nil
}
