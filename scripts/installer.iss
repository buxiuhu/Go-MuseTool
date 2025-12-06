; Inno Setup Script for Go-MuseTool
; 使用 Inno Setup 编译此脚本以创建安装程序
; 下载 Inno Setup: https://jrsoftware.org/isdl.php

#define MyAppName "Go MuseTool"
#define MyAppVersion "0.6.0"
#define MyAppPublisher "buxiuhu"
#define MyAppURL "https://github.com/buxiuhu/Go-MuseTool"
#define MyAppExeName "GoMuseTool.exe"

[Setup]
; 应用程序基本信息
AppId={{A1B2C3D4-E5F6-4A5B-8C9D-0E1F2A3B4C5D}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
AllowNoIcons=yes
; 输出设置
OutputDir=installer_output
OutputBaseFilename=GoMuseTool_Windows_setup_X64
Compression=lzma2/max
SolidCompression=yes
; 安装程序图标
SetupIconFile=..\icons\GoMuseTool.ico
; 卸载程序图标 - 使用安装目录下的图标文件
UninstallDisplayIcon={app}\GoMuseTool.ico
; 权限设置
PrivilegesRequired=lowest
; 界面设置
WizardStyle=modern
DisableWelcomePage=no
LicenseFile=..\LICENSE
; 如果没有 LICENSE 文件，可以注释掉上面这行

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"
Name: "chinesesimplified"; MessagesFile: "compiler:Languages\ChineseSimplified.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked
Name: "quicklaunchicon"; Description: "{cm:CreateQuickLaunchIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked; OnlyBelowVersion: 6.1; Check: not IsAdminInstallMode

[Files]
; 主程序文件, 从构建输出目录获取并重命名
Source: "..\release\GoMuseTool_Windows_X64.exe"; DestDir: "{app}"; DestName: "{#MyAppExeName}"; Flags: ignoreversion
; 包含图标文件
Source: "..\icons\GoMuseTool.ico"; DestDir: "{app}"; Flags: ignoreversion
; 语言文件已嵌入，无需单独分发
; 其他必要文件
; Source: "README.md"; DestDir: "{app}"; Flags: ignoreversion isreadme

[Icons]
; 开始菜单快捷方式 - 使用独立的图标文件
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; IconFilename: "{app}\GoMuseTool.ico"
Name: "{group}\{cm:UninstallProgram,`{#MyAppName}`}"; Filename: "{uninstallexe}"; IconFilename: "{app}\GoMuseTool.ico"
; 桌面快捷方式 - 使用独立的图标文件
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; IconFilename: "{app}\GoMuseTool.ico"; Tasks: desktopicon
; 快速启动栏快捷方式 - 使用独立的图标文件
Name: "{userappdata}\Microsoft\Internet Explorer\Quick Launch\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; IconFilename: "{app}\GoMuseTool.ico"; Tasks: quicklaunchicon

[Run]
; 安装完成后运行程序（可选）
Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,`{#StringChange(MyAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent

[Code]
// 检查是否已经有实例在运行
function InitializeSetup(): Boolean;
begin
  Result := True;
end;

// 卸载时的清理工作
procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
var
  ConfigPath: String;
begin
  if CurUninstallStep = usPostUninstall then
  begin
    // 询问是否删除配置文件
    ConfigPath := ExpandConstant('{app}\config.json');
    if FileExists(ConfigPath) then
    begin
      if MsgBox('是否删除配置文件？' + #13#10 + '(Do you want to delete configuration files?)', mbConfirmation, MB_YESNO) = IDYES then
      begin
        DeleteFile(ConfigPath);
      end;
    end;
  end;
end;