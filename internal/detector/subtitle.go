// Package used to detect aspects of files (missing audio, errors etc)
// detector/subtitle.go used to detect media files without English srt file
package detector

import (
	"log/slog"
	"os"
	"errors"
	"strings"

	"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/ENIACore/media_library_manager/internal/config"
)

// Returns list of entries (movie and episode files) missing english subtitle (srt) files
func DetectSubtitle(root *metadata.Entry, entries []*metadata.Entry, cfg *config.Config, logger *slog.Logger) []*metadata.Entry {
	if root.FileInfo.IsDir {
		for _, child := range root.Children {
			DetectSubtitle(child, entries, cfg, logger)
		}
	}

	if root.Role == metadata.MovieFile || root.Role == metadata.EpisodeFile {
		subtitlePath := root.FileInfo.SourcePath 
		ext := "." + strings.ToLower(root.FileInfo.Ext)
		subtitlePath = strings.TrimSuffix(subtitlePath, ext)
		subtitlePath += ".English.srt"
		
		if !fileExists(subtitlePath) {
			logger.Info("English SRT file missing, adding entry to download list", "SourcePath", root.Source(), "SubtitlePath", subtitlePath)
			root.FileInfo.DestPath = subtitlePath
			entries = append(entries, root)
		} else {
			logger.Info("English SRT file exists, ignoring media file", "SourcePath", root.Source())
		}
	}

	return entries
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
    return !errors.Is(err, os.ErrNotExist)
}
