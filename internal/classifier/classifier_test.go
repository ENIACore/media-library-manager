package classifier

import (
	"testing"
	"github.com/ENIACore/media_library_manager/internal/metadata"
)

func intPtr(i int) *int {
	return &i
}
var movieFile = metadata.Entry{
	Parent: nil,
	Children: nil,
	MediaInfo: metadata.MediaInfo{
		Title:		[]string{
			"TEST",
			"MOVIE",
		},
		Year:		intPtr(2025),
		Episode:	nil,
		Season:		nil,
		Resolution:	"1080P",
		Codec:		"X264",
		Source:		"REMUX",
		Audio:		"ATMOS",
		Language:	"ENGLISH",
	},
	PathInfo: metadata.PathInfo{
		IsDir: false,
		Dest: "",
		Source: "/test movie/test movie 2025 1080p.x264.remux.atmos.english.mp4",
		Ext: "MP4",	
		Type: metadata.Video,
	},
}


var episodeFile = metadata.Entry{
	Parent: nil,
	Children: nil,
	MediaInfo: metadata.MediaInfo{
		Title:		[]string{
			"TEST",
			"EPISODE",
		},
		Year:		intPtr(2025),
		Episode:	intPtr(1),
		Season:		nil,
		Resolution:	"1080P",
		Codec:		"X264",
		Source:		"REMUX",
		Audio:		"ATMOS",
		Language:	"ENGLISH",
	},
	PathInfo: metadata.PathInfo{
		IsDir: false,
		Dest: "",
		Source: "/test show/season 01/test episode E01 2025 1080p.x264.remux.atmos.english.mp4",
		Ext: "MP4",	
		Type: metadata.Video,
	},
}

var subtitleFile = metadata.Entry{
	Parent: nil,
	Children: nil,
	MediaInfo: metadata.MediaInfo{
		Title:		[]string{
			"SUBTITLE",
		},
		Year:		nil,
		Episode:	nil,
		Season:		nil,
		Resolution:	"",
		Codec:		"",
		Source:		"",
		Audio:		"",
		Language:	"ENGLISH",
	},
	PathInfo: metadata.PathInfo{
		IsDir: false,
		Dest: "",
		Source: "/test movie/subtitles/subtitle english.srt",
		Ext: "SRT",	
		Type: metadata.Subtitle,
	},
}

var bonusFile = metadata.Entry{
	Parent: nil,
	Children: nil,
	MediaInfo: metadata.MediaInfo{
		Title:		[]string{
			"TEST",
			"MOVIE",
			"BEHIND",
			"THE",
			"SCENES",
		},
		Year:		intPtr(2025),
		Episode:	nil,
		Season:		nil,
		Resolution:	"1080P",
		Codec:		"X264",
		Source:		"REMUX",
		Audio:		"ATMOS",
		Language:	"ENGLISH",
		Bonus:		"BEHIND_THE_SCENES",
	},
	PathInfo: metadata.PathInfo{
		IsDir: false,
		Dest: "",
		Source: "/test movie/test movie behind the scenes 2025 1080p.x264.remux.atmos.english.mp4",
		Ext: "MP4",	
		Type: metadata.Video,
	},
}

/*
Movie Directory
├── Movie File
├── Subtitle File(s) (optional)
├── Bonus File(s) (optional)
├── Subtitle Directory (optional)
└── Bonus Directory (optional)
*/
var movieDir = metadata.Entry{
	Children: []*metadata.Entry{
		&movieFile,
		&bonusFile,
		&subtitleFile,
		&subtitleFile,
	},
	MediaInfo: metadata.MediaInfo{
		Title:		[]string{
			"TEST",
			"MOVIE",
		},
	},
	PathInfo: metadata.PathInfo{
		IsDir: true,
		Source: "/test movie",
		Type: metadata.Unknown,
	},
}

/*
Series Directory
├── Season Directory(s)
├── Bonus Directory (optional)
└── Subtitle Directory (optional)
*/
var seriesDir = metadata.Entry{
	Children: []*metadata.Entry{
		&seasonDir,
		&bonusDir,
		&subtitleDir,
	},
	MediaInfo: metadata.MediaInfo{
		Title: []string{
			"TEST",
			"SHOW",
		},
	},
	PathInfo: metadata.PathInfo{
		IsDir: true,
		Source: "/test show",
		Type: metadata.Unknown,
	},
}

