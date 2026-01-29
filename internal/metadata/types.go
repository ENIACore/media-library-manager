package metadata

// EntryRole describes the purpose of a file or directory in relation to the media library.
type EntryRole int8

const (
	UnknownRole  EntryRole = iota
	SubtitleFile
	BonusFile
	EpisodeFile
	MovieFile

	SubtitleDir
	BonusDir
	SeasonDir
	SeriesDir
	MovieDir
)

// String returns the string representation of EntryRole for logging purposes.
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

// ContentType identifies the type of file.
// UnknownType can mean path is either unknown file type or a directory
// TODO: Add audio file and PDF file support
type ContentType int8

const (
	UnknownType ContentType = iota
	Video
	Subtitle
)
