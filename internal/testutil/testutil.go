package testutil

import (
	"log/slog"
	"testing"
	"path/filepath"
	"os"

	"github.com/ENIACore/media_library_manager/internal/config"
	"github.com/ENIACore/media_library_manager/internal/metadata"
)

var logger = slog.Default()

func IntPtr(i int) *int {
	return &i
}

func CreateTestDir(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()
	
	os.MkdirAll(filepath.Join(tempDir, "source"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "movies"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "shows"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "manager"), 0755)
	
	return tempDir
}

func CreateTestCfg(testDir string) config.Config {
	return config.Config{
		MoviePath:   filepath.Join(testDir, "movies"),
		ShowPath:    filepath.Join(testDir, "shows"),
		ManagerPath: filepath.Join(testDir, "manager"),
		DryRun:      false,
	}

}

func CreateTestSubFile(parent *metadata.Entry) *metadata.Entry {
	return &metadata.Entry{
		Parent: parent,
		Children: nil,
		MediaInfo: metadata.MediaInfo{
			Title:		[]string{
			},
			Year:		IntPtr(2025),
			Episode:	nil,
			Season:		nil,
			Resolution:	"",
			Codec:		"",
			Source:		"",
			Audio:		"",
			Language:	"English",
			Bonus: 		"",
		},
		PathInfo: metadata.PathInfo{
			Dest:		"",
			Source: 	"subtitle.english.srt",
			Ext: 		"SRT",	
			Type: 		metadata.Subtitle,
			IsDir: 		false,
		},

		Role: metadata.SubtitleFile,
	}
}

func CreateTestBonusFile(parent *metadata.Entry) *metadata.Entry {
	return &metadata.Entry{
		Parent: parent,
		Children: nil,

		MediaInfo: metadata.MediaInfo{
			Title:		[]string{
				"TEST",
				"TITLE",
			},
			Year:		IntPtr(2025),
			Episode:	nil,
			Season:		nil,
			Resolution:	"1080p",
			Codec:		"x264",
			Source:		"Remux",
			Audio:		"Atmos",
			Language:	"English",
			Bonus: 		"Behind.The.Scenes",
		},
		PathInfo: metadata.PathInfo{
			Dest:		"",
			Source: 	"test.title.2025.behind.the.scenes.1080p.x264.remux.atmos.english.mp4",

			Ext: 		"MP4",	
			Type: 		metadata.Video,
			IsDir: 		false,
		},

		Role: metadata.BonusFile,
	}
}

func CreateTestEpFile(parent *metadata.Entry) *metadata.Entry {
	return &metadata.Entry{
		Parent: parent,
		Children: nil,

		MediaInfo: metadata.MediaInfo{
			Title:		[]string{
				"TEST",
				"TITLE",
			},
			Year:		IntPtr(2025),
			Episode:	IntPtr(1),
			Season:		IntPtr(1),
			Resolution:	"1080p",
			Codec:		"x264",
			Source:		"Remux",
			Audio:		"Atmos",
			Language:	"English",
			Bonus: 		"",
		},
		PathInfo: metadata.PathInfo{
			Dest:		"",
			Source: 	"test.title.2025.S01E01.1080p.x264.remux.atmos.english.mp4",

			Ext: 		"MP4",	
			Type: 		metadata.Video,
			IsDir: 		false,
		},

		Role: metadata.EpisodeFile,
	}
}

func CreateTestMovieFile(parent *metadata.Entry) *metadata.Entry {
	return &metadata.Entry{
		Parent: parent,
		Children: nil,

		MediaInfo: metadata.MediaInfo{
			Title:		[]string{
				"TEST",
				"TITLE",
			},
			Year:		IntPtr(2025),
			Episode:	nil,
			Season:		nil,
			Resolution:	"1080p",
			Codec:		"x264",
			Source:		"Remux",
			Audio:		"Atmos",
			Language:	"English",
			Bonus: 		"",
		},
		PathInfo: metadata.PathInfo{
			Dest:		"",
			Source: 	"test.title.2025.1080p.x264.remux.atmos.english.mp4",
			Ext: 		"MP4",	
			Type: 		metadata.Video,
			IsDir: 		false,
		},

		Role: metadata.MovieFile,
	}
}

func CreateTestSubDir(parent *metadata.Entry, children ...*metadata.Entry) *metadata.Entry {
	root := &metadata.Entry{
		Parent: parent,
		Children: children,

		MediaInfo: metadata.MediaInfo{
			Title:		[]string{
			},
			Year:		nil,
			Episode:	nil,
			Season:		nil,
			Resolution:	"",
			Codec:		"",
			Source:		"",
			Audio:		"",
			Language:	"",
			Bonus: 		"",
		},
		PathInfo: metadata.PathInfo{
			Dest:		"",
			Source: 	"subtitles",
			Ext: 		"",	
			Type: 		metadata.UnknownType,
			IsDir: 		true,
		},

		Role: metadata.SubtitleDir,
	}

	for _, child := range children {
		child.Parent = root
		child.PathInfo.Source = filepath.Join("subtitles", child.PathInfo.Source)
	}

	return root
}

