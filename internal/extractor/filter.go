package extractor

import (
	"log/slog"
	"path/filepath"
	"os"
	"strings"
	"regexp"
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/ENIACore/media_library_manager/internal/patterns"
)

// Filter determines if path should be included in [metadata.Entry] tree stucture.
// # Validations
//	- Returns true if file is unrecognized type
//	- Returns true if file or directory is a 'sample'
// Returns true if path should be filtered and false otherwise.
func Filter(path string, logger *slog.Logger) bool {
	lg := logger.With("func", "Filter")
	lg.Debug("Determining if path should be filtered", "path", path)

	ext := filepath.Ext(path)
	if ext != "" {
		ext = strings.ToUpper(ext[1:]) // Remove leading dot and uppercase
	}

	info, err := os.Stat(path)
	if err != nil {
		lg.Error("Error occurred getting info for path, filtering path", "path", path)
		return true
	}

	pathType := ExtractType(ext)
	if pathType == metadata.UnknownType && !info.IsDir() {
		lg.Info("Path is a unknown file type, filtering path", "path", path)
		return true
	}


	sanitizedName := SanitizeName(path)
	
	// Considered sample if sample is at front of name 
	if isSample(sanitizedName) {
		lg.Info("Path is sample, filtering path", "path", path)
		return true
	}

	return false
}

// isSample determines if sanitized name contains pattern(s) indicating path is for sample(s) 
// A sample is a file that represents the video quality of a movie or show.
// Checks both the front and the back of the name.
func isSample(segments []string) bool {
	for i := range segments {
		candidates := segments[i:]
		if sample := parseSample(candidates); sample != "" {
			return true
		}
	}
	return false
}
 
// parseSample returns the sample pattern if the left most or right most segment(s) are a pattern match.
// Returns empty string if pattern not found.
// Used as helper function for isSample.
func parseSample(segments []string) string {
	for _, re := range patterns.GetSamplePatterns() {
		re := (*regexp.Regexp)(re)
		if match := matchSegments(segments, re); match != nil {
			return match[0]
		}
	}
	return ""
}
