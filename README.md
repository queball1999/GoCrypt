**This project is under development and may not work properly. Please report any issues**

_GoCrypt_ is a CLI and GUI file encryption tool specializing in fast, convenient, and secure file encryption. With seamless integration into the Windows context menu, _GoCrypt_ allows you to encrypt and decrypt files directly from your file explorer with ease. Leveraging Go and the ChaCha20-Poly1305 encryption algorithm, _GoCrypt_ boasts high performance and robust security.

### CLI Comands
`encrypt`, `enc`, `e` - encrypt provided files.

`decrypt`, `dec`, `d` - decrypt provided files.

### CLI Flags
`--output, -o` - Specify the output directory. By default _gocrypt_ will place the output file in the same directory as it was pulled from.

`--no-ui, -n` - Disable the GUI. By default _gocrypt_ will use the GUI for user interaction.

`--method, -m` - Specify the encryption method. Currently supports chacha20poly1305 and aes.

*IMPORTANT* - These flags MUST be passed _before_ the file arguments. Please refer to examples below.

### CLI Examples
```
./gocrypt encrypt secret.txt
```
```
./gocrypt --no-ui encrypt secret.txt contract.pdf music.mp3
```
```
./gocrypt -n encrypt secret.txt
```
```
./gocrypt -m chacha20poly1305 -o "C:/output.txt" encrypt secret.txt
```

### GUI Usage

Once installed, simply right-click on any file or folder in Windows Explorer and select the GoCrypt option from the context menu. You will be prompted to enter a password to either encrypt or decrypt the selected item.

_GoCrypt_ offers a GUI for interacting with files and by default is enabled for all interactions through the CLI, context menu, or clicking directly on the desktop icon. 

When launching _GoCrypt_ without any command line parameters, the user will be presented with a window allowing them to choose files for encryption.

### Installation

For the latest pre-built binaries of the application, check the [Releases](https://github.com/queball1999/GoCrypt/releases) page on the repository.

### Compiling

The installer is generated using [Inno Setup](https://jrsoftware.org/isinfo.php) and the config file can be found in this repo.

### Disclaimer

This software is provided "as is", without warranty of any kind, express or implied. Use it at your own risk.

### Acknowledgements

This software derives inspiration from [70sh1's](https://github.com/70sh1) project [eddy](https://github.com/70sh1/eddy/tree/main).