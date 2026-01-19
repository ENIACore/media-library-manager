package extractor

import (
	"path/filepath"
	"strings"
	"github.com/ENIACore/media_library_manager/internal/patterns"
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"log/slog"
	"regexp"
	"fmt"
)

func ExtractPath(path string, logger *slog.Logger) metadata.PathInfo {
	log := logger.With("func", "ExtractPath")
	log.Info("extracting path info from path", "path", path)

	filename := filepath.Base(path)	
	sanitizedName := strings.Split(sanitizeName(filename), ".")

	pathInfo := metadata.PathInfo{}
	pathInfo.Source = path
	pathInfo.Type, pathInfo.Ext = extractType(sanitizedName)
	if pathInfo.Ext == "" && pathInfo.Type == metadata.UnknownType {
		pathInfo.IsDir = true
	}

	log.Debug("successfully extracted path info", "path-info", fmt.Sprintf("%+v", pathInfo))
	return pathInfo
}

// Extracts both content type and extension
// Returns metadata.Unknown and "" if not found or unsupported
func extractType(segments []string) (metadata.ContentType, string) {
	for i := range segments {
		candidates := segments[i:]	
		if match := parseVideoExt(candidates); match != "" {
			return metadata.Video, match
		}
		if match := parseSubtitleExt(candidates); match != "" {
			return metadata.Subtitle, match
		}
		/* TODO: Audio files not yet supported
		if match := parseAudioExt(candidates); match != "" {
			return match
		}
		*/
	}
	return metadata.UnknownType, ""
}

func parseVideoExt(segments []string) string {
	for _, re := range patterns.GetVideoExtensionPatterns() {
		match := matchSegments(segments, (*regexp.Regexp)(re))
		if match != nil {
			return match[0]
		}
	}
	return ""
}

func parseSubtitleExt(segments []string) string {
	for _, re := range patterns.GetSubtitleExtensionPatterns() {
		match := matchSegments(segments, (*regexp.Regexp)(re))
		if match != nil {
			return match[0]
		}
	}
	return ""
}

func parseAudioExt(segments []string) string {
	for _, re := range patterns.GetAudioExtensionPatterns() {
		match := matchSegments(segments, (*regexp.Regexp)(re))
		if match != nil {
			return match[0]
		}
	}
	return ""
}
