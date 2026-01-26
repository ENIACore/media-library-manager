package extractor

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/ENIACore/media_library_manager/internal/patterns"
)

// ExtractMedia extracts media metadata from a file or directory path.
// Returns a populated MediaInfo containing title, year, episode, season,
// resolution, codec, source, audio, language, and bonus information.
func ExtractMedia(path string, logger *slog.Logger) metadata.MediaInfo {
	log := logger.With("func", "ExtractMedia")
	log.Info("extracting media info from path", "path", path)
	filename := filepath.Base(path)

	sanitizedName := strings.Split(sanitizeName(filename), ".")
	sanitizedName = sanitizePrefix(sanitizedName)

	mediaInfo := metadata.MediaInfo{}
	
	title := extractTitle(sanitizedName)
	sanitizedName = sanitizedName[len(title):]

	mediaInfo.Title = title
	mediaInfo.Year = extractYear(sanitizedName)
	mediaInfo.Episode = extractEpisode(sanitizedName)
	mediaInfo.Season = extractSeason(sanitizedName)
	mediaInfo.Resolution = extractResolution(sanitizedName)
	mediaInfo.Codec = extractCodec(sanitizedName)
	mediaInfo.Source = extractSource(sanitizedName)
	mediaInfo.Audio = extractAudio(sanitizedName)
	mediaInfo.Language = extractLanguage(sanitizedName)
	mediaInfo.Bonus = extractBonus(sanitizedName)
	log.Debug("successfully extracted media info", "media-info", fmt.Sprintf("%+v", mediaInfo))

	return mediaInfo
}

func sanitizePrefix(segments []string) []string {
	if match := parseWebsites(segments); match != "" {
		numParts := len(strings.Split(match, "."))
		return segments[numParts:]
	}
	return segments
}

// extractTitle returns the title starting from the leftmost segment.
// Extraction stops at the first recognized pattern (resolution, codec,
// source, audio, season, episode, extension, misc, or bonus).
//
// Segment order:
//
//	<title>.<year>.<pattern>...
//	<title>.<year>
//
// Year is optional in all patterns.
func extractTitle(segments []string) []string {
	var title []string
	var year *int
	for i, segment := range segments {
		candidates := segments[i:]
		if parseResolution(candidates) != "" ||
			parseCodec(candidates) != "" ||
			parseSource(candidates) != "" ||
			parseAudio(candidates) != "" ||
			parseSeason(candidates) != nil ||
			parseEpisode(candidates) != nil ||
			parseVideoExt(candidates) != "" ||
			parseSubtitleExt(candidates) != "" ||
			parseMisc(candidates) != "" ||
			parseAudioExt(candidates) != "" ||
			parseBonus(candidates) != "" {
			break
		}

		if year != nil {
			title = append(title, strconv.Itoa(*year))
			year = nil
		}
		if year = parseYear(segment); year == nil {
			title = append(title, segment)
		}
	}
	return title
}

// extractYear returns the year from segments, or nil if not found.
// Scans segments until a recognized pattern is encountered.
func extractYear(segments []string) *int {
	var year *int
	for i, segment := range segments {
		candidates := segments[i:]
		if parseResolution(candidates) != "" ||
			parseCodec(candidates) != "" ||
			parseSource(candidates) != "" ||
			parseAudio(candidates) != "" ||
			parseSeason(candidates) != nil ||
			parseEpisode(candidates) != nil ||
			parseVideoExt(candidates) != "" ||
			parseSubtitleExt(candidates) != "" ||
			parseMisc(candidates) != "" ||
			parseAudioExt(candidates) != "" ||
			parseBonus(candidates) != "" {
			return year
		}
		year = parseYear(segment)
	}
	return year
}

// extractSeason returns the season number from segments.
// Returns nil if no pattern, 0 if pattern without number, >0 for season number.
func extractSeason(segments []string) *int {
	for i := range segments {
		candidates := segments[i:]
		if season := parseSeason(candidates); season != nil {
			return season
		}
	}
	return nil
}

// extractEpisode returns the episode number from segments.
// Returns nil if no pattern, 0 if pattern without number, >0 for episode number.
func extractEpisode(segments []string) *int {
	for i := range segments {
		candidates := segments[i:]
		if ep := parseEpisode(candidates); ep != nil {
			return ep
		}
	}
	return nil
}

// extractResolution returns the resolution pattern from segments, or empty string.
func extractResolution(segments []string) string {
	for i := range segments {
		candidates := segments[i:]
		if res := parseResolution(candidates); res != "" {
			return res
		}
	}
	return ""
}

// extractCodec returns the codec pattern from segments, or empty string.
func extractCodec(segments []string) string {
	for i := range segments {
		candidates := segments[i:]
		if res := parseCodec(candidates); res != "" {
			return res
		}
	}
	return ""
}

