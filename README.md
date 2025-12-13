# Go MuseTool

Go MuseTool is a lightweight, customizable application launcher built with Go and Fyne. It features a modern UI, group management, and system tray integration.

## Features

- **Application Launcher**: Quickly launch your favorite applications and files.
- **Group Management**: Organize shortcuts into customizable groups.
- **Drag & Drop**: Reorder shortcuts and groups with ease.
- **System Tray Integration**: Minimize to tray for background operation.
- **Customizable UI**: Support for light/dark themes and custom title bar colors.
- **Multi-language Support**: English and Chinese (Simplified) support.

## Build Instructions

### Prerequisites

- Go 1.21 or higher
- GCC (MinGW-w64 recommended for Windows)

### Building on Windows

1.  Clone the repository.
2.  Run the build script:
    ```cmd
    build.bat
    ```
    This will generate `GoMuseTool.exe`.

### Manual Build

```bash
go build -ldflags "-H windowsgui -s -w" -trimpath -o GoMuseTool.exe .
```

## Dependencies

- [Fyne](https://fyne.io/): Cross-platform GUI toolkit.
- [sqweek/dialog](https://github.com/sqweek/dialog): Native system dialogs.
- [akavel/rsrc](https://github.com/akavel/rsrc): Tool for embedding Windows resources.

## License

[Apache License](LICENSE)

