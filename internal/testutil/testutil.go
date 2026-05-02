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
			Title:   []string{},
			Year:    IntPtr(2025),
			Episode: nil,
			Season:  nil,
			Bonus:   "",
		},
		FileInfo: metadata.FileInfo{
			DestPath:    "",
			SourcePath:  "subtitle.english.srt",
			Ext:         "SRT",
			ContentType: metadata.Subtitle,
			IsDir:       false,
			Language:    []string{"English"},
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
			Title:   []string{"TEST", "TITLE"},
			Year:    IntPtr(2025),
			Episode: nil,
			Season:  nil,
			Bonus:   "Behind.The.Scenes",
		},
		FileInfo: metadata.FileInfo{
			DestPath:    "",
			SourcePath:  "test.title.2025.behind.the.scenes.1080p.x264.remux.atmos.english.mp4",
			Ext:         "MP4",
			ContentType: metadata.Video,
			IsDir:       false,
			Resolution:  "1080p",
			Codec:       "x264",
			Audio:       "Atmos",
			Language:    []string{"English"},
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
			Title:   []string{"TEST", "TITLE"},
			Year:    IntPtr(2025),
			Episode: IntPtr(1),
			Season:  IntPtr(1),
			Bonus:   "",
		},
		FileInfo: metadata.FileInfo{
			DestPath:    "",
			SourcePath:  "test.title.2025.S01E01.1080p.x264.remux.atmos.english.mp4",
			Ext:         "MP4",
			ContentType: metadata.Video,
			IsDir:       false,
			Resolution:  "1080p",
			Codec:       "x264",
			Audio:       "Atmos",
			Language:    []string{"English"},
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
			Title:   []string{"TEST", "TITLE"},
			Year:    IntPtr(2025),
			Episode: nil,
			Season:  nil,
			Bonus:   "",
		},
		FileInfo: metadata.FileInfo{
			DestPath:    "",
			SourcePath:  "test.title.2025.1080p.x264.remux.atmos.english.mp4",
			Ext:         "MP4",
			ContentType: metadata.Video,
			IsDir:       false,
			Resolution:  "1080p",
			Codec:       "x264",
			Audio:       "Atmos",
			Language:    []string{"English"},
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
			Title:   []string{},
			Year:    nil,
			Episode: nil,
			Season:  nil,
			Bonus:   "",
		},
		FileInfo: metadata.FileInfo{
			DestPath:    "",
			SourcePath:  "subtitles",
			Ext:         "",
			ContentType: metadata.UnknownType,
			IsDir:       true,
		},

		Role: metadata.SubtitleDir,
	}

	for _, child := range children {
		child.Parent = root
		child.FileInfo.SourcePath = filepath.Join("subtitles", child.FileInfo.SourcePath)
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
			Title:   []string{},
			Year:    nil,
			Episode: nil,
			Season:  nil,
			Bonus:   "Extra",
		},
		FileInfo: metadata.FileInfo{
			DestPath:    "",
			SourcePath:  "extras",
			Ext:         "",
			ContentType: metadata.UnknownType,
			IsDir:       true,
		},

		Role: metadata.BonusDir,
	}

	for _, child := range children {
		child.Parent = root
		child.FileInfo.SourcePath = filepath.Join("extras", child.FileInfo.SourcePath)
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
			Title:   []string{},
			Year:    nil,
			Episode: nil,
			Season:  IntPtr(1),
			Bonus:   "",
		},
		FileInfo: metadata.FileInfo{
			DestPath:    "",
			SourcePath:  "S01",
			Ext:         "",
			ContentType: metadata.UnknownType,
			IsDir:       true,
		},

		Role: metadata.SeasonDir,
	}

	for _, child := range children {
		child.Parent = root
		child.FileInfo.SourcePath = filepath.Join("S01", child.FileInfo.SourcePath)
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
			Title:   []string{"TEST", "TITLE"},
			Year:    nil,
			Episode: nil,
			Season:  nil,
			Bonus:   "",
		},
		FileInfo: metadata.FileInfo{
			DestPath:    "",
			SourcePath:  "test.title",
			Ext:         "",
			ContentType: metadata.UnknownType,
			IsDir:       true,
		},

		Role: metadata.SeriesDir,
	}

	for _, child := range children {
		child.Parent = root
		child.FileInfo.SourcePath = filepath.Join("test.title", child.FileInfo.SourcePath)
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
			Title:   []string{"TEST", "TITLE"},
			Year:    IntPtr(2025),
			Episode: nil,
			Season:  nil,
			Bonus:   "",
		},
		FileInfo: metadata.FileInfo{
			DestPath:    "",
			SourcePath:  "test.title.2025",
			Ext:         "",
			ContentType: metadata.UnknownType,
			IsDir:       true,
		},

		Role: metadata.MovieDir,
	}

	for _, child := range children {
		child.Parent = root
		child.FileInfo.SourcePath = filepath.Join("test.title.2025", child.FileInfo.SourcePath)
	}

	return root
}

// CreateTestFiles recursively creates the actual filesystem structure for test entries.
// Creates directories and empty files based on the Entry tree structure.
// Updates entry source paths to absolute paths within testDir. Panics on filesystem errors.
func CreateTestFiles(root *metadata.Entry, testDir string) {
	source := filepath.Join(testDir, root.FileInfo.SourcePath)
	root.FileInfo.SourcePath = source

	if root.FileInfo.IsDir {
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
