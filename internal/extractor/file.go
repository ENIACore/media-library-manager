package extractor

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	lingua "github.com/pemistahl/lingua-go"
	ffprobe "gopkg.in/vansante/go-ffprobe.v2"

	"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/ENIACore/media_library_manager/internal/patterns"
)

var getLanguageDetector = sync.OnceValue(func() lingua.LanguageDetector {
	return lingua.NewLanguageDetectorBuilder().
		FromAllLanguages().
		WithLowAccuracyMode().
		Build()
})

var linguaToKey = patterns.GetLinguaToKey()
var assDialogueRe = (*regexp.Regexp)(patterns.GetAssDialoguePattern())
var assOverrideRe = (*regexp.Regexp)(patterns.GetAssOverridePattern())
var srtSequenceRe = (*regexp.Regexp)(patterns.GetSrtSequencePattern())
var srtTimestampRe = (*regexp.Regexp)(patterns.GetSrtTimestampPattern())
var sdhBracketRe = (*regexp.Regexp)(patterns.GetSdhBracketPattern())
var sdhMusicRe = (*regexp.Regexp)(patterns.GetSdhMusicPattern())

// ExtractPath extracts path metadata from a file or directory path.
// Uses golang 'os' and 'filepath' library to accurately determine file or directory information.
func ExtractFile(path string, logger *slog.Logger) (metadata.FileInfo, error) {
	log := logger.With("func", "ExtractPath")
	fileInfo := metadata.FileInfo{}
	fileInfo.DestPath = ""
	fileInfo.SourcePath = path

	fileInfo.Ext = getExt(path)
	fileInfo.ContentType = ExtractType(fileInfo.Ext)

	info, err := os.Stat(path)
	if err != nil {
		return fileInfo, fmt.Errorf("Stat returned error: %w", err)
	}
	fileInfo.IsDir = info.IsDir()

	switch fileInfo.ContentType {
	case metadata.Video:
		data, err := ffprobe.ProbeURL(context.Background(), path)
		if err != nil {
			return fileInfo, fmt.Errorf("Stat returned error: %w", err)
		}
		fileInfo.Resolution = extractResolution(data)
		fileInfo.Codec = extractCodec(data)
		fileInfo.Audio = extractAudio(data)
		fileInfo.AspectRatio = extractAspectRatio(data)
		fileInfo.Language = extractLanguage(data)
		fileInfo.Bitrate = extractBitrate(data)
		fileInfo.BitDepth = extractBitDepth(data)
	case metadata.Subtitle:
		segments := SanitizeName(path)
		fileInfo.Language = append(fileInfo.Language, parseSubtitleModifiers(segments)...)

		lang, isSDH, err := extractSubtitleInfo(path, fileInfo.Ext)
		if err != nil {
			log.Warn("Could not extract subtitle info", "err", err)
		} else {
			if isSDH && !containsLang(fileInfo.Language, "SDH") {
				fileInfo.Language = append(fileInfo.Language, "SDH")
			}
			if lang != "" {
				fileInfo.Language = append(fileInfo.Language, lang)
			}
		}
	}

	log.Info("Extracted file info", "SourcePath", fileInfo.SourcePath, "Ext", fileInfo.Ext, "FileType", fileInfo.ContentType, "IsDir", fileInfo.IsDir, "Resolution", fileInfo.Resolution, "Codec", fileInfo.Codec, "Audio", fileInfo.Audio, "AspectRatio", fileInfo.AspectRatio, "Language", fileInfo.Language, "Bitrate", fileInfo.Bitrate, "BitDepth", fileInfo.BitDepth)

	return fileInfo, nil
}

// extractSubtitleInfo reads a subtitle file, detects its language via lingua-go,
// and determines whether it is SDH (contains sound descriptions / music cues).
func extractSubtitleInfo(path, ext string) (lang string, isSDH bool, err error) {
	lines, err := readSubtitleDialogue(path, ext)
	if err != nil {
		return "", false, fmt.Errorf("extractor: reading subtitle dialogue: %w", err)
	}

	sample := buildSample(lines, 3000)
	if sample != "" {
		detector := getLanguageDetector()
		if detected, ok := detector.DetectLanguageOf(sample); ok {
			lang = linguaToKey[detected]
		}
	}

	isSDH = detectSDH(lines)
	return lang, isSDH, nil
}

