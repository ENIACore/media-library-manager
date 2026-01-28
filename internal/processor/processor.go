package processor

import (
	"fmt"
	"log/slog"
	"strings"
	"strconv"
	"path/filepath"

	"unicode"
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/ENIACore/media_library_manager/internal/config"
)

// Resolve determines destination paths for all entries in the tree based on their classified roles.
// Routes the root entry to the appropriate resolver function based on its role.
// Returns an error if the root entry has an invalid role for top-level processing.
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

// resolveSubtitleFile - TODO documentation, subtitle files will determine their own path based on context of parent and self
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
	
	// Subtitle files will group themselves under the same Subtitle folder, seperated by season if necessary
	if entry.MediaInfo.Season != nil {
		seasonPath := buildSeasonPath(entry)	
		basePath = filepath.Join(basePath, seasonPath) 
	}
	// If parent is bonus dir, group subtitles under resulting bonus dir
	if entry.Parent != nil && entry.Parent.Role == metadata.BonusDir {
		basePath = filepath.Join(basePath, "Extras")
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

// resolveSubtitleDir sets the destination path for a subtitle directory and its children.
// Requires a basePath from parent. Creates a "Subtitles" subdirectory and resolves all subtitle file children.
// Sets entry.PathInfo.Dest to the subtitle directory path.
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

// resolveBonusFile sets the destination path for a bonus content file.
// Requires a basePath from parent. Builds filename from metadata and sets entry.PathInfo.Dest.
func resolveBonusFile(basePath string, entry *metadata.Entry, logger *slog.Logger) error {
	lg := logger.With("func", "resolveBonusFile")

	if entry.Role != metadata.BonusFile {
		lg.Debug("Expected BonusFile role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected BonusFile role, received %v for node %v", entry.Role, entry.PathInfo.Source)
	}
	if basePath == "" {
		lg.Debug("Expected base path for BonusFile role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected base path for BonusFile role, for node %v", entry.PathInfo.Source)
	}

	// Bonus files will group themselves under the same bonus folder, seperated by season if necessary
	if entry.MediaInfo.Season != nil {
		seasonPath := buildSeasonPath(entry)	
		basePath = filepath.Join(basePath, seasonPath) 
	}
	basePath = filepath.Join(basePath, "Extras")

	filename := buildFilename(entry)

	entry.PathInfo.Dest = filepath.Join(basePath, filename)
	lg.Debug("Resolved bonus file destination", "source", entry.PathInfo.Source, "destination", entry.PathInfo.Dest)

	return nil
}

// resolveBonusDir sets the destination path for a bonus content directory and its children.
// Requires a basePath from parent. Creates an "Extras" subdirectory and resolves all bonus and subtitle file children.
// Sets entry.PathInfo.Dest to the extras directory path.
func resolveBonusDir(basePath string, entry *metadata.Entry, logger *slog.Logger) error {
	lg := logger.With("func", "resolveBonusDir")

	if entry.Role != metadata.BonusDir {
		lg.Debug("Expected BonusDir role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected BonusDir role, received %v for node %v", entry.Role, entry.PathInfo.Source)
	}
	if basePath == "" {
		lg.Debug("Expected base path for BonusDir role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected base path for BonusDir role, for node %v", entry.PathInfo.Source)
	}

	for _, child := range entry.Children {
		var err error
		switch child.Role {
		case metadata.BonusFile:
			err = resolveBonusFile(basePath, child, logger)
		case metadata.SubtitleFile:
			err = resolveSubtitleFile(basePath, child, logger)
		case metadata.BonusDir:
			// Allows for intermediary bonus dirs, flattens dir structure
			err = resolveBonusDir(basePath, child, logger)
		default:
			err = fmt.Errorf("Unexpected child role for BonusDir, received role %v for node %v", entry.Role, entry.PathInfo.Source)
		}
		if err != nil {
			return err
		}
	}

	entry.PathInfo.Dest = basePath
	lg.Debug("Resolved bonus dir destination", "source", entry.PathInfo.Source, "destination", entry.PathInfo.Dest)

	return nil
}

// resolveEpisodeFile sets the destination path for an episode file.
// Always appends its own season path to basePath based on episode metadata.
func resolveEpisodeFile(basePath string, entry *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "resolveEpisodeFile")

	if entry.Role != metadata.EpisodeFile {
		lg.Debug("Expected EpisodeFile role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected EpisodeFile role, received %v for node %v", entry.Role, entry.PathInfo.Source)
	}

	if basePath == "" {
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

// resolveSeasonDir sets the destination path for a season directory and its children.
// Passes title-only basePath to episode children, letting them determine their own season paths.
func resolveSeasonDir(basePath string, entry *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "resolveSeasonDir")

	if entry.Role != metadata.SeasonDir {
		lg.Debug("Expected SeasonDir role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected SeasonDir role, received %v for node %v", entry.Role, entry.PathInfo.Source)
	}

	// Build title-only basePath for children - episodes will add their own season
	if basePath == "" {
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
		default:
			err = fmt.Errorf("Unexpected child role for SeasonDir, received role %v for node %v", child.Role, child.PathInfo.Source)
		}
		if err != nil {
			return err
		}
	}

	// SeasonDir dest is informational only - actual structure determined by children
	entry.PathInfo.Dest = basePath
	lg.Debug("Resolved season dir destination", "source", entry.PathInfo.Source, "destination", entry.PathInfo.Dest)

	return nil
}



