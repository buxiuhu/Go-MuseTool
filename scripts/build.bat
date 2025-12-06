@echo off
REM Build script for Go MuseTool
REM This script builds a standalone executable with embedded resources

echo Building Go MuseTool...

REM Set build flags
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64

REM Build flags:
REM -ldflags "-H windowsgui -s -w"
REM   -H windowsgui: Build as Windows GUI application (no console window)
REM   -s: Omit symbol table
REM   -w: Omit DWARF debug information
REM -trimpath: Remove file system paths from executable

echo Checking for rsrc tool...
if not exist "%USERPROFILE%\go\bin\rsrc.exe" (
    echo Installing rsrc tool...
    go install github.com/akavel/rsrc@latest
)

echo Generating embedded resources (icons)...
REM Try to use windres first (MinGW), fallback to rsrc
where windres >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    windres -i ..\GoMuseTool.rc -o ..\GoMuseTool.syso -O coff
) else (
    "%USERPROFILE%\go\bin\rsrc" -ico ..\internal\assets\real_icon.ico -o ..\GoMuseTool.syso
)

echo Compiling executable...
go build -ldflags "-H windowsgui -s -w" -trimpath -o ..\GoMuseTool.exe ..\cmd\Go-MuseTool

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo Build successful!
    echo Output: GoMuseTool.exe
    echo ========================================
    echo.
    echo The executable includes:
    echo - Embedded language files (en.json, zh.json)
    echo - Embedded application icons
    echo - All dependencies
    echo.
    echo You can now distribute GoMuseTool.exe as a standalone application.
) else (
    echo.
    echo ========================================
    echo Build failed! Please check the errors above.
    echo ========================================
)

pause