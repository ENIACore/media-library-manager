package transfer
/*

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/ENIACore/media_library_manager/internal/config"
	"github.com/ENIACore/media_library_manager/internal/metadata"
)

func createTestMovieFile(testDir string, children []*metadata.Entry) *metadata.Entry {

}

func TestTransfer_MovieFile(t *testing.T) {
	testDir := createTestDir(t)

	sourcePath := filepath.Join(testDir, "source", "movie.mp4")

	testCfg := createTestCfg(testDir)

	entry := &metadata.Entry{
		Role: metadata.MovieFile,
		PathInfo: metadata.PathInfo{
			Source: sourcePath,
			Dest:   "movie.mp4",
			IsDir:  false,
		},
		Children: []*metadata.Entry{},
	}

	err := Transfer(entry, cfg, logger)

	if err != nil {
		t.Errorf("Transfer failed: %v", err)
	}

	destPath := filepath.Join(cfg.MoviePath, "movie.mp4")
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Errorf("Expected file at %s, but it doesn't exist", destPath)
	}

	if _, err := os.Stat(sourcePath); err == nil {
		t.Errorf("Source file still exists at %s", sourcePath)
	}
}

func TestTransfer_SeriesFile(t *testing.T) {
	tempDir, cleanup := setupTestDirs(t)
	defer cleanup()

	sourcePath := filepath.Join(tempDir, "source", "show", "S01E01.mp4")
	createTestFile(t, sourcePath)

	cfg := &config.Config{
		MoviePath:   filepath.Join(tempDir, "movies"),
		ShowPath:    filepath.Join(tempDir, "shows"),
		ManagerPath: filepath.Join(tempDir, "manager"),
		DryRun:      false,
	}

	entry := &metadata.Entry{
		Role: metadata.EpisodeFile,
		PathInfo: metadata.PathInfo{
			Source: sourcePath,
			Dest:   "Show/Season 1/S01E01.mp4",
			IsDir:  false,
		},
		Children: []*metadata.Entry{},
	}

	err := Transfer(entry, cfg, logger)

	if err != nil {
		t.Errorf("Transfer failed: %v", err)
	}

	destPath := filepath.Join(cfg.ShowPath, "Show", "Season 1", "S01E01.mp4")
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Errorf("Expected file at %s, but it doesn't exist", destPath)
	}
}

func TestTransfer_DryRun(t *testing.T) {
	tempDir, cleanup := setupTestDirs(t)
	defer cleanup()

	sourcePath := filepath.Join(tempDir, "source", "movie.mp4")
	createTestFile(t, sourcePath)

	cfg := &config.Config{
		MoviePath:   filepath.Join(tempDir, "movies"),
		ShowPath:    filepath.Join(tempDir, "shows"),
		ManagerPath: filepath.Join(tempDir, "manager"),
		DryRun:      true,
	}

	entry := &metadata.Entry{
		Role: metadata.MovieFile,
		PathInfo: metadata.PathInfo{
			Source: sourcePath,
			Dest:   "movie.mp4",
			IsDir:  false,
		},
		Children: []*metadata.Entry{},
	}

	err := Transfer(entry, cfg, logger)

	if err != nil {
		t.Errorf("Transfer failed: %v", err)
	}

	// Source should still exist
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		t.Errorf("Source file was moved during dry run")
	}

	// Dest should not exist
	destPath := filepath.Join(cfg.MoviePath, "movie.mp4")
	if _, err := os.Stat(destPath); err == nil {
		t.Errorf("Destination file exists during dry run")
	}
}

func TestTransfer_InvalidRole(t *testing.T) {
	tempDir, cleanup := setupTestDirs(t)
	defer cleanup()

	cfg := &config.Config{
		MoviePath:   filepath.Join(tempDir, "movies"),
		ShowPath:    filepath.Join(tempDir, "shows"),
		ManagerPath: filepath.Join(tempDir, "manager"),
		DryRun:      false,
	}

	entry := &metadata.Entry{
		Role: -1,
		PathInfo: metadata.PathInfo{
			Source: filepath.Join(tempDir, "source", "file.mp4"),
			Dest:   "file.mp4",
			IsDir:  false,
		},
		Children: []*metadata.Entry{},
	}

	err := Transfer(entry, cfg, logger)

	if err == nil {
		t.Error("Expected error for invalid role, got nil")
	}
}

func TestTransfer_WithChildren(t *testing.T) {
	tempDir, cleanup := setupTestDirs(t)
	defer cleanup()

	// Create parent directory and child files
	parentSource := filepath.Join(tempDir, "source", "show")
	os.MkdirAll(parentSource, 0755)

	child1Source := filepath.Join(parentSource, "S01E01.mp4")
	child2Source := filepath.Join(parentSource, "S01E02.mp4")
	createTestFile(t, child1Source)
	createTestFile(t, child2Source)

	cfg := &config.Config{
		MoviePath:   filepath.Join(tempDir, "movies"),
		ShowPath:    filepath.Join(tempDir, "shows"),
		ManagerPath: filepath.Join(tempDir, "manager"),
		DryRun:      false,
	}

	entry := &metadata.Entry{
		Role: metadata.SeriesDir,
		PathInfo: metadata.PathInfo{
			Source: parentSource,
			Dest:   "Show",
			IsDir:  true,
		},
		Children: []*metadata.Entry{
			{
				Role: metadata.EpisodeFile,
				PathInfo: metadata.PathInfo{
					Source: child1Source,
					Dest:   "Show/S01E01.mp4",
					IsDir:  false,
				},
				Children: []*metadata.Entry{},
			},
			{
				Role: metadata.EpisodeFile,
				PathInfo: metadata.PathInfo{
					Source: child2Source,
					Dest:   "Show/S01E02.mp4",
					IsDir:  false,
				},
				Children: []*metadata.Entry{},
			},
		},
	}

	err := Transfer(entry, cfg, logger)

	if err != nil {
		t.Errorf("Transfer failed: %v", err)
	}

	// Check both children were moved
	dest1 := filepath.Join(cfg.ShowPath, "Show", "S01E01.mp4")
	dest2 := filepath.Join(cfg.ShowPath, "Show", "S01E02.mp4")

	if _, err := os.Stat(dest1); os.IsNotExist(err) {
		t.Errorf("Expected file at %s", dest1)
	}
	if _, err := os.Stat(dest2); os.IsNotExist(err) {
		t.Errorf("Expected file at %s", dest2)
	}
}

func TestError_MovesToErrorDir(t *testing.T) {
	tempDir, cleanup := setupTestDirs(t)
	defer cleanup()

	sourcePath := filepath.Join(tempDir, "source", "error.mp4")
	createTestFile(t, sourcePath)

	cfg := &config.Config{
		MoviePath:   filepath.Join(tempDir, "movies"),
		ShowPath:    filepath.Join(tempDir, "shows"),
		ManagerPath: filepath.Join(tempDir, "manager"),
		DryRun:      false,
	}

	entry := &metadata.Entry{
		Role: metadata.MovieFile,
		PathInfo: metadata.PathInfo{
			Source: sourcePath,
			Dest:   "error.mp4",
			IsDir:  false,
		},
		Children: []*metadata.Entry{},
	}

	Error(entry, cfg, logger)

	errorDir := filepath.Join(cfg.ManagerPath, "errors")
	destPath := filepath.Join(errorDir, "error.mp4")

	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Errorf("Expected file at %s, but it doesn't exist", destPath)
	}

	if _, err := os.Stat(sourcePath); err == nil {
		t.Errorf("Source file still exists at %s", sourcePath)
	}
}

func TestError_DryRun(t *testing.T) {
	tempDir, cleanup := setupTestDirs(t)
	defer cleanup()

	sourcePath := filepath.Join(tempDir, "source", "error.mp4")
	createTestFile(t, sourcePath)

	cfg := &config.Config{
		MoviePath:   filepath.Join(tempDir, "movies"),
		ShowPath:    filepath.Join(tempDir, "shows"),
		ManagerPath: filepath.Join(tempDir, "manager"),
		DryRun:      true,
	}

	entry := &metadata.Entry{
		Role: metadata.MovieFile,
		PathInfo: metadata.PathInfo{
			Source: sourcePath,
			Dest:   "error.mp4",
			IsDir:  false,
		},
		Children: []*metadata.Entry{},
	}

	Error(entry, cfg, logger)

	// Source should still exist in dry run
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		t.Errorf("Source file was moved during dry run")
	}
}

func TestResolveConflict_NoConflict(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "test.mp4")

	result, err := resolveConflict(path)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result != path {
		t.Errorf("Expected %s, got %s", path, result)
	}
}

func TestResolveConflict_SingleConflict(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "test.mp4")
	createTestFile(t, path)

	result, err := resolveConflict(path)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := filepath.Join(tempDir, "test_1.mp4")
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestResolveConflict_MultipleConflicts(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "test.mp4")
	
	// Create original and first 3 conflicts
	createTestFile(t, basePath)
	createTestFile(t, filepath.Join(tempDir, "test_1.mp4"))
	createTestFile(t, filepath.Join(tempDir, "test_2.mp4"))
	createTestFile(t, filepath.Join(tempDir, "test_3.mp4"))

	result, err := resolveConflict(basePath)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := filepath.Join(tempDir, "test_4.mp4")
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestResolveConflict_PreservesExtension(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "test.mkv")
	createTestFile(t, path)

	result, err := resolveConflict(path)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := filepath.Join(tempDir, "test_1.mkv")
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestResolveConflict_NoExtension(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "testfile")
	createTestFile(t, path)

	result, err := resolveConflict(path)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := filepath.Join(tempDir, "testfile_1")
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestExists_FileExists(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "exists.txt")
	createTestFile(t, path)

	if !exists(path) {
		t.Error("Expected exists() to return true for existing file")
	}
}

func TestExists_FileDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "nonexistent.txt")

	if exists(path) {
		t.Error("Expected exists() to return false for non-existent file")
	}
}

func TestExists_Directory(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "subdir")
	os.MkdirAll(path, 0755)

	if !exists(path) {
		t.Error("Expected exists() to return true for existing directory")
	}
}

func TestMoveEntries_CreatesDestinationDirectory(t *testing.T) {
	tempDir := t.TempDir()
	
	sourcePath := filepath.Join(tempDir, "source", "file.mp4")
	createTestFile(t, sourcePath)

	destDir := filepath.Join(tempDir, "dest", "nested", "path")

	entry := &metadata.Entry{
		Role: metadata.MovieFile,
		PathInfo: metadata.PathInfo{
			Source: sourcePath,
			Dest:   "file.mp4",
			IsDir:  false,
		},
		Children: []*metadata.Entry{},
	}

	err := moveEntries(entry, destDir, logger)

	if err != nil {
		t.Errorf("moveEntries failed: %v", err)
	}

	// Check directory was created
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		t.Errorf("Expected directory %s to be created", destDir)
	}

	// Check file was moved
	destPath := filepath.Join(destDir, "file.mp4")
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Errorf("Expected file at %s", destPath)
	}
}
*/
