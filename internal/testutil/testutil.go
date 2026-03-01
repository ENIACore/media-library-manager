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

// IntPtr returns a pointer to the provided integer value.
// Helper function for creating int pointers in test structs.
func IntPtr(i int) *int {
	return &i
}

// CreateTestDir creates a temporary directory structure for testing.
// Creates subdirectories: source, movies, shows, and manager.
// Returns the temporary directory path. Cleanup is automatic via t.TempDir().
func CreateTestDir(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()
	
	os.MkdirAll(filepath.Join(tempDir, "source"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "movies"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "shows"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "manager"), 0755)
	
	return tempDir
}

// CreateTestCfg creates a test configuration pointing to subdirectories within testDir.
// Sets up paths for movies, shows, and manager directories with DryRun disabled.
func CreateTestCfg(testDir string) config.Config {
	return config.Config{
		MoviePath:   filepath.Join(testDir, "movies"),
		ShowPath:    filepath.Join(testDir, "shows"),
		ManagerPath: filepath.Join(testDir, "manager"),
		DryRun:      false,
	}

}

// CreateTestSubFile creates a test Entry representing a subtitle file.
// Returns an Entry with SubtitleFile role, English language, and .srt extension.
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
			Language:	[]string{"English"},
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

// CreateTestBonusFile creates a test Entry representing a bonus content file.
// Returns an Entry with BonusFile role and "Behind.The.Scenes" bonus metadata.
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
			Language:	[]string{"English"},
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

// CreateTestEpFile creates a test Entry representing a TV episode file.
// Returns an Entry with EpisodeFile role, S01E01 metadata, and video file properties.
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
			Episode:	[]int{1},
			Season:		IntPtr(1),
			Resolution:	"1080p",
			Codec:		"x264",
			Source:		"Remux",
			Audio:		"Atmos",
			Language:	[]string{"English"},
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

// CreateTestMovieFile creates a test Entry representing a movie file.
// Returns an Entry with MovieFile role, no season/episode metadata, and video file properties.
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
			Language:	[]string{"English"},
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

// CreateTestSubDir creates a test Entry representing a subtitle directory.
// Returns an Entry with SubtitleDir role containing the provided children.
// Updates child paths to be relative to "subtitles" directory.
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
			Language:	nil,
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

// CreateTestBonusDir creates a test Entry representing a bonus content directory.
// Returns an Entry with BonusDir role containing the provided children.
// Updates child paths to be relative to "extras" directory.
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
			Language:	nil,
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

// CreateTestSeasonDir creates a test Entry representing a TV season directory.
// Returns an Entry with SeasonDir role for Season 1 containing the provided children.
// Updates child paths to be relative to "S01" directory.
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
			Language:	nil,
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


// CreateTestSeriesDir creates a test Entry representing a TV series directory.
// Returns an Entry with SeriesDir role containing the provided children.
// Updates child paths to be relative to "test.title" directory.
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
			Language:	nil,
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

// CreateTestMovieDir creates a test Entry representing a movie directory.
// Returns an Entry with MovieDir role for a 2025 release containing the provided children.
// Updates child paths to be relative to "test.title" directory.
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
			Language:	nil,
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
		child.PathInfo.Source = filepath.Join("test.title.2025", child.PathInfo.Source)
	}

	return root
}

// CreateTestFiles recursively creates the actual filesystem structure for test entries.
// Creates directories and empty files based on the Entry tree structure.
// Updates entry source paths to absolute paths within testDir. Panics on filesystem errors.
func CreateTestFiles(root *metadata.Entry, testDir string) {
	source := filepath.Join(testDir, root.PathInfo.Source)	
	root.PathInfo.Source = source

	if root.PathInfo.IsDir {
        // Create the directory itself
		if err := os.MkdirAll(source, 0755); err != nil {
			panic("Unable to make test directory")
		}
	} else {
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
