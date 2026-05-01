package extractor

import (
	//"fmt"
	"log/slog"
	//"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/ENIACore/media_library_manager/internal/patterns"
)

// ExtractPath extracts path metadata from a file or directory path.
// Uses golang 'os' and 'filepath' library to accurately determine file or directory information.
func ExtractFile(path string, logger *slog.Logger) metadata.FileInfo {
	/*
	log := logger.With("func", "ExtractPath")
	pathInfo := metadata.FileInfo{}
	pathInfo.Source = path
	
	pathInfo.Ext = getExt(path)
	pathInfo.Type = extractType(pathInfo.Ext)

	if info, err := os.Stat(path); err == nil {
		pathInfo.IsDir = info.IsDir()
	} else {
		pathInfo.IsDir = pathInfo.Ext == "" && pathInfo.Type == metadata.UnknownType
	}

	log.Info("successfully extracted path info", "path-info", fmt.Sprintf("%+v", pathInfo))
	return pathInfo
	*/
	fileInfo := metadata.FileInfo{}
	return fileInfo
}

// extractType returns the content type based on file extension
// Returns UnknownType if extension empty, not found or unsupported.
func extractType(ext string) metadata.ContentType {
	if ext == "" {
		return metadata.UnknownType
	}
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


// parseBonus returns the video ext pattern if the left most segment(s) are a pattern match.
// Returns empty string if pattern not found.
// Used as helper function for extractor.
func parseVideoExt(segments []string) string {
	for _, re := range patterns.GetVideoExtensionPatterns() {
		match := matchSegments(segments, (*regexp.Regexp)(re))
		if match != nil {
			return match[0]
		}
	}
	return ""
}

// parseSubtitleExt returns the subtitle ext pattern if the left most segment(s) are a pattern match.
// Returns empty string if pattern not found.
// Used as helper function for extractor.
func parseSubtitleExt(segments []string) string {
	for _, re := range patterns.GetSubtitleExtensionPatterns() {
		match := matchSegments(segments, (*regexp.Regexp)(re))
		if match != nil {
			return match[0]
		}
	}
	return ""
}

// parseAudioExt returns the audio ext pattern if the left most segment(s) are a pattern match.
// Returns empty string if pattern not found.
// Used as helper function for extractor.
func parseAudioExt(segments []string) string {
	for _, re := range patterns.GetAudioExtensionPatterns() {
		match := matchSegments(segments, (*regexp.Regexp)(re))
		if match != nil {
			return match[0]
		}
	}
	return ""
}

func getExt(path string) string {
	// Extract extension without the leading dot and uppercase it
	ext := filepath.Ext(path)
	if ext != "" {
		ext = strings.ToUpper(ext[1:]) // Remove leading dot and uppercase
	}
	return ext
}
