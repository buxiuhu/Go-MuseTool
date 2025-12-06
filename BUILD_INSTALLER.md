# 创建安装程序指南

本文档说明如何为 Go-MuseTool 创建 Windows 安装程序。

## 方法一：使用 Inno Setup（推荐）

### 1. 下载并安装 Inno Setup

访问 [Inno Setup 官网](https://jrsoftware.org/isdl.php) 下载最新版本。

推荐下载：**Inno Setup 6.x** (包含 Unicode 支持)

### 2. 准备文件

确保以下文件存在：
- `Go-MuseTool.exe` - 主程序（已编译）
- `language/en.json` - 英文语言文件
- `language/zh.json` - 中文语言文件
- `real_icon.ico` - 应用图标
- `LICENSE` - 许可证文件（可选，如果没有可以在 installer.iss 中注释掉）

### 3. 编译安装程序

1. 右键点击 `installer.iss` 文件
2. 选择 "Compile" (使用 Inno Setup 编译)
3. 等待编译完成

编译完成后，安装程序将生成在 `installer_output` 目录中：
- 文件名：`GoMuseTool_Windows_setup_X64.exe`

### 4. 测试安装程序

运行生成的安装程序，测试：
- 安装过程是否正常
- 程序是否能正确运行
- 卸载是否干净

## 方法二：使用 NSIS

如果您更喜欢 NSIS，可以使用以下步骤：

### 1. 下载 NSIS

访问 [NSIS 官网](https://nsis.sourceforge.io/Download) 下载。

### 2. 创建 NSIS 脚本

需要手动编写 `.nsi` 脚本文件（比 Inno Setup 复杂一些）。

## 方法三：使用 WiX Toolset

WiX 是微软推荐的安装程序制作工具，但学习曲线较陡。

适合需要高级功能的场景，如：
- MSI 安装包
- Windows Installer 集成
- 企业部署

## 安装程序功能

当前的 Inno Setup 脚本包含以下功能：

✅ **基本功能**
- 安装到 Program Files
- 创建开始菜单快捷方式
- 可选创建桌面快捷方式
- 卸载程序

✅ **语言支持**
- 英文界面
- 简体中文界面

✅ **智能卸载**
- 卸载时询问是否删除配置文件
- 可选保留用户数据

## 自定义安装程序

编辑 `installer.iss` 文件可以自定义：

1. **应用信息**
   ```
   #define MyAppVersion "0.5.0"  // 修改版本号
   ```

2. **安装选项**
   - 默认安装路径
   - 是否创建桌面图标
   - 是否自动启动

3. **包含的文件**
   - 在 `[Files]` 部分添加或删除文件

4. **安装后操作**
   - 在 `[Run]` 部分配置安装后自动运行

## 发布检查清单

在发布安装程序前，请确认：

- [ ] 已更新版本号
- [ ] 已测试安装过程
- [ ] 已测试卸载过程
- [ ] 已测试程序功能
- [ ] 已准备 README 和 LICENSE 文件
- [ ] 安装程序文件名清晰（包含版本号）

## 常见问题

**Q: 安装程序太大？**
A: Inno Setup 已使用最大压缩率。如果仍然太大，考虑：
- 移除不必要的文件
- 使用在线安装程序（下载组件）

**Q: 需要管理员权限？**
A: 当前配置为 `PrivilegesRequired=lowest`，不需要管理员权限。
   如果需要写入 Program Files，改为 `admin`。

**Q: 如何添加数字签名？**
A: 需要购买代码签名证书，然后在 `[Setup]` 部分添加：
   ```
   SignTool=signtool sign /f "path\to\certificate.pfx" /p "password" $f
   ```

## 其他资源

- [Inno Setup 文档](https://jrsoftware.org/ishelp/)
- [Inno Setup 示例脚本](https://jrsoftware.org/isinfo.php)
- [NSIS 文档](https://nsis.sourceforge.io/Docs/)