/*
Season Directory
├── Episode File(s)
├── Subtitle File(s) (optional)
└── Subtitle Directory (optional)
*/
var seasonDir = metadata.Entry{
	Children: []*metadata.Entry{
		&episodeFile,
		&episodeFile,
		&subtitleFile,
		&subtitleFile,
	},
	MediaInfo: metadata.MediaInfo{
		Season:		intPtr(1),
	},
	PathInfo: metadata.PathInfo{
		IsDir: true,
		Source: "/test show/season 01",
		Type: metadata.Unknown,
	},
}

/*
Subtitle Directory
└── Subtitle File(s)
*/
var subtitleDir = metadata.Entry{
	Children: []*metadata.Entry{
		&subtitleFile,
		&subtitleFile,
		&subtitleFile,
	},
	MediaInfo: metadata.MediaInfo{
		Title:		[]string{
			"SUBTITLES",
		},

	},
	PathInfo: metadata.PathInfo{
		IsDir: true,
		Source: "/test movie/subtitles",
		Type: metadata.Unknown,
	},
}

/*
Bonus Directory
├── Bonus File(s)
└── Subtitle File(s) (optional)
*/
var bonusDir = metadata.Entry{
	Children: []*metadata.Entry{
		&bonusFile,
		&subtitleFile,
		&subtitleFile,
	},
	MediaInfo: metadata.MediaInfo{
		Title:		[]string{
			"EXTRAS",
		},
		Bonus: "EXTRA",
	},
	PathInfo: metadata.PathInfo{
		IsDir: true,
		Source: "/test movie/extras",
		Type: metadata.Unknown,
	},
}

func TestIsMovieFile(t *testing.T) {

	tests := []struct{
		name		string
		node		metadata.Entry
		expected	bool
	}{
		{
			name:		"movie file",
			node:		movieFile,
			expected:	true,
		},
		{
			name:		"episode file",
			node:		episodeFile,
			expected:	false,
		},
		{
			name:		"subtitle file",
			node:		subtitleFile,
			expected:	false,
		},
		{
			name:		"bonus file",
			node:		bonusFile,
			expected:	false,
		},
		{
			name:		"movie directory", 
			node: movieDir,
			expected:	false,
		},
		{
			name:		"series directory", 
			node: seriesDir,
			expected:	false,
		},
		{
			name:		"season directory", 
			node: seasonDir,
			expected:	false,
		},
		{
			name:		"subtitle directory", 
			node: subtitleDir,
			expected:	false,
		},
		{
			name:		"bonus directory", 
			node:		bonusDir,
			expected:	false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := isMovieFile(&test.node)
			if res != test.expected {
				t.Errorf("isMovieFile = %v, want %v", res, test.expected)
			}
		})
	}

}
func TestIsEpisodeFile(t *testing.T) {

	tests := []struct{
		name		string
		node		metadata.Entry
		expected	bool
	}{
		{
			name:		"movie file",
			node:		movieFile,
			expected:	false,
		},
		{
			name:		"episode file",
			node:		episodeFile,
			expected:	true,
		},
		{
			name:		"subtitle file",
			node:		subtitleFile,
			expected:	false,
		},
		{
			name:		"bonus file",
			node:		bonusFile,
			expected:	false,
		},
		{
			name:		"movie directory", 
			node: movieDir,
			expected:	false,
		},
		{
			name:		"series directory", 
			node: seriesDir,
			expected:	false,
		},
		{
			name:		"season directory", 
			node: seasonDir,
			expected:	false,
		},
		{
			name:		"subtitle directory", 
			node: subtitleDir,
			expected:	false,
		},
		{
			name:		"bonus directory", 
			node:		bonusDir,
			expected:	false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := isEpisodeFile(&test.node)
			if res != test.expected {
				t.Errorf("isEpisodeFile = %v, want %v", res, test.expected)
			}
		})
	}

}
func TestIsSubtitleFile(t *testing.T) {

	tests := []struct{
		name		string
		node		metadata.Entry
		expected	bool
	}{
		{
			name:		"movie file",
			node:		movieFile,
			expected:	false,
		},
		{
			name:		"episode file",
			node:		episodeFile,
			expected:	false,
		},
		{
			name:		"subtitle file",
			node:		subtitleFile,
			expected:	true,
		},
		{
			name:		"bonus file",
			node:		bonusFile,
			expected:	false,
		},
		{
			name:		"movie directory", 
			node: movieDir,
			expected:	false,
		},
		{
			name:		"series directory", 
			node: seriesDir,
			expected:	false,
		},
		{
			name:		"season directory", 
			node: seasonDir,
			expected:	false,
		},
		{
			name:		"subtitle directory", 
			node: subtitleDir,
			expected:	false,
		},
		{
			name:		"bonus directory", 
			node:		bonusDir,
			expected:	false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := isSubtitleFile(&test.node)
			if res != test.expected {
				t.Errorf("isSubtitleFile = %v, want %v", res, test.expected)
			}
		})
	}

}
func TestIsBonusFile(t *testing.T) {

	tests := []struct{
		name		string
		node		metadata.Entry
		expected	bool
	}{
		{
			name:		"movie file",
			node:		movieFile,
			expected:	false,
		},
		{
			name:		"episode file",
			node:		episodeFile,
			expected:	false,
		},
		{
			name:		"subtitle file",
			node:		subtitleFile,
			expected:	false,
		},
		{
			name:		"bonus file",
			node:		bonusFile,
			expected:	true,
		},
		{
			name:		"movie directory", 
			node: movieDir,
			expected:	false,
		},
		{
			name:		"series directory", 
			node: seriesDir,
			expected:	false,
		},
		{
			name:		"season directory", 
			node: seasonDir,
			expected:	false,
		},
		{
			name:		"subtitle directory", 
			node: subtitleDir,
			expected:	false,
		},
		{
			name:		"bonus directory", 
			node:		bonusDir,
			expected:	false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := isBonusFile(&test.node)
			if res != test.expected {
				t.Errorf("isBonusFile = %v, want %v", res, test.expected)
			}
		})
	}

}

