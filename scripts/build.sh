#!/bin/bash
# Build script for Go MuseTool (Linux/Mac version)
# This script builds a standalone executable with embedded resources

echo "Building Go MuseTool..."

# Set build flags
export CGO_ENABLED=1

# Build flags:
# -ldflags "-s -w"
#   -s: Omit symbol table
#   -w: Omit DWARF debug information
# -trimpath: Remove file system paths from executable

echo "Compiling executable..."
go build -ldflags "-s -w" -trimpath -o ../GoMuseTool ../cmd/Go-MuseTool

if [ $? -eq 0 ]; then
    echo ""
    echo "========================================"
    echo "Build successful!"
    echo "Output: GoMuseTool"
    echo "========================================"
    echo ""
    echo "The executable includes:"
    echo "- Embedded language files (en.json, zh.json)"
    echo "- All dependencies"
    echo ""
    echo "You can now distribute GoMuseTool as a standalone application."
else
    echo ""
    echo "========================================"
    echo "Build failed! Please check the errors above."
    echo "========================================"
fi
