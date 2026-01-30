// Package enricher passes important contextual information through the [metadata.Entry] tree bi-directionally.
// Enrichment allows for more accurate resolutions by contextualizing entries.
// Package relies on classifier package for necessary information and produces output used by resolver package.  
package enricher

import (
	"fmt"
	"log/slog"
	"slices"
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/ENIACore/media_library_manager/internal/config"
)

// Enrich uses helper functions to propogate necessary information throughout tree. 
// Propogated information:
//	- Titles
//	- Year
//	- Season and episode numbers
// Returns error if invalid Role found at root.
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


	enrichEpisodeFiles(root)
	enrichEntries(root)
	enrichIntermediaryEntries(root)
	return nil
}

// enrichEpisodeFiles recursively passes season information to all children of season directory
func enrichEpisodeFiles(root *metadata.Entry) {
	if root.Role == metadata.SeriesDir {
		for _, child := range root.Children {
			enrichEpisodeFiles(child)
		}
	}
	if root.Role == metadata.SeasonDir && root.MediaInfo.Season != nil {
		for _, child := range root.Children {
			if child.Role == metadata.EpisodeFile && child.MediaInfo.Season != nil {
				seasonCopy := *root.MediaInfo.Season
				child.MediaInfo.Season = &seasonCopy
			}
		}
	}
}

// enrichIntermediaryEntries recursively passes season and episode information from intermediary entries to children
// An intermediary entry is a subtitle directory inside a bonus or subtitle directory, or a bonus directory inside a bonus directory
// Intermediary entries typically are used to encapsulate subtitle and/or bonus files of specific episodes
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

// enrichEntries passes title and year from root node to children nodes of specified type
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

// getTitle is a recursive function that returns the first valid title, from a identified Role 
// This enables series and movie titles to be passed to all children
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

// getYear is a recursive function that returns the first valid year, from a identified Role 
// This enables the inception year of a movie or series to be passed to all children
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


// propogateDown is a recursive helper function to perform deep copy from non-empty src fields to entry objects.
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

