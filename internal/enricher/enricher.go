package enricher

import (
	"fmt"
	"log/slog"
	"github.com/ENIACore/media_library_manager/internal/metadata"
)

// Enrich uses helper functions to propagate necessary information throughout tree.
// Subtitles are skipped entirely. The verified root title propagates downward, overriding existing titles.
func Enrich(root *metadata.Entry, logger *slog.Logger) error {
	if root == nil {
		return fmt.Errorf("Cannot enrich nil root entry")
	}

	if !root.FileInfo.IsDir {
		return nil
	}

	switch root.Role {
	case metadata.SubtitleDir, metadata.DSDir, metadata.BTSDir, metadata.BonusDir, metadata.UnknownRole:
		return fmt.Errorf("Unexpected entry %v at root level", root.FileInfo.SourcePath)
	}

	enrichSubtitleFiles(root, metadata.MediaInfo{})
	enrichMovieFile(root, metadata.MediaInfo{})
	enrichEpisodeFiles(root, metadata.MediaInfo{})
	enrichExtrasFiles(root, metadata.MediaInfo{})
	enrichSubtitleWithSibling(root)

	return nil
}

// Passes season and episode to extras files.
// Season: Uses season of closest parent with non-nil season IF extras file doesn't have a season.
// Episode: Uses episode of closest parent with non-nil episode IF extras file doesn't have an episode.
// Title and year are not propagated; extras files retain their own title and year.
func enrichExtrasFiles(entry *metadata.Entry, ctx metadata.MediaInfo) {
	switch entry.Role {
	case metadata.DSFile, metadata.BTSFile, metadata.BonusFile:
		if entry.MediaInfo.Episode != nil {
			ctx.Episode = nil
		}
		if entry.MediaInfo.Season != nil {
			ctx.Season = nil
		}
		deepCopy(&entry.MediaInfo, ctx)
	case metadata.MovieDir, metadata.SeriesDir:
		// Title and year intentionally not propagated to extras files
		if entry.MediaInfo.TMDBid != 0 {
			ctx.TMDBid = entry.MediaInfo.TMDBid
		}
	case metadata.SeasonDir:
		if entry.MediaInfo.Season != nil {
			ctx.Season = entry.MediaInfo.Season
		}
		if entry.MediaInfo.TMDBid != 0 {
			ctx.TMDBid = entry.MediaInfo.TMDBid
		}
	default:
		if entry.MediaInfo.Season != nil {
			ctx.Season = entry.MediaInfo.Season
		}
		if entry.MediaInfo.Episode != nil {
			ctx.Episode = entry.MediaInfo.Episode
		}
	}

	for _, child := range entry.Children {
		enrichExtrasFiles(child, ctx)
	}
}

// Passes season, episode, title and year to episode files.
// Season: Uses season of closest parent with non-nil season IF episode file doesn't have season.
// Episode: Uses episode of closest parent with non-nil episode IF episode file doesn't have episode.
// Title: Uses title of root dir, overriding existing titles on episode files.
// Year: Uses year of root dir IF episode file doesn't have a year.
func enrichEpisodeFiles(entry *metadata.Entry, ctx metadata.MediaInfo) {
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
		return
	case metadata.SeriesDir, metadata.SeasonDir:
		if len(entry.MediaInfo.Title) > 0 {
			ctx.Title = entry.MediaInfo.Title
		}
		if entry.MediaInfo.Year != nil {
			ctx.Year = entry.MediaInfo.Year
		}
		if entry.MediaInfo.TMDBid != 0 {
			ctx.TMDBid = entry.MediaInfo.TMDBid
		}
	default:
		if entry.MediaInfo.Season != nil {
			ctx.Season = entry.MediaInfo.Season
		}
		if entry.MediaInfo.Episode != nil {
			ctx.Episode = entry.MediaInfo.Episode
		}
	}

	for _, child := range entry.Children {
		enrichEpisodeFiles(child, ctx)
	}
}

// Passes title and year to movie file.
// Title: Uses title of root dir, overriding existing title on movie file.
// Year: Uses year of root dir IF movie file doesn't have a year.
func enrichMovieFile(entry *metadata.Entry, ctx metadata.MediaInfo) {
	switch entry.Role {
	case metadata.MovieFile:
		deepCopy(&entry.MediaInfo, ctx)
		return
	case metadata.SeriesDir, metadata.SeasonDir:
		return
	case metadata.MovieDir:
		if len(entry.MediaInfo.Title) > 0 {
			ctx.Title = entry.MediaInfo.Title
		}
		if entry.MediaInfo.Year != nil {
			ctx.Year = entry.MediaInfo.Year
		}
		if entry.MediaInfo.TMDBid != 0 {
			ctx.TMDBid = entry.MediaInfo.TMDBid
		}
		if entry.MediaInfo.Edition != "" {
			ctx.Edition = entry.MediaInfo.Edition
		}
	}

	for _, child := range entry.Children {
		enrichMovieFile(child, ctx)
	}
}

