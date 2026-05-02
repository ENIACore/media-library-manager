/// Package transfer moves media files to their final destinations or error directory.
// Package relies on resolver package for destination paths and is final stage of processing pipeline.
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

func TestTransfer(entry *metadata.Entry, tempDir string, logger *slog.Logger) {
	lg := logger.With("func", "TestTransfer")

	for _, child := range entry.Children {
		TestTransfer(child, tempDir, logger)
	}

	if entry.FileInfo.IsDir {
		entry.FileInfo.DestPath = filepath.Join(tempDir, entry.FileInfo.DestPath)
		return
	}

	// Ignore empty destination paths, these are failed but unnecessary files
	if entry.FileInfo.DestPath == "" {
		return
	}

	destDir := filepath.Dir(entry.FileInfo.DestPath)
	destDir = filepath.Join(tempDir, destDir)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		panic("Unable to create destination directory: " + destDir)
	}

	entry.FileInfo.DestPath = filepath.Join(tempDir, entry.FileInfo.DestPath)
	entry.FileInfo.DestPath = resolveConflict(entry.FileInfo.DestPath)

	lg.Debug("Creating dummy entry", "source", entry.Source(), "dest", entry.Dest())
	f, err := os.Create(entry.FileInfo.DestPath)
	if err != nil {
		panic("Unable to create dummy file at: " + entry.FileInfo.DestPath)
	}
	f.Close()
}

// Transfer recursively moves media entries to their final destinations.
// Uses [resolveConflict] to handle duplicate filenames. Returns error if move fails.
func Transfer(entry *metadata.Entry, cfg *config.Config, logger *slog.Logger) {
	lg := logger.With("func", "Transfer")

	if cfg.DryRun {
		return
	}

	for _, child := range entry.Children {
		Transfer(child, cfg, logger)
	}

	if entry.FileInfo.IsDir {
		if err := os.RemoveAll(entry.FileInfo.SourcePath); err != nil {
			panic("Unable to remove source path: " + entry.FileInfo.SourcePath)
		}
		return
	}

	// Ignore empty destination paths, these are failed but unnecessary files
	if entry.FileInfo.DestPath == "" {
		return
	}

	destDir := filepath.Dir(entry.FileInfo.DestPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		panic("Unable to make destination directory: " + destDir)
	}

	entry.FileInfo.DestPath = resolveConflict(entry.FileInfo.DestPath)

    lg.Info("Moving entry", "source", entry.Source(), "dest", entry.Dest())
	
	if err := os.Rename(entry.FileInfo.SourcePath, entry.FileInfo.DestPath); err != nil {
		panic("Unable to move source " + entry.FileInfo.SourcePath + " to new destination " + entry.FileInfo.DestPath)
	}
}

// Cleanup removes all remaining entries from torrent directory after processing.
// Skips incomplete directory specified in [config.Config.IncompletePath].
func Cleanup(cfg *config.Config, logger *slog.Logger) {
	lg := logger.With("func", "Cleanup")

	if cfg.DryRun {
		return
	}

	entries, err := os.ReadDir(cfg.TorrentPath)
	if err != nil {
		lg.Error("Error occurred during cleanup ReadDir call, returning", "error", err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() == filepath.Base(cfg.IncompletePath) {
        	continue
    	}
		if err := os.RemoveAll(filepath.Join(cfg.TorrentPath, entry.Name())); err != nil {
			lg.Error("Error occurred during cleanup RemoveAll call, returning", "error", err)
			return
		}
	}	
}

// Error moves failed entry to manager's error directory for manual review.
// Uses [resolveConflict] to handle duplicate names. Panics if unable to create or move to error directory.
func Error(root *metadata.Entry, cfg *config.Config, logger *slog.Logger) {
	lg := logger.With("func", "Error")
	if root == nil {
		return
	}
	if cfg.DryRun {
		return
	}

	errorPath := filepath.Join(cfg.ManagerPath, "errors")
    if err := os.MkdirAll(errorPath, 0755); err != nil {
		lg.Error("Unable to create error dir")
		panic("Unable to create error dir")
    }

    sourceName := filepath.Base(root.FileInfo.SourcePath)
    destPath := filepath.Join(errorPath, sourceName)
    destPath = resolveConflict(destPath)

    lg.Info("Moving to error dir", "source", root.Source(), "dest", destPath)
	err := os.Rename(root.FileInfo.SourcePath, destPath)
	if err != nil {
		lg.Error("Unable to move entry to error dir")
		panic("Unable to move entry to error dir")
	}
}

// resolveConflict generates a unique filepath by appending numeric suffixes if the path exists.
// Tries up to 10000 variations (file_1, file_2, ..., file_10000).
// Panics if all variations are taken.
func resolveConflict(path string) string {
	if !exists(path) {
		return path
	}

	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(filepath.Base(path), ext)

	for i := 1; i < 10000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s_%d%s", base, i, ext))
		if !exists(candidate) {
			return candidate
		}
	}

	panic("Unable to resolve conflict for path: " + path)
}

// exists checks if a file or directory exists at the given path.
// Returns true if the path exists, false otherwise.
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
