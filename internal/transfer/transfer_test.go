package transfer

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/ENIACore/media_library_manager/internal/testutil"
)

var logger = slog.Default()

// --- Transfer Tests ---

func TestTransfer_MovieFile(t *testing.T) {
	testDir := testutil.CreateTestDir(t)
	cfg := testutil.CreateTestCfg(testDir)

	entry := testutil.CreateTestMovieFile(nil)
	testutil.CreateTestFiles(entry, testDir)

	entry.FileInfo.DestPath = filepath.Join(cfg.MoviePath, filepath.Base(entry.FileInfo.SourcePath))
	sourcePath := entry.FileInfo.SourcePath

	Transfer(entry, &cfg, logger)

	if _, err := os.Stat(entry.FileInfo.DestPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s, but it doesn't exist", entry.FileInfo.DestPath)
	}
	if _, err := os.Stat(sourcePath); err == nil {
		t.Errorf("source file still exists at %s", sourcePath)
	}
}

func TestTransfer_EpisodeFile(t *testing.T) {
	testDir := testutil.CreateTestDir(t)
	cfg := testutil.CreateTestCfg(testDir)

	entry := testutil.CreateTestEpFile(nil)
	testutil.CreateTestFiles(entry, testDir)

	entry.FileInfo.DestPath = filepath.Join(cfg.ShowPath, filepath.Base(entry.FileInfo.SourcePath))
	sourcePath := entry.FileInfo.SourcePath

	Transfer(entry, &cfg, logger)

	if _, err := os.Stat(entry.FileInfo.DestPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s, but it doesn't exist", entry.FileInfo.DestPath)
	}
	if _, err := os.Stat(sourcePath); err == nil {
		t.Errorf("source file still exists at %s", sourcePath)
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

	movieFile.FileInfo.DestPath = filepath.Join(cfg.MoviePath, filepath.Base(movieFile.FileInfo.SourcePath))
	bonusFile.FileInfo.DestPath = filepath.Join(cfg.MoviePath, filepath.Base(bonusFile.FileInfo.SourcePath))
	subtitleFile.FileInfo.DestPath = filepath.Join(cfg.MoviePath, filepath.Base(subtitleFile.FileInfo.SourcePath))

	Transfer(entry, &cfg, logger)

	for _, dest := range []string{movieFile.FileInfo.DestPath, bonusFile.FileInfo.DestPath, subtitleFile.FileInfo.DestPath} {
		if _, err := os.Stat(dest); os.IsNotExist(err) {
			t.Errorf("expected file at %s", dest)
		}
	}
}

func TestTransfer_SeriesDir(t *testing.T) {
	testDir := testutil.CreateTestDir(t)
	cfg := testutil.CreateTestCfg(testDir)

	ep1 := testutil.CreateTestEpFile(nil)
	ep2 := testutil.CreateTestEpFile(nil)
	ep2.FileInfo.SourcePath = "test.title.2025.S01E02.1080p.x264.remux.atmos.english.mp4"

	seasonDir := testutil.CreateTestSeasonDir(nil, ep1, ep2)
	seriesDir := testutil.CreateTestSeriesDir(nil, seasonDir)
	testutil.CreateTestFiles(seriesDir, testDir)

	ep1.FileInfo.DestPath = filepath.Join(cfg.ShowPath, filepath.Base(ep1.FileInfo.SourcePath))
	ep2.FileInfo.DestPath = filepath.Join(cfg.ShowPath, filepath.Base(ep2.FileInfo.SourcePath))

	Transfer(seriesDir, &cfg, logger)

	for _, dest := range []string{ep1.FileInfo.DestPath, ep2.FileInfo.DestPath} {
		if _, err := os.Stat(dest); os.IsNotExist(err) {
			t.Errorf("expected episode at %s", dest)
		}
	}
}

func TestTransfer_CreatesNestedDirectories(t *testing.T) {
	testDir := testutil.CreateTestDir(t)
	cfg := testutil.CreateTestCfg(testDir)

	entry := testutil.CreateTestMovieFile(nil)
	testutil.CreateTestFiles(entry, testDir)

	entry.FileInfo.DestPath = filepath.Join(cfg.MoviePath, "Some Movie (2025)", "file.mp4")

	Transfer(entry, &cfg, logger)

	if _, err := os.Stat(entry.FileInfo.DestPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s", entry.FileInfo.DestPath)
	}
}

// --- Error Tests ---

func TestError_MovesFileToErrorDir(t *testing.T) {
	testDir := testutil.CreateTestDir(t)
	cfg := testutil.CreateTestCfg(testDir)

	entry := testutil.CreateTestMovieFile(nil)
	testutil.CreateTestFiles(entry, testDir)

	sourcePath := entry.FileInfo.SourcePath
	expectedDest := filepath.Join(cfg.ManagerPath, "errors", filepath.Base(sourcePath))

	Error(entry, &cfg, logger)

	if _, err := os.Stat(expectedDest); os.IsNotExist(err) {
		t.Errorf("expected file at %s", expectedDest)
	}
	if _, err := os.Stat(sourcePath); err == nil {
		t.Errorf("source file still exists at %s", sourcePath)
	}
}

func TestError_MovesDirectoryToErrorDir(t *testing.T) {
	testDir := testutil.CreateTestDir(t)
	cfg := testutil.CreateTestCfg(testDir)

	movieFile := testutil.CreateTestMovieFile(nil)
	bonusFile := testutil.CreateTestBonusFile(nil)
	entry := testutil.CreateTestMovieDir(nil, movieFile, bonusFile)
	testutil.CreateTestFiles(entry, testDir)

	sourcePath := entry.FileInfo.SourcePath
	expectedDest := filepath.Join(cfg.ManagerPath, "errors", filepath.Base(sourcePath))

	Error(entry, &cfg, logger)

	if _, err := os.Stat(expectedDest); os.IsNotExist(err) {
		t.Errorf("expected directory at %s", expectedDest)
	}
	if _, err := os.Stat(filepath.Join(expectedDest, filepath.Base(movieFile.FileInfo.SourcePath))); os.IsNotExist(err) {
		t.Errorf("expected movie file inside error dir")
	}
	if _, err := os.Stat(sourcePath); err == nil {
		t.Errorf("source directory still exists at %s", sourcePath)
	}
}

func TestError_ResolvesConflicts(t *testing.T) {
	testDir := testutil.CreateTestDir(t)
	cfg := testutil.CreateTestCfg(testDir)

	entry := testutil.CreateTestMovieFile(nil)
	testutil.CreateTestFiles(entry, testDir)

	errorDir := filepath.Join(cfg.ManagerPath, "errors")
	if err := os.MkdirAll(errorDir, 0755); err != nil {
		t.Fatalf("failed to create error dir: %v", err)
	}
	conflictPath := filepath.Join(errorDir, filepath.Base(entry.FileInfo.SourcePath))
	if _, err := os.Create(conflictPath); err != nil {
		t.Fatalf("failed to create conflict file: %v", err)
	}

	sourcePath := entry.FileInfo.SourcePath

	Error(entry, &cfg, logger)

	ext := filepath.Ext(sourcePath)
	base := filepath.Base(sourcePath)
	base = base[:len(base)-len(ext)]
	expectedDest := filepath.Join(errorDir, base+"_1"+ext)

	if _, err := os.Stat(expectedDest); os.IsNotExist(err) {
		t.Errorf("expected file at %s", expectedDest)
	}
}

// --- resolveConflict Tests ---

func TestResolveConflict_NoConflict(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.mp4")

	result := resolveConflict(path)

	if result != path {
		t.Errorf("resolveConflict() = %v, want %v", result, path)
	}
}

func TestResolveConflict_SingleConflict(t *testing.T) {
	testDir := t.TempDir()
	path := filepath.Join(testDir, "test.mp4")

	if _, err := os.Create(path); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	result := resolveConflict(path)

	if expected := filepath.Join(testDir, "test_1.mp4"); result != expected {
		t.Errorf("resolveConflict() = %v, want %v", result, expected)
	}
}

func TestResolveConflict_MultipleConflicts(t *testing.T) {
	testDir := t.TempDir()
	basePath := filepath.Join(testDir, "test.mp4")

	for _, name := range []string{"test.mp4", "test_1.mp4", "test_2.mp4", "test_3.mp4"} {
		if _, err := os.Create(filepath.Join(testDir, name)); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	result := resolveConflict(basePath)

	if expected := filepath.Join(testDir, "test_4.mp4"); result != expected {
		t.Errorf("resolveConflict() = %v, want %v", result, expected)
	}
}

func TestResolveConflict_NoExtension(t *testing.T) {
	testDir := t.TempDir()
	path := filepath.Join(testDir, "testfile")

	if _, err := os.Create(path); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	result := resolveConflict(path)

	if expected := filepath.Join(testDir, "testfile_1"); result != expected {
		t.Errorf("resolveConflict() = %v, want %v", result, expected)
	}
}

func TestResolveConflict_Directory(t *testing.T) {
	testDir := t.TempDir()
	path := filepath.Join(testDir, "somedir")

	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	result := resolveConflict(path)

	if expected := filepath.Join(testDir, "somedir_1"); result != expected {
		t.Errorf("resolveConflict() = %v, want %v", result, expected)
	}
}

// --- exists Tests ---

func TestExists_FileExists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "exists.txt")
	if _, err := os.Create(path); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if !exists(path) {
		t.Error("exists() = false, want true")
	}
}

func TestExists_FileDoesNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.txt")

	if exists(path) {
		t.Error("exists() = true, want false")
	}
}

func TestExists_Directory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "subdir")
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	if !exists(path) {
		t.Error("exists() = false, want true")
	}
}
