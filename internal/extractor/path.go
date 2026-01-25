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

	filename := filepath.Base(path)
	sanitizedName := strings.Split(sanitizeName(filename), ".")

	pathInfo := metadata.PathInfo{}
	pathInfo.Source = path
	pathInfo.Type, pathInfo.Ext = extractType(sanitizedName)

	if info, err := os.Stat(path); err == nil {
		pathInfo.IsDir = info.IsDir()
	} else {
		pathInfo.IsDir = pathInfo.Ext == "" && pathInfo.Type == metadata.UnknownType
	}

	log.Debug("successfully extracted path info", "path-info", fmt.Sprintf("%+v", pathInfo))
	return pathInfo
}

// extractType returns the content type and extension from segments.
// Returns UnknownType and empty string if not found or unsupported.
func extractType(segments []string) (metadata.ContentType, string) {
	for i := range segments {
		candidates := segments[i:]
		if match := parseVideoExt(candidates); match != "" {
			return metadata.Video, match
		}
		if match := parseSubtitleExt(candidates); match != "" {
			return metadata.Subtitle, match
		}
		// TODO: Audio files not yet supported
	}
	return metadata.UnknownType, ""
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
