# Inno Setup 图标问题诊断和修复指南

## 🔍 问题诊断

当 Inno Setup 编译时提示"不包含图标"错误,可能的原因有:

### 1. 图标文件路径问题

**当前配置**:
```ini
SetupIconFile=..\\internal\\assets\\real_icon.ico
```

这个路径是相对于 `scripts` 目录的。

**验证路径**:
```powershell
# 在 scripts 目录下运行
cd scripts
Test-Path "..\internal\assets\real_icon.ico"  # 应该返回 True
```

### 2. 图标文件格式问题

**常见问题**:
- 文件不是有效的 ICO 格式
- 文件损坏
- 文件是 PNG 但扩展名改成了 .ico
- 图标尺寸不符合要求

**验证图标文件**:
```powershell
$ico = "..\internal\assets\real_icon.ico"
$bytes = [System.IO.File]::ReadAllBytes((Resolve-Path $ico).Path)

# 检查 ICO 文件头 (应该是 00 00 01 00)
$header = $bytes[0..3]
Write-Host "Header: $([BitConverter]::ToString($header))"

if ($header[0] -eq 0 -and $header[1] -eq 0 -and $header[2] -eq 1 -and $header[3] -eq 0) {
    Write-Host "✓ Valid ICO header"
    $numImages = [BitConverter]::ToUInt16($bytes, 4)
    Write-Host "Number of images: $numImages"
} else {
    Write-Host "✗ Invalid ICO header - file is not a valid ICO file!"
}
```

### 3. Inno Setup 版本问题

某些旧版本的 Inno Setup 对图标格式要求更严格。

## ✅ 解决方案

### 方案 1: 使用绝对路径(临时测试)

修改 `installer.iss`:
```ini
; 使用绝对路径测试
SetupIconFile=C:\Users\manxi\.gemini\antigravity\scratch\gomusetool\Go-MuseTool\internal\assets\real_icon.ico
```

如果这样可以工作,说明是相对路径问题。

### 方案 2: 复制图标到 scripts 目录

```powershell
# 复制图标到 scripts 目录
Copy-Item "..\internal\assets\real_icon.ico" ".\app_icon.ico"
```

然后修改 `installer.iss`:
```ini
SetupIconFile=app_icon.ico
```

### 方案 3: 验证并重新创建图标文件

如果图标文件本身有问题:

1. **检查文件是否真的是 ICO 格式**:
   ```powershell
   # 查看文件头
   Format-Hex "..\internal\assets\real_icon.ico" -Count 16
   ```
   
   应该看到: `00 00 01 00 ...`

2. **使用在线工具重新转换**:
   - 访问: https://www.icoconverter.com/
   - 上传您的源图像
   - 选择尺寸: 16x16, 32x32, 48x48, 256x256
   - 下载新的 ICO 文件
   - 替换 `real_icon.ico`

3. **使用 ImageMagick 转换**(如果已安装):
   ```bash
   magick convert source.png -define icon:auto-resize=256,48,32,16 real_icon.ico
   ```

### 方案 4: 暂时移除图标配置

如果急需构建安装程序,可以暂时注释掉图标配置:

```ini
; 暂时注释掉,使用默认图标
; SetupIconFile=..\\internal\\assets\\real_icon.ico
```

这样安装程序会使用默认图标,但至少可以正常构建。

### 方案 5: 检查文件权限

确保图标文件没有被其他程序锁定:

```powershell
# 检查文件是否被占用
$file = "..\internal\assets\real_icon.ico"
try {
    $stream = [System.IO.File]::Open($file, 'Open', 'Read', 'None')
    $stream.Close()
    Write-Host "✓ File is accessible"
} catch {
    Write-Host "✗ File is locked or inaccessible: $_"
}
```

## 🔧 完整诊断脚本

将以下内容保存为 `diagnose_icon.ps1`:

