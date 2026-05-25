// Package resolver determines destination paths for [metadata.Entry] objects based on their classified roles.
// Package relies on enricher package for contextual information and produces output used by transfer package.
package resolver

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"unicode"

	"github.com/ENIACore/media_library_manager/internal/config"
	"github.com/ENIACore/media_library_manager/internal/extractor"
	"github.com/ENIACore/media_library_manager/internal/metadata"
)

// Resolve determines destination paths for all entries in tree based on classified roles.
// Returns error if root entry has invalid role for top-level processing.
func Resolve(root *metadata.Entry, cfg *config.Config) error {
	switch root.Role {
		case metadata.SubtitleFile, metadata.BonusFile, metadata.SubtitleDir, metadata.BonusDir:
			return fmt.Errorf("Entry %v cannot be processed at root level", root.FileInfo.SourcePath)
		case metadata.EpisodeFile:
			return resolveEpisodeFile("", root, cfg)
		case metadata.MovieFile:
			return resolveMovieFile("", root, cfg)
		case metadata.SeasonDir:
			return resolveSeasonDir("", root, cfg)
		case metadata.SeriesDir:
			return resolveSeriesDir(root, cfg)
		case metadata.MovieDir:
			return resolveMovieDir(root, cfg)
		default:
			return fmt.Errorf("Entry %v has unknown role", root.FileInfo.SourcePath)
	}
}
/*
	Resolvers
*/

// resolveSubtitleFile sets destination path for subtitle file at same level as movie/episode/extras files 
// Groups by season if applicable. Clears quality metadata before building filename.
// Subtitle files are non-essential and don't return errors if processing is not possible
func resolveSubtitleFile(basePath string, entry *metadata.Entry) {

	if entry.Role != metadata.SubtitleFile || basePath == "" {
		return
	}
	
	// If subtitle is a part of season, it will be placed in season dir
	if entry.MediaInfo.Season != nil {
		seasonPath := buildSeasonPath(entry)	
		basePath = filepath.Join(basePath, seasonPath) 
	}

	// If subtitle is a part of extras, it will be placed in extras dir (also in season dir if applicable)
	if entry.MediaInfo.DS != "" || entry.MediaInfo.BTS != "" || entry.MediaInfo.Bonus != "" {
		extrasPath := buildExtrasPath(entry)	
		basePath = filepath.Join(basePath, extrasPath)
	}

	if filename, err := buildSubtitleFileName(entry); err == nil {
		entry.FileInfo.DestPath = filepath.Join(basePath, filename)
	} else {
		entry.FileInfo.DestPath = ""
	}
}

// resolveSubtitleDir recursively sets destination path for subtitle directory and its children.
// Allows for intermediary subtitle directories, flattening directory structure.
// Subtitle directories are non-essential and don't return errors if processing is not possible
func resolveSubtitleDir(basePath string, entry *metadata.Entry) {

	if entry.Role != metadata.SubtitleDir || basePath == "" {
		return
	}

	for _, child := range entry.Children {
		switch child.Role {
			case metadata.SubtitleFile:
				resolveSubtitleFile(basePath, child)
			case metadata.SubtitleDir:
				resolveSubtitleDir(basePath, child)
			default:
				return
		}
	}

	entry.FileInfo.DestPath = basePath
}

func resolveExtrasFile(basePath string, entry *metadata.Entry) error {
	if entry.Role != metadata.DSFile && entry.Role != metadata.BTSFile && entry.Role != metadata.BonusFile {
		return fmt.Errorf("Expected DSFile or BTSFile or BonusFile role, received %v for node %v", entry.Role.String(), entry.Source())
	}
	if basePath == "" {
		return fmt.Errorf("Expected base path for extras file role, for node %v", entry.Source())
	}

	if entry.MediaInfo.Season != nil {
		seasonPath := buildSeasonPath(entry)	
		basePath = filepath.Join(basePath, seasonPath) 
	}
	extrasPath := buildExtrasPath(entry)
	basePath = filepath.Join(basePath, extrasPath)

	title := extractor.SanitizeName(entry.FileInfo.SourcePath)
	entry.MediaInfo.Title = title[:len(title)-1] // Remove file ext from filename
	filename := buildReadableFilename(entry)

	entry.FileInfo.DestPath = filepath.Join(basePath, filename)
	return nil
}

