### Description
This document outlines the cryptographic processes utilized by _GoCrypt_, a file encryption tool heavily influenced by [eddy](https://github.com/70sh1/eddy/tree/main).
It serves as a reference for reconstructing the program or developing similar applications, such as decrypting files encrypted by GoCrypt.

### Outline
_GoCrypt_ employs the ChaCha20-Poly1305 authenticated encryption algorithm, which combines the ChaCha20 stream cipher with the Poly1305 message authentication code (MAC). This provides both confidentiality and integrity. The encryption key is a 256-bit value derived from a user-provided passphrase. This passphrase, along with a random salt, is passed through the PBKDF2 key derivation function (KDF) with the following parameters: 4096 iterations, a 32-byte key length, and the SHA-256 hash function.

Upon generating the key, a 24-byte nonce is randomly generated for each encryption operation. The encryption process uses an "Encrypt-then-MAC" (EtM) construction, where the Poly1305 MAC is computed over the ciphertext to ensure data integrity and authenticity.

### Layered Encryption
(**Coming Soon**) _GoCrypt_ offers an optional layered encryption feature, where each data chunk is encrypted multiple times, each with a different salt creating a unique key for each layer. The number of layers can be specified by the user, with each layer adding an additional level of security.

### Data Chunks
To optimize memory usage, _GoCrypt_ chunks the data and "streams" it to the output file in a controlled manner. This method ensures that only a portion of the data is kept in memory at any given time, significantly reducing the application's overall memory footprint. Each chunk is encrypted separately, and in the case of layered encryption, each chunk undergoes multiple rounds of encryption before being written to the file.

### File Format
An encrypted file (.enc) has the following structure. The format is designed to provide plausible deniability, meaning the file is generally indistinguishable from other types of data, such as compressed or randomly generated files.

| nonce    | scrypt salt | ecnrypted file contents |
| -------- | ----------- | ----------------------- |
| 12 bytes | 16 bytes    | 0~256GiB                |
