package spotify

import (
	"os"
	"path/filepath"
	"runtime"
)

type SessionConfig struct {
	// ConfigHomeDir sets the parent directory for all configs & state files. CacheDir is nested inside this path if
	// CacheDir is left unset.
	// Defaults to ${HOME}/.apollo/spotify/ on linux/macos and %userprofile%\AppData\local\apollo\spotify\ on windows.
	ConfigHomeDir string `json:"config_home_dir"`

	// CacheSize sets the max file cache size in MB. File caching is disabled if 0.
	// Defaults to 0.
	CacheSize int `json:"cache_size"`

	// CacheDir sets the directory to be used for cache files. Is only used when CacheSize > 0.
	// Defaults to ${HOME}/.apollo/spotify/cache/ on linux/macos and %userprofile%\AppData\local\apollo\spotify\cache\ on windows.
	CacheDir string `json:"cache_dir"`

	// OAuthCallback sets the callback address for oauth logins
	// Defaults to "" (http://localhost:8888/callback).
	OAuthCallback string `json:"oauth_callback"`
}

func DefaultSessionConfig() SessionConfig {
	configHomeDir := ""

	switch runtime.GOOS {
	case "windows":
		if path, ok := os.LookupEnv("USERPROFILE"); ok {
			configHomeDir = filepath.Join(path, "AppData", "local", "apollo", "spotify")
		}
	default:
		if path, ok := os.LookupEnv("HOME"); ok {
			configHomeDir = filepath.Join(path, ".local", "apollo", "spotify")
		}
	}

	if configHomeDir == "" {
		configHomeDir = filepath.Join(".", ".apollo", "spotify")
	}

	return SessionConfig{
		ConfigHomeDir: configHomeDir,
		CacheSize:     0,
		CacheDir:      filepath.Join(configHomeDir, "cache"),
		OAuthCallback: "",
	}
}