func resolveExtrasDir(basePath string, entry *metadata.Entry) error {
	if entry.Role != metadata.DSDir && entry.Role != metadata.BTSDir && entry.Role != metadata.BonusDir {
		return fmt.Errorf("Expected DSDir or BTSDir or BonusDir role, received %v for node %v", entry.Role.String(), entry.Source())
	}
	if basePath == "" {
		return fmt.Errorf("Expected base path for extras dir role, for node %v", entry.Source())
	}

	for _, child := range entry.Children {
		switch child.Role {
		case metadata.DSFile, metadata.BTSFile, metadata.BonusFile:
			if err := resolveExtrasFile(basePath, child); err != nil {
				return err
			}
		case metadata.SubtitleFile:
			resolveSubtitleFile(basePath, child)
		case metadata.DSDir, metadata.BTSDir, metadata.BonusDir:
			if err := resolveExtrasDir(basePath, child); err != nil { // Allows for intermediary bonus dirs, flattens dir structure
				return err
			}
		default:
			return fmt.Errorf("Unexpected child role for extras dir, received role %v for node %v", entry.Role.String(), entry.Source())
		}
	}

	entry.FileInfo.DestPath = basePath
	return nil
}

// resolveEpisodeFile sets destination path for episode file.
// Always appends its own season path to basePath based on episode metadata.
func resolveEpisodeFile(basePath string, entry *metadata.Entry, cfg *config.Config) error {
	if entry.Role != metadata.EpisodeFile {
		return fmt.Errorf("Expected EpisodeFile role, received %v for node %v", entry.Role.String(), entry.Source())
	}

	if basePath == "" {
		title, err := buildTitle(entry)
		if err != nil {
			return fmt.Errorf("Unable to resolve episode file: %w", err)
		}
		basePath = filepath.Join(cfg.ShowPath, title)
	}

	seasonPath := buildSeasonPath(entry) // Episode always determines its own season directory
	basePath = filepath.Join(basePath, seasonPath)

	filename, err := buildEpisodeFileName(entry)
	if err != nil {
		return err
	}

	entry.FileInfo.DestPath = filepath.Join(basePath, filename)
	return nil
}

// resolveSeasonDir recursively sets destination path for season directory and its children.
// Passes title-only basePath to episode children, which determine their own season paths.
func resolveSeasonDir(basePath string, entry *metadata.Entry, cfg *config.Config) error {
	if entry.Role != metadata.SeasonDir {
		return fmt.Errorf("Expected SeasonDir role, received %v for node %v", entry.Role.String(), entry.Source())
	}

	// Build title-only basePath for children - episodes will add their own season
	if basePath == "" {
		title, err := buildTitle(entry) 
		if err != nil {
			return fmt.Errorf("Unable to resolve season dir: %w", err)
		}
		basePath = filepath.Join(cfg.ShowPath, title)
	}

	for _, child := range entry.Children {
		switch child.Role {
			case metadata.EpisodeFile:
				if err := resolveEpisodeFile(basePath, child, cfg); err != nil { // Pass title-only basePath; episode builds its own season path
					return err
				}
			case metadata.SubtitleFile:
				resolveSubtitleFile(basePath, child)
			case metadata.SubtitleDir:
				resolveSubtitleDir(basePath, child)
			case metadata.DSDir, metadata.BTSDir, metadata.BonusDir:
				if err := resolveExtrasDir(basePath, child); err != nil {
					return err
				}
			case metadata.DSFile, metadata.BTSFile, metadata.BonusFile:
				if err := resolveExtrasFile(basePath, child); err != nil {
					return err
				}
			default:
				return fmt.Errorf("Unexpected child role for SeasonDir, received role %v for node %v", child.Role.String(), child.Source())
		}
	}

	seasonPath := buildSeasonPath(entry) // SeasonDir dest should include season path to show actual directory structure
	entry.FileInfo.DestPath = filepath.Join(basePath, seasonPath)
	return nil
}



// resolveSeriesDir recursively sets destination path for series directory and its children.
// Top-level resolver that constructs basePath from cfg.ShowPath and title.
func resolveSeriesDir(entry *metadata.Entry, cfg *config.Config) error {
	if entry.Role != metadata.SeriesDir {
		return fmt.Errorf("Expected SeriesDir role, received %v for node %v", entry.Role.String(), entry.Source())
	}

	title, err := buildTitle(entry)
	if err != nil {
		return fmt.Errorf("Unable to resolve series dir: %w", err)
	}
	basePath := filepath.Join(cfg.ShowPath, title)

	for _, child := range entry.Children {
		switch child.Role {
			case metadata.SeasonDir:
				if err := resolveSeasonDir(basePath, child, cfg); err != nil {
					return err
				}
			case metadata.DSDir, metadata.BTSDir, metadata.BonusDir:
				if err := resolveExtrasDir(basePath, child); err != nil {
					return err
				}
			case metadata.SubtitleDir:
				resolveSubtitleDir(basePath, child)
			default:
				return fmt.Errorf("Unexpected child role for SeriesDir, received role %v for node %v", entry.Role.String(), entry.Source())
		}
	}

	entry.FileInfo.DestPath = basePath
	return nil
}