// resolveSeriesDir sets the destination path for a series directory and its children.
// Constructs basePath from cfg.ShowPath and title. Resolves all season, bonus, and subtitle directory children.
// Sets entry.PathInfo.Dest. Top-level resolver, does not accept basePath from parent.
func resolveSeriesDir(entry *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "resolveSeriesDir")

	if entry.Role != metadata.SeriesDir {
		lg.Debug("Expected SeriesDir role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected SeriesDir role, received %v for node %v", entry.Role, entry.PathInfo.Source)
	}

	titlePath := buildTitlePath(entry)
	basePath := filepath.Join(cfg.ShowPath, titlePath)

	for _, child := range entry.Children {
		var err error
		switch child.Role {
		case metadata.SeasonDir:
			err = resolveSeasonDir(basePath, child, cfg, logger)
		case metadata.BonusDir:
			err = resolveBonusDir(basePath, child, logger)
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

// resolveMovieFile sets the destination path for a movie file.
// If basePath is empty, constructs path from cfg.MoviePath and title.
// Clears season, episode, and bonus metadata before building filename. Sets entry.PathInfo.Dest.
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

// resolveMovieDir sets the destination path for a movie directory and its children.
// Constructs basePath from cfg.MoviePath and title. Resolves all movie file, subtitle, and bonus children.
// Sets entry.PathInfo.Dest. Top-level resolver, does not accept basePath from parent.
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
		case metadata.BonusFile:
			err = resolveBonusFile(basePath, child, logger)
		case metadata.SubtitleDir:
			err = resolveSubtitleDir(basePath, child, logger)
		case metadata.BonusDir:
			err = resolveBonusDir(basePath, child, logger)
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

// buildFilename constructs a standardized filename from entry metadata.
// Includes title, year, season, episode, resolution, codec, source, audio, language, and bonus fields.
// Returns filename with lowercase extension. Nil fields are omitted from output.
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
	if entry.MediaInfo.Language != "" {
		filename += "." + entry.MediaInfo.Language
	}
	if entry.MediaInfo.Bonus != "" {
		filename += "." + entry.MediaInfo.Bonus
	}

	return filename + "." + strings.ToLower(entry.PathInfo.Ext)
}

// buildTitlePath constructs a directory path from title and year metadata.
// Capitalizes each title component and joins with dots. Appends year if present.
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

// buildSeasonPath constructs a season directory name from season metadata.
// Returns formatted season (e.g., "S01") or defaults to "S01" if season is nil.
func buildSeasonPath(entry *metadata.Entry) string {
	if entry.MediaInfo.Season != nil {
		return fmt.Sprintf("S%02d", *entry.MediaInfo.Season)
	}
	return "S01"
}

// capitalize returns the string with its first rune converted to uppercase and all others to lowercase.
// Returns empty string if input is empty.
func capitalize(s string) string {
    if s == "" {
        return ""
    }
    s = strings.ToLower(s)
    r := []rune(s)
    r[0] = unicode.ToUpper(r[0])
    return string(r)
}
