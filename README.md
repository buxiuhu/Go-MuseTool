
# Go MuseTool

Go MuseTool 是一款轻量级、可定制的应用启动器，使用 **Go 语言**和 **Fyne** 跨平台 GUI 工具包开发。它具备现代化的用户界面、群组管理功能，并支持系统托盘集成。

## 预览图
<img width="1172" height="469" alt="image" src="https://github.com/user-attachments/assets/98ab3494-3a4d-4ed0-b030-65efb0a1a445" />


## 🚀 主要功能

  * **应用启动器 (Application Launcher)**：快速启动您收藏的应用和文件。
  * **群组管理 (Group Management)**：将快捷方式组织到可自定义的群组中。
  * **拖放支持 (Drag & Drop)**：轻松地重新排列快捷方式和群组的顺序。
  * **系统托盘集成 (System Tray Integration)**：可最小化到系统托盘，在后台运行。
  * **可定制 UI (Customizable UI)**：支持浅色/深色主题，以及自定义标题栏颜色。
  * **多语言支持 (Multi-language Support)**：支持英语和简体中文。

## ⚙️ 构建说明

### 先决条件

  * Go 1.21 或更高版本
  * MSYS2 补充mingw64\bin依赖 GCC 编译器（推荐 Windows 用户使用 MinGW-w64）

### 在 Windows 上构建

1.  克隆（Clone）本仓库。
2.  运行构建脚本：
```cmd
build.bat
```

执行后将生成 `GoMuseTool.exe` 可执行文件。

### 手动构建

```bash
go build -ldflags "-H windowsgui -s -w" -trimpath -o GoMuseTool.exe .
```

> *注：`-H windowsgui` 标志用于隐藏 Windows 上的命令行窗口；`-s -w` 用于减小可执行文件体积。*

## 源码运行
1、进入目录

```bash
cd GoMuseTool
```

2、执行启动命令

```bash
go run .
```

## 📦 依赖项

  * [Fyne](https://fyne.io/)：跨平台 GUI 工具包。
  * [sqweek/dialog](https://github.com/sqweek/dialog)：原生系统对话框支持。
  * [akavel/rsrc](https://github.com/akavel/rsrc)：用于嵌入 Windows 资源的工具。




