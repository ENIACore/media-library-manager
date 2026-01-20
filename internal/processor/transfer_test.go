package processor

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/ENIACore/media_library_manager/internal/metadata"
)

func TestMove(t *testing.T) {
	t.Run("move single file", func(t *testing.T) {
		tempDir := t.TempDir()
		sourceDir := filepath.Join(tempDir, "source")
		mediaDir := filepath.Join(tempDir, "media")

		if err := os.MkdirAll(sourceDir, 0755); err != nil {
			t.Fatalf("failed to create source dir: %v", err)
		}

		sourceFile := filepath.Join(sourceDir, "test.mp4")
		if err := os.WriteFile(sourceFile, []byte("test content"), 0644); err != nil {
			t.Fatalf("failed to create source file: %v", err)
		}

		entry := &metadata.Entry{
			PathInfo: metadata.PathInfo{
				IsDir:  false,
				Source: sourceFile,
				Dest:   "Test.Movie.2025/Test.Movie.2025.mp4",
			},
			Role: metadata.MovieFile,
		}

		err := Move(entry, mediaDir, slog.Default())
		if err != nil {
			t.Errorf("Move unexpected error: %v", err)
		}

		expectedPath := filepath.Join(mediaDir, "Test.Movie.2025", "Test.Movie.2025.mp4")
		if !exists(expectedPath) {
			t.Errorf("expected file not found at %v", expectedPath)
		}

		if exists(sourceFile) {
			t.Errorf("source file should have been moved: %v", sourceFile)
		}
	})

	t.Run("move nested structure", func(t *testing.T) {
		tempDir := t.TempDir()
		sourceDir := filepath.Join(tempDir, "source")
		mediaDir := filepath.Join(tempDir, "media")

		movieDir := filepath.Join(sourceDir, "test movie")
		subtitlesDir := filepath.Join(movieDir, "subs")
		if err := os.MkdirAll(subtitlesDir, 0755); err != nil {
			t.Fatalf("failed to create dirs: %v", err)
		}

		movieFile := filepath.Join(movieDir, "movie.mp4")
		subtitleFile := filepath.Join(subtitlesDir, "english.srt")
		if err := os.WriteFile(movieFile, []byte("movie"), 0644); err != nil {
			t.Fatalf("failed to create movie file: %v", err)
		}
		if err := os.WriteFile(subtitleFile, []byte("subtitle"), 0644); err != nil {
			t.Fatalf("failed to create subtitle file: %v", err)
		}

		subtitleEntry := &metadata.Entry{
			PathInfo: metadata.PathInfo{
				IsDir:  false,
				Source: subtitleFile,
				Dest:   "Test.Movie.2025/Subtitles/Test.Movie.English.srt",
			},
			Role: metadata.SubtitleFile,
		}

		movieEntry := &metadata.Entry{
			PathInfo: metadata.PathInfo{
				IsDir:  false,
				Source: movieFile,
				Dest:   "Test.Movie.2025/Test.Movie.2025.mp4",
			},
			Role: metadata.MovieFile,
		}

		entry := &metadata.Entry{
			Children: []*metadata.Entry{movieEntry, subtitleEntry},
			PathInfo: metadata.PathInfo{
				IsDir:  true,
				Source: movieDir,
				Dest:   "Test.Movie.2025",
			},
			Role: metadata.MovieDir,
		}

		err := Move(entry, mediaDir, slog.Default())
		if err != nil {
			t.Errorf("Move unexpected error: %v", err)
		}

		expectedMovie := filepath.Join(mediaDir, "Test.Movie.2025", "Test.Movie.2025.mp4")
		if !exists(expectedMovie) {
			t.Errorf("expected movie file not found at %v", expectedMovie)
		}

		expectedSubtitle := filepath.Join(mediaDir, "Test.Movie.2025", "Subtitles", "Test.Movie.English.srt")
		if !exists(expectedSubtitle) {
			t.Errorf("expected subtitle file not found at %v", expectedSubtitle)
		}
	})

	t.Run("move with conflict", func(t *testing.T) {
		tempDir := t.TempDir()
		sourceDir := filepath.Join(tempDir, "source")
		mediaDir := filepath.Join(tempDir, "media")

		if err := os.MkdirAll(sourceDir, 0755); err != nil {
			t.Fatalf("failed to create source dir: %v", err)
		}

		destDir := filepath.Join(mediaDir, "Test.Movie.2025")
		if err := os.MkdirAll(destDir, 0755); err != nil {
			t.Fatalf("failed to create dest dir: %v", err)
		}
		existingFile := filepath.Join(destDir, "Test.Movie.2025.mp4")
		if err := os.WriteFile(existingFile, []byte("existing"), 0644); err != nil {
			t.Fatalf("failed to create existing file: %v", err)
		}

		sourceFile := filepath.Join(sourceDir, "test.mp4")
		if err := os.WriteFile(sourceFile, []byte("new content"), 0644); err != nil {
			t.Fatalf("failed to create source file: %v", err)
		}

		entry := &metadata.Entry{
			PathInfo: metadata.PathInfo{
				IsDir:  false,
				Source: sourceFile,
				Dest:   "Test.Movie.2025/Test.Movie.2025.mp4",
			},
			Role: metadata.MovieFile,
		}

		err := Move(entry, mediaDir, slog.Default())
		if err != nil {
			t.Errorf("Move unexpected error: %v", err)
		}

		expectedPath := filepath.Join(mediaDir, "Test.Movie.2025", "Test.Movie.2025_1.mp4")
		if !exists(expectedPath) {
			t.Errorf("expected conflict-resolved file not found at %v", expectedPath)
		}

		if !exists(existingFile) {
			t.Errorf("existing file should still exist: %v", existingFile)
		}
	})

	t.Run("move with multiple conflicts", func(t *testing.T) {
		tempDir := t.TempDir()
		sourceDir := filepath.Join(tempDir, "source")
		mediaDir := filepath.Join(tempDir, "media")

		if err := os.MkdirAll(sourceDir, 0755); err != nil {
			t.Fatalf("failed to create source dir: %v", err)
		}

		destDir := filepath.Join(mediaDir, "Test.Movie.2025")
		if err := os.MkdirAll(destDir, 0755); err != nil {
			t.Fatalf("failed to create dest dir: %v", err)
		}

		existingFile := filepath.Join(destDir, "Test.Movie.2025.mp4")
		if err := os.WriteFile(existingFile, []byte("existing"), 0644); err != nil {
			t.Fatalf("failed to create existing file: %v", err)
		}
		existingFile1 := filepath.Join(destDir, "Test.Movie.2025_1.mp4")
		if err := os.WriteFile(existingFile1, []byte("existing1"), 0644); err != nil {
			t.Fatalf("failed to create existing file 1: %v", err)
		}
		existingFile2 := filepath.Join(destDir, "Test.Movie.2025_2.mp4")
		if err := os.WriteFile(existingFile2, []byte("existing2"), 0644); err != nil {
			t.Fatalf("failed to create existing file 2: %v", err)
		}

		sourceFile := filepath.Join(sourceDir, "test.mp4")
		if err := os.WriteFile(sourceFile, []byte("new content"), 0644); err != nil {
			t.Fatalf("failed to create source file: %v", err)
		}

		entry := &metadata.Entry{
			PathInfo: metadata.PathInfo{
				IsDir:  false,
				Source: sourceFile,
				Dest:   "Test.Movie.2025/Test.Movie.2025.mp4",
			},
			Role: metadata.MovieFile,
		}

		err := Move(entry, mediaDir, slog.Default())
		if err != nil {
			t.Errorf("Move unexpected error: %v", err)
		}

		expectedPath := filepath.Join(mediaDir, "Test.Movie.2025", "Test.Movie.2025_3.mp4")
		if !exists(expectedPath) {
			t.Errorf("expected conflict-resolved file not found at %v", expectedPath)
		}
	})

	t.Run("error on empty destination", func(t *testing.T) {
		entry := &metadata.Entry{
			PathInfo: metadata.PathInfo{
				IsDir:  false,
				Source: "/some/path/file.mp4",
				Dest:   "",
			},
			Role: metadata.MovieFile,
		}

		err := Move(entry, "/media", slog.Default())
		if err == nil {
			t.Errorf("Move expected error for empty destination")
		}
	})

	t.Run("error on missing source", func(t *testing.T) {
		tempDir := t.TempDir()
		mediaDir := filepath.Join(tempDir, "media")

		entry := &metadata.Entry{
			PathInfo: metadata.PathInfo{
				IsDir:  false,
				Source: "/nonexistent/file.mp4",
				Dest:   "Test.Movie.2025/Test.Movie.2025.mp4",
			},
			Role: metadata.MovieFile,
		}

		err := Move(entry, mediaDir, slog.Default())
		if err == nil {
			t.Errorf("Move expected error for missing source")
		}
	})
}

