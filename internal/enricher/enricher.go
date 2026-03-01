// Package enricher passes important contextual information through the [metadata.Entry] tree bi-directionally.
// Enrichment allows for more accurate resolutions by contextualizing entries.
// Package relies on classifier package for necessary information and produces output used by resolver package.  
package enricher

import (
	"fmt"
	"log/slog"
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/ENIACore/media_library_manager/internal/config"
)

// Enrich uses helper functions to propogate necessary information throughout tree. 
func Enrich(root *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "Enrich")

	if root == nil {
		return fmt.Errorf("Cannot enrich nil root entry")
	}


	if !root.PathInfo.IsDir {
		lg.Debug("Cannot enrich file at root level", "source", root.PathInfo.Source)
		return nil
	}

	switch root.Role {
		case metadata.SubtitleDir, metadata.DSDir, metadata.BTSDir, metadata.BonusDir, metadata.UnknownRole:
			return fmt.Errorf("Unexpected entry %v at root level", root.PathInfo.Source)
	}

	enrichMovieFile(root, nil)
	enrichEpisodeFiles(root, nil)
	enrichSubtitleFiles(root, nil)
	enrichExtrasFiles(root, nil)
	
	return nil
}

// Passes season, episode, extras, title and year to subtitle files
// Season: Uses season of closest parent with non-nil season IF extras file doesn't have a season
// Episode: Uses episode of closest parent with non-nil episode IF extras file doesn't have a season
// Extras (DS, BTS, Bonus): If extras dir is encountered, that means subtitle file is inside it therefore subtitle file will receive the pattern
// Title: Uses title of root dir (movie dir, series dir, or season dir)
// Year: Uses year of root dir (movie dir, series dir, or season dir)
func enrichSubtitleFiles(entry *metadata.Entry, ctx *metadata.MediaInfo) {
	if ctx == nil {
		ctx = &metadata.MediaInfo{}
	}

	switch entry.Role {
		case metadata.SubtitleFile:
			if entry.MediaInfo.Episode != nil {
				ctx.Episode = nil
			}
			if entry.MediaInfo.Season != nil {
				ctx.Season = nil
			}
			deepCopy(&entry.MediaInfo, ctx)		
			return
		case metadata.DSDir:
			ctx.DS = entry.MediaInfo.DS
		case metadata.BTSDir:
			ctx.BTS = entry.MediaInfo.BTS
		case metadata.BonusDir:
			ctx.Bonus = entry.MediaInfo.Bonus
		case metadata.MovieDir, metadata.SeriesDir, metadata.SeasonDir:
			if len(entry.MediaInfo.Title) > 0 {
				ctx.Title = entry.MediaInfo.Title
			}
			if entry.MediaInfo.Year != nil {
				ctx.Year = entry.MediaInfo.Year
			}
		default:
			if entry.MediaInfo.Season != nil {
				ctx.Season = entry.MediaInfo.Season
			}
			if entry.MediaInfo.Episode != nil {
				ctx.Episode = entry.MediaInfo.Episode
			}
	}

	if len(entry.Children) == 0 {
		return
	}

	for _, child := range entry.Children {
		enrichSubtitleFiles(child, ctx)
	}
}

// Passes season, episode and year to extras files
// Season: Uses season of closest parent with non-nil season IF extras file doesn't have a season
// Episode: Uses episode of closest parent with non-nil episode IF extras file doesn't have a episode
func enrichExtrasFiles(entry *metadata.Entry, ctx *metadata.MediaInfo) {
	if ctx == nil {
		ctx = &metadata.MediaInfo{}
	}

	switch entry.Role {
		case metadata.DSFile, metadata.BTSFile, metadata.BonusFile:
			if entry.MediaInfo.Episode != nil {
				ctx.Episode = nil
			}
			if entry.MediaInfo.Season != nil {
				ctx.Season = nil
			}
			deepCopy(&entry.MediaInfo, ctx)		
			return
		case metadata.MovieDir, metadata.SeriesDir, metadata.SeasonDir:
			// Do nothing, title and year will come from extra file itself, due to unique naming nature of extras files
		default:
			if entry.MediaInfo.Season != nil {
				ctx.Season = entry.MediaInfo.Season
			}
			if entry.MediaInfo.Episode != nil {
				ctx.Episode = entry.MediaInfo.Episode
			}
	}

	if len(entry.Children) == 0 {
		return
	}

	for _, child := range entry.Children {
		enrichExtrasFiles(child, ctx)
	}
}

// Passes season, episode, title and year to subtitle files
// Season: Uses season of closest parent with non-nil season IF episode file doesn't have season
// Episode: Uses episode of closest parent with non-nil episode IF episode file doesn't have episode
// Title: Uses title of root dir (movie dir, series dir, or season dir)
// Year: Uses year of root dir (movie dir, series dir, or season dir)
func enrichEpisodeFiles(entry *metadata.Entry, ctx *metadata.MediaInfo) {
	if ctx == nil {
		ctx = &metadata.MediaInfo{}
	}

	switch entry.Role {
		case metadata.EpisodeFile:
			if entry.MediaInfo.Episode != nil {
				ctx.Episode = nil
			}
			if entry.MediaInfo.Season != nil {
				ctx.Season = nil
			}
			deepCopy(&entry.MediaInfo, ctx)		
			return
		case metadata.MovieDir:
			// Return early if movie directory
			return
		case metadata.SeriesDir, metadata.SeasonDir:
			if len(entry.MediaInfo.Title) > 0 {
				ctx.Title = entry.MediaInfo.Title
			}
			if entry.MediaInfo.Year != nil {
				ctx.Year = entry.MediaInfo.Year
			}
		default:
			if entry.MediaInfo.Season != nil {
				ctx.Season = entry.MediaInfo.Season
			}
			if entry.MediaInfo.Episode != nil {
				ctx.Episode = entry.MediaInfo.Episode
			}
	}

	if len(entry.Children) == 0 {
		return
	}

	for _, child := range entry.Children {
		enrichEpisodeFiles(child, ctx)
	}
}

// Passes title and year to movie file
// Title: Uses title of root dir (movie dir, series dir, or season dir)
// Year: Uses year of root dir (movie dir, series dir, or season dir)
func enrichMovieFile(entry *metadata.Entry, ctx *metadata.MediaInfo) {
	if ctx == nil {
		ctx = &metadata.MediaInfo{}
	}

	switch entry.Role {
		case metadata.MovieFile:
			deepCopy(&entry.MediaInfo, ctx)		
			return
		case metadata.SeriesDir, metadata.SeasonDir:
			// Return early if series or season dir
			return
		case metadata.MovieDir:
			if len(entry.MediaInfo.Title) > 0 {
				ctx.Title = entry.MediaInfo.Title
			}
			if entry.MediaInfo.Year != nil {
				ctx.Year = entry.MediaInfo.Year
			}
		default:
			// Do nothing for entries that are not movie dir or file
	}

	if len(entry.Children) == 0 {
		return
	}

	for _, child := range entry.Children {
		enrichMovieFile(child, ctx)
	}
}

func deepCopy(dest *metadata.MediaInfo, src *metadata.MediaInfo) {
	if src.Title != nil {
		dest.Title = append([]string(nil), src.Title...)
	}
	if src.Season != nil {
		season := *src.Season
		dest.Season = &season
	}
	if src.Episode != nil {
		episode := *src.Episode
		dest.Episode = &episode
	}
	if src.Year != nil {
		year := *src.Year
		dest.Year = &year
	}
}
