package processor

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/ENIACore/media_library_manager/internal/metadata"
)

func Transfer(entry *metadata.Entry, libraryPath string, logger *slog.Logger) error {
	log := logger.With("func", "Move")
	log.Info("processing entry", "source", entry.PathInfo.Source, "mediaPath", libraryPath)

	if entry == nil || entry.PathInfo.Dest == "" {
		log.Error("empty destination", "path", entry.PathInfo.Source)
		return fmt.Errorf("entry %v has no destination set", entry.PathInfo.Source)
	}
	
	moveEntries(entry, libraryPath, logger)	
	log.Info("successfully processed entry", "source", entry.PathInfo.Source, "dest", entry.PathInfo.Dest)
	return nil
}

func Error(entry *metadata.Entry, errorPath string, logger *slog.Logger) {
	log := logger.With("func", "HandleError")

	if entry == nil {
		log.Info("Entry is nil, unable to move entry")
	}

	log.Info("moving entries to error path", "entry", entry.PathInfo.Source, "error-path", errorPath)
	moveEntries(entry, errorPath, logger)

}

func moveEntries(entry *metadata.Entry, path string, logger *slog.Logger) {
	moveEntry(entry, path, logger)
	for _, child := range entry.Children {
		moveEntries(child, path, logger)
	}
}

func moveEntry(entry *metadata.Entry, path string, logger *slog.Logger) {
	log := logger.With("func", "moveEntry")

	if entry.PathInfo.IsDir {
		log.Debug("skipping directory move", "path", entry.PathInfo.Source)
		return
	}

	destPath := filepath.Join(path, entry.PathInfo.Dest)
	destDir := filepath.Dir(destPath)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		log.Error("failed to create directory", "dir", destDir, "err", err)
		panic("failed to create directory")
	}

	finalPath, err := resolveConflict(destPath)
	if err != nil {
		log.Error("failed to resolve conflict", "path", destPath, "err", err)
		panic("failed to resolve conflict")
	}

	if err := os.Rename(entry.PathInfo.Source, finalPath); err != nil {
		log.Error("failed to move file", "source", entry.PathInfo.Source, "dest", finalPath, "err", err)
		panic("failed to move file")
	}

	log.Debug("moved file", "source", entry.PathInfo.Source, "dest", finalPath)
	entry.PathInfo.Dest = finalPath
}

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

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
