package spotify

import (
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"runtime"
)

type PlayerConfig struct {
	// CacheSize sets the max file cache size in MB. File caching is disabled if 0.
	// Defaults to 0.
	CacheSize int `json:"cache_size"`

	// ConfigHomeDir sets the parent directory for all configs & state files. CacheDir is nested inside this path if
	// CacheDir is left unset.
	// Defaults to ${HOME}/.apollo/spotify/ on linux/macos and %userprofile%\AppData\local\apollo\spotify\ on windows.
	ConfigHomeDir string `json:"config_home_dir"`

	// CacheDir sets the directory to be used for cache files. Is only used when CacheSize > 0.
	// Defaults to ${HOME}/.apollo/spotify/cache/ on linux/macos and %userprofile%\AppData\local\apollo\spotify\cache\ on windows.
	CacheDir string `json:"cache_dir"`

	// BitRate sets the bit rate of the downloaded files. This can be 96, 160, or 320.
	// Defaults to 160.
	BitRate int `json:"bit_rate"`

	// PreloadSize sets the number of enqueued songs to download ahead of time (stored in memory if CacheSize == 0)
	// Defaults to 2.
	PreloadSize int `json:"preload_size"`
}

func DefaultPlayerConfig() PlayerConfig {
	configHomeDir := ""

	switch runtime.GOOS {
	case "windows":
		if path, ok := os.LookupEnv("USERPROFILE"); ok {
			configHomeDir = filepath.Join(path, "AppData", "local", "apollo", "spotify")
		}
	default:
		if path, ok := os.LookupEnv("HOME"); ok {
			configHomeDir = filepath.Join(path, ".apollo", "spotify")
		}
	}

	if configHomeDir == "" {
		configHomeDir = filepath.Join(".", ".apollo", "spotify")
		log.Warn().Msg("could not parse home directory for cache, using ./ instead")
	}

	return PlayerConfig{
		CacheSize:     0,
		ConfigHomeDir: configHomeDir,
		CacheDir:      filepath.Join(configHomeDir, "cache"),
		BitRate:       160,
		PreloadSize:   2,
	}
}