func TestMoveEntry(t *testing.T) {
	t.Run("skip directory", func(t *testing.T) {
		tempDir := t.TempDir()
		sourceDir := filepath.Join(tempDir, "source", "movie")
		mediaDir := filepath.Join(tempDir, "media")

		if err := os.MkdirAll(sourceDir, 0755); err != nil {
			t.Fatalf("failed to create source dir: %v", err)
		}

		entry := &metadata.Entry{
			PathInfo: metadata.PathInfo{
				IsDir:  true,
				Source: sourceDir,
				Dest:   "Test.Movie.2025",
			},
			Role: metadata.MovieDir,
		}

		err := moveEntry(entry, mediaDir, slog.Default())
		if err != nil {
			t.Errorf("moveEntry unexpected error: %v", err)
		}

		if exists(filepath.Join(mediaDir, "Test.Movie.2025")) {
			t.Errorf("directory should not be created for dir entries")
		}
	})

	t.Run("move file creates parent dirs", func(t *testing.T) {
		tempDir := t.TempDir()
		sourceDir := filepath.Join(tempDir, "source")
		mediaDir := filepath.Join(tempDir, "media")

		if err := os.MkdirAll(sourceDir, 0755); err != nil {
			t.Fatalf("failed to create source dir: %v", err)
		}

		sourceFile := filepath.Join(sourceDir, "subtitle.srt")
		if err := os.WriteFile(sourceFile, []byte("subtitle"), 0644); err != nil {
			t.Fatalf("failed to create source file: %v", err)
		}

		entry := &metadata.Entry{
			PathInfo: metadata.PathInfo{
				IsDir:  false,
				Source: sourceFile,
				Dest:   "Test.Movie.2025/Subtitles/Test.Movie.English.srt",
			},
			Role: metadata.SubtitleFile,
		}

		err := moveEntry(entry, mediaDir, slog.Default())
		if err != nil {
			t.Errorf("moveEntry unexpected error: %v", err)
		}

		expectedPath := filepath.Join(mediaDir, "Test.Movie.2025", "Subtitles", "Test.Movie.English.srt")
		if !exists(expectedPath) {
			t.Errorf("expected file not found at %v", expectedPath)
		}

		subtitlesDir := filepath.Join(mediaDir, "Test.Movie.2025", "Subtitles")
		if !exists(subtitlesDir) {
			t.Errorf("expected Subtitles dir to be created: %v", subtitlesDir)
		}
	})
}

