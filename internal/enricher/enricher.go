package enricher

import (
	"fmt"
	"log/slog"
	"slices"
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/ENIACore/media_library_manager/internal/config"
)

// Propogates title and year of movie/show to all files
// Enriches subtitle and bonus files found inside intermediary directories
func Enrich(root *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "Enrich")

	if root == nil {
		return fmt.Errorf("Cannot enrich nil root entry")
	}

	lg.Debug("Enriching root entry", "entry", root.PathInfo.Source)
	switch root.Role {
	case metadata.SubtitleFile, metadata.BonusFile, metadata.SubtitleDir, metadata.BonusDir, metadata.UnknownRole:
		return fmt.Errorf("Entry %v cannot be enriched at root level", root.PathInfo.Source)
	}


	enrichEntries(root)
	enrichIntermediaryEntries(root)
	return nil
}

func enrichEntries(root *metadata.Entry) {
	var roles []metadata.EntryRole
	switch root.Role {
	case metadata.MovieFile, metadata.MovieDir:
		roles = []metadata.EntryRole{
			metadata.MovieDir,
			metadata.MovieFile,
			metadata.BonusFile,
			metadata.SubtitleFile,
		}
	case metadata.SeriesDir, metadata.SeasonDir, metadata.EpisodeFile:
		roles = []metadata.EntryRole{
			metadata.SeriesDir,
			metadata.SeasonDir,
			metadata.EpisodeFile,
			metadata.BonusFile,
			metadata.SubtitleFile,
		}
	}

	title := getTitle(root, roles)
	year := getYear(root, roles)

	mediaInfo := metadata.MediaInfo{}
	if title != nil {
		mediaInfo.Title = title
	}
	if year != nil {
		mediaInfo.Year = year
	}

	propogateDown(root, &mediaInfo, roles)
}

func getTitle(root *metadata.Entry, roles []metadata.EntryRole) []string {
	if root == nil {
		return nil
	}

	queue := []*metadata.Entry{root}

	for len(queue) > 0 {
		entry := queue[0]
		queue = queue[1:]

		for _, role := range roles {
			if entry.Role == role && len(entry.MediaInfo.Title) > 0 {
				return entry.MediaInfo.Title
			}
		}

		queue = append(queue, entry.Children...)
	}

	return nil
}

func getYear(root *metadata.Entry, roles []metadata.EntryRole) *int {
	if root == nil {
		return nil
	}

	queue := []*metadata.Entry{root}

	for len(queue) > 0 {
		entry := queue[0]
		queue = queue[1:]

		for _, role := range roles {
			if entry.Role == role && entry.MediaInfo.Year != nil {
				return entry.MediaInfo.Year
			}
		}

		queue = append(queue, entry.Children...)
	}

	return nil
}

func enrichIntermediaryEntries(entry *metadata.Entry) {
	switch entry.Role {
	case metadata.SubtitleDir, metadata.BonusDir:
		if entry.Height() == 2 {
			for _, child := range entry.Children {
				mediaInfo := metadata.MediaInfo{
					Season:  child.MediaInfo.Season,
					Episode: child.MediaInfo.Episode,
				}
				propogateDown(child, &mediaInfo, []metadata.EntryRole{metadata.BonusFile, metadata.SubtitleFile})
			}
		} else {
			for _, child := range entry.Children {
				enrichIntermediaryEntries(child)
			}
		}
	default:
		for _, child := range entry.Children {
			enrichIntermediaryEntries(child)
		}
	}
}

func propogateDown(entry *metadata.Entry,  src *metadata.MediaInfo, roles []metadata.EntryRole) {
	if slices.Contains(roles, entry.Role) || roles == nil {
		if src.Title != nil {
			entry.MediaInfo.Title = append([]string(nil), src.Title...)
		}
		if src.Year != nil {
			yearCopy := *src.Year
			entry.MediaInfo.Year = &yearCopy
		}
		if src.Season != nil {
			seasonCopy := *src.Season
			entry.MediaInfo.Season = &seasonCopy
		}
		if src.Episode != nil {
			episodeCopy := *src.Episode
			entry.MediaInfo.Episode = &episodeCopy
		}
	}

	for _, child := range entry.Children {
		propogateDown(child, src, roles)
	}
}

