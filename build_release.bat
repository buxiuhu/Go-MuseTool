@echo off
REM Release build script for Go MuseTool
REM This script creates a production-ready executable

echo ========================================
echo Building Go MuseTool - Release Version
echo ========================================
echo.

REM Set build flags
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64

REM Check for rsrc tool
echo [1/4] Checking build tools...
if not exist "%USERPROFILE%\go\bin\rsrc.exe" (
    echo Installing rsrc tool...
    go install github.com/akavel/rsrc@latest
)

REM Generate embedded resources
echo [2/4] Generating embedded resources...
where windres >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    windres -i resource.rc -o GoMuseTool.syso -O coff
) else (
    "%USERPROFILE%\go\bin\rsrc" -ico GOMuseTool.ico -o GoMuseTool.syso
)

REM Build release version
echo [3/4] Compiling release executable...
go build -ldflags "-H windowsgui -s -w" -trimpath -o GoMuseTool.exe .

if %ERRORLEVEL% EQU 0 (
    echo [4/4] Verifying build...
    if exist GoMuseTool.exe (
        for %%A in (GoMuseTool.exe) do set size=%%~zA
        echo.
        echo ========================================
        echo BUILD SUCCESSFUL!
        echo ========================================
        echo.
        echo Output file: GoMuseTool.exe
        echo File size: %size% bytes
        echo.
        echo The executable includes:
        echo   - Embedded language files (en.json, zh.json)
        echo   - Embedded application icon
        echo   - All dependencies
        echo   - Optimized for release (stripped symbols)
        echo.
        echo Ready for distribution!
        echo ========================================
    ) else (
        echo ERROR: Executable not found after build!
        exit /b 1
    )
) else (
    echo.
    echo ========================================
    echo BUILD FAILED!
    echo ========================================
    echo Please check the errors above.
    exit /b 1
)
