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
// Season, Episode, Resolution, Codec, Source, Audio, Language, BTS, DS, and Bonus are all deterministic patterns (i.e 'terminators')
// Title and Year are undeterministic patterns and matched first from left most segment until the first terminator is encountered
func ExtractMedia(path string, logger *slog.Logger) metadata.MediaInfo {
	log := logger.With("func", "ExtractMedia")
	log.Debug("extracting media info from path", "path", path)
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
	mediaInfo.DS = extractDS(sanitizedName)
	mediaInfo.BTS = extractBTS(sanitizedName)
	mediaInfo.Bonus = extractBonus(sanitizedName)
	mediaInfo.Edition = extractEdition(sanitizedName)


	// Second passes for unique cases
	sanitizedName = strings.Split(sanitizeName(filename), ".")

	// pass for language again if subtitle file, ensure language isn't recognized as title
	ext := getExt(path)
	if extractType(ext) == metadata.Subtitle {
		mediaInfo.Language = extractLanguage(sanitizedName)
	}

	// ensure beginning of title isn't extras pattern, prevent extras pattern being recognized as title
	// necessary to use parse instead of extract to prevent false positives in middle of title
	if mediaInfo.DS == "" && parseDS(sanitizedName) != "" {
    	mediaInfo.DS = parseDS(sanitizedName)
	}
	if mediaInfo.BTS == "" && parseBTS(sanitizedName) != "" {
    	mediaInfo.BTS = parseBTS(sanitizedName)
	}
	if mediaInfo.Bonus == "" && parseBonus(sanitizedName) != "" {
    	mediaInfo.Bonus = parseBonus(sanitizedName)
	}
	log.Info("successfully extracted media info", "media-info", fmt.Sprintf("%+v", mediaInfo))

	return mediaInfo
}


// sanitizePrefix finds common unwanted patterns at the beginning of file names (i.e websites) and removes them from the slice
func sanitizePrefix(segments []string) []string {
	if match := parseWebsite(segments); match != "" {
		numParts := len(strings.Split(match, "."))
		return segments[numParts:]
	}
	return segments
}


// isStrictTerminator returns true for patterns that are unambiguously NOT title words.
// Used before a year has been found in extractTitle.
func isStrictTerminator(segments []string) bool {
    return parseResolution(segments) != "" ||
        parseCodec(segments) != "" ||
        parseSource(segments) != "" ||
        parseAudio(segments) != "" ||
        parseSeason(segments) != nil ||
        parseEpisode(segments) != nil ||
        parseVideoExt(segments) != "" ||
        parseSubtitleExt(segments) != "" ||
        parseAudioExt(segments) != ""
}

// isTerminator returns true for any known metadata pattern.
// Used after a year has been found, where ambiguous words are safe to treat as terminators.
func isTerminator(segments []string) bool {
    return isStrictTerminator(segments) ||
        parseMisc(segments) != "" ||
        parseDS(segments) != "" ||
        parseBTS(segments) != "" ||
        parseBonus(segments) != "" ||
        parseEdition(segments) != "" ||
        parseLanguage(segments) != ""
}

