package extractor

import (
	"log/slog"
	"path/filepath"
	"os"
	"strings"
	"github.com/ENIACore/media_library_manager/internal/metadata"
)

// Filter 
func Filter(path string, logger *slog.Logger) bool {
	lg := logger.With("func", "Filter")
	lg.Info("Determining if path should be filtered", "path", path)

	ext := filepath.Ext(path)
	if ext != "" {
		ext = strings.ToUpper(ext[1:]) // Remove leading dot and uppercase
	}

	info, err := os.Stat(path)
	if err != nil {
		lg.Error("Error occurred getting info for path, filtering path", "path", path)
		return true
	}

	pathType := extractType(ext)
	if pathType == metadata.UnknownType && !info.IsDir() {
		lg.Debug("Path is a unknown file type, filtering path", "path", path)
		return true
	}


	filename := filepath.Base(path)
	sanitizedName := strings.Split(sanitizeName(filename), ".")
	
	if isSample(sanitizedName) {
		lg.Debug("Path is sample, filtering path", "path", path)
		return true
	}

	return false
}

func isSample(segments []string) bool {
	if len(segments) == 1 && parseSamples(segments) != "" {
		return true
	}
	return false
}
