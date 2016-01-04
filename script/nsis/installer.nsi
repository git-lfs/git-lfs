Name    "Git LFS"
OutFile "GitLFSInstaller.exe"

InstallDir $DESKTOP

VIProductVersion "1.1.0.0"
VIAddVersionKey FileVersion      "1.1.0.0"
VIAddVersionKey ProductVersion   "1.1.0.0"
VIAddVersionKey ProductName      "Git LFS"
VIAddVersionKey Comments         "Git LFS"
VIAddVersionKey CompanyName      "GitHub, Inc"
VIAddVersionKey FileDescription  "Git LFS extension"
VIAddVersionKey InternalName     "Git LFS"
VIAddVersionKey LegalCopyright   "(c) GitHub, Inc. and Git LFS contributors"
VIAddVersionKey LegalTrademarks  "(c) GitHub, Inc. and Git LFS contributors"
VIAddVersionKey OriginalFilename "GitLFSInstaller.exe"

Function .onGUIInit
  Var /GLOBAL HASPROXYGIT
  StrCpy $HASPROXYGIT "no"
  IfFileExists $PROGRAMFILES64\Git\cmd\git.exe 0 +2
    StrCpy $HASPROXYGIT "yes"

  ; Git for Windows location
  IfFileExists $PROGRAMFILES64\Git\mingw64\bin\git.exe 0 +3
    StrCpy $INSTDIR "$PROGRAMFILES64\Git\mingw64\bin"
    Goto exit

  IfFileExists $PROGRAMFILES64\Git\bin\git.exe 0 +3
    StrCpy $INSTDIR "$PROGRAMFILES64\Git\bin"
    Goto exit

  IfFileExists $PROGRAMFILES\Git\bin\git.exe 0 +2
    StrCpy $INSTDIR "$PROGRAMFILES\Git\bin"

  exit:
    SetOutPath $INSTDIR
FunctionEnd

Function .onInstSuccess
  IfSilent +2
    MessageBox MB_OK "Open Git Bash and run 'git lfs install' to get started."
FunctionEnd

Page license
Page directory
Page instfiles
LicenseData ..\..\LICENSE.md

Section
  File git-lfs.exe

  StrCmp $HASPROXYGIT "yes" 0 +3
    SetOutPath $PROGRAMFILES64\Git\cmd
    File git-lfs.exe

  WriteUninstaller $INSTDIR\git-lfs-uninstaller.exe
SectionEnd


Section "Uninstall"
  Delete $INSTDIR\git-lfs-uninstaller.exe
  Delete $INSTDIR\git-lfs.exe

  StrCmp $HASPROXYGIT "yes" 0
    Delete $PROGRAMFILES64\Git\cmd\git-lfs.exe
SectionEnd
