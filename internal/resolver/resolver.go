// Package resolver determines destination paths for [metadata.Entry] objects based on their classified roles.
// Package relies on enricher package for contextual information and produces output used by transfer package.
package resolver

import (
	"fmt"
	"log/slog"
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
func Resolve(root *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	switch root.Role {
	case metadata.SubtitleFile, metadata.BonusFile, metadata.SubtitleDir, metadata.BonusDir:
		return fmt.Errorf("Entry %v cannot be processed at root level", root.PathInfo.Source)
	case metadata.EpisodeFile:
		return resolveEpisodeFile("", root, cfg, logger)
	case metadata.MovieFile:
		return resolveMovieFile("", root, cfg, logger)
	case metadata.SeasonDir:
		return resolveSeasonDir("", root, cfg, logger)
	case metadata.SeriesDir:
		return resolveSeriesDir(root, cfg, logger)
	case metadata.MovieDir:
		return resolveMovieDir(root, cfg, logger)
	default:
		return fmt.Errorf("Entry %v has unknown role", root.PathInfo.Source)
	}
}
/*
	Resolvers
*/

// resolveSubtitleFile sets destination path for subtitle file under "Subtitles" subdirectory.
// Groups by season if applicable. Clears quality metadata before building filename.
func resolveSubtitleFile(basePath string, entry *metadata.Entry, logger *slog.Logger) error {
	lg := logger.With("func", "resolveSubtitleFile")

	if entry.Role != metadata.SubtitleFile {
		lg.Debug("Expected SubtitleFile role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected SubtitleFile role, received %v for node %v", entry.Role, entry.PathInfo.Source)
	}
	if basePath == "" {
		lg.Debug("Expected base path for SubtitleFile role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected base path for SubtitlFile role, for node %v", entry.PathInfo.Source)
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
	basePath = filepath.Join(basePath, "Subtitles")

	entry.MediaInfo.Resolution = ""
	entry.MediaInfo.Codec = ""
	entry.MediaInfo.Source = ""
	entry.MediaInfo.Audio = ""
	entry.MediaInfo.Bonus = ""
	filename := buildFilename(entry)

	entry.PathInfo.Dest = filepath.Join(basePath, filename)
	lg.Debug("Resolved bonus file destination", "source", entry.PathInfo.Source, "destination", entry.PathInfo.Dest)

	return nil
}

// resolveSubtitleDir recursively sets destination path for subtitle directory and its children.
// Allows for intermediary subtitle directories, flattening directory structure.
func resolveSubtitleDir(basePath string, entry *metadata.Entry, logger *slog.Logger) error {
	lg := logger.With("func", "resolveSubtitleDir")

	if entry.Role != metadata.SubtitleDir {
		lg.Debug("Expected SubtitleDir role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected SubtitleDir role, received %v for node %v", entry.Role, entry.PathInfo.Source)
	}
	if basePath == "" {
		lg.Debug("Expected base path for SubtitleDir role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected base path for SubtitleDir role, for node %v", entry.PathInfo.Source)
	}

	for _, child := range entry.Children {
		var err error
		switch child.Role {
		case metadata.SubtitleFile:
			err = resolveSubtitleFile(basePath, child, logger)
		case metadata.SubtitleDir:
			// Allows for intermediary subtitle dirs, flattens dir structure
			err = resolveSubtitleDir(basePath, child, logger)
		default:
			err = fmt.Errorf("Unexpected child role for SubtitleDir, received role %v for node %v", entry.Role, entry.PathInfo.Source)
		}
		if err != nil {
			return err
		}
	}

	entry.PathInfo.Dest = basePath
	lg.Debug("Resolved subtitle dir destination", "source", entry.PathInfo.Source, "destination", entry.PathInfo.Dest)

	return nil
}

func resolveExtrasFile(basePath string, entry *metadata.Entry, logger *slog.Logger) error {
	lg := logger.With("func", "resolveExtrasFile")

	if entry.Role != metadata.DSFile && entry.Role != metadata.BTSFile && entry.Role != metadata.BonusFile {
		lg.Debug("Expected DSFile or BTSFile or BonusFile role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected DSFile or BTSFile or BonusFile role, received %v for node %v", entry.Role, entry.PathInfo.Source)
	}
	if basePath == "" {
		lg.Debug("Expected base path for extras file role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected base path for extras file role, for node %v", entry.PathInfo.Source)
	}

	// Bonus files will group themselves under the same bonus folder, seperated by season if necessary
	if entry.MediaInfo.Season != nil {
		seasonPath := buildSeasonPath(entry)	
		basePath = filepath.Join(basePath, seasonPath) 
	}
	extrasPath := buildExtrasPath(entry)
	basePath = filepath.Join(basePath, extrasPath)

	title := extractor.SanitizeName(entry.PathInfo.Source)
	entry.MediaInfo.Title = title[:len(title)-1]
	filename := buildReadableFilename(entry)

	entry.PathInfo.Dest = filepath.Join(basePath, filename)
	lg.Debug("Resolved bonus file destination", "source", entry.PathInfo.Source, "destination", entry.PathInfo.Dest)

	return nil
}

func resolveExtrasDir(basePath string, entry *metadata.Entry, logger *slog.Logger) error {
	lg := logger.With("func", "resolveExtrasDir")

	if entry.Role != metadata.DSDir && entry.Role != metadata.BTSDir && entry.Role != metadata.BonusDir {
		lg.Debug("Expected DSDir or BTSDir or BonusDir role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected DSDir or BTSDir or BonusDir role, received %v for node %v", entry.Role, entry.PathInfo.Source)
	}
	if basePath == "" {
		lg.Debug("Expected base path for extras dir role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected base path for extras dir role, for node %v", entry.PathInfo.Source)
	}

	for _, child := range entry.Children {
		var err error
		switch child.Role {
		case metadata.DSFile, metadata.BTSFile, metadata.BonusFile:
			err = resolveExtrasFile(basePath, child, logger)
		case metadata.SubtitleFile:
			err = resolveSubtitleFile(basePath, child, logger)
		case metadata.DSDir, metadata.BTSDir, metadata.BonusDir:
			// Allows for intermediary bonus dirs, flattens dir structure
			err = resolveExtrasDir(basePath, child, logger)
		default:
			err = fmt.Errorf("Unexpected child role for extras dir, received role %v for node %v", entry.Role, entry.PathInfo.Source)
		}
		if err != nil {
			return err
		}
	}

	entry.PathInfo.Dest = basePath
	lg.Debug("Resolved bonus dir destination", "source", entry.PathInfo.Source, "destination", entry.PathInfo.Dest)

	return nil
}

// resolveEpisodeFile sets destination path for episode file.
// Always appends its own season path to basePath based on episode metadata.
func resolveEpisodeFile(basePath string, entry *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "resolveEpisodeFile")

	if entry.Role != metadata.EpisodeFile {
		lg.Debug("Expected EpisodeFile role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected EpisodeFile role, received %v for node %v", entry.Role, entry.PathInfo.Source)
	}

	if basePath == "" {
		// Main series path should not include year, to ensure seasons are more likely to fall under same directory
		entry.MediaInfo.Year = nil

		titlePath := buildTitlePath(entry)
		basePath = filepath.Join(cfg.ShowPath, titlePath)
	}

	// Episode always determines its own season directory
	seasonPath := buildSeasonPath(entry)
	basePath = filepath.Join(basePath, seasonPath)

	entry.MediaInfo.Bonus = ""
	filename := buildFilename(entry)

	entry.PathInfo.Dest = filepath.Join(basePath, filename)
	lg.Debug("Resolved episode file destination", "source", entry.PathInfo.Source, "destination", entry.PathInfo.Dest)

	return nil
}

// resolveSeasonDir recursively sets destination path for season directory and its children.
// Passes title-only basePath to episode children, which determine their own season paths.
func resolveSeasonDir(basePath string, entry *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "resolveSeasonDir")

	if entry.Role != metadata.SeasonDir {
		lg.Debug("Expected SeasonDir role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected SeasonDir role, received %v for node %v", entry.Role, entry.PathInfo.Source)
	}

	// Build title-only basePath for children - episodes will add their own season
	if basePath == "" {
		// Main series path should not include year, to ensure seasons are more likely to fall under same directory
		entry.MediaInfo.Year = nil

		titlePath := buildTitlePath(entry)
		basePath = filepath.Join(cfg.ShowPath, titlePath)
	}

	for _, child := range entry.Children {
		var err error
		switch child.Role {
		case metadata.EpisodeFile:
			// Pass title-only basePath; episode builds its own season path
			err = resolveEpisodeFile(basePath, child, cfg, logger)
		case metadata.SubtitleFile:
			err = resolveSubtitleFile(basePath, child, logger)
		case metadata.SubtitleDir:
			err = resolveSubtitleDir(basePath, child, logger)
		case metadata.DSDir, metadata.BTSDir, metadata.BonusDir:
			err = resolveExtrasDir(basePath, child, logger)
		case metadata.DSFile, metadata.BTSFile, metadata.BonusFile:
			err = resolveExtrasFile(basePath, child, logger)
		default:
			err = fmt.Errorf("Unexpected child role for SeasonDir, received role %v for node %v", child.Role, child.PathInfo.Source)
		}
		if err != nil {
			return err
		}
	}

	// SeasonDir dest should include season path to show actual directory structure
	seasonPath := buildSeasonPath(entry)
	entry.PathInfo.Dest = filepath.Join(basePath, seasonPath)
	lg.Debug("Resolved season dir destination", "source", entry.PathInfo.Source, "destination", entry.PathInfo.Dest)

	return nil
}



