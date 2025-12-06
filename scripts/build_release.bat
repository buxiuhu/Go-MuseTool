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
    windres -i ..\GoMuseTool.rc -o ..\GoMuseTool.syso -O coff
) else (
    "%USERPROFILE%\go\bin\rsrc" -ico ..\icons\GoMuseTool.ico -o ..\GoMuseTool.syso
)

REM Move .syso file to the package directory so go build picks it up
REM Standard behavior: syso must be in the package directory being built
echo Moving resource file to package directory...
copy ..\GoMuseTool.syso ..\cmd\Go-MuseTool\GoMuseTool.syso /Y

REM Set output directory and filename
set OUTPUT_DIR=..\release
set OUTPUT_FILE=GoMuseTool_Windows_X64.exe

REM Create release directory if it doesn't exist
if not exist "%OUTPUT_DIR%" (
    echo Creating release directory...
    mkdir "%OUTPUT_DIR%"
)

REM Build release version
echo [3/4] Compiling release executable...
go build -ldflags "-H windowsgui -s -w" -trimpath -o "%OUTPUT_DIR%\%OUTPUT_FILE%" ..\cmd\Go-MuseTool

if %ERRORLEVEL% EQU 0 (
    echo [4/4] Verifying build...
    if exist "%OUTPUT_DIR%\%OUTPUT_FILE%" (
        for %%A in ("%OUTPUT_DIR%\%OUTPUT_FILE%") do set size=%%~zA
        echo.
        echo ========================================
        echo BUILD SUCCESSFUL!
        echo ========================================
        echo.
        echo Output file: %OUTPUT_DIR%\%OUTPUT_FILE%
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