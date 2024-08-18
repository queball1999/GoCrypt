### Description
This file describes the cryptography done by _GoCrypt_, which is heavily influenced by [eddy](https://github.com/70sh1/eddy/tree/main).
It can be used to recreate the program (or build a new one) to, for example, decrypt already encrypted files.

### Outline
_GoCrypt_ uses the ChaCha20-Poly1305 authenticated encryption algorithm, which combines the ChaCha20 stream cipher with the Poly1305 message authentication code (MAC). The 256-bit ChaCha20 key is derived by passing a user-provided or randomly generated password (passphrase) and a random 16-byte salt to the scrypt key derivation function (KDF) with the following parameters: n=65536, r=8, p=1, keyLen=32.

After deriving the key, a random 12-byte nonce is generated for each encryption operation. The Poly1305 MAC tag is calculated over the ciphertext, ensuring data integrity and authenticity in an "Encrypt-then-MAC" (EtM) construction.

### Layered Encryption
_GoCrypt_ offers an optional layered encryption feature, where each data chunk is encrypted multiple times, each with a different salt creating a unique key for each layer. The number of layers can be specified by the user, with each layer adding an additional level of security.

### Data Chunks
To optimize memory usage, _GoCrypt_ chunks the data and "streams" it to the output file in a controlled manner. This method ensures that only a portion of the data is kept in memory at any given time, significantly reducing the application's overall memory footprint. Each chunk is encrypted separately, and in the case of layered encryption, each chunk undergoes multiple rounds of encryption before being written to the file.

### File Format
An encrypted file (.enc) has the following structure. The format is designed to provide plausible deniability, meaning the file is generally indistinguishable from other types of data, such as compressed or randomly generated files.

| nonce    | scrypt salt | MAC tag  | ecnrypted file contents |
| -------- | ----------- | -------- | ----------------------- |
| 12 bytes | 16 bytes    | 64 bytes | 0~256GiB                |
