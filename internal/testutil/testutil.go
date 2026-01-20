package testutil

import (
	"log/slog"
	"testing"

	"github.com/ENIACore/media_library_manager/internal/config"
	"github.com/ENIACore/media_library_manager/internal/metadata"
)

var logger = slog.Default()

func createTestDir(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()
	
	os.MkdirAll(filepath.Join(tempDir, "source"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "movies"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "shows"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "manager"), 0755)
	
	return tempDir
}

func createTestCfg(testDir string) config.Config {
	return config.Config{
		MoviePath:   filepath.Join(testDir, "movies"),
		ShowPath:    filepath.Join(testDir, "shows"),
		ManagerPath: filepath.Join(testDir, "manager"),
		DryRun:      false,
	}

}
