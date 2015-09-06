Name    "Git LFS"
OutFile "GitLFSInstaller.exe"


InstallDir $DESKTOP

VIProductVersion "0.6.0.0"
VIAddVersionKey ProductName      "Git LFS"
VIAddVersionKey Comments         "Git LFS tool"
VIAddVersionKey CompanyName      "GitHub, Inc"
VIAddVersionKey LegalCopyright   "GitHub, Inc"
VIAddVersionKey FileDescription  "Git LFS Tool"
VIAddVersionKey FileVersion      1
VIAddVersionKey ProductVersion   1
VIAddVersionKey InternalName     "Git LFS"
VIAddVersionKey LegalTrademarks  "something, something, darkside"
VIAddVersionKey OriginalFilename "GitLFSInstaller.exe"

Function .onGUIInit
  IfFileExists $PROGRAMFILES64\Git\mingw64\bin\git.exe 0 +3
    StrCpy $INSTDIR "$PROGRAMFILES64\Git\mingw64\bin"
    Goto exit

  IfFileExists $PROGRAMFILES64\Git\bin\git.exe 0 +3
    StrCpy $INSTDIR "$PROGRAMFILES64\Git\bin"
    Goto exit

  IfFileExists $PROGRAMFILES\Git\bin\git.exe 0 exit
    StrCpy $INSTDIR "$PROGRAMFILES\Git\bin\git.exe"

  exit:
    SetOutPath $INSTDIR
FunctionEnd

Function .onInstSuccess
  MessageBox MB_OK "Open your Git Bash and run 'git lfs init' to get started."
FunctionEnd

Page license
Page directory
Page instfiles
LicenseData ..\..\LICENSE.md

Section
  File git-lfs.exe
  WriteUninstaller $INSTDIR\git-lfs-uninstaller.exe
SectionEnd


Section "Uninstall"
  Delete $INSTDIR\git-lfs-uninstaller.exe
  Delete $INSTDIR\git-lfs.exe
SectionEnd