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

const Unknown = -1

// Classification of node that describes purpose of file or directory
const (
	SubtitleFile	EntryRole = iota
	BonusFile
	EpisodeFile
	MovieFile			
	
	SubtitleDir
	BonusDir
	SeasonDir
	SeriesDir
	MovieDir
)


type ContentType int8

// Type of file
const (
    Video			ContentType = iota
    Subtitle		
	//Audio			- Audio classification not yet included		
)
