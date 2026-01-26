package transfer

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/ENIACore/media_library_manager/internal/testutil"
	"github.com/ENIACore/media_library_manager/internal/config"
)

var logger = slog.Default()

// --- Transfer Tests ---

func TestTransfer_MovieFile(t *testing.T) {
	testDir := testutil.CreateTestDir(t)
	cfg := testutil.CreateTestCfg(testDir)

	entry := testutil.CreateTestMovieFile(nil)
	testutil.CreateTestFiles(entry, testDir)

	// Set full destination path
	entry.PathInfo.Dest = filepath.Join(cfg.MoviePath, filepath.Base(entry.PathInfo.Source))
	sourcePath := entry.PathInfo.Source

	if err := Transfer(entry, &cfg, logger); err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}

	if _, err := os.Stat(entry.PathInfo.Dest); os.IsNotExist(err) {
		t.Errorf("expected file at %s, but it doesn't exist", entry.PathInfo.Dest)
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

	entry.PathInfo.Dest = filepath.Join(cfg.ShowPath, filepath.Base(entry.PathInfo.Source))
	sourcePath := entry.PathInfo.Source

	if err := Transfer(entry, &cfg, logger); err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}

	if _, err := os.Stat(entry.PathInfo.Dest); os.IsNotExist(err) {
		t.Errorf("expected file at %s, but it doesn't exist", entry.PathInfo.Dest)
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

	// Set full destination paths for each file
	movieFile.PathInfo.Dest = filepath.Join(cfg.MoviePath, filepath.Base(movieFile.PathInfo.Source))
	bonusFile.PathInfo.Dest = filepath.Join(cfg.MoviePath, filepath.Base(bonusFile.PathInfo.Source))
	subtitleFile.PathInfo.Dest = filepath.Join(cfg.MoviePath, filepath.Base(subtitleFile.PathInfo.Source))

	if err := Transfer(entry, &cfg, logger); err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}

	for _, dest := range []string{movieFile.PathInfo.Dest, bonusFile.PathInfo.Dest, subtitleFile.PathInfo.Dest} {
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
	ep2.PathInfo.Source = "test.title.2025.S01E02.1080p.x264.remux.atmos.english.mp4"

	seasonDir := testutil.CreateTestSeasonDir(nil, ep1, ep2)
	seriesDir := testutil.CreateTestSeriesDir(nil, seasonDir)
	testutil.CreateTestFiles(seriesDir, testDir)

	ep1.PathInfo.Dest = filepath.Join(cfg.ShowPath, filepath.Base(ep1.PathInfo.Source))
	ep2.PathInfo.Dest = filepath.Join(cfg.ShowPath, filepath.Base(ep2.PathInfo.Source))

	if err := Transfer(seriesDir, &cfg, logger); err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}

	for _, dest := range []string{ep1.PathInfo.Dest, ep2.PathInfo.Dest} {
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

	// Dest includes nested directories that don't exist yet
	entry.PathInfo.Dest = filepath.Join(cfg.MoviePath, "Some Movie (2025)", "file.mp4")

	if err := Transfer(entry, &cfg, logger); err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}

	if _, err := os.Stat(entry.PathInfo.Dest); os.IsNotExist(err) {
		t.Errorf("expected file at %s", entry.PathInfo.Dest)
	}
}

func TestTransfer_DryRun(t *testing.T) {
	testDir := testutil.CreateTestDir(t)
	cfg := testutil.CreateTestCfg(testDir)
	cfg.DryRun = true

	entry := testutil.CreateTestMovieFile(nil)
	testutil.CreateTestFiles(entry, testDir)

	sourcePath := entry.PathInfo.Source
	entry.PathInfo.Dest = filepath.Join(cfg.MoviePath, filepath.Base(entry.PathInfo.Source))

	if err := Transfer(entry, &cfg, logger); err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}

	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		t.Errorf("source file was moved during dry run")
	}
	if _, err := os.Stat(entry.PathInfo.Dest); err == nil {
		t.Errorf("destination file exists during dry run")
	}
}

// --- Error Tests ---

