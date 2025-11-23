# Go MuseTool - 打包说明 / Packaging Guide

## 简体中文

### 概述
Go MuseTool 现已完全打包为独立可执行文件，所有语言文件已嵌入到 EXE 中。

### 已嵌入的资源
- ✅ 语言文件 (en.json, zh.json) - 使用 Go embed 包嵌入
- ✅ 应用程序图标 (GoMuseTool.syso) - 通过资源文件嵌入
- ✅ 所有依赖项 - 静态编译

### 构建方法

#### Windows 系统
运行构建脚本:
```batch
.\build.bat
```

或手动构建:
```batch
go build -ldflags "-H windowsgui -s -w" -trimpath -o GoMuseTool.exe .
```

#### Linux/Mac 系统
运行构建脚本:
```bash
chmod +x build.sh
./build.sh
```

或手动构建:
```bash
go build -ldflags "-s -w" -trimpath -o GoMuseTool .
```

### 编译参数说明
- `-H windowsgui`: 构建为 Windows GUI 应用程序(无控制台窗口)
- `-s`: 省略符号表
- `-w`: 省略 DWARF 调试信息
- `-trimpath`: 从可执行文件中移除文件系统路径

### 分发
生成的 `GoMuseTool.exe` 是一个完全独立的可执行文件,可以直接分发,无需附带任何额外文件。

---

## English

### Overview
Go MuseTool is now fully packaged as a standalone executable with all language files embedded in the EXE.

### Embedded Resources
- ✅ Language files (en.json, zh.json) - Embedded using Go embed package
- ✅ Application icon (GoMuseTool.syso) - Embedded via resource file
- ✅ All dependencies - Statically compiled

### Build Instructions

#### Windows
Run the build script:
```batch
.\build.bat
```

Or build manually:
```batch
go build -ldflags "-H windowsgui -s -w" -trimpath -o GoMuseTool.exe .
```

#### Linux/Mac
Run the build script:
```bash
chmod +x build.sh
./build.sh
```

Or build manually:
```bash
go build -ldflags "-s -w" -trimpath -o GoMuseTool .
```

### Build Flags Explained
- `-H windowsgui`: Build as Windows GUI application (no console window)
- `-s`: Omit symbol table
- `-w`: Omit DWARF debug information
- `-trimpath`: Remove file system paths from executable

### Distribution
The generated `GoMuseTool.exe` is a fully standalone executable that can be distributed directly without any additional files.

---

## Technical Details / 技术细节

### Code Changes / 代码修改
The following changes were made to support embedded resources:

修改了以下代码以支持嵌入式资源:

1. **language/manager.go**
   - Added `//go:embed *.json` directive
   - Changed from `os.ReadFile()` to `languageFiles.ReadFile()`
   - Removed dependency on external language files

2. **Build Scripts**
   - `build.bat` - Windows build script
   - `build.sh` - Linux/Mac build script

### File Structure / 文件结构
```
go musetool/
├── language/
│   ├── en.json          (embedded in exe / 嵌入到exe中)
│   ├── zh.json          (embedded in exe / 嵌入到exe中)
│   └── manager.go       (modified / 已修改)
├── GoMuseTool.syso      (icon resource / 图标资源)
├── build.bat            (Windows build script / Windows构建脚本)
├── build.sh             (Linux/Mac build script / Linux/Mac构建脚本)
└── GoMuseTool.exe       (output / 输出文件)
```

### Requirements / 要求
- Go 1.16+ (for embed support / 支持embed功能)
- CGO enabled for Windows (for Fyne / Fyne需要)
- GCC/MinGW on Windows (for CGO / CGO需要)
