package transfer

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/ENIACore/media_library_manager/internal/testutil"
)

var logger = slog.Default()

func TestTransfer_MovieFile(t *testing.T) {
	testDir := testutil.CreateTestDir(t)
	cfg := testutil.CreateTestCfg(testDir)

	entry := testutil.CreateTestMovieFile(nil)
	testutil.CreateTestFiles(entry, testDir)

	err := Transfer(entry, &cfg, logger)
	if err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}

	destPath := filepath.Join(cfg.MoviePath, entry.PathInfo.Dest)
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s, but it doesn't exist", destPath)
	}

	if _, err := os.Stat(entry.PathInfo.Source); err == nil {
		t.Errorf("source file still exists at %s", entry.PathInfo.Source)
	}
}

func TestTransfer_EpisodeFile(t *testing.T) {
	testDir := testutil.CreateTestDir(t)
	cfg := testutil.CreateTestCfg(testDir)

	entry := testutil.CreateTestEpFile(nil)
	testutil.CreateTestFiles(entry, testDir)

	err := Transfer(entry, &cfg, logger)
	if err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}

	destPath := filepath.Join(cfg.ShowPath, entry.PathInfo.Dest)
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s, but it doesn't exist", destPath)
	}

	if _, err := os.Stat(entry.PathInfo.Source); err == nil {
		t.Errorf("source file still exists at %s", entry.PathInfo.Source)
	}
}

func TestTransfer_MovieDir(t *testing.T) {
	testDir := testutil.CreateTestDir(t)
	cfg := testutil.CreateTestCfg(testDir)

	movieFile := testutil.CreateTestMovieFile(nil)
	bonusFile := testutil.CreateTestBonusFile(nil)
	subtitleFile := testutil.CreateTestSubFile(nil)

	entry := testutil.CreateTestMovieDir(nil, movieFile, bonusFile, subtitleFile)
	testutil.CreateTestFiles(entry, testDir)

	err := Transfer(entry, &cfg, logger)
	if err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}

	// Verify movie file was transferred
	movieDest := filepath.Join(cfg.MoviePath, movieFile.PathInfo.Dest)
	if _, err := os.Stat(movieDest); os.IsNotExist(err) {
		t.Errorf("expected movie file at %s", movieDest)
	}

	// Verify bonus file was transferred
	bonusDest := filepath.Join(cfg.MoviePath, bonusFile.PathInfo.Dest)
	if _, err := os.Stat(bonusDest); os.IsNotExist(err) {
		t.Errorf("expected bonus file at %s", bonusDest)
	}

	// Verify subtitle file was transferred
	subtitleDest := filepath.Join(cfg.MoviePath, subtitleFile.PathInfo.Dest)
	if _, err := os.Stat(subtitleDest); os.IsNotExist(err) {
		t.Errorf("expected subtitle file at %s", subtitleDest)
	}
}

func TestTransfer_SeriesDir(t *testing.T) {
	testDir := testutil.CreateTestDir(t)
	cfg := testutil.CreateTestCfg(testDir)

	ep1 := testutil.CreateTestEpFile(nil)
	ep2 := testutil.CreateTestEpFile(nil)
	ep2.PathInfo.Source = "test.title.2025.S01E02.1080p.x264.remux.atmos.english.mp4"

	seasonDir := testutil.CreateTestSeasonDir(nil, ep1, ep2)
	seriesDir := testutil.CreateTestSeriesDir(nil, seasonDir)
	testutil.CreateTestFiles(seriesDir, testDir)

	err := Transfer(seriesDir, &cfg, logger)
	if err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}

	// Verify episodes were transferred
	ep1Dest := filepath.Join(cfg.ShowPath, ep1.PathInfo.Dest)
	if _, err := os.Stat(ep1Dest); os.IsNotExist(err) {
		t.Errorf("expected episode 1 at %s", ep1Dest)
	}

	ep2Dest := filepath.Join(cfg.ShowPath, ep2.PathInfo.Dest)
	if _, err := os.Stat(ep2Dest); os.IsNotExist(err) {
		t.Errorf("expected episode 2 at %s", ep2Dest)
	}
}

func TestTransfer_DryRun(t *testing.T) {
	testDir := testutil.CreateTestDir(t)
	cfg := testutil.CreateTestCfg(testDir)
	cfg.DryRun = true

	entry := testutil.CreateTestMovieFile(nil)
	testutil.CreateTestFiles(entry, testDir)

	sourcePath := entry.PathInfo.Source

	err := Transfer(entry, &cfg, logger)
	if err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}

	// Source should still exist in dry run
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		t.Errorf("source file was moved during dry run")
	}

	// Destination should not exist
	destPath := filepath.Join(cfg.MoviePath, entry.PathInfo.Dest)
	if _, err := os.Stat(destPath); err == nil {
		t.Errorf("destination file exists during dry run")
	}
}

func TestTransfer_InvalidRole(t *testing.T) {
	testDir := testutil.CreateTestDir(t)
	cfg := testutil.CreateTestCfg(testDir)

	entry := testutil.CreateTestMovieFile(nil)
	entry.Role = -1
	testutil.CreateTestFiles(entry, testDir)

	err := Transfer(entry, &cfg, logger)
	if err == nil {
		t.Error("expected error for invalid role, got nil")
	}
}