func TestIsSubtitleDir(t *testing.T) {

	tests := []struct{
		name		string
		node		metadata.Entry
		expected	bool
	}{
		{
			name:		"movie file",
			node:		movieFile,
			expected:	false,
		},
		{
			name:		"episode file",
			node:		episodeFile,
			expected:	false,
		},
		{
			name:		"subtitle file",
			node:		subtitleFile,
			expected:	false,
		},
		{
			name:		"bonus file",
			node:		bonusFile,
			expected:	false,
		},
		{
			name:		"movie directory", 
			node: movieDir,
			expected:	false,
		},
		{
			name:		"series directory", 
			node: seriesDir,
			expected:	false,
		},
		{
			name:		"season directory", 
			node: seasonDir,
			expected:	false,
		},
		{
			name:		"subtitle directory", 
			node: subtitleDir,
			expected:	true,
		},
		{
			name:		"bonus directory", 
			node:		bonusDir,
			expected:	false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := isSubtitleDir(&test.node)
			if res != test.expected {
				t.Errorf("isSubtitleDir = %v, want %v", res, test.expected)
			}
		})
	}
}

func TestIsBonusDir(t *testing.T) {
	tests := []struct{
		name		string
		node		metadata.Entry
		expected	bool
	}{
		{
			name:		"movie file",
			node:		movieFile,
			expected:	false,
		},
		{
			name:		"episode file",
			node:		episodeFile,
			expected:	false,
		},
		{
			name:		"subtitle file",
			node:		subtitleFile,
			expected:	false,
		},
		{
			name:		"bonus file",
			node:		bonusFile,
			expected:	false,
		},
		{
			name:		"movie directory", 
			node: movieDir,
			expected:	false,
		},
		{
			name:		"series directory", 
			node: seriesDir,
			expected:	false,
		},
		{
			name:		"season directory", 
			node: seasonDir,
			expected:	false,
		},
		{
			name:		"subtitle directory", 
			node: subtitleDir,
			expected:	false,
		},
		{
			name:		"bonus directory", 
			node:		bonusDir,
			expected:	true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := isBonusDir(&test.node)
			if res != test.expected {
				t.Errorf("isBonusDir = %v, want %v", res, test.expected)
			}
		})
	}
}

func TestIsMovieDir(t *testing.T) {
	tests := []struct{
		name		string
		node		metadata.Entry
		expected	bool
	}{
		{
			name:		"movie file",
			node:		movieFile,
			expected:	false,
		},
		{
			name:		"episode file",
			node:		episodeFile,
			expected:	false,
		},
		{
			name:		"subtitle file",
			node:		subtitleFile,
			expected:	false,
		},
		{
			name:		"bonus file",
			node:		bonusFile,
			expected:	false,
		},
		{
			name:		"movie directory", 
			node: 		movieDir,
			expected:	true,
		},
		{
			name:		"series directory", 
			node: 		seriesDir,
			expected:	false,
		},
		{
			name:		"season directory", 
			node: 		seasonDir,
			expected:	false,
		},
		{
			name:		"subtitle directory", 
			node: 		subtitleDir,
			expected:	false,
		},
		{
			name:		"bonus directory", 
			node:		bonusDir,
			expected:	false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := isMovieDir(&test.node)
			if res != test.expected {
				t.Errorf("isMovieDir = %v, want %v", res, test.expected)
			}
		})
	}
}

