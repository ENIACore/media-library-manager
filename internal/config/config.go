package config

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	// Both modes
	Mode        string
	Limit       int
	MoviePath   string
	ShowPath    string
	ManagerPath string
	LogStdout   bool

	// Ingest mode
	TorrentPath    string
	IncompletePath string
	Interactive    bool
	DryRun         bool
	TMDBApiKey     string

	// Subtitle mode
	SubtitlePath           string
	OpenSubtitlesApiKey    string
	OpenSubtitlesUserAgent string
	OpenSubtitlesUser      string
	OpenSubtitlesPass      string
}

var Load = sync.OnceValue(New)

// New creates and returns a new Config. Values are resolved in precedence order:
// command-line flags > .env file > hardcoded defaults.
//
// The .env file is searched in order:
//
//	./.env
//	~/.config/media-library-manager/.env
//	/etc/media-library-manager/.env
//	/usr/local/etc/media-library-manager/.env
func New() *Config {
	env := loadEnvFile()

	defaults := &Config{
		Mode:                   getEnv(env, "ENIACORE_MODE", "ingest"),
		TorrentPath:            getEnv(env, "ENIACORE_TORRENT_PATH", "/opt/qbit/downloads"),
		IncompletePath:         getEnv(env, "ENIACORE_INCOMPLETE_PATH", "/opt/qbit/downloads/temp"),
		MoviePath:              getEnv(env, "ENIACORE_MOVIE_PATH", "/opt/jellyfin/media/movies"),
		ShowPath:               getEnv(env, "ENIACORE_SHOW_PATH", "/opt/jellyfin/media/shows"),
		ManagerPath:            getEnv(env, "ENIACORE_MANAGER_PATH", "/opt/media_manager"),
		LogStdout:              getEnvBool(env, "ENIACORE_LOG_STDOUT", true),
		Interactive:            getEnvBool(env, "ENIACORE_INTERACTIVE", true),
		DryRun:                 getEnvBool(env, "ENIACORE_DRY_RUN", true),
		TMDBApiKey:             getEnv(env, "ENIACORE_TMDB_API_KEY", ""),
		Limit:                  getEnvInt(env, "ENIACORE_LIMIT", 10),
		OpenSubtitlesApiKey:    getEnv(env, "ENIACORE_OS_API_KEY", ""),
		OpenSubtitlesUserAgent: getEnv(env, "ENIACORE_OS_USER_AGENT", ""),
		OpenSubtitlesUser:      getEnv(env, "ENIACORE_OS_USER", ""),
		OpenSubtitlesPass:      getEnv(env, "ENIACORE_OS_PASS", ""),
	}

	cfg := &Config{}
	flag.StringVar(&cfg.Mode, "mode", defaults.Mode, "Application mode (ingest or subtitle)")
	flag.StringVar(&cfg.TorrentPath, "torrent-path", defaults.TorrentPath, "Path to downloaded torrents")
	flag.StringVar(&cfg.IncompletePath, "incomplete-path", defaults.IncompletePath, "Path to in-progress (incomplete) torrents")
	flag.StringVar(&cfg.MoviePath, "movie-path", defaults.MoviePath, "Path to movie library")
	flag.StringVar(&cfg.ShowPath, "show-path", defaults.ShowPath, "Path to show library")
	flag.StringVar(&cfg.ManagerPath, "manager-path", defaults.ManagerPath, "Path to program directory")
	flag.BoolVar(&cfg.LogStdout, "log-stdout", defaults.LogStdout, "Log to standard output")
	flag.BoolVar(&cfg.Interactive, "interactive", defaults.Interactive, "User can interactively correct program")
	flag.BoolVar(&cfg.DryRun, "dry-run", defaults.DryRun, "Run without writing files to disk")
	flag.StringVar(&cfg.TMDBApiKey, "tmdb-api-key", defaults.TMDBApiKey, "TMDb API read access token or v3 key")
	flag.IntVar(&cfg.Limit, "limit", defaults.Limit, "Limits number of entries to process (0 = unlimited)")
	flag.StringVar(&cfg.SubtitlePath, "path", "", "Walk a specific directory for missing subtitles instead of the full library")
	flag.StringVar(&cfg.OpenSubtitlesApiKey, "os-api-key", defaults.OpenSubtitlesApiKey, "OpenSubtitles REST API key")
	flag.StringVar(&cfg.OpenSubtitlesUserAgent, "os-user-agent", defaults.OpenSubtitlesUserAgent, "OpenSubtitles user agent")
	flag.StringVar(&cfg.OpenSubtitlesUser, "os-user", defaults.OpenSubtitlesUser, "OpenSubtitles username")
	flag.StringVar(&cfg.OpenSubtitlesPass, "os-pass", defaults.OpenSubtitlesPass, "OpenSubtitles password")
	flag.Parse()

	return cfg
}

// OverLimit reports whether count has exceeded the configured limit.
// A limit of 0 means unlimited.
func (c *Config) OverLimit(count int) bool {
	return c.Limit != 0 && count > c.Limit
}

// loadEnvFile searches the standard locations for a .env file and returns its
// contents as a key→value map. Returns an empty map if no file is found.
func loadEnvFile() map[string]string {
	home, _ := os.UserHomeDir()
	candidates := []string{
		".env",
		filepath.Join(home, ".config", "media-library-manager", ".env"),
		"/etc/media-library-manager/.env",
		"/usr/local/etc/media-library-manager/.env",
	}
	for _, p := range candidates {
		env, err := godotenv.Read(p)
		if err == nil {
			return env
		}
	}
	return map[string]string{}
}

func getEnv(env map[string]string, key, defaultVal string) string {
	if v, ok := env[key]; ok {
		return v
	}
	return defaultVal
}

func getEnvBool(env map[string]string, key string, defaultVal bool) bool {
	if v, ok := env[key]; ok {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return defaultVal
}

func getEnvInt(env map[string]string, key string, defaultVal int) int {
	if v, ok := env[key]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}
