// Package used to detect aspects of files (missing audio, errors etc)
// detector/subtitle.go used to detect media files without English srt file
package detector

import (
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

var videoExts = map[string]bool{
	"MP4": true, "MKV": true, "AVI": true, "MOV": true, "FLV": true,
	"WMV": true, "WEBM": true, "M4V": true, "TS": true, "M2TS": true,
	"MPG": true, "MPEG": true, "VOB": true, "3GP": true, "OGV": true,
	"RMVB": true, "RM": true, "DIVX": true, "F4V": true,
}

// DetectSubtitle walks basePath and returns full paths of video files
// that do not have a sibling .English.srt file.
func DetectSubtitle(basePath string, logger *slog.Logger) ([]string, error) {
	var missing []string

	err := filepath.WalkDir(basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToUpper(strings.TrimPrefix(filepath.Ext(path), "."))
		if !videoExts[ext] {
			return nil
		}

		subtitlePath := strings.TrimSuffix(path, filepath.Ext(path)) + ".English.srt"
		if fileExists(subtitlePath) {
			logger.Info("English SRT file exists, skipping", "path", filepath.Base(path))
			return nil
		}

		logger.Info("English SRT missing, adding to list", "path", filepath.Base(path))
		missing = append(missing, path)
		return nil
	})

	return missing, err
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}