// resolveMovieFile sets destination path for movie file.
// Clears season, episode, and bonus metadata before building filename.
func resolveMovieFile(basePath string, entry *metadata.Entry, cfg *config.Config) error {
	if entry.Role != metadata.MovieFile {
		return fmt.Errorf("Expected MovieFile role, received %v for node %v", entry.Role.String(), entry.Source())
	}
	if basePath == "" {
		titlePath, err := buildTitle(entry)
		if err != nil {
			return err
		}
		basePath = filepath.Join(cfg.MoviePath, titlePath)
	}

	filename, err := buildMovieFileName(entry)
	if err != nil {
		return err
	}

	entry.FileInfo.DestPath = filepath.Join(basePath, filename)
	return nil
}

// resolveMovieDir recursively sets destination path for movie directory and its children.
// Top-level resolver that constructs basePath from cfg.MoviePath and title.
func resolveMovieDir(entry *metadata.Entry, cfg *config.Config) error {
	if entry.Role != metadata.MovieDir {
		return fmt.Errorf("Expected MovieDir role, received %v for node %v", entry.Role.String(), entry.Source())
	}

	titlePath, err := buildTitle(entry)
	if err != nil {
		return fmt.Errorf("Unable to resolve movie dir: %w", err)
	}
	basePath := filepath.Join(cfg.MoviePath, titlePath)

	for _, child := range entry.Children {
		switch child.Role {
			case metadata.MovieFile:
				if err := resolveMovieFile(basePath, child, cfg); err != nil {
					return err
				}
			case metadata.SubtitleFile:
				resolveSubtitleFile(basePath, child)
			case metadata.DSFile, metadata.BTSFile, metadata.BonusFile:
				if err := resolveExtrasFile(basePath, child); err != nil {
					return err
				}
			case metadata.SubtitleDir:
				resolveSubtitleDir(basePath, child)
			case metadata.DSDir, metadata.BTSDir, metadata.BonusDir:
				if err := resolveExtrasDir(basePath, child); err != nil {
					return err
				}
			default:
				return fmt.Errorf("Unexpected child role for MovieDir, received role %v for node %v", entry.Role.String(), entry.Source())
		}
	}

	entry.FileInfo.DestPath = basePath
	return nil
}

/*
	Resolver Helpers
*/

// buildFilename constructs standardized filename from entry metadata.
// Nil fields are omitted from output. Returns filename with lowercase extension.
func buildFilename(entry *metadata.Entry) string {
	filename := ""

    capitalized := make([]string, len(entry.MediaInfo.Title))
    for i, part := range entry.MediaInfo.Title {
        capitalized[i] = capitalize(part)
    }
    filename = strings.Join(capitalized, ".")
    
    if entry.MediaInfo.Year != nil {
        filename += "." + strconv.Itoa(*entry.MediaInfo.Year)
    }
	if entry.MediaInfo.Season != nil || entry.MediaInfo.Episode != nil {
		filename += "."
		if entry.MediaInfo.Season != nil {
			filename += fmt.Sprintf("S%02d", *entry.MediaInfo.Season)
		}
		if entry.MediaInfo.Episode != nil {
			filename += fmt.Sprintf("E%02d", *entry.MediaInfo.Episode)
		}
	}
	if entry.FileInfo.Resolution != "" {
		filename += "." + entry.FileInfo.Resolution
	}
	if entry.FileInfo.Codec != "" {
		filename += "." + entry.FileInfo.Codec
	}
	if entry.FileInfo.Audio != "" {
		filename += "." + entry.FileInfo.Audio
	}
	if len(entry.FileInfo.Language) > 0 {
		filename += "." + strings.Join(entry.FileInfo.Language, ".")
	}
	if entry.MediaInfo.Bonus != "" {
		filename += "." + entry.MediaInfo.Bonus
	}
	if entry.MediaInfo.Edition != "" {
		filename += "." + entry.MediaInfo.Edition
	}

	return filename + "." + strings.ToLower(entry.FileInfo.Ext)
}

func buildReadableFilename(entry *metadata.Entry) string {
    capitalized := make([]string, len(entry.MediaInfo.Title))
    for i, part := range entry.MediaInfo.Title {
        capitalized[i] = capitalize(part)
    }
    filename := strings.Join(capitalized, " ")

	filename += "." + strings.ToLower(entry.FileInfo.Ext) 
    
    return filename

}

