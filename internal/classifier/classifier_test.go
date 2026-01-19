package classifier

import (
	"testing"
	"github.com/ENIACore/media_library_manager/internal/metadata"
)

func TestClassifyEntries(t *testing.T) {
	/*
		Testing movie file
	*/
	err := ClassifyEntries(&movieFile)
	if err != nil {
		t.Errorf("ClassifyEntries has err for %v, want no error", movieFile.PathInfo.Source)
	}
	if movieFile.Role != metadata.MovieFile {
		t.Errorf("Role = %v, want %v", movieFile.Role, metadata.MovieFile)
	}

	/*
		Testing episode file
	*/
	err = ClassifyEntries(&episodeFile)
	if err != nil {
		t.Errorf("ClassifyEntries has err for %v, want no error", episodeFile.PathInfo.Source)
	}
	if episodeFile.Role != metadata.EpisodeFile {
		t.Errorf("Role = %v, want %v", episodeFile.Role, metadata.EpisodeFile)
	}

	/*
		Testing subtitle file
	*/
	err = ClassifyEntries(&subtitleFile)
	if err != nil {
		t.Errorf("ClassifyEntries has err for %v, want no error", subtitleFile.PathInfo.Source)
	}
	if subtitleFile.Role != metadata.SubtitleFile {
		t.Errorf("Role = %v, want %v", subtitleFile.Role, metadata.SubtitleFile)
	}
	
	/*
		Testing bonus file
	*/
	err = ClassifyEntries(&bonusFile)
	if err != nil {
		t.Errorf("ClassifyEntries has err for %v, want no error", bonusFile.PathInfo.Source)
	}
	if bonusFile.Role != metadata.BonusFile {
		t.Errorf("Role = %v, want %v", bonusFile.Role, metadata.BonusFile)
	}



	/*
		Testing movie dir
		Movie Dir Children: movieFile, bonusFile, subtitleFile, subtitleFile
	*/
	err = ClassifyEntries(&movieDir)
	if err != nil {
		t.Errorf("ClassifyEntries has err for %v, want no error", movieDir.PathInfo.Source)
	}
	if movieDir.Role != metadata.MovieDir || movieDir.Children[0].Role != metadata.MovieFile || movieDir.Children[1].Role != metadata.BonusFile || movieDir.Children[2].Role != metadata.SubtitleFile || movieDir.Children[3].Role != metadata.SubtitleFile {
		t.Errorf("Classification for %v failed, view logging for more information", movieDir.PathInfo.Source)
	}

	/*
		Testing season dir
		Season Dir Children: episodeFile, episodeFile, subtitleFile, subtitleFile
	*/
	err = ClassifyEntries(&seasonDir)
	if err != nil {
		t.Errorf("ClassifyEntries has err for %v, want no error", seasonDir.PathInfo.Source)
	}
	if seasonDir.Role != metadata.SeasonDir || seasonDir.Children[0].Role != metadata.EpisodeFile || seasonDir.Children[1].Role != metadata.EpisodeFile || seasonDir.Children[2].Role != metadata.SubtitleFile || seasonDir.Children[3].Role != metadata.SubtitleFile {
		t.Errorf("Classification for %v failed, view logging for more information", seasonDir.PathInfo.Source)
	}

	/*
		Testing series dir
		Series Dir Children: seasonDir, bonusDir, subtitleDir,
		Season Dir Children: episodeFile, episodeFile, subtitleFile, subtitleFile
	*/
	err = ClassifyEntries(&seriesDir)
	if err != nil {
		t.Errorf("ClassifyEntries has err for %v, want no error", seriesDir.PathInfo.Source)
	}
	if seriesDir.Role != metadata.SeriesDir || seriesDir.Children[0].Role != metadata.SeasonDir || seriesDir.Children[1].Role != metadata.BonusDir || seriesDir.Children[2].Role != metadata.SubtitleDir {
		t.Errorf("Classification for %v failed, view logging for more information", seriesDir.PathInfo.Source)
	}
	if seriesDir.Children[0].Role != metadata.SeasonDir || seriesDir.Children[0].Children[0].Role != metadata.EpisodeFile || seriesDir.Children[0].Children[1].Role != metadata.EpisodeFile || seriesDir.Children[0].Children[2].Role != metadata.SubtitleFile || seriesDir.Children[0].Children[3].Role != metadata.SubtitleFile {
		t.Errorf("Classification for %v failed, view logging for more information", seriesDir.Children[0].PathInfo.Source)
	}
	if seriesDir.Children[1].Role != metadata.BonusDir || seriesDir.Children[1].Children[0].Role != metadata.BonusFile  || seriesDir.Children[1].Children[1].Role != metadata.SubtitleFile  || seriesDir.Children[1].Children[2].Role != metadata.SubtitleFile {
		t.Errorf("Classification for %v failed, view logging for more information", seriesDir.Children[1].PathInfo.Source)
	}
	if seriesDir.Children[2].Role != metadata.SubtitleDir || seriesDir.Children[2].Children[0].Role != metadata.SubtitleFile || seriesDir.Children[2].Children[1].Role != metadata.SubtitleFile || seriesDir.Children[2].Children[2].Role != metadata.SubtitleFile {
		t.Errorf("Classification for %v failed, view logging for more information", seriesDir.Children[2].PathInfo.Source)
	}

	/*
		Testing subtitle dir
		Subtitle Dir Children: subtitleFile, subtitleFile, subtitleFile,
	*/
	err = ClassifyEntries(&subtitleDir)
	if err != nil {
		t.Errorf("ClassifyEntries has err for %v, want no error", subtitleDir.PathInfo.Source)
	}
	if subtitleDir.Role != metadata.SubtitleDir || subtitleDir.Children[0].Role != metadata.SubtitleFile || subtitleDir.Children[1].Role != metadata.SubtitleFile || subtitleDir.Children[2].Role != metadata.SubtitleFile {
		t.Errorf("Classification for %v failed, view logging for more information", subtitleDir.PathInfo.Source)
	}
	
	/*
		Testing bonus dir
		Bonus Dir Children: bonusFile, subtitleFile, subtitleFile,
	*/
	err = ClassifyEntries(&bonusDir)
	if err != nil {
		t.Errorf("ClassifyEntries has err for %v, want no error", bonusDir.PathInfo.Source)
	}
	if bonusDir.Role != metadata.BonusDir || bonusDir.Children[0].Role != metadata.BonusFile  || bonusDir.Children[1].Role != metadata.SubtitleFile  || bonusDir.Children[2].Role != metadata.SubtitleFile {
		t.Errorf("Classification for %v failed, view logging for more information", bonusDir.PathInfo.Source)
	}
}


