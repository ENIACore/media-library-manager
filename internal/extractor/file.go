package extractor

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"path/filepath"
	"regexp"
	"strings"

	ffprobe "gopkg.in/vansante/go-ffprobe.v2"

	"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/ENIACore/media_library_manager/internal/patterns"
)

// ExtractPath extracts path metadata from a file or directory path.
// Uses golang 'os' and 'filepath' library to accurately determine file or directory information.
func ExtractFile(path string, logger *slog.Logger) (metadata.FileInfo, error) {
	log := logger.With("func", "ExtractPath")
	fileInfo := metadata.FileInfo{}
	fileInfo.DestPath = ""
	fileInfo.SourcePath = path
	
	fileInfo.Ext = getExt(path)
	fileInfo.ContentType = extractType(fileInfo.Ext)

	info, err := os.Stat(path)
	if err != nil {
		return fileInfo, fmt.Errorf("Stat returned error: %w", err)
	}
	fileInfo.IsDir = info.IsDir()

	if fileInfo.ContentType == metadata.Video {
		data, err := ffprobe.ProbeURL(context.Background(), path)
		if err != nil {
			return fileInfo, fmt.Errorf("Stat returned error: %w", err)
		}
		fileInfo.Resolution = extractResolution(data)
		fileInfo.Codec = extractCodec(data)
		fileInfo.Audio = extractAudio(data)
		fileInfo.Language = extractLanguage(data)
		fileInfo.Bitrate = extractBitrate(data)
	} 

	log.Info("Extracted file info", "SourcePath", fileInfo.SourcePath, "Ext", fileInfo.Ext, "FileType", fileInfo.ContentType, "IsDir", fileInfo.IsDir, "Resolution", fileInfo.Resolution, "Codec", fileInfo.Codec, "Audio", fileInfo.Audio, "Language", fileInfo.Language, "Bitrate", fileInfo.Bitrate)

	return fileInfo, nil
}

func extractCodec(data *ffprobe.ProbeData) string {
	if vs := data.FirstVideoStream(); vs != nil {
		codec := SanitizeName(vs.CodecName)
		return parseCodec(codec)
	}
	return ""
}

// parseCodec returns the codec if the left most segment(s) are a pattern match.
// Returns empty string if pattern not found.
// Used as helper function for extractor.
func parseCodec(segments []string) string {
	for _, group := range patterns.GetCodecPatternGroups() {
		for _, re := range group.Patterns {
			if matchSegments(segments, (*regexp.Regexp)(re)) != nil {
				return group.Key
			}
		}
	}
	return ""
}

func extractResolution(data *ffprobe.ProbeData) string {
	if vs := data.FirstVideoStream(); vs != nil {
		h := vs.Height
		w := vs.Width
		switch {
		case h >= 4320 || w >= 7680:
			return "8K"
		case h >= 2160 || w >= 3840:
			return "4K"
		case h >= 1440 || w >= 2560:
			return "2K"
		case h >= 1080 || w >= 1920:
			return "1080p"
		case h >= 720 || w >= 1280:
			return "720p"
		case h >= 576 || w >= 720:
			return "576p"
		case h >= 480 || w >= 640:
			return "480p"
		}
	}
	return ""
}

func extractAudio(data *ffprobe.ProbeData) string {
	if as := data.FirstAudioStream(); as != nil {
		audio := SanitizeName(as.CodecName)
		return parseAudio(audio)
	}
	return ""
}

// parseAudio returns the audio if the left most segment(s) are a pattern match.
// Returns empty string if pattern not found.
// Used as helper function for extractor.
func parseAudio(segments []string) string {
	for _, group := range patterns.GetAudioPatternGroups() {
		for _, re := range group.Patterns {
			if matchSegments(segments, (*regexp.Regexp)(re)) != nil {
				return group.Key
			}
		}
	}
	return ""
}

func extractLanguage(data *ffprobe.ProbeData) []string {
	var languages []string
	for _, stream := range data.Streams {
		lang := ParseLanguage([]string{strings.ToUpper(stream.Tags.Language)})
		if stream.CodecType == "audio" && lang != "" {
			languages = append(languages, lang)
		}
	}
	return languages
}

// ParseLanguage returns the language key if the left most segment(s) are a pattern match.
// Returns empty string if pattern not found.
func ParseLanguage(segments []string) string {
	for _, group := range patterns.GetLanguagePatternGroups() {
		for _, re := range group.Patterns {
			if matchSegments(segments, (*regexp.Regexp)(re)) != nil {
				return group.Key
			}
		}
	}
	return ""
}

func extractBitrate(data *ffprobe.ProbeData) string {
	if data.Format.BitRate == "" {
		return ""
	}
	bps, err := strconv.Atoi(data.Format.BitRate)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%d kbps", bps/1000)
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
