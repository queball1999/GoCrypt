[Registry]
Root: HKCR; Subkey: "*\shell\QCrypt"; ValueName: MUIVerb; Flags: uninsdeletevalue uninsdeletekeyifempty; ValueType: string; ValueData: "QCrypt";
Root: HKCR; Subkey: "*\shell\QCrypt"; ValueName: SubCommands; Flags: uninsdeletevalue uninsdeletekeyifempty; ValueType: string; ValueData: "QCryptEncrypt;QCryptDecrypt";
Root: HKCR; Subkey: "*\shell\QCrypt"; ValueName: Icon; Flags: uninsdeletevalue uninsdeletekeyifempty; ValueType: string; ValueData: "C:\Program Files (x86)\GoCrypt\gocrypt.exe";
Root: HKCR; Subkey: "Directory\shell\QCrypt"; ValueName: MUIVerb; Flags: uninsdeletevalue uninsdeletekeyifempty; ValueType: string; ValueData: "QCrypt";
Root: HKCR; Subkey: "Directory\shell\QCrypt"; ValueName: SubCommands; Flags: uninsdeletevalue uninsdeletekeyifempty; ValueType: string; ValueData: "QCryptEncrypt;QCryptDecrypt";
Root: HKCR; Subkey: "Directory\shell\QCrypt"; ValueName: Icon; Flags: uninsdeletevalue uninsdeletekeyifempty; ValueType: string; ValueData: "C:\Program Files (x86)\GoCrypt\gocrypt.exe";
Root: HKLM; Subkey: "SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\CommandStore\shell\QCryptEncrypt"; Flags: uninsdeletevalue uninsdeletekeyifempty; ValueType: string; ValueData: "Encrypt";
Root: HKLM; Subkey: "SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\CommandStore\shell\QCryptEncrypt"; ValueName: Icon; Flags: uninsdeletevalue uninsdeletekeyifempty; ValueType: string; ValueData: "C:\Program Files (x86)\GoCrypt\gocrypt.exe";
Root: HKLM; Subkey: "SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\CommandStore\shell\QCryptEncrypt\command"; Flags: uninsdeletevalue uninsdeletekeyifempty; ValueType: string; ValueData: "\";
Root: HKLM; Subkey: "SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\CommandStore\shell\QCryptDecrypt"; Flags: uninsdeletevalue uninsdeletekeyifempty; ValueType: string; ValueData: "Decrypt";
Root: HKLM; Subkey: "SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\CommandStore\shell\QCryptDecrypt"; ValueName: Icon; Flags: uninsdeletevalue uninsdeletekeyifempty; ValueType: string; ValueData: "C:\Program Files (x86)\GoCrypt\gocrypt.exe";
Root: HKLM; Subkey: "SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\CommandStore\shell\QCryptDecrypt\command"; Flags: uninsdeletevalue uninsdeletekeyifempty; ValueType: string; ValueData: "\";