// extractTitle returns the title starting from the leftmost segment.
// Extraction stops at the first terminator
// A Year is determined to be part of the title if it is not succeeded by a terminator or is is succeeded by a valid movie year
func extractTitle(segments []string) []string {
	var title []string
	var year *int
	for i, segment := range segments {
		candidates := segments[i:]
		
		// If year found, use unstrict terminator
		if year != nil && isTerminator(candidates) {
            break
        }
		// If year not found, use strict terminator without english words	
        if year == nil && isStrictTerminator(candidates) {
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

// extractYear returns the year from segments or nil if not found.
// Scans segments until a terminator is encountered
func extractYear(segments []string) *int {
	var year *int
	for i, segment := range segments {
		candidates := segments[i:]
		if isTerminator(candidates) {
			return year
		}
		year = parseYear(segment)
	}
	return year
}

// extractSeason returns the season number from segments by iterating through each segment.
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

// extractEpisode returns the episode number from segments by iterating through each segment.
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

// extractResolution returns the resolution pattern from segments by iterating through each segment.
// Returns empty string if resolution is not found.
func extractResolution(segments []string) string {
	for i := range segments {
		candidates := segments[i:]
		if res := parseResolution(candidates); res != "" {
			return res
		}
	}
	return ""
}

// extractCodec returns the codec pattern from segments by iterating through each segment.
// Returns empty string if codec is not found.
func extractCodec(segments []string) string {
	for i := range segments {
		candidates := segments[i:]
		if res := parseCodec(candidates); res != "" {
			return res
		}
	}
	return ""
}

// extractSource returns the source pattern from segments by iterating through each segment.
// Returns empty string if source is not found.
func extractSource(segments []string) string {
	for i := range segments {
		candidates := segments[i:]
		if res := parseSource(candidates); res != "" {
			return res
		}
	}
	return ""
}

// extractAudio returns the audio pattern from segments by iterating through each segment.
// Returns empty string if audio is not found.
func extractAudio(segments []string) string {
	for i := range segments {
		candidates := segments[i:]
		if res := parseAudio(candidates); res != "" {
			return res
		}
	}
	return ""
}

// extractLanguage returns the language pattern from segments by iterating through each segment.
// Returns empty string if language is not found.
func extractLanguage(segments []string) string {
	for i := range segments {
		candidates := segments[i:]
		if language := parseLanguage(candidates); language != "" {
			return language
		}
	}
	return ""
}

// extractBTS returns the behind-the-scenes pattern from segments by iterating through each segment.
// Returns empty string if no pattern is found.
func extractBTS(segments []string) string {
	for i := range segments {
		candidates := segments[i:]
		if bts := parseBTS(candidates); bts != "" {
			return bts
		}
	}
	return ""
}

// extractDS returns the deleted scenes pattern from segments by iterating through each segment.
// Returns empty string if no pattern is found.
func extractDS(segments []string) string {
	for i := range segments {
		candidates := segments[i:]
		if ds := parseDS(candidates); ds != "" {
			return ds
		}
	}
	return ""
}

// extractBonus returns the bonus pattern from segments by iterating through each segment.
// Returns empty string if bonus is not found.
func extractBonus(segments []string) string {
	for i := range segments {
		candidates := segments[i:]
		if bonus := parseBonus(candidates); bonus != "" {
			return bonus
		}
	}
	return ""
}

func extractEdition(segments []string) string {
	for i := range segments {
		candidates := segments[i:]
		if edition := parseEdition(candidates); edition != "" {
			return edition
		}
	}
	return ""
}

// parseResolution returns the resolution if the left most segment(s) are a pattern match.
// Returns empty string if pattern not found.
// Used as helper function for extractor.
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

// parseSource returns the source if the left most segment(s) are a pattern match.
// Returns empty string if pattern not found.
// Used as helper function for extractor.
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

// parseYear returns the year if the input string is a valid year.
// A valid year is a 4 digit number between current year and 1930.
// Used as helper function for extractor.
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

// parseSeason returns the season number if the left most segment(s) are a pattern match.
// Returns nil if no pattern, 0 if pattern without number, >0 for season number.
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

// isRomanNumeral returns true if the string is a Roman numeral I through XX.
func isRomanNumeral(s string) bool {
	romanPattern := regexp.MustCompile(`^(?i)(I|II|III|IV|V|VI|VII|VIII|IX|X|XI|XII|XIII|XIV|XV|XVI|XVII|XVIII|XIX|XX)$`)
	return romanPattern.MatchString(s)
}

// parseEpisode returns the episode number if the left most segment(s) are a pattern match.
// Returns nil if no pattern, 0 if pattern without number, >0 for episode number.
func parseEpisode(segments []string) *int {
	unknown := 0
	for _, re := range patterns.GetEpisodePatterns() {
		match := matchSegments(segments, (*regexp.Regexp)(re))

		if match == nil {
			continue
		}

		// If episode followed by roman numeral, it is a part of the title (e.g Star Wars Episode VII...)
		if len(match) == 1 && len(segments) > 2 && isRomanNumeral(segments[1]) {
			return nil
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

// parseLanguage returns the language if the left most segment(s) are a pattern match.
// Returns empty string if pattern not found.
// Used as helper function for extractor.
func parseLanguage(segments []string) string {
	for _, group := range patterns.GetLanguagePatternGroups() {
		for _, re := range group.Patterns {
			if matchSegments(segments, (*regexp.Regexp)(re)) != nil {
				return group.Key
			}
		}
	}
	return ""
}

// parseMisc returns the misc pattern if the left most segment(s) are a pattern match.
// Returns empty string if pattern not found.
// Used as helper function for isTerminator.
func parseMisc(segments []string) string {
	for _, re := range patterns.GetMiscPatterns() {
		if match := matchSegments(segments, (*regexp.Regexp)(re)); match != nil {
			return match[0]
		}
	}
	return ""
}

// parseBTS returns the behind-the-scenes pattern key if the left most segment(s) are a pattern match.
// Returns empty string if pattern not found.
// Used as helper function for extractor.
func parseBTS(segments []string) string {
	for _, group := range patterns.GetBehindTheScenesPatternGroups() {
		for _, re := range group.Patterns {
			if matchSegments(segments, (*regexp.Regexp)(re)) != nil {
				return group.Key
			}
		}
	}
	return ""
}

// parseDS returns the deleted scenes pattern key if the left most segment(s) are a pattern match.
// Returns empty string if pattern not found.
// Used as helper function for extractor.
func parseDS(segments []string) string {
	for _, group := range patterns.GetDeletedScenesPatternGroups() {
		for _, re := range group.Patterns {
			if matchSegments(segments, (*regexp.Regexp)(re)) != nil {
				return group.Key
			}
		}
	}
	return ""
}

// parseBonus returns the bonus pattern key if the left most segment(s) are a pattern match.
// Returns empty string if pattern not found.
// Used as helper function for extractor.
func parseBonus(segments []string) string {
	for _, group := range patterns.GetBonusPatternGroups() {
		for _, re := range group.Patterns {
			if matchSegments(segments, (*regexp.Regexp)(re)) != nil {
				return group.Key
			}
		}
	}
	return ""
}

func parseEdition(segments []string) string {
	for _, group := range patterns.GetEditionPatterns() {
		for _, re := range group.Patterns {
			if matchSegments(segments, (*regexp.Regexp)(re)) != nil {
				return group.Key
			}
		}
	}
	return ""
}

// parseWebsite returns the website pattern if the left most segment(s) are a pattern match.
// Returns empty string if pattern not found.
// Used as helper function for sanitizePrefix.
func parseWebsite(segments []string) string {
	for _, re := range patterns.GetWebsitePatterns() {
		if match := matchSegments(segments, (*regexp.Regexp)(re)); match != nil {
			return match[0]
		}
	}
	return ""
}
