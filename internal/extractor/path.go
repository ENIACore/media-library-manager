package extractor

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/ENIACore/media_library_manager/internal/patterns"
)

// ExtractPath extracts filesystem metadata from a file or directory path.
// Returns a populated PathInfo containing source path, content type,
// file extension, and whether the path is a directory.
func ExtractPath(path string, logger *slog.Logger) metadata.PathInfo {
	log := logger.With("func", "ExtractPath")
	log.Info("extracting path info from path", "path", path)

	pathInfo := metadata.PathInfo{}
	pathInfo.Source = path
	
	// Extract extension without the leading dot and uppercase it
	ext := filepath.Ext(path)
	if ext != "" {
		ext = strings.ToUpper(ext[1:]) // Remove leading dot and uppercase
	}
	pathInfo.Ext = ext
	pathInfo.Type = extractType(ext)

	if info, err := os.Stat(path); err == nil {
		pathInfo.IsDir = info.IsDir()
	} else {
		pathInfo.IsDir = pathInfo.Ext == "" && pathInfo.Type == metadata.UnknownType
	}

	log.Debug("successfully extracted path info", "path-info", fmt.Sprintf("%+v", pathInfo))
	return pathInfo
}

// extractType returns the content type from an extension string.
// Returns UnknownType if not found or unsupported.
func extractType(ext string) metadata.ContentType {
	if ext == "" {
		return metadata.UnknownType
	}

	// Create a single-element slice for matching
	extSegments := []string{ext}
	
	if match := parseVideoExt(extSegments); match != "" {
		return metadata.Video
	}
	if match := parseSubtitleExt(extSegments); match != "" {
		return metadata.Subtitle
	}
	// TODO: Audio files not yet supported
	
	return metadata.UnknownType
}

// parseVideoExt checks if segments match a video extension pattern.
// Returns the matched extension or empty string.
func parseVideoExt(segments []string) string {
	for _, re := range patterns.GetVideoExtensionPatterns() {
		match := matchSegments(segments, (*regexp.Regexp)(re))
		if match != nil {
			return match[0]
		}
	}
	return ""
}

// parseSubtitleExt checks if segments match a subtitle extension pattern.
// Returns the matched extension or empty string.
func parseSubtitleExt(segments []string) string {
	for _, re := range patterns.GetSubtitleExtensionPatterns() {
		match := matchSegments(segments, (*regexp.Regexp)(re))
		if match != nil {
			return match[0]
		}
	}
	return ""
}

// parseAudioExt checks if segments match an audio extension pattern.
// Returns the matched extension or empty string.
func parseAudioExt(segments []string) string {
	for _, re := range patterns.GetAudioExtensionPatterns() {
		match := matchSegments(segments, (*regexp.Regexp)(re))
		if match != nil {
			return match[0]
		}
	}
	return ""
}
