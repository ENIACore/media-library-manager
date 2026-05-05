package enhancer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ENIACore/media_library_manager/internal/config"
)

const (
	sessionExpiry       = 24 * time.Hour
	sessionSafetyMargin = time.Hour
)

type sessionCache struct {
	Token     string    `json:"token"`
	BaseURL   string    `json:"base_url"`
	ExpiresAt time.Time `json:"expires_at"`
}

func cachePath(cfg *config.Config) string {
	return filepath.Join(cfg.ManagerPath, "cache", "opensubtitles.json")
}

// LoadCachedSession reads the session cache from disk and returns a Session if the
// stored token hasn't expired (with a 1-hour safety margin). Returns nil if the
// cache is missing, malformed, or expired.
func LoadCachedSession(cfg *config.Config) *Session {
	data, err := os.ReadFile(cachePath(cfg))
	if err != nil {
		return nil
	}

	var cache sessionCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil
	}

	if time.Until(cache.ExpiresAt) < sessionSafetyMargin {
		return nil
	}

	return &Session{Token: cache.Token, BaseURL: cache.BaseURL}
}

// SaveSession persists the session to the cache file with a 24-hour expiry.
func SaveSession(session *Session, cfg *config.Config) error {
	path := cachePath(cfg)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("enhancer: failed to create cache directory: %w", err)
	}

	cache := sessionCache{
		Token:     session.Token,
		BaseURL:   session.BaseURL,
		ExpiresAt: time.Now().Add(sessionExpiry),
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

// InvalidateCache removes the session cache file.
func InvalidateCache(cfg *config.Config) {
	os.Remove(cachePath(cfg))
}
