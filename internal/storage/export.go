package storage

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"go-musetool/internal/model"
)

// ExportConfigWithIcons exports the configuration and all icon files to a zip archive
func ExportConfigWithIcons(zipPath string, config *model.Config) error {
	// Create the zip file
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Write config.json to zip
	configWriter, err := zipWriter.Create("config.json")
	if err != nil {
		return fmt.Errorf("failed to create config.json in zip: %w", err)
	}

	encoder := json.NewEncoder(configWriter)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false) // 保留中文字符，不转义为Unicode
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	// Collect all unique icon paths
	iconPaths := make(map[string]bool)
	for _, group := range config.Groups {
		for _, shortcut := range group.Shortcuts {
			if shortcut.IconPath != "" {
				iconPaths[shortcut.IconPath] = true
			}
		}
	}

	// Copy each icon file to the zip archive
	for iconPath := range iconPaths {
		// Check if icon file exists
		if _, err := os.Stat(iconPath); os.IsNotExist(err) {
			// Skip missing icons but log it
			continue
		}

		// Open the icon file
		iconFile, err := os.Open(iconPath)
		if err != nil {
			// Skip files that can't be opened
			continue
		}

		// Get the base filename
		iconFileName := filepath.Base(iconPath)

		// Create entry in zip under icons/ directory
		iconWriter, err := zipWriter.Create("icons/" + iconFileName)
		if err != nil {
			iconFile.Close()
			return fmt.Errorf("failed to create icon entry in zip: %w", err)
		}

		// Copy icon data
		if _, err := io.Copy(iconWriter, iconFile); err != nil {
			iconFile.Close()
			return fmt.Errorf("failed to copy icon data: %w", err)
		}

		iconFile.Close()
	}

	return nil
}
