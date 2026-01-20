package processor

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/ENIACore/media_library_manager/internal/metadata"
)

func TestTransfer(t *testing.T) {
	t.Run("move single file", func(t *testing.T) {
		tempDir := t.TempDir()
		sourceDir := filepath.Join(tempDir, "source")
		mediaDir := filepath.Join(tempDir, "media")

		os.MkdirAll(sourceDir, 0755)

		sourceFile := filepath.Join(sourceDir, "test.mp4")
		os.WriteFile(sourceFile, []byte("test content"), 0644)

		entry := &metadata.Entry{
			PathInfo: metadata.PathInfo{
				IsDir:  false,
				Source: sourceFile,
				Dest:   "Test.Movie.2025/Test.Movie.2025.mp4",
			},
		}

		err := Transfer(entry, mediaDir, slog.Default())
		if err != nil {
			t.Fatalf("Transfer error: %v", err)
		}

		expected := filepath.Join(mediaDir, "Test.Movie.2025", "Test.Movie.2025.mp4")
		if !exists(expected) {
			t.Errorf("expected file not found: %v", expected)
		}
		if exists(sourceFile) {
			t.Errorf("source should be moved: %v", sourceFile)
		}
	})

	t.Run("nested structure", func(t *testing.T) {
		tempDir := t.TempDir()
		sourceDir := filepath.Join(tempDir, "source")
		mediaDir := filepath.Join(tempDir, "media")

		movieDir := filepath.Join(sourceDir, "test movie")
		subDir := filepath.Join(movieDir, "subs")
		os.MkdirAll(subDir, 0755)

		movieFile := filepath.Join(movieDir, "movie.mp4")
		subFile := filepath.Join(subDir, "english.srt")
		os.WriteFile(movieFile, []byte("movie"), 0644)
		os.WriteFile(subFile, []byte("sub"), 0644)

		movieEntry := &metadata.Entry{
			PathInfo: metadata.PathInfo{
				IsDir:  false,
				Source: movieFile,
				Dest:   "Test.Movie.2025/Test.Movie.2025.mp4",
			},
		}
		subEntry := &metadata.Entry{
			PathInfo: metadata.PathInfo{
				IsDir:  false,
				Source: subFile,
				Dest:   "Test.Movie.2025/Subtitles/Test.Movie.English.srt",
			},
		}

		root := &metadata.Entry{
			PathInfo: metadata.PathInfo{
				IsDir:  true,
				Source: movieDir,
				Dest:   "Test.Movie.2025",
			},
			Children: []*metadata.Entry{movieEntry, subEntry},
		}

		err := Transfer(root, mediaDir, slog.Default())
		if err != nil {
			t.Fatalf("Transfer error: %v", err)
		}

		if !exists(filepath.Join(mediaDir, "Test.Movie.2025", "Test.Movie.2025.mp4")) {
			t.Errorf("movie not moved")
		}
		if !exists(filepath.Join(mediaDir, "Test.Movie.2025", "Subtitles", "Test.Movie.English.srt")) {
			t.Errorf("subtitle not moved")
		}
	})

	t.Run("conflict resolution", func(t *testing.T) {
		tempDir := t.TempDir()
		sourceDir := filepath.Join(tempDir, "source")
		mediaDir := filepath.Join(tempDir, "media")
		destDir := filepath.Join(mediaDir, "Test.Movie.2025")

		os.MkdirAll(sourceDir, 0755)
		os.MkdirAll(destDir, 0755)
		os.WriteFile(filepath.Join(destDir, "Test.Movie.2025.mp4"), []byte("old"), 0644)

		sourceFile := filepath.Join(sourceDir, "test.mp4")
		os.WriteFile(sourceFile, []byte("new"), 0644)

		entry := &metadata.Entry{
			PathInfo: metadata.PathInfo{
				IsDir:  false,
				Source: sourceFile,
				Dest:   "Test.Movie.2025/Test.Movie.2025.mp4",
			},
		}

		Transfer(entry, mediaDir, slog.Default())

		if !exists(filepath.Join(destDir, "Test.Movie.2025_1.mp4")) {
			t.Errorf("conflict file not created")
		}
	})

	t.Run("empty destination error", func(t *testing.T) {
		entry := &metadata.Entry{
			PathInfo: metadata.PathInfo{
				IsDir:  false,
				Source: "/x/file.mp4",
				Dest:   "",
			},
		}
		if err := Transfer(entry, "/media", slog.Default()); err == nil {
			t.Errorf("expected error for empty destination")
		}
	})

	t.Run("missing source panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic for missing source")
			}
		}()

		entry := &metadata.Entry{
			PathInfo: metadata.PathInfo{
				IsDir:  false,
				Source: "/nonexistent/file.mp4",
				Dest:   "Test.Movie.2025/Test.Movie.2025.mp4",
			},
		}

		Transfer(entry, "/media", slog.Default())
	})
}

