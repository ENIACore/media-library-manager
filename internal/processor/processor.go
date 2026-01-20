package processor

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/ENIACore/media_library_manager/internal/metadata"
)

/*
Destination Format:

Movie File (root):
	<Title>.<Year>/<Title>.<Year>.<Resolution>.<Codec>.<Source>.<Audio>.<Language>.<ext>

Movie Directory:
	<Title>.<Year>/
		Subtitles/<Title>.<Language>.<ext>
		Extras/<Title>.<Bonus>.<Resolution>.<Codec>.<Source>.<Audio>.<Language>.<ext>
		<Title>.<Year>.<Resolution>.<Codec>.<Source>.<Audio>.<Language>.<ext>

Episode File (root):
	<Title>.<Year>/<Season>/
		<Title>.<S##E##>.<Resolution>.<Codec>.<Source>.<Audio>.<Language>.<ext>

Season Directory:
	<Title>.<Year>/<Season>/
		Subtitles/<Title>.<S##E##>.<Language>.<ext>
		<Title>.<S##E##>.<Resolution>.<Codec>.<Source>.<Audio>.<Language>.<ext>

Series Directory:
	<Title>.<Year>/
		<Season>/
			Subtitles/<Title>.<S##E##>.<Language>.<ext>
			<Title>.<S##E##>.<Resolution>.<Codec>.<Source>.<Audio>.<Language>.<ext>
		Extras/<Title>.<Bonus>.<Resolution>.<Codec>.<Source>.<Audio>.<Language>.<ext>
		Subtitles/<Title>.<Language>.<ext>

Notes:
- Title is capitalized (e.g., "Test.Movie")
- Year is included when available
- Missing metadata fields are omitted from the filename
- All subtitle files go into a "Subtitles" subdirectory
- All bonus files go into an "Extras" subdirectory
- Extensions are lowercase
*/

func ResolveEntries(root *metadata.Entry, logger *slog.Logger) error {
	log := logger.With("func", "ResolveEntries")
	log.Info("resolving root", "path", root.PathInfo.Source)

	var err error

	switch root.Role {
		case metadata.SubtitleFile, metadata.BonusFile, metadata.SubtitleDir, metadata.BonusDir:
			log.Error("invalid root entry", "path", root.PathInfo.Source, "role", root.Role)
			return fmt.Errorf("entry %v cannot be processed alone at root level", root.PathInfo.Source)
		case metadata.MovieFile:
			err = resolveMovieFile("", root, logger)
		case metadata.EpisodeFile:
			err = resolveEpisodeFile("", nil, root, logger)
		case metadata.SeasonDir:
			err = resolveSeasonDir("", root, logger)
		case metadata.SeriesDir:
			err = resolveSeriesDir(root, logger)
		case metadata.MovieDir:
			err = resolveMovieDir(root, logger)
		default:
			log.Error("unknown role", "path", root.PathInfo.Source, "role", root.Role)
			return fmt.Errorf("entry %v has unknown role", root.PathInfo.Source)
	}

	if err != nil {
		log.Error("resolution failed", "path", root.PathInfo.Source, "err", err)
		return fmt.Errorf("failed to resolve root %v: %w", root.PathInfo.Source, err)
	}

	log.Info("resolved root", "path", root.PathInfo.Source, "dest", root.PathInfo.Dest)
	return nil
}

func resolveMovieDir(entry *metadata.Entry, logger *slog.Logger) error {
	log := logger.With("func", "resolveMovieDir")

	// Use dir for base path and title if movie file with title not present
	title := buildTitle(entry.MediaInfo.Title)
	basePath := buildBasePath(entry.MediaInfo.Title, entry.MediaInfo.Year)
	for _, child := range entry.Children {
		if child.Role == metadata.MovieFile && len(child.MediaInfo.Title) > 0{
			title = buildTitle(child.MediaInfo.Title)
			basePath = buildBasePath(child.MediaInfo.Title, child.MediaInfo.Year)
			break
		}
	}

	for _, child := range entry.Children {
		var err error
		switch child.Role {
			case metadata.MovieFile:
				err = resolveMovieFile(basePath, child, logger)
			case metadata.SubtitleFile:
				err = resolveSubtitleFile(basePath, title, child, logger)
			case metadata.SubtitleDir:
				err = resolveSubtitleDir(basePath, title, child, logger)
			case metadata.BonusFile:
				err = resolveBonusFile(basePath, title, child, logger)
			case metadata.BonusDir:
				err = resolveBonusDir(basePath, title, child, logger)
			default:
				log.Debug("unexpected child", "path", child.PathInfo.Source, "role", child.Role)
				return fmt.Errorf("unexpected child role %v in movie dir", child.Role)
		}
		if err != nil {
			return err
		}
	}

	log.Debug("resolved dir", "path", entry.PathInfo.Source, "dest", basePath)
	entry.PathInfo.Dest = basePath
	return nil
}