// buildVideoFileStem constructs the filename stem shared by episode, movie, and subtitle files.
// Format: Name (Year) [tmdbid] [ S##E##] [ - audio - bitrate - codec - resolution]
// Season/episode suffix is only added when both are non-nil.
func buildVideoFileStem(entry *metadata.Entry) (string, error) {
	stem, err := buildTitle(entry)
	if err != nil {
		return "", err
	}

	if entry.MediaInfo.Season != nil && entry.MediaInfo.Episode != nil {
		stem += fmt.Sprintf(" S%02d", *entry.MediaInfo.Season)
		stem += fmt.Sprintf("E%02d", *entry.MediaInfo.Episode)
	}

	if entry.MediaInfo.Source != "" {
		stem += " - " + entry.MediaInfo.Source
	}
	if entry.FileInfo.Audio != "" {
		stem += " - " + entry.FileInfo.Audio
	}
	if entry.FileInfo.AspectRatio != "" {
		stem += " - " + entry.FileInfo.AspectRatio
	}
	if entry.FileInfo.Bitrate != "" {
		stem += " - " + entry.FileInfo.Bitrate
	}
	if entry.FileInfo.BitDepth != "" {
		stem += " - " + entry.FileInfo.BitDepth
	}
	if entry.FileInfo.Codec != "" {
		stem += " - " + entry.FileInfo.Codec
	}
	if entry.FileInfo.Resolution != "" {
		stem += " - " + entry.FileInfo.Resolution
	}

	return stem, nil
}

// buildEpisodeFileName builds: Name (Year) [tmdbid]  S##E## - audio - bitrate - codec - resolution.ext
func buildEpisodeFileName(entry *metadata.Entry) (string, error) {
	if entry.MediaInfo.Season == nil || entry.MediaInfo.Episode == nil {
		return "", fmt.Errorf("No episode or season in episode file")
	}

	stem, err := buildVideoFileStem(entry)
	if err != nil {
		return "", nil
	}

	return stem + "." + strings.ToLower(entry.FileInfo.Ext), nil
}

// buildSubtitleFileName builds the same stem as the sibling video file and appends .language.ext
// Returns error if language is not set.
func buildSubtitleFileName(entry *metadata.Entry) (string, error) {
	if len(entry.FileInfo.Language) == 0 {
		return "", fmt.Errorf("No language detected for subtitle")
	}

	stem, err := buildVideoFileStem(entry)
	if err != nil {
		return "", err
	}

	return stem + "." + entry.FileInfo.Language[0] + "." + strings.ToLower(entry.FileInfo.Ext), nil
}

// buildMovieFileName builds: Name (Year) [tmdbid] - audio - bitrate - codec - resolution.ext
func buildMovieFileName(entry *metadata.Entry) (string, error) {
	stem, err := buildVideoFileStem(entry)
	if err != nil {
		return "", nil
	}

	return stem + "." + strings.ToLower(entry.FileInfo.Ext), nil
}

// Builds title using title, year, and tmdbid: Name (Year) [tmdbid]
func buildTitle(entry *metadata.Entry) (string, error) {
	if len(entry.MediaInfo.Title) == 0 {
		return "", fmt.Errorf("Entry %v does not have title, unable to build title", entry.Source())
	}
	if entry.MediaInfo.Year == nil {
		return "", fmt.Errorf("Entry %v does not have year, unable to build title", entry.Source())
	}

    capitalized := make([]string, len(entry.MediaInfo.Title))
    for i, part := range entry.MediaInfo.Title {
        capitalized[i] = capitalize(part)
    }
	title := strings.Join(capitalized, " ")
	title += " (" + strconv.Itoa(*entry.MediaInfo.Year) + ")"

	if entry.MediaInfo.TMDBid != 0 {
		title = title + " [" + entry.TMDB() + "]"
	}

	return title, nil
}

// buildSeasonPath constructs season directory name from season metadata.
// Defaults to "S01" if season is nil.
func buildSeasonPath(entry *metadata.Entry) string {
	if entry.MediaInfo.Season != nil {
		return fmt.Sprintf("Season %02d", *entry.MediaInfo.Season)
	}
	return "Season 01"
}

func buildExtrasPath(entry *metadata.Entry) string {
	if entry.MediaInfo.DS != "" {
		return "deleted scenes"
	}
	if entry.MediaInfo.BTS != "" {
		return "behind the scenes"
	}
	if entry.MediaInfo.Bonus != "" {
		return "extras"
	}
	return "extras"
}

// capitalize returns string with first rune uppercase and all others lowercase.
// If the string is a Roman numeral, all letters are capitalized.
func capitalize(s string) string {
    if s == "" {
        return ""
    }
    if isRomanNumeral(s) {
        return strings.ToUpper(s)
    }
    s = strings.ToLower(s)
    r := []rune(s)
    r[0] = unicode.ToUpper(r[0])
    return string(r)
}

// isRomanNumeral returns true if s is a valid Roman numeral.
func isRomanNumeral(s string) bool {
    matched, _ := regexp.MatchString(`(?i)^M{0,4}(CM|CD|D?C{0,3})(XC|XL|L?X{0,3})(IX|IV|V?I{0,3})$`, s)
    return matched && s != ""
}
