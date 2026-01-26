package config

import (
    "flag"
    "os"
    "strconv"
    "sync"
)

type Config struct {
    TorrentPath string
    MoviePath   string
    ShowPath    string
    ManagerPath string
    LogStdout	bool
    DryRun      bool
}

var Load = sync.OnceValue(New)

// New creates and returns a new Config with values loaded from command-line flags and environment variables.
// Precedence order: command-line args > environment variables > defaults.
func New() *Config {
    defaults := &Config{
        TorrentPath: getEnv("ENIACORE_TORRENT_PATH", "/opt/qbit/downloads"),
        MoviePath:   getEnv("ENIACORE_MOVIE_PATH", "/opt/jellyfin/media/movies"),
        ShowPath:    getEnv("ENIACORE_SHOW_PATH", "/opt/jellyfin/media/shows"),
        ManagerPath: getEnv("ENIACORE_MANAGER_PATH", "/opt/media_manager"),
        LogStdout:   getEnvBool("ENIACORE_LOG_STDOUT", true),
        DryRun:      getEnvBool("ENIACORE_DRY_RUN", true),
    }

    // Parse flags with env defaults
    cfg := &Config{}
    flag.StringVar(&cfg.TorrentPath, "torrent-path", defaults.TorrentPath, "Path to downloaded torrents")
    flag.StringVar(&cfg.MoviePath, "movie-path", defaults.MoviePath, "Path to movie library")
    flag.StringVar(&cfg.ShowPath, "show-path", defaults.ShowPath, "Path to show library")
    flag.StringVar(&cfg.ManagerPath, "manager-path", defaults.ManagerPath, "Path to program directory")
    flag.BoolVar(&cfg.LogStdout, "log-stdout", defaults.LogStdout, "Log to standard output")
    flag.BoolVar(&cfg.DryRun, "dry-run", defaults.DryRun, "Run without moving files")
    flag.Parse()

    return cfg
}

// getEnv retrieves an environment variable value or returns the default if not set.
func getEnv(key, defaultVal string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultVal
}

// getEnvBool retrieves a boolean environment variable value or returns the default if not set or invalid.
func getEnvBool(key string, defaultVal bool) bool {
    if value := os.Getenv(key); value != "" {
        if b, err := strconv.ParseBool(value); err == nil {
            return b
        }
    }
    return defaultVal
}
