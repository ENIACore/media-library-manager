//go:build integration

package enhancer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ENIACore/media_library_manager/internal/config"
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/joho/godotenv"
)

// TestFetchSubtitle_integration downloads a real subtitle from OpenSubtitles using
// credentials from the nearest .env file. Run with: go test -tags integration ./internal/enhancer/
func TestFetchSubtitle_integration(t *testing.T) {
	env, err := godotenv.Read(findEnvFile(t))
	if err != nil {
		t.Fatalf("failed to read .env: %v", err)
	}

	cfg := &config.Config{
		OpenSubtitlesApiKey:    env["ENIACORE_OS_API_KEY"],
		OpenSubtitlesUserAgent: env["ENIACORE_OS_USER_AGENT"],
		OpenSubtitlesUser:      env["ENIACORE_OS_USER"],
		OpenSubtitlesPass:      env["ENIACORE_OS_PASS"],
		DryRun:                 false,
	}

	if cfg.OpenSubtitlesApiKey == "" || cfg.OpenSubtitlesUser == "" {
		t.Skip("OpenSubtitles credentials not set in .env")
	}

	session, err := Login(cfg, noopLogger())
	if err != nil {
		t.Fatalf("Login() error: %v", err)
	}

	// The Dark Knight (2008) — well-known, reliable subtitle coverage on OpenSubtitles.
	dest := filepath.Join(t.TempDir(), "The Dark Knight.English.srt")
	entry := &metadata.Entry{
		Role:      metadata.MovieFile,
		MediaInfo: metadata.MediaInfo{TMDBid: 155},
		FileInfo:  metadata.FileInfo{DestPath: dest},
	}

	if err := FetchSubtitle(entry, session, cfg, noopLogger()); err != nil {
		t.Fatalf("FetchSubtitle() error: %v", err)
	}

	info, err := os.Stat(dest)
	if err != nil {
		t.Fatalf("subtitle file not written: %v", err)
	}
	if info.Size() == 0 {
		t.Error("subtitle file is empty")
	}
	t.Logf("downloaded %d bytes → %s", info.Size(), dest)
}

// findEnvFile walks up the directory tree from the test's working directory
// until it finds a .env file, skipping the test if none exists.
func findEnvFile(t *testing.T) string {
	t.Helper()
	dir, _ := os.Getwd()
	for {
		candidate := filepath.Join(dir, ".env")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Skip("no .env file found, skipping integration test")
	return ""
}
