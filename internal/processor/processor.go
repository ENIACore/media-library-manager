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

func resolveSubtitleFile(basePath string, entry *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "resolveSubtitleFile")

	if entry.Role != metadata.SubtitleFile {
		lg.Debug("Expected SubtitleFile role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected SubtitleFile role, received %v for node %v", entry.Role, entry.PathInfo.Source)
	}
	if basePath == "" {
		lg.Debug("Expected base path for SubtitleFile role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected base path for SubtitlFile role, for node %v", entry.PathInfo.Source)
	}

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

func resolveSubtitleDir(basePath string, entry *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "resolveSubtitleDir")

	if entry.Role != metadata.SubtitleDir {
		lg.Debug("Expected SubtitleDir role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected SubtitleDir role, received %v for node %v", entry.Role, entry.PathInfo.Source)
	}
	if basePath == "" {
		lg.Debug("Expected base path for SubtitleDir role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected base path for SubtitleDir role, for node %v", entry.PathInfo.Source)
	}

	basePath = filepath.Join(basePath, "Subtitles")

	for _, child := range entry.Children {
		var err error
		switch child.Role {
		case metadata.SubtitleFile:
			err = resolveSubtitleFile(basePath, child, cfg, logger)
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

func resolveBonusFile(basePath string, entry *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "resolveBonusFile")

	if entry.Role != metadata.BonusFile {
		lg.Debug("Expected BonusFile role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected BonusFile role, received %v for node %v", entry.Role, entry.PathInfo.Source)
	}
	if basePath == "" {
		lg.Debug("Expected base path for BonusFile role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected base path for BonusFile role, for node %v", entry.PathInfo.Source)
	}

	filename := buildFilename(entry)

	entry.PathInfo.Dest = filepath.Join(basePath, filename)
	lg.Debug("Resolved bonus file destination", "source", entry.PathInfo.Source, "destination", entry.PathInfo.Dest)

	return nil
}

func resolveBonusDir(basePath string, entry *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "resolveBonusDir")

	if entry.Role != metadata.BonusDir {
		lg.Debug("Expected BonusDir role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected BonusDir role, received %v for node %v", entry.Role, entry.PathInfo.Source)
	}
	if basePath == "" {
		lg.Debug("Expected base path for BonusDir role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected base path for BonusDir role, for node %v", entry.PathInfo.Source)
	}

	basePath = filepath.Join(basePath, "Extras")

	for _, child := range entry.Children {
		var err error
		switch child.Role {
		case metadata.BonusFile:
			err = resolveBonusFile(basePath, child, cfg, logger)
		case metadata.SubtitleFile:
			err = resolveSubtitleFile(basePath, child, cfg, logger)
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

func resolveEpisodeFile(basePath string, entry *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "resolveEpisodeFile")

	if entry.Role != metadata.EpisodeFile {
		lg.Debug("Expected EpisodeFile role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected EpisodeFile role, received %v for node %v", entry.Role, entry.PathInfo.Source)
	}

	if basePath == "" {
		titlePath := buildTitlePath(entry)	
		seasonPath := buildSeasonPath(entry)
		basePath = filepath.Join(cfg.ShowPath, titlePath, seasonPath)
	}
	entry.MediaInfo.Bonus = ""
	filename := buildFilename(entry)

	entry.PathInfo.Dest = filepath.Join(basePath, filename)
	lg.Debug("Resolved episode file destination", "source", entry.PathInfo.Source, "destination", entry.PathInfo.Dest)

	return nil
}

func resolveSeasonDir(basePath string, entry *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "resolveSeasonDir")

	if entry.Role != metadata.SeasonDir {
		lg.Debug("Expected SeasonDir role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("Expected SeasonDir role, received %v for node %v", entry.Role, entry.PathInfo.Source)
	}

	seasonPath := buildSeasonPath(entry)
	if basePath == "" {
		titlePath := buildTitlePath(entry)
		basePath = filepath.Join(cfg.ShowPath, titlePath, seasonPath)	
	} else {
		basePath = filepath.Join(basePath, seasonPath)
	}

	for _, child := range entry.Children {
		var err error
		switch child.Role {
		case metadata.EpisodeFile:
			err = resolveEpisodeFile(basePath, child, cfg, logger)
		case metadata.SubtitleFile:
			err = resolveSubtitleFile(basePath, child, cfg, logger)
		case metadata.SubtitleDir:
			err = resolveSubtitleDir(basePath, child, cfg, logger)
		default:
			err = fmt.Errorf("Unexpected child role for SeasonDir, received role %v for node %v", entry.Role, entry.PathInfo.Source)
		}
		if err != nil {
			return err
		}
	}

	entry.PathInfo.Dest = basePath
	lg.Debug("Resolved season dir destination", "source", entry.PathInfo.Source, "destination", entry.PathInfo.Dest)

	return nil
}

// Top level resolver cannot have basePath from parent
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
			err = resolveBonusDir(basePath, child, cfg, logger)
		case metadata.SubtitleDir:
			err = resolveSubtitleDir(basePath, child, cfg, logger)
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

// Top level resolver cannot have basePath from parent
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
			err = resolveSubtitleFile(basePath, child, cfg, logger)
		case metadata.BonusFile:
			err = resolveBonusFile(basePath, child, cfg, logger)
		case metadata.SubtitleDir:
			err = resolveSubtitleDir(basePath, child, cfg, logger)
		case metadata.BonusDir:
			err = resolveBonusDir(basePath, child, cfg, logger)
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

// Set Entry fields to nil to omit from filename
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

func buildSeasonPath(entry *metadata.Entry) string {
	if entry.MediaInfo.Season != nil {
		return fmt.Sprintf("S%02d", *entry.MediaInfo.Season)
	}
	return "S01"
}

func capitalize(s string) string {
    if s == "" {
        return ""
    }
    r := []rune(s)
    r[0] = unicode.ToUpper(r[0])
    return string(r)
}