// resolveSeriesDir recursively sets destination path for series directory and its children.
// Top-level resolver that constructs basePath from cfg.ShowPath and title.
func resolveSeriesDir(entry *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "resolveSeriesDir")

	if entry.Role != metadata.SeriesDir {
		lg.Debug("Expected SeriesDir role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected SeriesDir role, received %v for node %v", entry.Role, entry.PathInfo.Source)
	}

	// Main series path should not include year, to ensure seasons are more likely to fall under same directory
	entry.MediaInfo.Year = nil

	titlePath := buildTitlePath(entry)
	basePath := filepath.Join(cfg.ShowPath, titlePath)

	for _, child := range entry.Children {
		var err error
		switch child.Role {
		case metadata.SeasonDir:
			err = resolveSeasonDir(basePath, child, cfg, logger)
		case metadata.DSDir, metadata.BTSDir, metadata.BonusDir:
			err = resolveExtrasDir(basePath, child, logger)
		case metadata.SubtitleDir:
			err = resolveSubtitleDir(basePath, child, logger)
		default:
			return fmt.Errorf("Unexpected child role for SeriesDir, received role %v for node %v", entry.Role, entry.PathInfo.Source)
		}
		if err != nil {
			return err
		}
	}

	entry.PathInfo.Dest = basePath
	lg.Debug("Resolved series dir destination", "source", entry.PathInfo.Source, "destination", entry.PathInfo.Dest)

	return nil
}

