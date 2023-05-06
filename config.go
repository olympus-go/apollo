package apollo

type PlayerConfig struct {
	// CacheSize sets the max file cache size in MB. File caching is disabled if 0.
	// Defaults to 0.
	CacheSize int `json:"cache_size"`

	// CacheDir sets the directory to be used for cache files. Is only used when CacheSize > 0.
	// Defaults to ${HOME}/.apollo/spotify/cache/ on linux/macos and %userprofile%\AppData\local\apollo\spotify\cache\ on windows.
	CacheDir string `json:"cache_dir"`

	// TargetAudioFormat sets the target audio format when the default ffmpeg transcoder is used. This is ignored if
	// a custom transcoder is supplied.
	TargetAudioFormat string

	// PacketBuffer sets the size of the byte buffer used to read packets from enqueued Playable
	PacketBuffer int `json:"packet_buffer"`
}
