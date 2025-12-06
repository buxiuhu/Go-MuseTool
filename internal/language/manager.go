package language

import (
	"embed"
	"encoding/json"
	"log"
	"sync"
)

//go:embed *.json
var languageFiles embed.FS

// Translations holds all translatable strings
type Translations struct {
	// Window
	WindowTitle string

	// Toolbar
	ToolbarAddShortcut string
	ToolbarNewGroup    string
	ToolbarEditGroup   string
	ToolbarDeleteGroup string
	ToolbarSettings    string

	// Settings Dialog
	SettingsTitle                string
	SettingsTheme                string
	SettingsTabPosition          string
	SettingsOpacity              string
	SettingsOpacityTitle         string
	SettingsOpacityHint          string
	SettingsLanguage             string
	SettingsDebugLog             string
	SettingsAutoStart            string
	SettingsMinimizeToTray       string
	SettingsResetCloseDialog     string
	SettingsResetCloseDialogDesc string
	SettingsDataManagement       string
	SettingsExport               string
	SettingsImport               string
	SettingsExportSuccess        string
	SettingsImportSuccess        string
	SettingsImportSuccessRestart string
	SettingsImportError          string
	SettingsSave                 string
	SettingsCancel               string
	SettingsClose                string

	// Theme Options
	ThemeSystem string
	ThemeLight  string
	ThemeDark   string

	// Tab Position Options
	TabPosTop    string
	TabPosLeft   string
	TabPosRight  string
	TabPosBottom string

	// Group Management
	GroupAdd           string
	GroupAddTitle      string
	GroupAddLabel      string
	GroupAddButton     string
	GroupRename        string
	GroupRenameTitle   string
	GroupRenameLabel   string
	GroupDelete        string
	GroupDeleteTitle   string
	GroupDeleteConfirm string
	GroupCannotDelete  string
	GroupMustHaveOne   string
	GroupAlreadyExists string
	GroupNameEmpty     string

	// Shortcut Management
	ShortcutEdit          string
	ShortcutDelete        string
	ShortcutDeleteTitle   string
	ShortcutDeleteConfirm string
	ShortcutMoveTo        string
	ShortcutAdd           string
	ShortcutAddTitle      string
	ShortcutEditTitle     string
	ShortcutName          string
	ShortcutPath          string
	ShortcutIcon          string
	ShortcutBrowse        string
	ShortcutBrowseExe     string
	ShortcutBrowseIcon    string
	ShortcutSave          string
	ShortcutCancel        string
	ShortcutNameRequired  string

	// Context Menu
	ContextMenuNewGroup     string
	ContextMenuRenameGroup  string
	ContextMenuDeleteGroup  string
	ContextMenuMoveLeft     string
	ContextMenuMoveRight    string
	ContextMenuOpenLocation string

	// Common
	Error   string
	Success string
	Confirm string
	Add     string
	Save    string
	Cancel  string
	Close   string
	Browse  string

	// Tray
	TrayShow string
	TrayExit string

	// Close Dialog
	CloseDialogTitle    string
	CloseDialogMessage  string
	CloseDialogMinimize string
	CloseDialogExit     string
	CloseDialogRemember string

	// About Dialog
	AboutTitle   string
	AboutVersion string
	AboutAuthor  string
	AboutLink    string
}

var (
	currentLang  = "en"
	translations = &Translations{}
	mu           sync.RWMutex
)

// Load loads the translations for the specified language
func Load(lang string) error {
	mu.Lock()
	defer mu.Unlock()

	// Default to English if empty
	if lang == "" {
		lang = "en"
	}

	filename := lang + ".json"
	data, err := languageFiles.ReadFile(filename)
	if err != nil {
		// Try to load English as fallback if requested language fails
		if lang != "en" {
			log.Printf("failed to load language %s: %v, falling back to en", lang, err)
			return Load("en")
		}
		return err
	}

	var t Translations
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}

	translations = &t
	currentLang = lang
	return nil
}

// T returns the current translations
func T() *Translations {
	mu.RLock()
	defer mu.RUnlock()
	return translations
}

// GetLanguage returns the current language code
func GetLanguage() string {
	mu.RLock()
	defer mu.RUnlock()
	return currentLang
}

// CurrentLanguage returns the current language code (alias for GetLanguage)
func CurrentLanguage() string {
	return GetLanguage()
}