// resolveMovieFile sets destination path for movie file.
// Clears season, episode, and bonus metadata before building filename.
func resolveMovieFile(basePath string, entry *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "resolveMovieFile")

	if entry.Role != metadata.MovieFile {
		lg.Debug("Expected MovieFile role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected MovieFile role, received %v for node %v", entry.Role, entry.PathInfo.Source)
	}

	if basePath == "" {
		titlePath := buildTitlePath(entry)
		basePath = filepath.Join(cfg.MoviePath, titlePath)
	}

	entry.MediaInfo.Season = nil
	entry.MediaInfo.Episode = nil
	entry.MediaInfo.Bonus = ""
	filename := buildFilename(entry)

	entry.PathInfo.Dest = filepath.Join(basePath, filename)
	lg.Debug("Resolved movie file destination", "source", entry.PathInfo.Source, "destination", entry.PathInfo.Dest)

	return nil
}

// resolveMovieDir recursively sets destination path for movie directory and its children.
// Top-level resolver that constructs basePath from cfg.MoviePath and title.
func resolveMovieDir(entry *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "resolveMovieDir")

	if entry.Role != metadata.MovieDir {
		lg.Debug("Expected MovieDir role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected MovieDir role, received %v for node %v", entry.Role, entry.PathInfo.Source)
	}

	titlePath := buildTitlePath(entry)
	basePath := filepath.Join(cfg.MoviePath, titlePath)

	for _, child := range entry.Children {
		var err error
		switch child.Role {
		case metadata.MovieFile:
			err = resolveMovieFile(basePath, child, cfg, logger)
		case metadata.SubtitleFile:
			err = resolveSubtitleFile(basePath, child, logger)
		case metadata.DSFile, metadata.BTSFile, metadata.BonusFile:
			err = resolveExtrasFile(basePath, child, logger)
		case metadata.SubtitleDir:
			err = resolveSubtitleDir(basePath, child, logger)
		case metadata.DSDir, metadata.BTSDir, metadata.BonusDir:
			err = resolveExtrasDir(basePath, child, logger)
		default:
			return fmt.Errorf("Unexpected child role for MovieDir, received role %v for node %v", entry.Role, entry.PathInfo.Source)
		}
		
		if err != nil {
			return err
		}
	}

	entry.PathInfo.Dest = basePath
	lg.Debug("Resolved movie dir destination", "source", entry.PathInfo.Source, "destination", entry.PathInfo.Dest)

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
	if entry.MediaInfo.Resolution != "" {
		filename += "." + entry.MediaInfo.Resolution
	}
	if entry.MediaInfo.Codec != "" {
		filename += "." + entry.MediaInfo.Codec
	}
	if entry.MediaInfo.Source != "" {
		filename += "." + entry.MediaInfo.Source
	}
	if entry.MediaInfo.Audio != "" {
		filename += "." + entry.MediaInfo.Audio
	}
	if len(entry.MediaInfo.Language) > 0 {
		filename += "." + strings.Join(entry.MediaInfo.Language, ".")
	}
	if entry.MediaInfo.Bonus != "" {
		filename += "." + entry.MediaInfo.Bonus
	}
	if entry.MediaInfo.Edition != "" {
		filename += "." + entry.MediaInfo.Edition
	}

	return filename + "." + strings.ToLower(entry.PathInfo.Ext)
}

func buildReadableFilename(entry *metadata.Entry) string {
    capitalized := make([]string, len(entry.MediaInfo.Title))
    for i, part := range entry.MediaInfo.Title {
        capitalized[i] = capitalize(part)
    }
    filename := strings.Join(capitalized, " ")

	filename += "." + strings.ToLower(entry.PathInfo.Ext) 
    
    return filename

}

// buildTitlePath constructs directory path from title and year metadata.
// Capitalizes each title component and joins with dots.
func buildTitlePath(entry *metadata.Entry) string {
    capitalized := make([]string, len(entry.MediaInfo.Title))
    for i, part := range entry.MediaInfo.Title {
        capitalized[i] = capitalize(part)
    }
    basePath := strings.Join(capitalized, ".")
    
    if entry.MediaInfo.Year != nil {
        basePath += "." + strconv.Itoa(*entry.MediaInfo.Year)
    }
    return basePath
}

// buildSeasonPath constructs season directory name from season metadata.
// Defaults to "S01" if season is nil.
func buildSeasonPath(entry *metadata.Entry) string {
	if entry.MediaInfo.Season != nil {
		return fmt.Sprintf("S%02d", *entry.MediaInfo.Season)
	}
	return "S01"
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


