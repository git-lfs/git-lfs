#define MyAppName "Git LFS"

#define PathToX86Binary "..\..\git-lfs-x86.exe"
#ifnexist PathToX86Binary
  #pragma error PathToX86Binary + " does not exist, please build it first."
#endif

#define PathToX64Binary "..\..\git-lfs-x64.exe"
#ifnexist PathToX64Binary
  #pragma error PathToX64Binary + " does not exist, please build it first."
#endif

#define PathToARM64Binary "..\..\git-lfs-arm64.exe"
#ifnexist PathToARM64Binary
  #pragma error PathToARM64Binary + " does not exist, please build it first."
#endif

; Arbitrarily choose the x86 executable here as both have the version embedded.
#define MyVersionInfoVersion GetFileVersion(PathToX86Binary)

; Misuse RemoveFileExt to strip the 4th patch-level version number.
#define MyAppVersion RemoveFileExt(MyVersionInfoVersion)

#define MyAppPublisher "GitHub, Inc."
#define MyAppURL "https://git-lfs.github.com/"
#define MyAppFilePrefix "git-lfs-windows"

[Setup]
; NOTE: The value of AppId uniquely identifies this application.
; Do not use the same AppId value in installers for other applications.
; (To generate a new GUID, click Tools | Generate GUID inside the IDE.)
AppCopyright=GitHub, Inc. and Git LFS contributors
AppId={{286391DE-F778-44EA-9375-1B21AAA04FF0}
AppName={#MyAppName}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
AppVersion={#MyAppVersion}
ArchitecturesInstallIn64BitMode=x64 arm64
ChangesEnvironment=yes
Compression=lzma
DefaultDirName={code:GetDefaultDirName}
DirExistsWarning=no
DisableReadyPage=True
LicenseFile=..\..\LICENSE.md
OutputBaseFilename={#MyAppFilePrefix}-{#MyAppVersion}
OutputDir=..\..\
PrivilegesRequired=none
SetupIconFile=git-lfs-logo.ico
SolidCompression=yes
UninstallDisplayIcon={app}\git-lfs.exe
UsePreviousAppDir=no
VersionInfoVersion={#MyVersionInfoVersion}
WizardImageFile=git-lfs-wizard-image.bmp
WizardSmallImageFile=git-lfs-logo.bmp

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Files]
Source: {#PathToX86Binary}; DestDir: "{app}"; Flags: ignoreversion; DestName: "git-lfs.exe"; AfterInstall: InstallGitLFS; Check: IsX86
Source: {#PathToX64Binary}; DestDir: "{app}"; Flags: ignoreversion; DestName: "git-lfs.exe"; AfterInstall: InstallGitLFS; Check: IsX64
Source: {#PathToARM64Binary}; DestDir: "{app}"; Flags: ignoreversion; DestName: "git-lfs.exe"; AfterInstall: InstallGitLFS; Check: IsARM64

[Registry]
Root: HKLM; Subkey: "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"; Check: IsAdminLoggedOn and NeedsAddPath('{app}')
Root: HKLM; Subkey: "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"; ValueType: string; ValueName: "GIT_LFS_PATH"; ValueData: "{app}"; Check: IsAdminLoggedOn
Root: HKCU; Subkey: "Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"; Check: (not IsAdminLoggedOn) and NeedsAddPath('{app}')
Root: HKCU; Subkey: "Environment"; ValueType: string; ValueName: "GIT_LFS_PATH"; ValueData: "{app}"; Check: not IsAdminLoggedOn

[Code]
function GetDefaultDirName(Dummy: string): string;
begin
  if IsAdminLoggedOn then begin
    Result:=ExpandConstant('{pf}\{#MyAppName}');
  end else begin
    Result:=ExpandConstant('{userpf}\{#MyAppName}');
  end;
end;

// Checks to see if we need to add the dir to the env PATH variable.
function NeedsAddPath(Param: string): boolean;
var
  OrigPath: string;
  ParamExpanded: string;
begin
  //expand the setup constants like {app} from Param
  ParamExpanded := ExpandConstant(Param);
  if not RegQueryStringValue(HKEY_LOCAL_MACHINE,
    'SYSTEM\CurrentControlSet\Control\Session Manager\Environment',
    'Path', OrigPath)
  then begin
    Result := True;
    exit;
  end;
  // look for the path with leading and trailing semicolon and with or without \ ending
  // Pos() returns 0 if not found
  Result := Pos(';' + UpperCase(ParamExpanded) + ';', ';' + UpperCase(OrigPath) + ';') = 0;
  if Result = True then
    Result := Pos(';' + UpperCase(ParamExpanded) + '\;', ';' + UpperCase(OrigPath) + ';') = 0;
end;

// Runs the lfs initialization.
procedure InstallGitLFS();
var
  ResultCode: integer;
begin
  Exec(
    ExpandConstant('{cmd}'),
    ExpandConstant('/C ""{app}\git-lfs.exe" install"'),
    '', SW_HIDE, ewWaitUntilTerminated, ResultCode
  );
  if not ResultCode = 1 then
    MsgBox(
    'Git LFS was not able to automatically initialize itself. ' +
    'Please run "git lfs install" from the commandline.', mbInformation, MB_OK);
end;

// Event function automatically called when uninstalling:
function InitializeUninstall(): Boolean;
var
  ResultCode: integer;
begin
  Exec(
    ExpandConstant('{cmd}'),
    ExpandConstant('/C ""{app}\git-lfs.exe" uninstall"'),
    '', SW_HIDE, ewWaitUntilTerminated, ResultCode
  );
  Result := True;
end;