func TestError_MovesFileToErrorDir(t *testing.T) {
	testDir := testutil.CreateTestDir(t)
	cfg := testutil.CreateTestCfg(testDir)

	entry := testutil.CreateTestMovieFile(nil)
	testutil.CreateTestFiles(entry, testDir)

	sourcePath := entry.PathInfo.Source
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

	// Create a directory with multiple files
	movieFile := testutil.CreateTestMovieFile(nil)
	bonusFile := testutil.CreateTestBonusFile(nil)
	entry := testutil.CreateTestMovieDir(nil, movieFile, bonusFile)
	testutil.CreateTestFiles(entry, testDir)

	sourcePath := entry.PathInfo.Source
	expectedDest := filepath.Join(cfg.ManagerPath, "errors", filepath.Base(sourcePath))

	Error(entry, &cfg, logger)

	// Entire directory should be moved
	if _, err := os.Stat(expectedDest); os.IsNotExist(err) {
		t.Errorf("expected directory at %s", expectedDest)
	}
	// Children should exist inside moved directory
	if _, err := os.Stat(filepath.Join(expectedDest, filepath.Base(movieFile.PathInfo.Source))); os.IsNotExist(err) {
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

	// Pre-create conflicting file in error dir
	errorDir := filepath.Join(cfg.ManagerPath, "errors")
	if err := os.MkdirAll(errorDir, 0755); err != nil {
		t.Fatalf("failed to create error dir: %v", err)
	}
	conflictPath := filepath.Join(errorDir, filepath.Base(entry.PathInfo.Source))
	if _, err := os.Create(conflictPath); err != nil {
		t.Fatalf("failed to create conflict file: %v", err)
	}

	sourcePath := entry.PathInfo.Source

	Error(entry, &cfg, logger)

	// Should be moved with _1 suffix
	ext := filepath.Ext(sourcePath)
	base := filepath.Base(sourcePath)
	base = base[:len(base)-len(ext)]
	expectedDest := filepath.Join(errorDir, base+"_1"+ext)

	if _, err := os.Stat(expectedDest); os.IsNotExist(err) {
		t.Errorf("expected file at %s", expectedDest)
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

	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		t.Errorf("source file was moved during dry run")
	}
}

// --- resolveConflict Tests ---

func TestResolveConflict_NoConflict(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.mp4")

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

	if _, err := os.Create(path); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	result, err := resolveConflict(path)

	if err != nil {
		t.Fatalf("resolveConflict() error = %v", err)
	}
	if expected := filepath.Join(testDir, "test_1.mp4"); result != expected {
		t.Errorf("resolveConflict() = %v, want %v", result, expected)
	}
}

func TestResolveConflict_MultipleConflicts(t *testing.T) {
	testDir := t.TempDir()
	basePath := filepath.Join(testDir, "test.mp4")

	// Create original and conflicts _1, _2, _3
	for _, name := range []string{"test.mp4", "test_1.mp4", "test_2.mp4", "test_3.mp4"} {
		if _, err := os.Create(filepath.Join(testDir, name)); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	result, err := resolveConflict(basePath)

	if err != nil {
		t.Fatalf("resolveConflict() error = %v", err)
	}
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

	result, err := resolveConflict(path)

	if err != nil {
		t.Fatalf("resolveConflict() error = %v", err)
	}
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

	result, err := resolveConflict(path)

	if err != nil {
		t.Fatalf("resolveConflict() error = %v", err)
	}
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

func TestCleanup(t *testing.T) {
	t.Run("removes all entries in torrent path", func(t *testing.T) {
		testDir := t.TempDir()
		torrentPath := filepath.Join(testDir, "torrents")
		os.MkdirAll(torrentPath, 0755)
		
		cfg := config.Config{
			TorrentPath: torrentPath,
			DryRun:      false,
		}

		// Create test files and directories in torrent path
		file1 := filepath.Join(torrentPath, "file1.txt")
		file2 := filepath.Join(torrentPath, "file2.mp4")
		dir1 := filepath.Join(torrentPath, "dir1")
		dir2 := filepath.Join(torrentPath, "dir2")

		if _, err := os.Create(file1); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
		if _, err := os.Create(file2); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
		if err := os.MkdirAll(dir1, 0755); err != nil {
			t.Fatalf("failed to create test dir: %v", err)
		}
		if err := os.MkdirAll(dir2, 0755); err != nil {
			t.Fatalf("failed to create test dir: %v", err)
		}

		Cleanup(&cfg, logger)

		// Verify all entries removed
		entries, err := os.ReadDir(torrentPath)
		if err != nil {
			t.Fatalf("failed to read torrent path: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("expected empty directory, found %d entries", len(entries))
		}

		// Verify torrent path itself still exists
		if _, err := os.Stat(torrentPath); os.IsNotExist(err) {
			t.Error("torrent path was removed")
		}
	})

	t.Run("removes nested directories", func(t *testing.T) {
		testDir := t.TempDir()
		torrentPath := filepath.Join(testDir, "torrents")
		os.MkdirAll(torrentPath, 0755)
		
		cfg := config.Config{
			TorrentPath: torrentPath,
			DryRun:      false,
		}

		// Create nested structure
		nestedDir := filepath.Join(torrentPath, "parent", "child", "grandchild")
		if err := os.MkdirAll(nestedDir, 0755); err != nil {
			t.Fatalf("failed to create nested dir: %v", err)
		}
		nestedFile := filepath.Join(nestedDir, "file.txt")
		if _, err := os.Create(nestedFile); err != nil {
			t.Fatalf("failed to create nested file: %v", err)
		}

		Cleanup(&cfg, logger)

		entries, err := os.ReadDir(torrentPath)
		if err != nil {
			t.Fatalf("failed to read torrent path: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("expected empty directory, found %d entries", len(entries))
		}
	})

	t.Run("dry run does not remove files", func(t *testing.T) {
		testDir := t.TempDir()
		torrentPath := filepath.Join(testDir, "torrents")
		os.MkdirAll(torrentPath, 0755)
		
		cfg := config.Config{
			TorrentPath: torrentPath,
			DryRun:      true,
		}

		testFile := filepath.Join(torrentPath, "test.txt")
		if _, err := os.Create(testFile); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		Cleanup(&cfg, logger)

		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			t.Error("file was removed during dry run")
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		testDir := t.TempDir()
		torrentPath := filepath.Join(testDir, "torrents")
		os.MkdirAll(torrentPath, 0755)
		
		cfg := config.Config{
			TorrentPath: torrentPath,
			DryRun:      false,
		}

		Cleanup(&cfg, logger)

		// Should not error, just do nothing
		entries, err := os.ReadDir(torrentPath)
		if err != nil {
			t.Fatalf("failed to read torrent path: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("expected empty directory, found %d entries", len(entries))
		}
	})
}
