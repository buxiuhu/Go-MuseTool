package ui

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// ResolveLnkTarget uses PowerShell to find the target path of a Windows shortcut (.lnk) file.
func ResolveLnkTarget(lnkPath string) string {
	// PowerShell command to get the target path of an LNK file
	psScript := fmt.Sprintf(`
        $shell = New-Object -ComObject WScript.Shell
        # 创建快捷方式对象
        $shortcut = $shell.CreateShortcut('%s')
        # 输出目标路径
        Write-Host $shortcut.TargetPath
    `, strings.ReplaceAll(lnkPath, "'", "''"))

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	// Hide the PowerShell window
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Failed to resolve LNK target for %s: %v", lnkPath, err)
		return ""
	}
	// 清理输出中的空格和换行符
	return strings.TrimSpace(string(output))
}

// ExtractIconFromExe extracts the icon from a Windows executable using PowerShell
// Returns the ABSOLUTE path to the extracted icon file (PNG format), or empty string if extraction fails
func ExtractIconFromExe(exePath string) string {
	// Check if file is an exe (此函数只接受 .exe 路径)
	if !strings.HasSuffix(strings.ToLower(exePath), ".exe") {
		return ""
	}

	// Check if exe exists
	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		return ""
	}

	// Create icons directory if it doesn't exist
	iconsDir := filepath.Join(".", "icons")
	if err := os.MkdirAll(iconsDir, 0755); err != nil {
		log.Printf("Failed to create icons directory: %v", err)
		return ""
	}

	// Generate icon filename based on exe name
	exeName := filepath.Base(exePath)
	iconName := strings.TrimSuffix(exeName, filepath.Ext(exeName)) + ".png"
	iconPath := filepath.Join(iconsDir, iconName)

	// 关键修正：获取绝对路径，保证 Fyne 始终能找到图标
	absIconPath, err := filepath.Abs(iconPath)
	if err != nil {
		log.Printf("Failed to get absolute path for icon: %v", err)
		return ""
	}

	// Check if icon already exists (使用绝对路径)
	if _, err := os.Stat(absIconPath); err == nil {
		return absIconPath
	}

	// Use PowerShell to extract icon and convert to PNG at higher resolution
	psScript := fmt.Sprintf(`
		Add-Type -AssemblyName System.Drawing
		$icon = [System.Drawing.Icon]::ExtractAssociatedIcon('%s')
		if ($icon) {
			# Try to get the largest available icon size
			$sizes = @(256, 128, 64, 48, 32)
			$bestIcon = $null
			foreach ($size in $sizes) {
				try {
					$bestIcon = New-Object System.Drawing.Icon($icon, $size, $size)
					break
				} catch {
					continue
				}
			}
			if ($bestIcon -eq $null) {
				$bestIcon = $icon
			}
			$bitmap = $bestIcon.ToBitmap()
			$bitmap.Save('%s', [System.Drawing.Imaging.ImageFormat]::Png)
			$bitmap.Dispose()
			$bestIcon.Dispose()
			if ($bestIcon -ne $icon) {
				$icon.Dispose()
			}
		}
	`, strings.ReplaceAll(exePath, "'", "''"), strings.ReplaceAll(absIconPath, "'", "''"))

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	// Hide the PowerShell window
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to extract icon using PowerShell: %v, output: %s", err, string(output))
		return ""
	}

	// Verify icon was created (使用绝对路径)
	if _, err := os.Stat(absIconPath); err != nil {
		log.Printf("Icon file was not created: %v", err)
		return ""
	}

	log.Printf("Successfully extracted icon to: %s", absIconPath)
	return absIconPath
}

// GetExecutableInfo returns basic info about an executable
func GetExecutableInfo(exePath string) (name string, version string) {
	// Get base name
	name = filepath.Base(exePath)

	// 循环移除后缀，直到没有 .lnk 或 .exe 为止
	// 这样做可以处理 "Game.exe.lnk" 这种情况，同时保留文件名的大小写
	for {
		ext := filepath.Ext(name)
		if ext == "" {
			break
		}
		lowerExt := strings.ToLower(ext)
		if lowerExt == ".lnk" || lowerExt == ".exe" {
			name = strings.TrimSuffix(name, ext)
		} else {
			break
		}
	}

	// Remove spaces, newlines, and other whitespace characters
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "\n", "")
	name = strings.ReplaceAll(name, "\r", "")
	name = strings.ReplaceAll(name, "\t", "")

	return name, ""
}
