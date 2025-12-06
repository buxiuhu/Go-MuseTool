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

// ImportConfigWithIcons imports configuration and icons from a zip archive
func ImportConfigWithIcons(zipPath, appDataDir string) (*model.Config, error) {
	// Open the zip file
	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip file: %w", err)
	}
	defer zipReader.Close()

	var config *model.Config
	iconMapping := make(map[string]string) // old filename -> new path

	// Ensure icons directory exists in app data
	iconsDir := filepath.Join(appDataDir, "icons")
	if err := os.MkdirAll(iconsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create icons directory: %w", err)
	}

	// Process each file in the zip
	for _, file := range zipReader.File {
		if file.Name == "config.json" {
			// Read and parse config.json
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open config.json: %w", err)
			}

			decoder := json.NewDecoder(rc)
			if err := decoder.Decode(&config); err != nil {
				rc.Close()
				return nil, fmt.Errorf("failed to decode config: %w", err)
			}
			rc.Close()

		} else if filepath.Dir(file.Name) == "icons" {
			// Extract icon file
			iconFileName := filepath.Base(file.Name)
			newIconPath := filepath.Join(iconsDir, iconFileName)

			// Open source file in zip
			rc, err := file.Open()
			if err != nil {
				continue // Skip files that can't be opened
			}

			// Create destination file
			destFile, err := os.Create(newIconPath)
			if err != nil {
				rc.Close()
				continue // Skip files that can't be created
			}

			// Copy data
			if _, err := io.Copy(destFile, rc); err != nil {
				rc.Close()
				destFile.Close()
				continue
			}

			rc.Close()
			destFile.Close()

			// Store mapping from filename to new path
			iconMapping[iconFileName] = newIconPath
		}
	}

	if config == nil {
		return nil, fmt.Errorf("config.json not found in zip archive")
	}

	// Update icon paths in config to point to new locations
	for i := range config.Groups {
		for j := range config.Groups[i].Shortcuts {
			oldIconPath := config.Groups[i].Shortcuts[j].IconPath
			if oldIconPath != "" {
				oldFileName := filepath.Base(oldIconPath)
				if newPath, exists := iconMapping[oldFileName]; exists {
					config.Groups[i].Shortcuts[j].IconPath = newPath
				}
			}
		}
	}

	return config, nil
}