```powershell
Write-Host "=== Inno Setup Icon Diagnostics ===" -ForegroundColor Cyan
Write-Host ""

# 1. Check file existence
$iconPath = "..\internal\assets\real_icon.ico"
Write-Host "[1] Checking file existence..." -ForegroundColor Yellow
if (Test-Path $iconPath) {
    Write-Host "  ✓ Icon file exists" -ForegroundColor Green
    $fullPath = (Resolve-Path $iconPath).Path
    Write-Host "  Path: $fullPath" -ForegroundColor Gray
} else {
    Write-Host "  ✗ Icon file NOT found!" -ForegroundColor Red
    exit 1
}

# 2. Check file size
Write-Host ""
Write-Host "[2] Checking file size..." -ForegroundColor Yellow
$fileInfo = Get-Item $iconPath
Write-Host "  Size: $($fileInfo.Length) bytes" -ForegroundColor Gray
if ($fileInfo.Length -lt 100) {
    Write-Host "  ⚠️  WARNING: File is very small, may be invalid" -ForegroundColor Red
}

# 3. Check ICO format
Write-Host ""
Write-Host "[3] Checking ICO format..." -ForegroundColor Yellow
$bytes = [System.IO.File]::ReadAllBytes($fullPath)
$header = $bytes[0..3]

if ($header[0] -eq 0 -and $header[1] -eq 0 -and $header[2] -eq 1 -and $header[3] -eq 0) {
    Write-Host "  ✓ Valid ICO file header" -ForegroundColor Green
    $numImages = [BitConverter]::ToUInt16($bytes, 4)
    Write-Host "  Number of images: $numImages" -ForegroundColor Gray
    
    if ($numImages -eq 0 -or $numImages -gt 20) {
        Write-Host "  ✗ Invalid number of images!" -ForegroundColor Red
    } else {
        # List image sizes
        for ($i = 0; $i -lt $numImages; $i++) {
            $offset = 6 + ($i * 16)
            $width = $bytes[$offset]
            $height = $bytes[$offset + 1]
            if ($width -eq 0) { $width = 256 }
            if ($height -eq 0) { $height = 256 }
            Write-Host "    Image $($i+1): ${width}x${height}" -ForegroundColor Gray
        }
    }
} else {
    Write-Host "  ✗ NOT a valid ICO file!" -ForegroundColor Red
    Write-Host "  Header bytes: $([BitConverter]::ToString($header))" -ForegroundColor Gray
    Write-Host "  Expected: 00-00-01-00" -ForegroundColor Gray
}

# 4. Check file accessibility
Write-Host ""
Write-Host "[4] Checking file accessibility..." -ForegroundColor Yellow
try {
    $stream = [System.IO.File]::Open($fullPath, 'Open', 'Read', 'None')
    $stream.Close()
    Write-Host "  ✓ File is accessible" -ForegroundColor Green
} catch {
    Write-Host "  ✗ File is locked or inaccessible" -ForegroundColor Red
    Write-Host "  Error: $_" -ForegroundColor Gray
}

# 5. Recommendations
Write-Host ""
Write-Host "=== Recommendations ===" -ForegroundColor Cyan
if ($numImages -eq 1) {
    Write-Host "⚠️  Icon only contains 1 image size" -ForegroundColor Yellow
    Write-Host "   Recommendation: Create a multi-resolution icon with 16x16, 32x32, 48x48, 256x256" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "=== Quick Fixes ===" -ForegroundColor Cyan
Write-Host "1. Try using absolute path in installer.iss:" -ForegroundColor White
Write-Host "   SetupIconFile=$fullPath" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Or copy icon to scripts directory:" -ForegroundColor White
Write-Host "   Copy-Item '$iconPath' '.\app_icon.ico'" -ForegroundColor Gray
Write-Host "   Then use: SetupIconFile=app_icon.ico" -ForegroundColor Gray
```

运行诊断:
```powershell
cd scripts
powershell -ExecutionPolicy Bypass -File diagnose_icon.ps1
```

## 📝 推荐的修复步骤

1. **运行诊断脚本**确定具体问题
2. **如果是格式问题**,使用在线工具重新创建 ICO 文件
3. **如果是路径问题**,尝试使用绝对路径或复制到 scripts 目录
4. **如果是单一尺寸问题**,创建包含多个尺寸的 ICO 文件

## ⚠️ 注意事项

- Inno Setup 要求 ICO 文件必须是标准的 Windows ICO 格式
- 不能使用 PNG 文件改扩展名为 .ico
- 建议包含多个尺寸以获得最佳显示效果
- 路径中不要包含中文或特殊字符
