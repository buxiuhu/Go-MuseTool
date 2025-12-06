package version

// Version information for Go-MuseTool
const (
	// Version is the current version of the application
	Version = "0.6.0"

	// AppName is the application name
	AppName = "Go MuseTool"

	// Author is the application author
	Author = "buxiuhu"

	// GitHubURL is the project's GitHub repository URL
	GitHubURL = "https://github.com/buxiuhu/Go-MuseTool"
)

// GetVersion returns the version string with 'v' prefix
func GetVersion() string {
	return "v" + Version
}

// GetFullVersion returns the full version string with app name
func GetFullVersion() string {
	return AppName + " v" + Version
}
