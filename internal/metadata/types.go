package metadata

// Structure of media torrents
/*
Subtitle File
Bonus File
Episode File
Movie File

Subtitle Directory
└── Subtitle File(s)

Bonus Directory
├── Bonus File(s)
└── Subtitle File(s) (optional)

Season Directory
├── Episode File(s)
├── Subtitle File(s) (optional)
└── Subtitle Directory (optional)

Series Directory
├── Season Directory(s)
├── Bonus Directory (optional)
└── Subtitle Directory (optional)

Movie Directory
├── Movie File
├── Subtitle File(s) (optional)
├── Bonus File(s) (optional)
├── Subtitle Directory (optional)
└── Bonus Directory (optional)
*/

type EntryRole int8

// Classification of node that describes purpose of file or directory
const (
	UnknownRole		EntryRole = iota
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

type ContentType int8

// Type of file
const (
	UnknownType		ContentType = iota
    Video			
    Subtitle		
	//Audio			- Audio classification not yet included		
)
