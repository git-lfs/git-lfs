:::::::::::::::::::::::::::::::::::::::::
:: Automatically check & get admin rights
:::::::::::::::::::::::::::::::::::::::::
@echo off
CLS

:: Check if git.exe is in the user's path before continuing
where /q git.exe
if %errorlevel% neq 0 (ECHO Unable to find git.exe, exiting... & EXIT /b %errorlevel%)

:checkPrivileges
NET FILE 1>NUL 2>NUL
if '%errorlevel%' == '0' ( goto gotPrivileges ) else ( goto getPrivileges )

:getPrivileges
if '%1'=='ELEV' (shift & goto gotPrivileges)
echo.
echo **************************************
echo Installing Git LFS as Administrator
echo **************************************

setlocal DisableDelayedExpansion
set "batchPath=%~0"
setlocal EnableDelayedExpansion
echo Set UAC = CreateObject^("Shell.Application"^) > "%temp%\OEgetPrivileges.vbs"
echo UAC.ShellExecute "!batchPath!", "ELEV", "", "runas", 1 >> "%temp%\OEgetPrivileges.vbs"
"%temp%\OEgetPrivileges.vbs"
exit /B

:gotPrivileges

setlocal & cd /d %~dp0

set GIT_LFS_BIN_PATH="%LOCALAPPDATA%\GitLFS\bin"
IF EXIST %GIT_LFS_BIN_PATH% GOTO DIRECTORY_EXISTS
mkdir %GIT_LFS_BIN_PATH%
set "path=%PATH%;%GIT_LFS_BIN_PATH:"=%"
1>NUL setx PATH "%PATH%" /M
:DIRECTORY_EXISTS

:: Delete any existing git-lfs programs
2>NUL del /q %GIT_LFS_BIN_PATH%\git-lfs*

1>NUL copy git-lfs.exe %GIT_LFS_BIN_PATH%\git-lfs.exe

git lfs init

cmd /k
