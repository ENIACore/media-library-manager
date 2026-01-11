package extractor

import (
	"fmt"
	"strings"
	"regexp"
	"strconv"
	"time"
	"log/slog"
	"github.com/ENIACore/media_library_manager/internal/patterns"
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"path/filepath"
)



func ExtractMedia(path string, logger *slog.Logger) metadata.MediaInfo {
	log := logger.With("func", "ExtractMedia")
	log.Info("extracting media info from path", "path", path)
	filename := filepath.Base(path)	

	sanitizedName := strings.Split(sanitizeName(filename), ".")
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

// Returns title starting from left most segment
// Extracts using segment order:
//		-	<title>.<year (optional)>.<misc pattern>...
//		-	<title>.<year (optional)>.<bonus pattern>...
//		-	<title>.<year (optional)>.<resolution, codec, source, or audio>...
//		-	<title>.<year (optional)>.<season or ep>...
//		-	<title>.<year (optional)>.<file ext>
//		-	<title>.<year (optional)>
func extractTitle(segments []string) []string {
	var title []string
	var year *int
	for i, segment := range segments {
		candidates := segments[i:]
		if	parseResolution(candidates) != "" 	||
			parseCodec(candidates) != ""		||
			parseSource(candidates) != "" 		||
			parseAudio(candidates) != "" 		||
			parseSeason(candidates) != nil 		||
			parseEpisode(candidates) != nil 	||
			parseVideoExt(candidates) != "" 	||
			parseSubtitleExt(candidates) != "" 	||
			parseMisc(candidates) != "" 		||
			parseAudioExt(candidates) != ""		||
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

// Returns year and -1 for no year found
// Extracts using segment order:
//		- <...>.<year (optional)>.<misc pattern>...
//		- <...>.<year (optional)>.<bonus pattern>...
//		- <...>.<year (optional)>.<resolution, codec, source, or audio>...
//		- <...>.<year (optional)>.<season or ep>...
//		- <...>.<year (optional)>.<file ext>...
//		- <...>.<year (optional)>
func extractYear(segments []string) *int {
	var year *int
	for i, segment := range segments {

		candidates := segments[i:]
		if	parseResolution(candidates) != "" 	||
			parseCodec(candidates) != ""		||
			parseSource(candidates) != "" 		||
			parseAudio(candidates) != "" 		||
			parseSeason(candidates) != nil 		||
			parseEpisode(candidates) != nil 	||
			parseVideoExt(candidates) != "" 	||
			parseSubtitleExt(candidates) != "" 	||
			parseMisc(candidates) != "" 		||
			parseAudioExt(candidates) != ""		||
			parseBonus(candidates) != "" {
			return year
		}
		year = parseYear(segment)
	}
	return year
}

// Returns nil for no season pattern, 0 for season without number, >0 for season number found
// Extracts without using expected segment order
func extractSeason(segments []string) *int {
	for i := range segments {
		candidates := segments[i:]
		if season := parseSeason(candidates); season != nil {
			return season
		}
	}
	return nil
}

// Returns nil for no episode pattern, 0 for episode without number, >0 for episode number found
// Extracts without using expected segment order
func extractEpisode(segments []string) *int {
	for i := range segments {
		candidates := segments[i:]
		if ep := parseEpisode(candidates); ep != nil {
			return ep
		}
	}

	return nil
}

// Returns resolution pattern or "" for no resolution pattern
// Extracts without using expected segment order
func extractResolution(segments []string) string {
	for i := range segments {
		candidates := segments[i:]
		if res := parseResolution(candidates); res != "" {
			return res
		}
	}
	return ""
}

// Returns codec pattern or "" for no codec pattern
// Extracts without using expected segment order
func extractCodec(segments []string) string {
	for i := range segments {
		candidates := segments[i:]
		if res := parseCodec(candidates); res != "" {
			return res
		}
	}
	return ""
}

// Returns source pattern or "" for no source pattern
// Extracts without using expected segment order
func extractSource(segments []string) string {
	for i := range segments {
		candidates := segments[i:]
		if res := parseSource(candidates); res != "" {
			return res
		}
	}
	return ""
}

// Returns audio pattern or "" for no audio pattern
// Extracts without using expected segment order
func extractAudio(segments []string) string {
	for i := range segments {
		candidates := segments[i:]
		if res := parseAudio(candidates); res != "" {
			return res
		}
	}
	return ""
}

func extractLanguage(segments []string) string {
	for i := range segments {
		candidates := segments[i:]
		if language := parseLanguage(candidates); language != "" {
			return language
		}
	}
	return ""
}

func extractBonus(segments []string) string {
	for i := range segments {
		candidates := segments[i:]
		if bonus := parseBonus(candidates); bonus != "" {
			return bonus
		}
	}
	return ""
}

// Helper function to return resolution if left most segments are a resolution or empty string if not
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

// Helper function to return codec if left most segments are a codec or empty string if not
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

// Helper function to return media source if left most segments are a media source or empty string if not
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

// Helper function to return audio if left most segments are a audio or empty string if not
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

// Returns nil for invalid year, otherwise returns year
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


// Returns nil for SEASON pattern not matched, 0 for match without number, >0 for season number
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
		
		return &unknown // matched but couldn't parse number
		
	}
	return nil
}


// Returns nil for EPISODE pattern not matched, 0 for match without number, >0 for EPISODE number
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
		return &unknown // matched but couldn't parse number
		
	}
	return nil
}

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

func parseMisc(segments []string) string {
	for _, re := range patterns.GetMiscPatterns() {
		if match := matchSegments(segments, (*regexp.Regexp)(re)); match != nil {
			return match[0]
		}
	}
	return ""
}

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
