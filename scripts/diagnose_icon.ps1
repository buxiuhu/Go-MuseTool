Write-Host "=== Inno Setup Icon Diagnostics ===" -ForegroundColor Cyan
Write-Host ""

# 1. Check file existence
$iconPath = "..\internal\assets\real_icon.ico"
Write-Host "[1] Checking file existence..." -ForegroundColor Yellow
if (Test-Path $iconPath) {
    Write-Host "  ✓ Icon file exists" -ForegroundColor Green
    $fullPath = (Resolve-Path $iconPath).Path
    Write-Host "  Path: $fullPath" -ForegroundColor Gray
}
else {
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
    }
    else {
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
}
else {
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
}
catch {
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
Write-Host ""
Write-Host "3. Test with Inno Setup compiler:" -ForegroundColor White
Write-Host "   ISCC.exe installer.iss" -ForegroundColor Gray
