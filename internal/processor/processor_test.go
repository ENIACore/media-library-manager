package processor

import (
	"testing"
	"github.com/ENIACore/media_library_manager/internal/metadata"
)

func TestResolveEntries(t *testing.T) {
	/*	
		subtitle file
	*/
	err := ResolveEntries(&subtitleFile)
	if err == nil {
		//t.Errorf("ResolveEntries != error, expect error for subtitle file")
	}
	/*	
		bonus file
	*/
	err = ResolveEntries(&bonusFile)
	if err == nil {
		//t.Errorf("ResolveEntries != error, expect error for bonus file")
	}
	/*	
		episode file
	*/
	err = ResolveEntries(&episodeFile)
	/*
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
	*/
	if episodeFile.PathInfo.Dest != "Test.Episode.2025.E01.1080P.X264.REMUX.ATMOS.ENGLISH" {

	}
	if err != nil {
		//t.Errorf("ResolveEntries = error, expect no error for episode file")
	}
	/*	
		movie file
	*/
	/*	
		subtitle directory
	*/
	/*	
		bonus directory
	*/
	/*	
		season directory
	*/
	/*	
		series directory
	*/
	/*	
		movie directory
	*/
}

/*
	test helper functions
*/
func resetEntryRoles(entry *metadata.Entry) {
	entry.Role = metadata.UnknownRole
	for _, child := range entry.Children {
		child.Role = metadata.UnknownRole
		resetEntryRoles(child)
	}
}

func intPtr(i int) *int {
	return &i
}

/*
	test files
*/

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

/*
	test directories
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
		Type: metadata.UnknownType,
	},
}

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
		Type: metadata.UnknownType,
	},
}

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
		Type: metadata.UnknownType,
	},
}

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
		Type: metadata.UnknownType,
	},
}

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
		Type: metadata.UnknownType,
	},
}
