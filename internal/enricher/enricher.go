package enricher

import (
	"fmt"
	"log/slog"
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/ENIACore/media_library_manager/internal/config"
)

func Enrich(root *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "Enrich")
	lg.Debug("Enriching root entry", "entry", root.PathInfo.Source)
	var title []string
	var year *int

	switch root.Role {
	case metadata.SubtitleFile, metadata.BonusFile, metadata.SubtitleDir, metadata.BonusDir:
		return fmt.Errorf("Entry %v cannot be enriched at root level", root.PathInfo.Source)
	case metadata.EpisodeFile, metadata.SeasonDir, metadata.SeriesDir:
		title = getShowTitle(root)	
		year = getShowYear(root)
	case metadata.MovieFile, metadata.MovieDir:
		title = getMovieTitle(root)
		year = getShowYear(root)
	default:
		return fmt.Errorf("Entry %v has unknown role", root.PathInfo.Source)
	}

	lg.Info("Successfully extracted title and year for root entry", "entry", root.PathInfo.Source, "title", title, "year", year)
	setEntryValues(root, title, year)
	return nil
}

// Determines title using order of precedence: Series > Season > Episode > nil
func getShowTitle(root *metadata.Entry) []string {
	if root == nil {
		return nil
	}
	if root.Role == metadata.EpisodeFile {
		return root.MediaInfo.Title
	}
	if len(root.MediaInfo.Title) > 0 {
		return root.MediaInfo.Title
	}
	for _, child := range root.Children {
		if title := getShowTitle(child); len(title) > 0 {
			return title
		}
	}
	return root.MediaInfo.Title
}

func getShowYear(root *metadata.Entry) *int {
	if root == nil {
		return nil
	}
	if root.Role == metadata.EpisodeFile {
		return root.MediaInfo.Year
	}
	if root.MediaInfo.Year != nil {
		return root.MediaInfo.Year
	}
	for _, child := range root.Children {
		if year := getShowYear(child); year != nil {
			return year
		}
	}
	return root.MediaInfo.Year
}

func getMovieTitle(root *metadata.Entry) []string {
	if root == nil {
		return nil
	}
	if root.Role == metadata.MovieFile {
		return root.MediaInfo.Title
	}
	if len(root.MediaInfo.Title) > 0 {
		return root.MediaInfo.Title
	}
	for _, child := range root.Children {
		if title := getMovieTitle(child); len(title) > 0 {
			return title
		}
	}
	return root.MediaInfo.Title
}

func getMovieYear(root *metadata.Entry) *int {
	if root == nil {
		return nil
	}
	if root.Role == metadata.MovieFile {
		return root.MediaInfo.Year
	}
	if root.MediaInfo.Year != nil {
		return root.MediaInfo.Year
	}
	for _, child := range root.Children {
		if year := getMovieYear(child); year != nil {
			return year
		}
	}
	return root.MediaInfo.Year
}

func setEntryValues(root *metadata.Entry, title []string, year *int) {
	root.MediaInfo.Title = title
	root.MediaInfo.Year = year
	for _, child := range root.Children {
		setEntryValues(child, title, year)
	}
}