func TestIsSeasonDir(t *testing.T) {
	tests := []struct{
		name		string
		node		metadata.Entry
		expected	bool
	}{
		{
			name:		"movie file",
			node:		movieFile,
			expected:	false,
		},
		{
			name:		"episode file",
			node:		episodeFile,
			expected:	false,
		},
		{
			name:		"subtitle file",
			node:		subtitleFile,
			expected:	false,
		},
		{
			name:		"bonus file",
			node:		bonusFile,
			expected:	false,
		},
		{
			name:		"movie directory", 
			node: 		movieDir,
			expected:	false,
		},
		{
			name:		"series directory", 
			node: 		seriesDir,
			expected:	false,
		},
		{
			name:		"season directory", 
			node: 		seasonDir,
			expected:	true,
		},
		{
			name:		"subtitle directory", 
			node: 		subtitleDir,
			expected:	false,
		},
		{
			name:		"bonus directory", 
			node:		bonusDir,
			expected:	false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := isSeasonDir(&test.node)
			if res != test.expected {
				t.Errorf("isSeasonDir = %v, want %v", res, test.expected)
			}
		})
	}
}

func TestIsSeriesDir(t *testing.T) {
	tests := []struct{
		name		string
		node		metadata.Entry
		expected	bool
	}{
		{
			name:		"movie file",
			node:		movieFile,
			expected:	false,
		},
		{
			name:		"episode file",
			node:		episodeFile,
			expected:	false,
		},
		{
			name:		"subtitle file",
			node:		subtitleFile,
			expected:	false,
		},
		{
			name:		"bonus file",
			node:		bonusFile,
			expected:	false,
		},
		{
			name:		"movie directory", 
			node: movieDir,
			expected:	false,
		},
		{
			name:		"series directory", 
			node: seriesDir,
			expected:	true,
		},
		{
			name:		"season directory", 
			node: seasonDir,
			expected:	false,
		},
		{
			name:		"subtitle directory", 
			node: subtitleDir,
			expected:	false,
		},
		{
			name:		"bonus directory", 
			node:		bonusDir,
			expected:	false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := isSeriesDir(&test.node)
			if res != test.expected {
				t.Errorf("isSeriesDir = %v, want %v", res, test.expected)
			}
		})
	}
}

func TestClassifyEntryFile(t *testing.T) {
	tests := []struct{
		name		string
		node		metadata.Entry
		expected	metadata.EntryRole
	}{
		{
			name:		"movie file",
			node:		movieFile,
			expected:	metadata.MovieFile,
		},
		{
			name:		"episode file",
			node:		episodeFile,
			expected:	metadata.EpisodeFile,
		},
		{
			name:		"subtitle file",
			node:		subtitleFile,
			expected:	metadata.SubtitleFile,
		},
		{
			name:		"bonus file",
			node:		bonusFile,
			expected:	metadata.BonusFile,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := classifyEntryFile(&test.node)
			if res != test.expected {
				t.Errorf("classifyEntryFile = %v, want %v", res, test.expected)
			}
		})
	}
}


func TestClassifyEntryDir(t *testing.T) {
	tests := []struct{
		name		string
		node		metadata.Entry
		expected	metadata.EntryRole
	}{
		{
			name:		"movie directory", 
			node: movieDir,
			expected:	metadata.MovieDir,
		},
		{
			name:		"series directory", 
			node: seriesDir,
			expected:	metadata.SeriesDir,
		},
		{
			name:		"season directory", 
			node: seasonDir,
			expected:	metadata.SeasonDir,
		},
		{
			name:		"subtitle directory", 
			node: subtitleDir,
			expected:	metadata.SubtitleDir,
		},
		{
			name:		"bonus directory", 
			node:		bonusDir,
			expected:	metadata.BonusDir,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := classifyEntryDir(&test.node)
			if res != test.expected {
				t.Errorf("classifyEntryDir = %v, want %v", res, test.expected)
			}
		})
	}
}