// extractSource returns the source pattern from segments, or empty string.
func extractSource(segments []string) string {
	for i := range segments {
		candidates := segments[i:]
		if res := parseSource(candidates); res != "" {
			return res
		}
	}
	return ""
}

// extractAudio returns the audio pattern from segments, or empty string.
func extractAudio(segments []string) string {
	for i := range segments {
		candidates := segments[i:]
		if res := parseAudio(candidates); res != "" {
			return res
		}
	}
	return ""
}

// extractLanguage returns the language pattern from segments, or empty string.
func extractLanguage(segments []string) string {
	for i := range segments {
		candidates := segments[i:]
		if language := parseLanguage(candidates); language != "" {
			return language
		}
	}
	return ""
}

// extractBonus returns the bonus pattern from segments, or empty string.
func extractBonus(segments []string) string {
	for i := range segments {
		candidates := segments[i:]
		if bonus := parseBonus(candidates); bonus != "" {
			return bonus
		}
	}
	return ""
}

// parseResolution checks if the leftmost segments match a resolution pattern.
// Returns the resolution key or empty string.
func parseResolution(segments []string) string {
	for _, group := range patterns.GetResolutionPatternGroups() {
		for _, re := range group.Patterns {
			if matchSegments(segments, (*regexp.Regexp)(re)) != nil {
				return group.Key
			}
		}
	}
	return ""
}

// parseCodec checks if the leftmost segments match a codec pattern.
// Returns the codec key or empty string.
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

// parseSource checks if the leftmost segments match a source pattern.
// Returns the source key or empty string.
func parseSource(segments []string) string {
	for _, group := range patterns.GetSourcePatternGroups() {
		for _, re := range group.Patterns {
			if matchSegments(segments, (*regexp.Regexp)(re)) != nil {
				return group.Key
			}
		}
	}
	return ""
}

// parseAudio checks if the leftmost segments match an audio pattern.
// Returns the audio key or empty string.
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

// parseYear validates and returns a year from a segment.
// Returns nil if invalid (not 4 digits, not between 1930 and current year).
func parseYear(s string) *int {
	if len(s) != 4 {
		return nil
	}

	year, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	if year < 1930 || year > time.Now().Year() {
		return nil
	}

	return &year
}

// parseSeason checks if segments match a season pattern.
// Returns nil if no match, 0 if pattern without number, >0 for season number.
func parseSeason(segments []string) *int {
	unknown := 0
	for _, re := range patterns.GetSeasonPatterns() {
		match := matchSegments(segments, (*regexp.Regexp)(re))

		if match == nil {
			continue
		}
		if len(match) == 1 {
			return &unknown
		}

		if season, err := strconv.Atoi(match[1]); err == nil {
			return &season
		}

		return &unknown
	}
	return nil
}

// parseEpisode checks if segments match an episode pattern.
// Returns nil if no match, 0 if pattern without number, >0 for episode number.
func parseEpisode(segments []string) *int {
	unknown := 0
	for _, re := range patterns.GetEpisodePatterns() {
		match := matchSegments(segments, (*regexp.Regexp)(re))

		if match == nil {
			continue
		}
		if len(match) == 1 {
			return &unknown
		}

		if ep, err := strconv.Atoi(match[1]); err == nil {
			return &ep
		}
		return &unknown
	}
	return nil
}

// parseLanguage checks if segments match a language pattern.
// Returns the language key or empty string.
func parseLanguage(segments []string) string {
	for _, group := range patterns.GetLanguagePatternGroups() {
		for _, re := range group.Patterns {
			match := matchSegments(segments, (*regexp.Regexp)(re))
			if match != nil {
				return group.Key
			}
		}
	}
	return ""
}

// parseMisc checks if segments match a miscellaneous pattern.
// Returns the matched string or empty string.
func parseMisc(segments []string) string {
	for _, re := range patterns.GetMiscPatterns() {
		if match := matchSegments(segments, (*regexp.Regexp)(re)); match != nil {
			return match[0]
		}
	}
	return ""
}

// parseBonus checks if segments match a bonus content pattern.
// Returns the bonus key or empty string.
func parseBonus(segments []string) string {
	for _, group := range patterns.GetBonusPatternGroups() {
		for _, re := range group.Patterns {
			match := matchSegments(segments, (*regexp.Regexp)(re))
			if match != nil {
				return group.Key
			}
		}
	}
	return ""
}

func parseWebsites(segments []string) string {
	for _, re := range patterns.GetWebsitePatterns() {
		match := matchSegments(segments, (*regexp.Regexp)(re))
		if match != nil {
			return match[0]
		}
	}
	return ""
}
