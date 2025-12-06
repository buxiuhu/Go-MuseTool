# 版本管理说明

## 版本号位置

为了确保版本号的一致性，本项目采用集中式版本管理方案。

### 主要版本定义

**唯一真实来源**: `internal/version/version.go`

```go
const (
    Version = "0.5.0"  // 应用程序版本号
    AppName = "Go MuseTool"
    Author = "buxiuhu"
    GitHubURL = "https://github.com/buxiuhu/Go-MuseTool"
)
```

### 版本号使用位置

1. **应用程序代码** - `internal/ui/gui.go`
   - 使用 `version.GetVersion()` 获取带 "v" 前缀的版本号
   - 使用 `version.Author` 获取作者信息
   - 使用 `version.GitHubURL` 获取项目链接

2. **安装程序脚本** - `scripts/installer.iss`
   - 手动同步版本号: `#define MyAppVersion "0.5.0"`
   - **注意**: 需要手动更新以保持与 `version.go` 一致

3. **文档** - `BUILD_INSTALLER.md`
   - 示例代码中的版本号引用
   - 仅用于文档说明

## 如何更新版本号

### 步骤 1: 更新主版本定义

编辑 `internal/version/version.go`:

```go
const (
    Version = "0.6.0"  // 修改为新版本号
    // ...
)
```

### 步骤 2: 更新安装程序脚本

编辑 `scripts/installer.iss`:

```inno
#define MyAppVersion "0.6.0"  // 修改为与 version.go 相同的版本号
```

### 步骤 3: (可选) 更新文档

如果需要，更新 `BUILD_INSTALLER.md` 中的示例版本号。

## 版本号格式

本项目遵循 [语义化版本 2.0.0](https://semver.org/lang/zh-CN/) 规范：

```
主版本号.次版本号.修订号 (MAJOR.MINOR.PATCH)
```

- **主版本号**: 不兼容的 API 修改
- **次版本号**: 向下兼容的功能性新增
- **修订号**: 向下兼容的问题修正

### 示例

- `0.5.0` - 当前版本
- `0.5.1` - 修复 bug
- `0.6.0` - 新增功能
- `1.0.0` - 正式发布版本

## 验证版本一致性

在发布前，请确认以下位置的版本号一致：

- [ ] `internal/version/version.go` - Version 常量
- [ ] `scripts/installer.iss` - MyAppVersion 定义
- [ ] 应用程序"关于"对话框显示正确版本
- [ ] 安装程序显示正确版本

## 自动化建议

未来可以考虑：

1. 使用构建脚本自动同步版本号到 `installer.iss`
2. 添加 Git 标签与版本号的关联
3. 使用 CI/CD 自动验证版本号一致性
