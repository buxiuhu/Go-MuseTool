package model

// Theme constants
const (
	ThemeSystem = "System"
	ThemeLight  = "Light"
	ThemeDark   = "Dark"
)

// Config represents the application configuration structure.
type Config struct {
	ThemePreference string `json:"theme_preference"` // "dark", "light", "system"
	Language        string `json:"language"`         // "en" or "zh"
	TabPosition     string `json:"tab_position"`     // "top", "bottom", "left", "right"
	WindowDecorated bool   `json:"window_decorated"` // true = 有边框, false = 无边框

	// 窗口位置和大小
	WindowX      int `json:"window_x"`      // 窗口X坐标
	WindowY      int `json:"window_y"`      // 窗口Y坐标
	WindowWidth  int `json:"window_width"`  // 窗口宽度
	WindowHeight int `json:"window_height"` // 窗口高度

	// 标题栏颜色 (BGR格式: 0xBBGGRR)
	TitleBarColor uint32 `json:"title_bar_color"` // 自定义标题栏颜色，0表示使用主题默认

	// 窗口透明度 (0.0 - 1.0, 1.0 表示完全不透明)
	Opacity float64 `json:"opacity"` // 窗口透明度，默认 1.0

	// 调试模式
	DebugMode bool `json:"debug_mode"` // 是否开启调试日志

	// 开机自启动
	AutoStart bool `json:"auto_start"` // 是否开机自启动

	// 最小化到托盘
	MinimizeToTray bool `json:"minimize_to_tray"` // 关闭窗口时是否最小化到托盘而不是退出

	// 关闭对话框已显示
	CloseDialogShown bool `json:"close_dialog_shown"` // 是否已显示过首次关闭对话框

	Groups []Group `json:"groups"`
}

// Group represents a tab/group in the launcher.
type Group struct {
	Name      string     `json:"name"`
	Shortcuts []Shortcut `json:"shortcuts"`
}

// Shortcut represents an executable or URL item.
type Shortcut struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	IconPath string `json:"iconPath"` // 绝对路径到提取的图标
}
