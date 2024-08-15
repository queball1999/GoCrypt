# Lightweight Context Menu File Encryption Tool

## Overview

This is a lightweight, user-friendly file encryption utility written in Go, designed to integrate seamlessly into your Windows context menu. With just a right-click, you can securely encrypt or decrypt files directly from the file explorer. The tool supports fast and secure encryption algorithms, including ChaCha20-Poly1305, and provides options to automatically delete the original file after encryption for added security.

## Features

- **Context Menu Integration:** Encrypt or decrypt files directly from the Windows context menu.
- **Secure Encryption:** Utilizes ChaCha20-Poly1305 for fast and secure encryption.
- **File Deletion Option:** Automatically delete original files after encryption.
- **Open Source:** The project is licensed under the MIT License, allowing for free use, modification, and distribution.

## Installation

For the latest pre-built versions of the application, check the [Releases](https://github.com/queball1999/GoCrypt/releases) page on the repository.

## Usage

Once installed, simply right-click on any file or folder in Windows Explorer and select the GoCrypt option from the context menu. You will be prompted to enter a password to either encrypt or decrypt the selected item.

## Contributing

Contributions are welcome! Feel free to submit a pull request or open an issue if you encounter any problems.

## Compiling

The installer is generated using [Inno Setup](https://jrsoftware.org/isinfo.php) and the config file can be found in this repo.

## License

This project is licensed under the MIT License. See the LICENSE[https://github.com/queball1999/GoCrypt/blob/main/LICENSE] file for more details.

## Disclaimer

This software is provided "as is", without warranty of any kind, express or implied. Use it at your own risk.