func TestResolveConflict(t *testing.T) {
	t.Run("no conflict", func(t *testing.T) {
		tempDir := t.TempDir()
		path := filepath.Join(tempDir, "file.mp4")

		result, err := resolveConflict(path)
		if err != nil {
			t.Errorf("resolveConflict unexpected error: %v", err)
		}
		if result != path {
			t.Errorf("resolveConflict = %v, want %v", result, path)
		}
	})

	t.Run("single conflict", func(t *testing.T) {
		tempDir := t.TempDir()
		path := filepath.Join(tempDir, "file.mp4")

		if err := os.WriteFile(path, []byte("existing"), 0644); err != nil {
			t.Fatalf("failed to create existing file: %v", err)
		}

		result, err := resolveConflict(path)
		if err != nil {
			t.Errorf("resolveConflict unexpected error: %v", err)
		}

		expected := filepath.Join(tempDir, "file_1.mp4")
		if result != expected {
			t.Errorf("resolveConflict = %v, want %v", result, expected)
		}
	})

	t.Run("multiple conflicts", func(t *testing.T) {
		tempDir := t.TempDir()
		path := filepath.Join(tempDir, "file.mp4")

		if err := os.WriteFile(path, []byte("existing"), 0644); err != nil {
			t.Fatalf("failed to create existing file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tempDir, "file_1.mp4"), []byte("existing1"), 0644); err != nil {
			t.Fatalf("failed to create existing file 1: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tempDir, "file_2.mp4"), []byte("existing2"), 0644); err != nil {
			t.Fatalf("failed to create existing file 2: %v", err)
		}

		result, err := resolveConflict(path)
		if err != nil {
			t.Errorf("resolveConflict unexpected error: %v", err)
		}

		expected := filepath.Join(tempDir, "file_3.mp4")
		if result != expected {
			t.Errorf("resolveConflict = %v, want %v", result, expected)
		}
	})

	t.Run("file without extension", func(t *testing.T) {
		tempDir := t.TempDir()
		path := filepath.Join(tempDir, "README")

		if err := os.WriteFile(path, []byte("existing"), 0644); err != nil {
			t.Fatalf("failed to create existing file: %v", err)
		}

		result, err := resolveConflict(path)
		if err != nil {
			t.Errorf("resolveConflict unexpected error: %v", err)
		}

		expected := filepath.Join(tempDir, "README_1")
		if result != expected {
			t.Errorf("resolveConflict = %v, want %v", result, expected)
		}
	})

	t.Run("nested path conflict", func(t *testing.T) {
		tempDir := t.TempDir()
		nestedDir := filepath.Join(tempDir, "Movies", "Test.Movie.2025")
		if err := os.MkdirAll(nestedDir, 0755); err != nil {
			t.Fatalf("failed to create nested dir: %v", err)
		}

		path := filepath.Join(nestedDir, "movie.mp4")
		if err := os.WriteFile(path, []byte("existing"), 0644); err != nil {
			t.Fatalf("failed to create existing file: %v", err)
		}

		result, err := resolveConflict(path)
		if err != nil {
			t.Errorf("resolveConflict unexpected error: %v", err)
		}

		expected := filepath.Join(nestedDir, "movie_1.mp4")
		if result != expected {
			t.Errorf("resolveConflict = %v, want %v", result, expected)
		}
	})
}

func TestExists(t *testing.T) {
	t.Run("file exists", func(t *testing.T) {
		tempDir := t.TempDir()
		path := filepath.Join(tempDir, "file.txt")

		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}

		if !exists(path) {
			t.Errorf("exists = false, want true for existing file")
		}
	})

	t.Run("file does not exist", func(t *testing.T) {
		tempDir := t.TempDir()
		path := filepath.Join(tempDir, "nonexistent.txt")

		if exists(path) {
			t.Errorf("exists = true, want false for nonexistent file")
		}
	})

	t.Run("directory exists", func(t *testing.T) {
		tempDir := t.TempDir()

		if !exists(tempDir) {
			t.Errorf("exists = false, want true for existing directory")
		}
	})
}