/*
	Classifier helper function tests
*/

func TestClassifySubtitleDir(t *testing.T) {
	/*
		Testing subtitle dir
		Subtitle Dir Children: subtitleFile, subtitleFile, subtitleFile,
	*/
	err := classifySubtitleDir(&subtitleDir)
	if err != nil {
		t.Errorf("classifySubtitleDir has err for %v, want no error", subtitleDir.PathInfo.Source)
	}
	if subtitleDir.Role != metadata.SubtitleDir || subtitleDir.Children[0].Role != metadata.SubtitleFile || subtitleDir.Children[1].Role != metadata.SubtitleFile || subtitleDir.Children[2].Role != metadata.SubtitleFile {
		t.Errorf("Classification for %v failed, view logging for more information", subtitleDir.PathInfo.Source)
	}

	err = classifySubtitleDir(&movieDir)
	if err == nil {
		t.Errorf("classifySubtitleDir has no error, expected error")
	}
}

func TestClassifyBonusDir(t *testing.T) {
	/*
		Testing bonus dir
		Bonus Dir Children: bonusFile, subtitleFile, subtitleFile,
	*/
	err := classifyBonusDir(&bonusDir)
	if err != nil {
		t.Errorf("classifyBonusDir has err for %v, want no error", bonusDir.PathInfo.Source)
	}
	if bonusDir.Role != metadata.BonusDir || bonusDir.Children[0].Role != metadata.BonusFile  || bonusDir.Children[1].Role != metadata.SubtitleFile  || bonusDir.Children[2].Role != metadata.SubtitleFile {
		t.Errorf("Classification for %v failed, view logging for more information", bonusDir.PathInfo.Source)
	}


	err = classifyBonusDir(&movieDir)
	if err == nil {
		t.Errorf("classifyBonusDir has no error, expected error")
	}
}

func TestClassifyFile(t *testing.T) {
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
			res := classifyFile(&test.node)
			if res != test.expected {
				t.Errorf("classifyEntryFile = %v, want %v", res, test.expected)
			}
		})
	}
}


func TestClassifyDir(t *testing.T) {
	tests := []struct{
		name		string
		node		metadata.Entry
		expected	metadata.EntryRole
	}{
		{
			name:		"movie directory", 
			node: 		movieDir,
			expected:	metadata.MovieDir,
		},
		{
			name:		"series directory", 
			node: 		seriesDir,
			expected:	metadata.SeriesDir,
		},
		{
			name:		"season directory", 
			node: 		seasonDir,
			expected:	metadata.SeasonDir,
		},
		{
			name:		"subtitle directory", 
			node: 		subtitleDir,
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
			res := classifyDir(&test.node)
			if res != test.expected {
				t.Errorf("classifyEntryDir = %v, want %v", res, test.expected)
			}
		})
	}
}

/*
	file helper tests
*/

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

/*
	dir helper tests
*/

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


/*
	test helper functions
*/

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
		Type: metadata.Unknown,
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
		Type: metadata.Unknown,
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
		Type: metadata.Unknown,
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
		Type: metadata.Unknown,
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
		Type: metadata.Unknown,
	},
}