func resolveSeriesDir(entry *metadata.Entry, logger *slog.Logger) error {
	log := logger.With("func", "resolveSeriesDir")

	basePath := buildBasePath(entry.MediaInfo.Title, entry.MediaInfo.Year)
	title := buildTitle(entry.MediaInfo.Title)

	for _, child := range entry.Children {
		var err error
		switch child.Role {
			case metadata.SeasonDir:
				err = resolveSeasonDir(basePath, child, logger)
			case metadata.SubtitleDir:
				err = resolveSubtitleDir(basePath, title, child, logger)
			case metadata.BonusDir:
				err = resolveBonusDir(basePath, title, child, logger)
			default:
				log.Debug("unexpected child", "path", child.PathInfo.Source, "role", child.Role)
				return fmt.Errorf("unexpected child role %v in series dir", child.Role)
		}
		if err != nil {
			return err
		}
	}

	log.Debug("resolved dir", "path", entry.PathInfo.Source, "dest", basePath)
	entry.PathInfo.Dest = basePath
	return nil
}

func resolveSeasonDir(basePath string, entry *metadata.Entry, logger *slog.Logger) error {
	log := logger.With("func", "resolveSeasonDir")

	seasonNum := 1
	if entry.MediaInfo.Season != nil {
		seasonNum = *entry.MediaInfo.Season
	}
	seasonPath := fmt.Sprintf("Season %02d", seasonNum)

	// Assign both series title and base path (if base path is empty)
	title := buildTitle(entry.MediaInfo.Title)
	for _, child := range entry.Children {
		if child.Role == metadata.EpisodeFile {
			if basePath == "" {
				basePath = buildBasePath(child.MediaInfo.Title, child.MediaInfo.Year)
			}
			if len(child.MediaInfo.Title) > 0 {
				title = buildTitle(child.MediaInfo.Title)
				break
			}
		}
	}

	basePath = joinPath(basePath, seasonPath)

	for _, child := range entry.Children {
		var err error
		switch child.Role {
		case metadata.EpisodeFile:
			err = resolveEpisodeFile(basePath, entry.MediaInfo.Season, child, logger)
		case metadata.SubtitleFile:
			err = resolveSubtitleFile(basePath, title, child, logger)
		case metadata.SubtitleDir:
			err = resolveSubtitleDir(basePath, title, child, logger)
		default:
			log.Debug("unexpected child", "path", child.PathInfo.Source, "role", child.Role)
			return fmt.Errorf("unexpected child role %v in season dir", child.Role)
		}
		if err != nil {
			return err
		}
	}

	log.Debug("resolved dir", "path", entry.PathInfo.Source, "dest", basePath)
	entry.PathInfo.Dest = basePath
	return nil
}

