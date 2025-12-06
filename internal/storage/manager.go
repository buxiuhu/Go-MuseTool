package storage

import (
	"encoding/json"
	"errors"
	"os"

	"go-musetool/internal/model"
)

func LoadConfig(path string) (*model.Config, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &model.Config{Groups: []model.Group{}}, nil
		}
		return nil, err
	}
	defer file.Close()

	var config model.Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func SaveConfig(path string, config *model.Config) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false) // 保留中文字符，不转义为Unicode
	return encoder.Encode(config)
}

func AddShortcut(config *model.Config, groupName string, shortcut model.Shortcut) error {
	for i, group := range config.Groups {
		if group.Name == groupName {
			config.Groups[i].Shortcuts = append(config.Groups[i].Shortcuts, shortcut)
			return nil
		}
	}
	return errors.New("group not found")
}

func UpdateShortcut(config *model.Config, groupName string, oldName string, newShortcut model.Shortcut) error {
	for i, group := range config.Groups {
		if group.Name == groupName {
			for j, s := range group.Shortcuts {
				if s.Name == oldName {
					config.Groups[i].Shortcuts[j] = newShortcut
					return nil
				}
			}
			return errors.New("shortcut not found")
		}
	}
	return errors.New("group not found")
}

func RemoveShortcut(config *model.Config, groupName string, shortcutName string) error {
	for i, group := range config.Groups {
		if group.Name == groupName {
			for j, s := range group.Shortcuts {
				if s.Name == shortcutName {
					config.Groups[i].Shortcuts = append(config.Groups[i].Shortcuts[:j], config.Groups[i].Shortcuts[j+1:]...)
					return nil
				}
			}
			return errors.New("shortcut not found")
		}
	}
	return errors.New("group not found")
}
