package transfer

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/ENIACore/media_library_manager/internal/config"
	"github.com/ENIACore/media_library_manager/internal/metadata"
)

// Transfer moves a media entry tree to the appropriate media library directory.
// Routes to movie or show library based on entry role. Returns immediately if DryRun is enabled.
// Returns an error if transfer fails or entry has an invalid root role.
func Transfer(entry *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "Transfer")
	lg.Info("Moving entries to media dir")

	if cfg.DryRun {
		lg.Info("Dry run true, returning")
		return nil
	}

	var mediaPath string
	switch(entry.Role) {
	case metadata.MovieDir, metadata.MovieFile:
		mediaPath = cfg.MoviePath
	case metadata.SeriesDir, metadata.SeasonDir, metadata.EpisodeFile:
		mediaPath = cfg.ShowPath
	default:
		return fmt.Errorf("Attempted to transfer root entry %v with invalid role %v", entry.PathInfo.Source, entry.Role)
	}

	if err := moveEntries(entry, mediaPath, logger); err != nil {
		lg.Error("Error occurred while transfering entries", "error", err)
		return err
	}

	return nil
}

// Error moves a failed entry to the error directory for manual review.
// Returns immediately if DryRun is enabled. Panics if error handling itself fails.
func Error(entry *metadata.Entry, cfg *config.Config, logger *slog.Logger) {
	lg := logger.With("func", "Error")
	lg.Info("Moving entries to error dir")

	if cfg.DryRun {
		lg.Info("Dry run true, returning")
		return
	}

	errorPath := filepath.Join(cfg.ManagerPath, "errors")
	if err := moveEntries(entry, errorPath, logger); err != nil {
		lg.Error("Error occurred while error handling; non-recoverable", "error", err)
		panic("Error occurred during error handling")
	}
}

// moveEntries recursively moves files in the entry tree to their destination paths.
// Processes children first, then moves files. Skips directories (only moves files).
// Resolves filename conflicts by appending numeric suffixes. Creates destination directories as needed.
func moveEntries(entry *metadata.Entry, destDir string, logger *slog.Logger) error {
	lg := logger.With("func", "moveEntries")

	var err error
	for _, child := range entry.Children {
		if err = moveEntries(child, destDir, logger); err != nil {
			return err
		}
	}

	if entry.PathInfo.IsDir {
		return nil
	}

	destPath := filepath.Join(destDir, entry.PathInfo.Dest)
	if err = os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	if destPath, err = resolveConflict(destPath); err != nil {
		return err
	}


	lg.Info("Moving entry", "source", entry.PathInfo.Source, "dest", entry.PathInfo.Dest)
	if err := os.Rename(entry.PathInfo.Source, destPath); err != nil {
		return err
	}

	return nil
}

// resolveConflict generates a unique filepath by appending numeric suffixes if the path exists.
// Tries up to 10000 variations (file_1, file_2, ..., file_10000).
// Returns the original path if no conflict, or an error if all variations are taken.
func resolveConflict(path string) (string, error) {
	if !exists(path) {
		return path, nil
	}

	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(filepath.Base(path), ext)

	for i := 1; i < 10000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s_%d%s", base, i, ext))
		if !exists(candidate) {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("could not resolve conflict for %v after 10000 attempts", path)
}

// exists checks if a file or directory exists at the given path.
// Returns true if the path exists, false otherwise.
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