func resolveMovieFile(basePath string, entry *metadata.Entry, logger *slog.Logger) error {
	log := logger.With("func", "resolveMovieFile")

	if entry.Role != metadata.MovieFile {
		log.Debug("invalid role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("entry %v is not a movie file", entry.PathInfo.Source)
	}

	if basePath == "" {
		basePath = buildBasePath(entry.MediaInfo.Title, entry.MediaInfo.Year)
	}

	filename := buildVideoFilename(entry.MediaInfo, entry.PathInfo.Ext)
	basePath = joinPath(basePath, filename)

	log.Debug("resolved file", "path", entry.PathInfo.Source, "dest", basePath)
	entry.PathInfo.Dest = basePath
	return nil
}

func resolveEpisodeFile(basePath string, parentSeason *int, entry *metadata.Entry, logger *slog.Logger) error {
	log := logger.With("func", "resolveEpisodeFile")

	if entry.Role != metadata.EpisodeFile {
		log.Debug("invalid role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("entry %v is not an episode file", entry.PathInfo.Source)
	}

	if basePath == "" {
		season := 1
		if entry.MediaInfo.Season != nil {
			season = *entry.MediaInfo.Season
		} else if parentSeason != nil {
			season = *parentSeason
		}
		basePath = buildBasePath(entry.MediaInfo.Title, entry.MediaInfo.Year)
		seasonPath := fmt.Sprintf("S%02d", season)
		basePath = joinPath(basePath, seasonPath)
	}

	filename := buildEpisodeFilename(entry.MediaInfo, parentSeason, entry.PathInfo.Ext)
	basePath = joinPath(basePath, filename)

	log.Debug("resolved file", "path", entry.PathInfo.Source, "dest", basePath)
	entry.PathInfo.Dest = basePath
	return nil
}

func resolveSubtitleFile(basePath string, title string, entry *metadata.Entry, logger *slog.Logger) error {
	log := logger.With("func", "resolveSubtitleFile")

	if entry.Role != metadata.SubtitleFile {
		log.Debug("invalid role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("entry %v is not a subtitle file", entry.PathInfo.Source)
	}

	filename := buildSubtitleFilename(title, entry.MediaInfo, entry.PathInfo.Ext)
	basePath = joinPath(basePath, "Subtitles")
	basePath = joinPath(basePath, filename)

	log.Debug("resolved file", "path", entry.PathInfo.Source, "dest", basePath)
	entry.PathInfo.Dest = basePath
	return nil
}

func resolveSubtitleDir(basePath string, title string, entry *metadata.Entry, logger *slog.Logger) error {
	log := logger.With("func", "resolveSubtitleDir")

	if entry.Role != metadata.SubtitleDir {
		log.Debug("invalid role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("entry %v is not a subtitle dir", entry.PathInfo.Source)
	}

	for _, child := range entry.Children {
		if err := resolveSubtitleFile(basePath, title, child, logger); err != nil {
			return err
		}
	}

	basePath = joinPath(basePath, "Subtitles")
	log.Debug("resolved dir", "path", entry.PathInfo.Source, "dest", basePath)
	entry.PathInfo.Dest = basePath
	return nil
}

func resolveBonusFile(basePath string, title string, entry *metadata.Entry, logger *slog.Logger) error {
	log := logger.With("func", "resolveBonusFile")

	if entry.Role != metadata.BonusFile {
		log.Debug("invalid role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("entry %v is not a bonus file", entry.PathInfo.Source)
	}

	filename := buildBonusFilename(title, entry.MediaInfo, entry.PathInfo.Ext)
	basePath = joinPath(basePath, "Extras")
	basePath = joinPath(basePath, filename)

	log.Debug("resolved file", "path", entry.PathInfo.Source, "dest", basePath)
	entry.PathInfo.Dest = basePath
	return nil
}

func resolveBonusDir(basePath string, title string, entry *metadata.Entry, logger *slog.Logger) error {
	log := logger.With("func", "resolveBonusDir")

	if entry.Role != metadata.BonusDir {
		log.Debug("invalid role", "path", entry.PathInfo.Source, "role", entry.Role)
		return fmt.Errorf("entry %v is not a bonus dir", entry.PathInfo.Source)
	}

	for _, child := range entry.Children {
		var err error
		switch child.Role {
		case metadata.BonusFile:
			err = resolveBonusFile(basePath, title, child, logger)
		case metadata.SubtitleFile:
			err = resolveSubtitleFile(basePath, title, child, logger)
		default:
			log.Debug("unexpected child", "path", child.PathInfo.Source, "role", child.Role)
			return fmt.Errorf("unexpected child role %v in bonus dir", child.Role)
		}
		if err != nil {
			return err
		}
	}

	basePath = joinPath(basePath, "Extras")
	log.Debug("resolved dir", "path", entry.PathInfo.Source, "dest", basePath)
	entry.PathInfo.Dest = basePath
	return nil
}

/*
	filename builders
*/

func buildBasePath(titleParts []string, year *int) string {
	title := buildTitle(titleParts)
	if year != nil {
		return fmt.Sprintf("%s.%d", title, *year)
	}
	return title
}

func buildTitle(parts []string) string {
	if len(parts) == 0 {
		return ""
	}

	capitalized := make([]string, len(parts))
	for i, part := range parts {
		capitalized[i] = capitalize(part)
	}
	return strings.Join(capitalized, ".")
}

func buildVideoFilename(info metadata.MediaInfo, ext string) string {
	parts := []string{buildTitle(info.Title)}

	if info.Year != nil {
		parts = append(parts, fmt.Sprintf("%d", *info.Year))
	}
	if info.Resolution != "" {
		parts = append(parts, capitalize(info.Resolution))
	}
	if info.Codec != "" {
		parts = append(parts, capitalize(info.Codec))
	}
	if info.Source != "" {
		parts = append(parts, capitalize(info.Source))
	}
	if info.Audio != "" {
		parts = append(parts, capitalize(info.Audio))
	}
	if info.Language != "" {
		parts = append(parts, capitalize(info.Language))
	}

	filename := strings.Join(parts, ".")
	return filename + "." + strings.ToLower(ext)
}

func buildEpisodeFilename(info metadata.MediaInfo, parentSeason *int, ext string) string {
	parts := []string{buildTitle(info.Title)}

	seasonNum := 0
	if info.Season != nil {
		seasonNum = *info.Season
	} else if parentSeason != nil {
		seasonNum = *parentSeason
	}

	episodeNum := 0
	if info.Episode != nil {
		episodeNum = *info.Episode
	}

	parts = append(parts, fmt.Sprintf("S%02dE%02d", seasonNum, episodeNum))

	if info.Resolution != "" {
		parts = append(parts, capitalize(info.Resolution))
	}
	if info.Codec != "" {
		parts = append(parts, capitalize(info.Codec))
	}
	if info.Source != "" {
		parts = append(parts, capitalize(info.Source))
	}
	if info.Audio != "" {
		parts = append(parts, capitalize(info.Audio))
	}
	if info.Language != "" {
		parts = append(parts, capitalize(info.Language))
	}

	filename := strings.Join(parts, ".")
	return filename + "." + strings.ToLower(ext)
}

func buildSubtitleFilename(title string, info metadata.MediaInfo, ext string) string {
	parts := []string{title}

	if info.Language != "" {
		parts = append(parts, capitalize(info.Language))
	}

	filename := strings.Join(parts, ".")
	return filename + "." + strings.ToLower(ext)
}

func buildBonusFilename(title string, info metadata.MediaInfo, ext string) string {
	parts := []string{title}

	if info.Bonus != "" {
		parts = append(parts, formatBonus(info.Bonus))
	}
	if info.Resolution != "" {
		parts = append(parts, capitalize(info.Resolution))
	}
	if info.Codec != "" {
		parts = append(parts, capitalize(info.Codec))
	}
	if info.Source != "" {
		parts = append(parts, capitalize(info.Source))
	}
	if info.Audio != "" {
		parts = append(parts, capitalize(info.Audio))
	}
	if info.Language != "" {
		parts = append(parts, capitalize(info.Language))
	}

	filename := strings.Join(parts, ".")
	return filename + "." + strings.ToLower(ext)
}

/*
	string helpers
*/

func capitalize(s string) string {
	if s == "" {
		return ""
	}
	lower := strings.ToLower(s)
	return strings.ToUpper(string(lower[0])) + lower[1:]
}

func formatBonus(s string) string {
	parts := strings.Split(s, "_")
	capitalized := make([]string, len(parts))
	for i, part := range parts {
		capitalized[i] = capitalize(part)
	}
	return strings.Join(capitalized, ".")
}

func joinPath(parts ...string) string {
	nonEmpty := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	return strings.Join(nonEmpty, "/")
}