// Passes season, episode, extras, title and year to subtitle files.
// Season: Uses season of closest parent with non-nil season IF subtitle file doesn't have a season.
// Episode: Uses episode of closest parent with non-nil episode IF subtitle file doesn't have an episode.
// Extras (DS, BTS, Bonus): If an extras dir is encountered, the subtitle file inside inherits its pattern.
// Title: Uses title of root dir (MovieDir, SeriesDir, or SeasonDir), overriding existing title.
// Year: Uses year of root dir IF subtitle file doesn't have a year.
func enrichSubtitleFiles(entry *metadata.Entry, ctx metadata.MediaInfo) {
	switch entry.Role {
	case metadata.SubtitleFile:
		if entry.MediaInfo.Episode != nil {
			ctx.Episode = nil
		}
		if entry.MediaInfo.Season != nil {
			ctx.Season = nil
		}
		deepCopy(&entry.MediaInfo, ctx)
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
		if entry.MediaInfo.TMDBid != 0 {
			ctx.TMDBid = entry.MediaInfo.TMDBid
		}
	default:
		if entry.MediaInfo.Season != nil {
			ctx.Season = entry.MediaInfo.Season
		}
		if entry.MediaInfo.Episode != nil {
			ctx.Episode = entry.MediaInfo.Episode
		}
	}

	for _, child := range entry.Children {
		enrichSubtitleFiles(child, ctx)
	}
}

// enrichSubtitleWithSibling matches each subtitle file to its sibling video file at the same parent
// level and copies FileInfo quality metadata (Audio, Codec, Bitrate, Resolution) from the video to
// the subtitle. Subtitles with no matching sibling are removed from the entry tree.
func enrichSubtitleWithSibling(entry *metadata.Entry) {
	if !entry.FileInfo.IsDir {
		return
	}

	var videoFiles []*metadata.Entry
	for _, child := range entry.Children {
		if child.Role == metadata.EpisodeFile || child.Role == metadata.MovieFile {
			videoFiles = append(videoFiles, child)
		}
	}

	if len(videoFiles) > 0 {
		kept := make([]*metadata.Entry, 0, len(entry.Children))
		for _, child := range entry.Children {
			switch child.Role {
			case metadata.SubtitleFile:
				if matchSubtitleToVideo(child, videoFiles) {
					kept = append(kept, child)
				}
			case metadata.SubtitleDir:
				pruneSubtitleDir(child, videoFiles)
				if len(child.Children) > 0 {
					kept = append(kept, child)
				}
			default:
				kept = append(kept, child)
			}
		}
		entry.Children = kept
	}

	for _, child := range entry.Children {
		enrichSubtitleWithSibling(child)
	}
}

// pruneSubtitleDir removes subtitle files from a subtitle directory that have no matching video sibling.
func pruneSubtitleDir(dir *metadata.Entry, videoFiles []*metadata.Entry) {
	kept := make([]*metadata.Entry, 0, len(dir.Children))
	for _, child := range dir.Children {
		switch child.Role {
		case metadata.SubtitleFile:
			if matchSubtitleToVideo(child, videoFiles) {
				kept = append(kept, child)
			}
		case metadata.SubtitleDir:
			pruneSubtitleDir(child, videoFiles)
			if len(child.Children) > 0 {
				kept = append(kept, child)
			}
		}
	}
	dir.Children = kept
}

// matchSubtitleToVideo finds the video file for a subtitle (by episode number if available,
// or the sole video file otherwise) and copies its FileInfo quality fields to the subtitle.
// Returns false if no match is found.
func matchSubtitleToVideo(sub *metadata.Entry, videoFiles []*metadata.Entry) bool {
	var match *metadata.Entry

	if sub.MediaInfo.Episode != nil {
		for _, v := range videoFiles {
			if v.MediaInfo.Episode != nil && *v.MediaInfo.Episode == *sub.MediaInfo.Episode {
				match = v
				break
			}
		}
	}

	if match == nil && len(videoFiles) == 1 {
		match = videoFiles[0]
	}

	if match == nil {
		return false
	}

	if sub.MediaInfo.Edition == "" {
		sub.MediaInfo.Edition = match.MediaInfo.Edition
	}

	if sub.FileInfo.Audio == "" {
		sub.FileInfo.Audio = match.FileInfo.Audio
	}
	if sub.FileInfo.AspectRatio == "" {
		sub.FileInfo.AspectRatio = match.FileInfo.AspectRatio
	}
	if sub.MediaInfo.Source == "" {
		sub.MediaInfo.Source = match.MediaInfo.Source
	}
	if sub.FileInfo.Codec == "" {
		sub.FileInfo.Codec = match.FileInfo.Codec
	}
	if sub.FileInfo.Bitrate == "" {
		sub.FileInfo.Bitrate = match.FileInfo.Bitrate
	}
	if sub.FileInfo.BitDepth == "" {
		sub.FileInfo.BitDepth = match.FileInfo.BitDepth
	}
	if sub.FileInfo.Resolution == "" {
		sub.FileInfo.Resolution = match.FileInfo.Resolution
	}

	return true
}

func deepCopy(dest *metadata.MediaInfo, src metadata.MediaInfo) {
	if len(src.Title) > 0 {
		dest.Title = append([]string(nil), src.Title...)
	}
	if src.Season != nil && dest.Season == nil {
		season := *src.Season
		dest.Season = &season
	}
	if src.Episode != nil && dest.Episode == nil {
		episode := *src.Episode
		dest.Episode = &episode
	}
	if src.Year != nil && dest.Year == nil {
		year := *src.Year
		dest.Year = &year
	}
	if src.TMDBid != 0 {
		dest.TMDBid = src.TMDBid
	}
	if src.Edition != "" {
		dest.Edition = src.Edition
	}
}
