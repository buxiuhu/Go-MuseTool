package main

import (
	_ "embed"
	"log"

	"go-musetool/internal/assets"
	"go-musetool/internal/logger"
	"go-musetool/internal/storage"
	"go-musetool/internal/ui"

	"fyne.io/fyne/v2"
)

// IconData is now imported from assets package
var iconData = assets.IconData

func main() {
	// Set standard log format to match logger package
	log.SetFlags(log.Ldate | log.Ltime)
	log.SetPrefix("【SYSTEM】")

	// Check if another instance is already running
	if !ui.CheckSingleInstance() {
		// Another instance is running, show dialog and exit
		ui.ShowAlreadyRunningDialog()
		return
	}
	defer ui.ReleaseSingleInstance()

	// Load configuration first to get debug mode setting
	config, err := storage.LoadConfig("config.json")
	if err != nil {
		log.Printf("Warning: Could not load config: %v. Starting with default config.", err)
		// Proceed with empty config if load fails
	}

	// Initialize logging with debug mode from config
	if err := logger.Setup(config.DebugMode); err != nil {
		log.Fatalf("Failed to setup logger: %v", err)
	}
	defer logger.Close()

	logger.Info("Starting Go MuseTool...")
	logger.Debug("Debug mode: %v", config.DebugMode)

	// Initialize and run UI
	logger.Info("Initializing UI...")

	// Pass icon data to NewLauncherApp so it's available when tray initializes
	app := ui.NewLauncherApp(config, "config.json", iconData)

	// Load and set application icon from embedded resource
	iconResource := fyne.NewStaticResource("icon.ico", iconData)
	app.App.SetIcon(iconResource)
	app.Window.SetIcon(iconResource)
	logger.Info("Application icon loaded from embedded resource")

	logger.Info("UI Initialized. Running app...")
	app.Run()
}
