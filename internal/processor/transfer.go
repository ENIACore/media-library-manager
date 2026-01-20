package processor

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/ENIACore/media_library_manager/internal/metadata"
)

func Move(entry *metadata.Entry, mediaPath string, logger *slog.Logger) error {
	log := logger.With("func", "Move")
	log.Info("moving entry", "source", entry.PathInfo.Source, "mediaPath", mediaPath)

	if entry.PathInfo.Dest == "" {
		log.Error("empty destination", "path", entry.PathInfo.Source)
		return fmt.Errorf("entry %v has no destination set", entry.PathInfo.Source)
	}

	err := moveEntry(entry, mediaPath, logger)
	if err != nil {
		log.Error("move failed", "path", entry.PathInfo.Source, "err", err)
		return fmt.Errorf("failed to move entry %v: %w", entry.PathInfo.Source, err)
	}

	log.Info("moved entry", "source", entry.PathInfo.Source, "dest", entry.PathInfo.Dest)
	return nil
}

func moveEntry(entry *metadata.Entry, mediaPath string, logger *slog.Logger) error {
	log := logger.With("func", "moveEntry")

	for _, child := range entry.Children {
		if err := moveEntry(child, mediaPath, logger); err != nil {
			return err
		}
	}

	if entry.PathInfo.IsDir {
		log.Debug("skipping directory move", "path", entry.PathInfo.Source)
		return nil
	}

	destPath := filepath.Join(mediaPath, entry.PathInfo.Dest)
	destDir := filepath.Dir(destPath)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		log.Error("failed to create directory", "dir", destDir, "err", err)
		return fmt.Errorf("failed to create directory %v: %w", destDir, err)
	}

	finalPath, err := resolveConflict(destPath)
	if err != nil {
		log.Error("failed to resolve conflict", "path", destPath, "err", err)
		return fmt.Errorf("failed to resolve conflict for %v: %w", destPath, err)
	}

	if err := os.Rename(entry.PathInfo.Source, finalPath); err != nil {
		log.Error("failed to move file", "source", entry.PathInfo.Source, "dest", finalPath, "err", err)
		return fmt.Errorf("failed to move %v to %v: %w", entry.PathInfo.Source, finalPath, err)
	}

	log.Debug("moved file", "source", entry.PathInfo.Source, "dest", finalPath)
	entry.PathInfo.Dest = finalPath
	return nil
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
