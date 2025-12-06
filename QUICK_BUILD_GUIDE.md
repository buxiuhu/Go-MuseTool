# 快速构建和测试指南

## 🚀 快速开始

### 步骤 1: 构建安装程序

如果您已安装 Inno Setup,可以使用以下命令:

```bash
# 方法 1: 使用 Inno Setup 编译器(命令行)
"C:\Program Files (x86)\Inno Setup 6\ISCC.exe" scripts\installer.iss

# 方法 2: 使用 Inno Setup GUI
# 1. 打开 Inno Setup
# 2. 文件 → 打开 → 选择 scripts\installer.iss
# 3. 构建 → 编译
```

安装程序将生成在: `scripts\installer_output\GoMuseTool_Windows_setup_X64.exe`

### 步骤 2: 测试安装

1. **卸载旧版本** (如果已安装):
   ```
   设置 → 应用 → 应用和功能 → Go MuseTool → 卸载
   ```

2. **安装新版本**:
   ```
   运行: scripts\installer_output\GoMuseTool_Windows_setup_X64.exe
   ```

3. **验证图标**:
   - ✅ 桌面快捷方式图标
   - ✅ 开始菜单图标
   - ✅ 应用程序窗口图标
   - ✅ 控制面板卸载程序图标

## 📝 验证清单

- [ ] 桌面快捷方式显示图标
- [ ] 开始菜单程序显示图标
- [ ] 运行程序后窗口标题栏显示图标
- [ ] 任务栏显示图标
- [ ] 控制面板"程序和功能"中显示图标

## 🔧 如果需要重新构建应用程序

```bash
cd scripts
.\build_release.bat
```

这将:
1. 生成 `GoMuseTool.syso` (包含图标资源)
2. 编译 `release\GoMuseTool_Windows_X64.exe`

## ⚠️ 常见问题

### 图标缓存问题

如果安装后图标仍不显示,尝试:

1. **刷新图标缓存**:
   ```powershell
   # 重启 Windows Explorer
   Stop-Process -Name explorer -Force
   Start-Process explorer
   ```

2. **或者重启电脑**

### Inno Setup 未安装

下载地址: https://jrsoftware.org/isdl.php

安装后,将安装路径添加到 PATH 环境变量,或使用完整路径调用 ISCC.exe。

## 📦 发布文件

构建完成后,以下文件可用于分发:

- `scripts\installer_output\GoMuseTool_Windows_setup_X64.exe` - 安装程序
- `release\GoMuseTool_Windows_X64.exe` - 独立可执行文件(无需安装)

## 🎯 版本信息

- **应用程序版本**: 0.6.0
- **安装程序版本**: 0.6.0
- **构建日期**: 2025-12-06