func TestError_MovesToErrorDir(t *testing.T) {
	testDir := testutil.CreateTestDir(t)
	cfg := testutil.CreateTestCfg(testDir)

	entry := testutil.CreateTestMovieFile(nil)
	testutil.CreateTestFiles(entry, testDir)

	sourcePath := entry.PathInfo.Source

	Error(entry, &cfg, logger)

	// Check file was moved to error directory
	errorDir := filepath.Join(cfg.ManagerPath, "errors")
	destPath := filepath.Join(errorDir, filepath.Base(entry.PathInfo.Source))

	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s, but it doesn't exist", destPath)
	}

	if _, err := os.Stat(sourcePath); err == nil {
		t.Errorf("source file still exists at %s", sourcePath)
	}
}

func TestError_DryRun(t *testing.T) {
	testDir := testutil.CreateTestDir(t)
	cfg := testutil.CreateTestCfg(testDir)
	cfg.DryRun = true

	entry := testutil.CreateTestMovieFile(nil)
	testutil.CreateTestFiles(entry, testDir)

	sourcePath := entry.PathInfo.Source

	Error(entry, &cfg, logger)

	// Source should still exist in dry run
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		t.Errorf("source file was moved during dry run")
	}
}

func TestResolveConflict_NoConflict(t *testing.T) {
	testDir := t.TempDir()
	path := filepath.Join(testDir, "test.mp4")

	result, err := resolveConflict(path)
	if err != nil {
		t.Fatalf("resolveConflict() error = %v", err)
	}

	if result != path {
		t.Errorf("resolveConflict() = %v, want %v", result, path)
	}
}

func TestResolveConflict_SingleConflict(t *testing.T) {
	testDir := t.TempDir()
	path := filepath.Join(testDir, "test.mp4")

	// Create conflicting file
	if _, err := os.Create(path); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	result, err := resolveConflict(path)
	if err != nil {
		t.Fatalf("resolveConflict() error = %v", err)
	}

	expected := filepath.Join(testDir, "test_1.mp4")
	if result != expected {
		t.Errorf("resolveConflict() = %v, want %v", result, expected)
	}
}

func TestResolveConflict_MultipleConflicts(t *testing.T) {
	testDir := t.TempDir()
	basePath := filepath.Join(testDir, "test.mp4")

	// Create original and first 3 conflicts
	for i := 0; i <= 3; i++ {
		var path string
		if i == 0 {
			path = basePath
		} else {
			path = filepath.Join(testDir, "test_"+string(rune('0'+i))+".mp4")
		}
		if _, err := os.Create(path); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	result, err := resolveConflict(basePath)
	if err != nil {
		t.Fatalf("resolveConflict() error = %v", err)
	}

	expected := filepath.Join(testDir, "test_4.mp4")
	if result != expected {
		t.Errorf("resolveConflict() = %v, want %v", result, expected)
	}
}

func TestResolveConflict_PreservesExtension(t *testing.T) {
	testDir := t.TempDir()
	path := filepath.Join(testDir, "test.mkv")

	if _, err := os.Create(path); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	result, err := resolveConflict(path)
	if err != nil {
		t.Fatalf("resolveConflict() error = %v", err)
	}

	expected := filepath.Join(testDir, "test_1.mkv")
	if result != expected {
		t.Errorf("resolveConflict() = %v, want %v", result, expected)
	}
}

func TestResolveConflict_NoExtension(t *testing.T) {
	testDir := t.TempDir()
	path := filepath.Join(testDir, "testfile")

	if _, err := os.Create(path); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	result, err := resolveConflict(path)
	if err != nil {
		t.Fatalf("resolveConflict() error = %v", err)
	}

	expected := filepath.Join(testDir, "testfile_1")
	if result != expected {
		t.Errorf("resolveConflict() = %v, want %v", result, expected)
	}
}

func TestExists_FileExists(t *testing.T) {
	testDir := t.TempDir()
	path := filepath.Join(testDir, "exists.txt")

	if _, err := os.Create(path); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if !exists(path) {
		t.Error("exists() = false, want true for existing file")
	}
}

func TestExists_FileDoesNotExist(t *testing.T) {
	testDir := t.TempDir()
	path := filepath.Join(testDir, "nonexistent.txt")

	if exists(path) {
		t.Error("exists() = true, want false for non-existent file")
	}
}

func TestExists_Directory(t *testing.T) {
	testDir := t.TempDir()
	path := filepath.Join(testDir, "subdir")

	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	if !exists(path) {
		t.Error("exists() = false, want true for existing directory")
	}
}

func TestMoveEntries_CreatesDestinationDirectory(t *testing.T) {
	testDir := t.TempDir()

	sourcePath := filepath.Join(testDir, "source", "file.mp4")
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0755); err != nil {
		t.Fatalf("failed to create source directory: %v", err)
	}
	if _, err := os.Create(sourcePath); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	destDir := filepath.Join(testDir, "dest", "nested", "path")
	entry := testutil.CreateTestMovieFile(nil)
	entry.PathInfo.Source = sourcePath
	entry.PathInfo.Dest = "file.mp4"

	err := moveEntries(entry, destDir, logger)
	if err != nil {
		t.Fatalf("moveEntries() error = %v", err)
	}

	// Check directory was created
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		t.Errorf("expected directory %s to be created", destDir)
	}

	// Check file was moved
	destPath := filepath.Join(destDir, "file.mp4")
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s", destPath)
	}
}
