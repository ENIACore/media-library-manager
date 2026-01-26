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

func Transfer(entry *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "Transfer")
	if cfg.DryRun {
		lg.Info("Dry run true, returning")
		return nil
	}

	for _, child := range entry.Children {
		if err := Transfer(child, cfg, logger); err != nil {
			return err
		}
	}

	if entry.PathInfo.IsDir {
    	lg.Debug("Removing empty source directory", "source", entry.PathInfo.Source)
    	return os.RemoveAll(entry.PathInfo.Source)
	}

	destDir := filepath.Dir(entry.PathInfo.Dest)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	destPath, err := resolveConflict(entry.PathInfo.Dest)
	if err != nil {
		return err
	}

    lg.Info("Moving entry", "source", entry.PathInfo.Source, "dest", destPath)
	return os.Rename(entry.PathInfo.Source, destPath)
}

// Function used to remove any remaining files caused nil root directories (i.e invalid and empty directories at root)
func Cleanup(cfg *config.Config, logger *slog.Logger) {
	lg := logger.With("func", "Cleanup")
	if cfg.DryRun {
		lg.Info("Dry run true, returning")
		return
	}

	entries, err := os.ReadDir(cfg.TorrentPath)
	if err != nil {
		lg.Error("Error occurred during cleanup ReadDir call, returning", "error", err)
		return
	}

	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(cfg.TorrentPath, entry.Name())); err != nil {
			lg.Error("Error occurred during cleanup RemoveAll call, returning", "error", err)
			return
		}
	}	
}

func Error(root *metadata.Entry, cfg *config.Config, logger *slog.Logger) {
	lg := logger.With("func", "Error")
	lg.Info("Moving entries to error dir")
	if root == nil {
		lg.Info("Root is nil, ignoring root entry error handling")
		return
	}
	if cfg.DryRun {
		lg.Info("Dry run true, returning")
		return
	}

	errorPath := filepath.Join(cfg.ManagerPath, "errors")
    if err := os.MkdirAll(errorPath, 0755); err != nil {
		lg.Error("Unable to create error dir")
		panic("Unable to create error dir")
    }

    sourceName := filepath.Base(root.PathInfo.Source)
    destPath := filepath.Join(errorPath, sourceName)
    destPath, err := resolveConflict(destPath)
    if err != nil {
		lg.Error("Unable to resolve conflicting paths in error dir")
		panic("Unable to resolve conflicting paths in error dir")
    }

    lg.Info("Moving to error dir", "source", root.PathInfo.Source, "dest", destPath)
    err = os.Rename(root.PathInfo.Source, destPath)
	if err != nil {
		lg.Error("Unable to move entry to error dir")
		panic("Unable to move entry to error dir")
	}
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