func CreateTestBonusDir(parent *metadata.Entry, children ...*metadata.Entry) *metadata.Entry {
	root := &metadata.Entry{
		Parent: parent,
		Children: children,

		MediaInfo: metadata.MediaInfo{
			Title:		[]string{
			},
			Year:		nil,
			Episode:	nil,
			Season:		nil,
			Resolution:	"",
			Codec:		"",
			Source:		"",
			Audio:		"",
			Language:	"",
			Bonus: 		"Extra",
		},
		PathInfo: metadata.PathInfo{
			Dest:		"",
			Source: 	"extras",
			Ext: 		"",	
			Type: 		metadata.UnknownType,
			IsDir: 		true,
		},

		Role: metadata.BonusDir,
	}

	for _, child := range children {
		child.Parent = root
		child.PathInfo.Source = filepath.Join("extras", child.PathInfo.Source)
	}

	return root
}

func CreateTestSeasonDir(parent *metadata.Entry, children ...*metadata.Entry) *metadata.Entry {
	root := &metadata.Entry{
		Parent: parent,
		Children: children,

		MediaInfo: metadata.MediaInfo{
			Title:		[]string{
			},
			Year:		nil,
			Episode:	nil,
			Season:		IntPtr(1),
			Resolution:	"",
			Codec:		"",
			Source:		"",
			Audio:		"",
			Language:	"",
			Bonus: 		"",
		},
		PathInfo: metadata.PathInfo{
			Dest:		"",
			Source: 	"S01",
			Ext: 		"",	
			Type: 		metadata.UnknownType,
			IsDir: 		true,
		},

		Role: metadata.SeasonDir,
	}

	for _, child := range children {
		child.Parent = root
		child.PathInfo.Source = filepath.Join("S01", child.PathInfo.Source)
	}

	return root
}


func CreateTestSeriesDir(parent *metadata.Entry, children ...*metadata.Entry) *metadata.Entry {
	root := &metadata.Entry{
		Parent: parent,
		Children: children,

		MediaInfo: metadata.MediaInfo{
			Title:		[]string{
				"TEST",
				"TITLE",
			},
			Year:		nil,
			Episode:	nil,
			Season:		nil,
			Resolution:	"",
			Codec:		"",
			Source:		"",
			Audio:		"",
			Language:	"",
			Bonus: 		"",
		},
		PathInfo: metadata.PathInfo{
			Dest:		"",
			Source: 	"test.title",
			Ext: 		"",	
			Type: 		metadata.UnknownType,
			IsDir: 		true,
		},

		Role: metadata.SeriesDir,
	}

	for _, child := range children {
		child.Parent = root
		child.PathInfo.Source = filepath.Join("test.title", child.PathInfo.Source)
	}

	return root
}

func CreateTestMovieDir(parent *metadata.Entry, children ...*metadata.Entry) *metadata.Entry {
	root := &metadata.Entry{
		Parent: parent,
		Children: children,

		MediaInfo: metadata.MediaInfo{
			Title:		[]string{
				"TEST",
				"TITLE",
			},
			Year:		IntPtr(2025),
			Episode:	nil,
			Season:		nil,
			Resolution:	"",
			Codec:		"",
			Source:		"",
			Audio:		"",
			Language:	"",
			Bonus: 		"",
		},
		PathInfo: metadata.PathInfo{
			Dest:		"",
			Source: 	"test.title.2025",
			Ext: 		"",	
			Type: 		metadata.UnknownType,
			IsDir: 		true,
		},

		Role: metadata.MovieDir,
	}

	for _, child := range children {
		child.Parent = root
		child.PathInfo.Source = filepath.Join("test.title", child.PathInfo.Source)
	}

	return root
}

func CreateTestFiles(root *metadata.Entry, testDir string) {
	source := filepath.Join(testDir, root.PathInfo.Source)	
	root.PathInfo.Source = source

	if !root.PathInfo.IsDir {
		dir := filepath.Dir(source)
		if err := os.MkdirAll(dir, 0755); err != nil {
			panic("Unable to make test directory")
    	}


		if _, err := os.Create(source); err != nil {
			panic("Unable to make test file")
		}		
	}

	for _, child := range root.Children {
		CreateTestFiles(child, testDir)
	}
}