// readSubtitleDialogue reads a subtitle file and returns only the dialogue text lines,
// stripping timestamps, sequence numbers, and format-specific markup.
func readSubtitleDialogue(path, ext string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch ext {
		case "ASS", "SSA":
			if !assDialogueRe.MatchString(line) {
				continue
			}
			// Dialogue lines: "Dialogue: 0,0:00:01.00,0:00:04.00,Default,,0,0,0,,text"
			// Text is everything after the 9th comma.
			parts := strings.SplitN(line, ",", 10)
			if len(parts) < 10 {
				continue
			}
			text := assOverrideRe.ReplaceAllString(parts[9], "")
			text = strings.TrimSpace(text)
			if text != "" {
				lines = append(lines, text)
			}
		default: // SRT, VTT, and unknowns
			if line == "" || srtSequenceRe.MatchString(line) || srtTimestampRe.MatchString(line) {
				continue
			}
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}

// buildSample concatenates dialogue lines up to maxBytes for language detection.
func buildSample(lines []string, maxBytes int) string {
	var b strings.Builder
	for _, l := range lines {
		if b.Len() >= maxBytes {
			break
		}
		b.WriteString(l)
		b.WriteByte(' ')
	}
	return strings.TrimSpace(b.String())
}

// detectSDH returns true when the proportion of lines containing sound-effect
// brackets or music symbols exceeds 10% — the hallmark of SDH subtitles.
func detectSDH(lines []string) bool {
	total, indicators := 0, 0
	for _, line := range lines {
		if line == "" {
			continue
		}
		total++
		if sdhBracketRe.MatchString(line) || sdhMusicRe.MatchString(line) {
			indicators++
		}
	}
	if total == 0 {
		return false
	}
	return float64(indicators)/float64(total) > 0.10
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
	sdh := false
	forced := false
	for _, stream := range data.Streams {
		switch stream.CodecType {
		case "audio":
			if lang := ParseLanguage([]string{strings.ToUpper(stream.Tags.Language)}); lang != "" {
				languages = append(languages, lang)
			}
		case "subtitle":
			if stream.Disposition.HearingImpaired == 1 {
				sdh = true
			}
			if stream.Disposition.Forced == 1 {
				forced = true
			}
		}
	}
	if forced {
		languages = append(languages, "Forced")
	}
	if sdh {
		languages = append(languages, "SDH")
	}
	return languages
}

// parseSubtitleModifiers scans all segments for subtitle modifier patterns (Forced, SDH).
// Returns found modifier keys in SubtitleModifierPatternGroups order so the Language array
// has modifiers in a stable order regardless of where they appear in the filename.
func parseSubtitleModifiers(segments []string) []string {
	var modifiers []string
	for _, group := range patterns.GetSubtitleModifierPatternGroups() {
		found := false
		for i := range segments {
			if found {
				break
			}
			for _, re := range group.Patterns {
				if matchSegments(segments[i:], (*regexp.Regexp)(re)) != nil {
					modifiers = append(modifiers, group.Key)
					found = true
					break
				}
			}
		}
	}
	return modifiers
}

func containsLang(langs []string, target string) bool {
	for _, l := range langs {
		if l == target {
			return true
		}
	}
	return false
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
	return fmt.Sprintf("%dkbps", bps/1000)
}

func extractAspectRatio(data *ffprobe.ProbeData) string {
	if vs := data.FirstVideoStream(); vs != nil && vs.Width > 0 && vs.Height > 0 {
		return fmt.Sprintf("%dx%d", vs.Width, vs.Height)
	}
	return ""
}

func extractBitDepth(data *ffprobe.ProbeData) string {
	if vs := data.FirstVideoStream(); vs != nil {
		if vs.BitsPerRawSample != "" && vs.BitsPerRawSample != "0" {
			return vs.BitsPerRawSample + "bit"
		}
	}
	return ""
}

// ExtractType returns the content type based on file extension.
// Returns UnknownType if extension is empty, not found, or unsupported.
func ExtractType(ext string) metadata.ContentType {
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
