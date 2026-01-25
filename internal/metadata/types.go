package metadata

// EntryRole classifies the purpose of a file or directory in the media hierarchy.
type EntryRole int8

// Entry role constants for files and directories.
const (
	UnknownRole  EntryRole = iota
	SubtitleFile           // subtitle file (.srt, .sub, etc.)
	BonusFile              // bonus/extra content file
	EpisodeFile            // TV episode file
	MovieFile              // feature film file

	SubtitleDir // directory containing subtitles
	BonusDir    // directory containing bonus content
	SeasonDir   // directory representing a TV season
	SeriesDir   // directory representing a TV series
	MovieDir    // directory representing a movie and its assets
)

// String returns the string representation of the EntryRole.
func (r EntryRole) String() string {
	switch r {
	case UnknownRole:
		return "UNKNOWN"
	case SubtitleFile:
		return "SUBTITLE_FILE"
	case BonusFile:
		return "BONUS_FILE"
	case EpisodeFile:
		return "EPISODE_FILE"
	case MovieFile:
		return "MOVIE_FILE"
	case SubtitleDir:
		return "SUBTITLE_DIR"
	case BonusDir:
		return "BONUS_DIR"
	case SeasonDir:
		return "SEASON_DIR"
	case SeriesDir:
		return "SERIES_DIR"
	case MovieDir:
		return "MOVIE_DIR"
	default:
		return "INVALID"
	}
}

// ContentType identifies the type of media file.
type ContentType int8

// Content type constants.
const (
	UnknownType ContentType = iota
	Video
	Subtitle
	// Audio - not yet implemented
)
