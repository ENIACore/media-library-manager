package config

import (
    "os"
    "strconv"
	"sync"
)

type Config struct {
    MediaPath	string // Initial location of media files & dirs
    ManagerPath	string // Location of manager dir
    LibraryPath	string // Location to place processed media files & dirs in
    DryRun		bool
}

// Load reads configuration from environment variables with defaults
var Load = sync.OnceValue(New)


func New() *Config {
    return &Config {
        MediaPath:		getEnv("TORRENT_DOWNLOAD_PATH", "/mnt/RAID/qbit-data/downloads"),
		ManagerPath:	getEnv("TORRENT_MANAGER_PATH", "/mnt/RAID/media-manager"),
        LibraryPath:	getEnv("MEDIA_SERVER_PATH", "/mnt/RAID/jelly/media"),
        DryRun:			getEnvBool("TORRENT_MANAGER_DRY_RUN", true),
	}
}

func getEnv(key, defaultVal string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
    if value := os.Getenv(key); value != "" {
        if b, err := strconv.ParseBool(value); err == nil {
            return b
        }
    }
    return defaultVal
}
