package processor

import (
	"log/slog"
	"testing"

	"github.com/ENIACore/media_library_manager/internal/metadata"
)

var logger = slog.Default()

func TestResolveEntries(t *testing.T) {
	/*
		Testing error cases for entries that cannot be alone at root
	*/
	tests := []struct {
		name  string
		entry *metadata.Entry
	}{
		{
			name:  "subtitle file at root",
			entry: makeSubtitleFile(),
		},
		{
			name:  "bonus file at root",
			entry: makeBonusFile(),
		},
		{
			name:  "subtitle dir at root",
			entry: makeSubtitleDir(),
		},
		{
			name:  "bonus dir at root",
			entry: makeBonusDir(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ResolveEntries(test.entry, logger)
			if err == nil {
				t.Errorf("ResolveEntries expected error for %v, got nil", test.name)
			}
		})
	}

	/*
		Testing movie file at root
	*/
	t.Run("movie file at root", func(t *testing.T) {
		entry := makeMovieFile()
		err := ResolveEntries(entry, logger)
		if err != nil {
			t.Errorf("ResolveEntries unexpected error: %v", err)
		}
		expected := "Test.Movie.2025/Test.Movie.2025.1080p.X264.Remux.Atmos.English.mp4"
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})

	/*
		Testing episode file at root
	*/
	t.Run("episode file at root", func(t *testing.T) {
		entry := makeEpisodeFile()
		err := ResolveEntries(entry, logger)
		if err != nil {
			t.Errorf("ResolveEntries unexpected error: %v", err)
		}
		expected := "Test.Show.2025/S01/Test.Show.S01E01.1080p.X264.Remux.Atmos.English.mp4"
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})

	/*
		Testing movie directory
	*/
	t.Run("movie directory", func(t *testing.T) {
		entry := makeMovieDir()
		err := ResolveEntries(entry, logger)
		if err != nil {
			t.Errorf("ResolveEntries unexpected error: %v", err)
		}

		expectedDir := "Test.Movie.2025"
		if entry.PathInfo.Dest != expectedDir {
			t.Errorf("Dir Dest = %v, want %v", entry.PathInfo.Dest, expectedDir)
		}

		expectedMovie := "Test.Movie.2025/Test.Movie.2025.1080p.X264.Remux.Atmos.English.mp4"
		if entry.Children[0].PathInfo.Dest != expectedMovie {
			t.Errorf("Movie Dest = %v, want %v", entry.Children[0].PathInfo.Dest, expectedMovie)
		}

		expectedBonus := "Test.Movie.2025/Extras/Test.Movie.Behind.The.Scenes.1080p.X264.Remux.Atmos.English.mp4"
		if entry.Children[1].PathInfo.Dest != expectedBonus {
			t.Errorf("Bonus Dest = %v, want %v", entry.Children[1].PathInfo.Dest, expectedBonus)
		}

		expectedSubtitle := "Test.Movie.2025/Subtitles/Test.Movie.English.srt"
		if entry.Children[2].PathInfo.Dest != expectedSubtitle {
			t.Errorf("Subtitle Dest = %v, want %v", entry.Children[2].PathInfo.Dest, expectedSubtitle)
		}
	})

	/*
		Testing season directory
	*/
	t.Run("season directory", func(t *testing.T) {
		entry := makeSeasonDir()
		err := ResolveEntries(entry, logger)
		if err != nil {
			t.Errorf("ResolveEntries unexpected error: %v", err)
		}

		expectedDir := "Test.Show.2025/Season 01"
		if entry.PathInfo.Dest != expectedDir {
			t.Errorf("Dir Dest = %v, want %v", entry.PathInfo.Dest, expectedDir)
		}

		expectedEpisode := "Test.Show.2025/Season 01/Test.Show.S01E01.1080p.X264.Remux.Atmos.English.mp4"
		if entry.Children[0].PathInfo.Dest != expectedEpisode {
			t.Errorf("Episode Dest = %v, want %v", entry.Children[0].PathInfo.Dest, expectedEpisode)
		}

		expectedSubtitle := "Test.Show.2025/Season 01/Subtitles/Test.Show.English.srt"
		if entry.Children[2].PathInfo.Dest != expectedSubtitle {
			t.Errorf("Subtitle Dest = %v, want %v", entry.Children[2].PathInfo.Dest, expectedSubtitle)
		}
	})

	/*
		Testing series directory
	*/
	t.Run("series directory", func(t *testing.T) {
		entry := makeSeriesDir()
		err := ResolveEntries(entry, logger)
		if err != nil {
			t.Errorf("ResolveEntries unexpected error: %v", err)
		}

		expectedDir := "Test.Show.2025"
		if entry.PathInfo.Dest != expectedDir {
			t.Errorf("Dir Dest = %v, want %v", entry.PathInfo.Dest, expectedDir)
		}

		expectedSeasonDir := "Test.Show.2025/Season 01"
		if entry.Children[0].PathInfo.Dest != expectedSeasonDir {
			t.Errorf("Season Dir Dest = %v, want %v", entry.Children[0].PathInfo.Dest, expectedSeasonDir)
		}

		expectedEpisode := "Test.Show.2025/Season 01/Test.Show.S01E01.1080p.X264.Remux.Atmos.English.mp4"
		if entry.Children[0].Children[0].PathInfo.Dest != expectedEpisode {
			t.Errorf("Episode Dest = %v, want %v", entry.Children[0].Children[0].PathInfo.Dest, expectedEpisode)
		}

		expectedBonusDir := "Test.Show.2025/Extras"
		if entry.Children[1].PathInfo.Dest != expectedBonusDir {
			t.Errorf("Bonus Dir Dest = %v, want %v", entry.Children[1].PathInfo.Dest, expectedBonusDir)
		}

		expectedSubtitleDir := "Test.Show.2025/Subtitles"
		if entry.Children[2].PathInfo.Dest != expectedSubtitleDir {
			t.Errorf("Subtitle Dir Dest = %v, want %v", entry.Children[2].PathInfo.Dest, expectedSubtitleDir)
		}
	})
}

func TestResolveMovieFile(t *testing.T) {
	t.Run("with base path", func(t *testing.T) {
		entry := makeMovieFile()
		err := resolveMovieFile("Test.Movie.2025", entry, logger)
		if err != nil {
			t.Errorf("resolveMovieFile unexpected error: %v", err)
		}
		expected := "Test.Movie.2025/Test.Movie.2025.1080p.X264.Remux.Atmos.English.mp4"
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})

	t.Run("without base path", func(t *testing.T) {
		entry := makeMovieFile()
		err := resolveMovieFile("", entry, logger)
		if err != nil {
			t.Errorf("resolveMovieFile unexpected error: %v", err)
		}
		expected := "Test.Movie.2025/Test.Movie.2025.1080p.X264.Remux.Atmos.English.mp4"
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})

	t.Run("wrong role", func(t *testing.T) {
		entry := makeSubtitleFile()
		err := resolveMovieFile("Test.Movie.2025", entry, logger)
		if err == nil {
			t.Errorf("resolveMovieFile expected error for subtitle file")
		}
	})

	t.Run("missing metadata", func(t *testing.T) {
		entry := makeMovieFileMinimal()
		err := resolveMovieFile("", entry, logger)
		if err != nil {
			t.Errorf("resolveMovieFile unexpected error: %v", err)
		}
		expected := "Test.Movie/Test.Movie.mp4"
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})
}

func TestResolveEpisodeFile(t *testing.T) {
	t.Run("with base path and parent season", func(t *testing.T) {
		entry := makeEpisodeFile()
		season := 1
		err := resolveEpisodeFile("Test.Show.2025/Season 01", &season, entry, logger)
		if err != nil {
			t.Errorf("resolveEpisodeFile unexpected error: %v", err)
		}
		expected := "Test.Show.2025/Season 01/Test.Show.S01E01.1080p.X264.Remux.Atmos.English.mp4"
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})

	t.Run("without base path", func(t *testing.T) {
		entry := makeEpisodeFile()
		entry.MediaInfo.Season = intPtr(2)
		err := resolveEpisodeFile("", nil, entry, logger)
		if err != nil {
			t.Errorf("resolveEpisodeFile unexpected error: %v", err)
		}
		expected := "Test.Show.2025/S02/Test.Show.S02E01.1080p.X264.Remux.Atmos.English.mp4"
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})

	t.Run("wrong role", func(t *testing.T) {
		entry := makeMovieFile()
		err := resolveEpisodeFile("Test.Show.2025/Season 01", nil, entry, logger)
		if err == nil {
			t.Errorf("resolveEpisodeFile expected error for movie file")
		}
	})
}

func TestResolveSubtitleFile(t *testing.T) {
	t.Run("valid subtitle file", func(t *testing.T) {
		entry := makeSubtitleFile()
		err := resolveSubtitleFile("Test.Movie.2025", "Test.Movie", entry, logger)
		if err != nil {
			t.Errorf("resolveSubtitleFile unexpected error: %v", err)
		}
		expected := "Test.Movie.2025/Subtitles/Test.Movie.English.srt"
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})

	t.Run("wrong role", func(t *testing.T) {
		entry := makeMovieFile()
		err := resolveSubtitleFile("Test.Movie.2025", "Test.Movie", entry, logger)
		if err == nil {
			t.Errorf("resolveSubtitleFile expected error for movie file")
		}
	})
}

func TestResolveSubtitleDir(t *testing.T) {
	t.Run("valid subtitle dir", func(t *testing.T) {
		entry := makeSubtitleDir()
		err := resolveSubtitleDir("Test.Movie.2025", "Test.Movie", entry, logger)
		if err != nil {
			t.Errorf("resolveSubtitleDir unexpected error: %v", err)
		}

		expected := "Test.Movie.2025/Subtitles"
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dir Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}

		expectedChild := "Test.Movie.2025/Subtitles/Test.Movie.English.srt"
		if entry.Children[0].PathInfo.Dest != expectedChild {
			t.Errorf("Child Dest = %v, want %v", entry.Children[0].PathInfo.Dest, expectedChild)
		}
	})

	t.Run("wrong role", func(t *testing.T) {
		entry := makeMovieDir()
		err := resolveSubtitleDir("Test.Movie.2025", "Test.Movie", entry, logger)
		if err == nil {
			t.Errorf("resolveSubtitleDir expected error for movie dir")
		}
	})
}

func TestResolveBonusFile(t *testing.T) {
	t.Run("valid bonus file", func(t *testing.T) {
		entry := makeBonusFile()
		err := resolveBonusFile("Test.Movie.2025", "Test.Movie", entry, logger)
		if err != nil {
			t.Errorf("resolveBonusFile unexpected error: %v", err)
		}
		expected := "Test.Movie.2025/Extras/Test.Movie.Behind.The.Scenes.1080p.X264.Remux.Atmos.English.mp4"
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})

	t.Run("wrong role", func(t *testing.T) {
		entry := makeMovieFile()
		err := resolveBonusFile("Test.Movie.2025", "Test.Movie", entry, logger)
		if err == nil {
			t.Errorf("resolveBonusFile expected error for movie file")
		}
	})
}

func TestResolveBonusDir(t *testing.T) {
	t.Run("valid bonus dir", func(t *testing.T) {
		entry := makeBonusDir()
		err := resolveBonusDir("Test.Movie.2025", "Test.Movie", entry, logger)
		if err != nil {
			t.Errorf("resolveBonusDir unexpected error: %v", err)
		}

		expected := "Test.Movie.2025/Extras"
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dir Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}

		expectedBonus := "Test.Movie.2025/Extras/Test.Movie.Behind.The.Scenes.1080p.X264.Remux.Atmos.English.mp4"
		if entry.Children[0].PathInfo.Dest != expectedBonus {
			t.Errorf("Bonus Dest = %v, want %v", entry.Children[0].PathInfo.Dest, expectedBonus)
		}

		expectedSubtitle := "Test.Movie.2025/Subtitles/Test.Movie.English.srt"
		if entry.Children[1].PathInfo.Dest != expectedSubtitle {
			t.Errorf("Subtitle Dest = %v, want %v", entry.Children[1].PathInfo.Dest, expectedSubtitle)
		}
	})

	t.Run("wrong role", func(t *testing.T) {
		entry := makeMovieDir()
		err := resolveBonusDir("Test.Movie.2025", "Test.Movie", entry, logger)
		if err == nil {
			t.Errorf("resolveBonusDir expected error for movie dir")
		}
	})
}

func TestResolveMovieDir(t *testing.T) {
	t.Run("full movie dir", func(t *testing.T) {
		entry := makeMovieDir()
		err := resolveMovieDir(entry, logger)
		if err != nil {
			t.Errorf("resolveMovieDir unexpected error: %v", err)
		}

		expected := "Test.Movie.2025"
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dir Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})

	t.Run("movie dir with subdirs", func(t *testing.T) {
		entry := makeMovieDirWithSubdirs()
		err := resolveMovieDir(entry, logger)
		if err != nil {
			t.Errorf("resolveMovieDir unexpected error: %v", err)
		}

		expectedBonusDir := "Test.Movie.2025/Extras"
		if entry.Children[1].PathInfo.Dest != expectedBonusDir {
			t.Errorf("Bonus Dir Dest = %v, want %v", entry.Children[1].PathInfo.Dest, expectedBonusDir)
		}

		expectedSubtitleDir := "Test.Movie.2025/Subtitles"
		if entry.Children[2].PathInfo.Dest != expectedSubtitleDir {
			t.Errorf("Subtitle Dir Dest = %v, want %v", entry.Children[2].PathInfo.Dest, expectedSubtitleDir)
		}
	})
}

func TestResolveSeasonDir(t *testing.T) {
	t.Run("with base path", func(t *testing.T) {
		entry := makeSeasonDir()
		err := resolveSeasonDir("Test.Show.2025", entry, logger)
		if err != nil {
			t.Errorf("resolveSeasonDir unexpected error: %v", err)
		}

		expected := "Test.Show.2025/Season 01"
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dir Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})

	t.Run("without base path", func(t *testing.T) {
		entry := makeSeasonDir()
		err := resolveSeasonDir("", entry, logger)
		if err != nil {
			t.Errorf("resolveSeasonDir unexpected error: %v", err)
		}

		expected := "Test.Show.2025/Season 01"
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dir Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})
}

func TestResolveSeriesDir(t *testing.T) {
	t.Run("full series dir", func(t *testing.T) {
		entry := makeSeriesDir()
		err := resolveSeriesDir(entry, logger)
		if err != nil {
			t.Errorf("resolveSeriesDir unexpected error: %v", err)
		}

		expected := "Test.Show.2025"
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dir Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}

		expectedSeason := "Test.Show.2025/Season 01"
		if entry.Children[0].PathInfo.Dest != expectedSeason {
			t.Errorf("Season Dir Dest = %v, want %v", entry.Children[0].PathInfo.Dest, expectedSeason)
		}
	})
}

/*
	filename builder tests
*/

func TestBuildBasePath(t *testing.T) {
	tests := []struct {
		name     string
		title    []string
		year     *int
		expected string
	}{
		{
			name:     "with year",
			title:    []string{"TEST", "MOVIE"},
			year:     intPtr(2025),
			expected: "Test.Movie.2025",
		},
		{
			name:     "without year",
			title:    []string{"TEST", "MOVIE"},
			year:     nil,
			expected: "Test.Movie",
		},
		{
			name:     "empty title",
			title:    []string{},
			year:     intPtr(2025),
			expected: ".2025",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := buildBasePath(test.title, test.year)
			if result != test.expected {
				t.Errorf("buildBasePath = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestBuildTitle(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		expected string
	}{
		{
			name:     "multiple parts",
			parts:    []string{"TEST", "MOVIE"},
			expected: "Test.Movie",
		},
		{
			name:     "single part",
			parts:    []string{"MOVIE"},
			expected: "Movie",
		},
		{
			name:     "empty",
			parts:    []string{},
			expected: "",
		},
		{
			name:     "mixed case",
			parts:    []string{"TeSt", "MoViE"},
			expected: "Test.Movie",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := buildTitle(test.parts)
			if result != test.expected {
				t.Errorf("buildTitle = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestBuildVideoFilename(t *testing.T) {
	tests := []struct {
		name     string
		info     metadata.MediaInfo
		ext      string
		expected string
	}{
		{
			name: "full metadata",
			info: metadata.MediaInfo{
				Title:      []string{"TEST", "MOVIE"},
				Year:       intPtr(2025),
				Resolution: "1080P",
				Codec:      "X264",
				Source:     "REMUX",
				Audio:      "ATMOS",
				Language:   "ENGLISH",
			},
			ext:      "MP4",
			expected: "Test.Movie.2025.1080p.X264.Remux.Atmos.English.mp4",
		},
		{
			name: "minimal metadata",
			info: metadata.MediaInfo{
				Title: []string{"TEST", "MOVIE"},
			},
			ext:      "MKV",
			expected: "Test.Movie.mkv",
		},
		{
			name: "partial metadata",
			info: metadata.MediaInfo{
				Title:      []string{"TEST", "MOVIE"},
				Year:       intPtr(2025),
				Resolution: "4K",
			},
			ext:      "MP4",
			expected: "Test.Movie.2025.4k.mp4",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := buildVideoFilename(test.info, test.ext)
			if result != test.expected {
				t.Errorf("buildVideoFilename = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestBuildEpisodeFilename(t *testing.T) {
	tests := []struct {
		name         string
		info         metadata.MediaInfo
		parentSeason *int
		ext          string
		expected     string
	}{
		{
			name: "with episode season and parent season",
			info: metadata.MediaInfo{
				Title:      []string{"TEST", "SHOW"},
				Season:     intPtr(1),
				Episode:    intPtr(5),
				Resolution: "1080P",
				Codec:      "X264",
				Source:     "REMUX",
				Audio:      "ATMOS",
				Language:   "ENGLISH",
			},
			parentSeason: intPtr(2),
			ext:          "MP4",
			expected:     "Test.Show.S01E05.1080p.X264.Remux.Atmos.English.mp4",
		},
		{
			name: "with parent season only",
			info: metadata.MediaInfo{
				Title:      []string{"TEST", "SHOW"},
				Episode:    intPtr(3),
				Resolution: "1080P",
			},
			parentSeason: intPtr(2),
			ext:          "MKV",
			expected:     "Test.Show.S02E03.1080p.mkv",
		},
		{
			name: "no season info",
			info: metadata.MediaInfo{
				Title:   []string{"TEST", "SHOW"},
				Episode: intPtr(1),
			},
			parentSeason: nil,
			ext:          "MP4",
			expected:     "Test.Show.S00E01.mp4",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := buildEpisodeFilename(test.info, test.parentSeason, test.ext)
			if result != test.expected {
				t.Errorf("buildEpisodeFilename = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestBuildSubtitleFilename(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		info     metadata.MediaInfo
		ext      string
		expected string
	}{
		{
			name:  "with language",
			title: "Test.Movie",
			info: metadata.MediaInfo{
				Language: "ENGLISH",
			},
			ext:      "SRT",
			expected: "Test.Movie.English.srt",
		},
		{
			name:     "without language",
			title:    "Test.Movie",
			info:     metadata.MediaInfo{},
			ext:      "SRT",
			expected: "Test.Movie.srt",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := buildSubtitleFilename(test.title, test.info, test.ext)
			if result != test.expected {
				t.Errorf("buildSubtitleFilename = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestBuildBonusFilename(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		info     metadata.MediaInfo
		ext      string
		expected string
	}{
		{
			name:  "full metadata",
			title: "Test.Movie",
			info: metadata.MediaInfo{
				Bonus:      "BEHIND_THE_SCENES",
				Resolution: "1080P",
				Codec:      "X264",
				Source:     "REMUX",
				Audio:      "ATMOS",
				Language:   "ENGLISH",
			},
			ext:      "MP4",
			expected: "Test.Movie.Behind.The.Scenes.1080p.X264.Remux.Atmos.English.mp4",
		},
		{
			name:  "minimal metadata",
			title: "Test.Movie",
			info: metadata.MediaInfo{
				Bonus: "DELETED_SCENES",
			},
			ext:      "MKV",
			expected: "Test.Movie.Deleted.Scenes.mkv",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := buildBonusFilename(test.title, test.info, test.ext)
			if result != test.expected {
				t.Errorf("buildBonusFilename = %v, want %v", result, test.expected)
			}
		})
	}
}

/*
	string helper tests
*/

func TestCapitalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "uppercase",
			input:    "TEST",
			expected: "Test",
		},
		{
			name:     "lowercase",
			input:    "test",
			expected: "Test",
		},
		{
			name:     "mixed",
			input:    "TeSt",
			expected: "Test",
		},
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
		{
			name:     "single char",
			input:    "t",
			expected: "T",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := capitalize(test.input)
			if result != test.expected {
				t.Errorf("capitalize = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestFormatBonus(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with underscores",
			input:    "BEHIND_THE_SCENES",
			expected: "Behind.The.Scenes",
		},
		{
			name:     "single word",
			input:    "INTERVIEW",
			expected: "Interview",
		},
		{
			name:     "lowercase with underscores",
			input:    "deleted_scenes",
			expected: "Deleted.Scenes",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := formatBonus(test.input)
			if result != test.expected {
				t.Errorf("formatBonus = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestJoinPath(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		expected string
	}{
		{
			name:     "multiple parts",
			parts:    []string{"Test.Movie.2025", "Subtitles", "file.srt"},
			expected: "Test.Movie.2025/Subtitles/file.srt",
		},
		{
			name:     "with empty parts",
			parts:    []string{"Test.Movie.2025", "", "file.srt"},
			expected: "Test.Movie.2025/file.srt",
		},
		{
			name:     "single part",
			parts:    []string{"Test.Movie.2025"},
			expected: "Test.Movie.2025",
		},
		{
			name:     "all empty",
			parts:    []string{"", "", ""},
			expected: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := joinPath(test.parts...)
			if result != test.expected {
				t.Errorf("joinPath = %v, want %v", result, test.expected)
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
	test entry factory functions
*/

func makeSubtitleFile() *metadata.Entry {
	return &metadata.Entry{
		Parent:   nil,
		Children: nil,
		MediaInfo: metadata.MediaInfo{
			Title:    []string{"SUBTITLE"},
			Language: "ENGLISH",
		},
		PathInfo: metadata.PathInfo{
			IsDir:  false,
			Dest:   "",
			Source: "/test movie/subtitles/subtitle english.srt",
			Ext:    "SRT",
			Type:   metadata.Subtitle,
		},
		Role: metadata.SubtitleFile,
	}
}

func makeBonusFile() *metadata.Entry {
	return &metadata.Entry{
		Parent:   nil,
		Children: nil,
		MediaInfo: metadata.MediaInfo{
			Title:      []string{"TEST", "MOVIE", "BEHIND", "THE", "SCENES"},
			Year:       intPtr(2025),
			Resolution: "1080P",
			Codec:      "X264",
			Source:     "REMUX",
			Audio:      "ATMOS",
			Language:   "ENGLISH",
			Bonus:      "BEHIND_THE_SCENES",
		},
		PathInfo: metadata.PathInfo{
			IsDir:  false,
			Dest:   "",
			Source: "/test movie/test movie behind the scenes 2025 1080p.x264.remux.atmos.english.mp4",
			Ext:    "MP4",
			Type:   metadata.Video,
		},
		Role: metadata.BonusFile,
	}
}

func makeEpisodeFile() *metadata.Entry {
	return &metadata.Entry{
		Parent:   nil,
		Children: nil,
		MediaInfo: metadata.MediaInfo{
			Title:      []string{"TEST", "SHOW"},
			Year:       intPtr(2025),
			Episode:    intPtr(1),
			Season:     intPtr(1),
			Resolution: "1080P",
			Codec:      "X264",
			Source:     "REMUX",
			Audio:      "ATMOS",
			Language:   "ENGLISH",
		},
		PathInfo: metadata.PathInfo{
			IsDir:  false,
			Dest:   "",
			Source: "/test show/season 01/test episode E01 2025 1080p.x264.remux.atmos.english.mp4",
			Ext:    "MP4",
			Type:   metadata.Video,
		},
		Role: metadata.EpisodeFile,
	}
}

func makeMovieFile() *metadata.Entry {
	return &metadata.Entry{
		Parent:   nil,
		Children: nil,
		MediaInfo: metadata.MediaInfo{
			Title:      []string{"TEST", "MOVIE"},
			Year:       intPtr(2025),
			Resolution: "1080P",
			Codec:      "X264",
			Source:     "REMUX",
			Audio:      "ATMOS",
			Language:   "ENGLISH",
		},
		PathInfo: metadata.PathInfo{
			IsDir:  false,
			Dest:   "",
			Source: "/test movie/test movie 2025 1080p.x264.remux.atmos.english.mp4",
			Ext:    "MP4",
			Type:   metadata.Video,
		},
		Role: metadata.MovieFile,
	}
}

func makeMovieFileMinimal() *metadata.Entry {
	return &metadata.Entry{
		Parent:   nil,
		Children: nil,
		MediaInfo: metadata.MediaInfo{
			Title: []string{"TEST", "MOVIE"},
		},
		PathInfo: metadata.PathInfo{
			IsDir:  false,
			Dest:   "",
			Source: "/test movie/test movie.mp4",
			Ext:    "MP4",
			Type:   metadata.Video,
		},
		Role: metadata.MovieFile,
	}
}

func makeSubtitleDir() *metadata.Entry {
	return &metadata.Entry{
		Children: []*metadata.Entry{
			makeSubtitleFile(),
			makeSubtitleFile(),
			makeSubtitleFile(),
		},
		MediaInfo: metadata.MediaInfo{
			Title: []string{"SUBTITLES"},
		},
		PathInfo: metadata.PathInfo{
			IsDir:  true,
			Source: "/test movie/subtitles",
			Type:   metadata.UnknownType,
		},
		Role: metadata.SubtitleDir,
	}
}

func makeBonusDir() *metadata.Entry {
	return &metadata.Entry{
		Children: []*metadata.Entry{
			makeBonusFile(),
			makeSubtitleFile(),
			makeSubtitleFile(),
		},
		MediaInfo: metadata.MediaInfo{
			Title: []string{"EXTRAS"},
			Bonus: "EXTRA",
		},
		PathInfo: metadata.PathInfo{
			IsDir:  true,
			Source: "/test movie/extras",
			Type:   metadata.UnknownType,
		},
		Role: metadata.BonusDir,
	}
}

func makeSeasonDir() *metadata.Entry {
	return &metadata.Entry{
		Children: []*metadata.Entry{
			makeEpisodeFile(),
			makeEpisodeFile(),
			makeSubtitleFile(),
			makeSubtitleFile(),
		},
		MediaInfo: metadata.MediaInfo{
			Season: intPtr(1),
		},
		PathInfo: metadata.PathInfo{
			IsDir:  true,
			Source: "/test show/season 01",
			Type:   metadata.UnknownType,
		},
		Role: metadata.SeasonDir,
	}
}

func makeSeriesDir() *metadata.Entry {
	return &metadata.Entry{
		Children: []*metadata.Entry{
			makeSeasonDir(),
			makeBonusDir(),
			makeSubtitleDir(),
		},
		MediaInfo: metadata.MediaInfo{
			Title: []string{"TEST", "SHOW"},
			Year:  intPtr(2025),
		},
		PathInfo: metadata.PathInfo{
			IsDir:  true,
			Source: "/test show",
			Type:   metadata.UnknownType,
		},
		Role: metadata.SeriesDir,
	}
}

func makeMovieDir() *metadata.Entry {
	return &metadata.Entry{
		Children: []*metadata.Entry{
			makeMovieFile(),
			makeBonusFile(),
			makeSubtitleFile(),
			makeSubtitleFile(),
		},
		MediaInfo: metadata.MediaInfo{
			Title: []string{"TEST", "MOVIE"},
			Year:  intPtr(2025),
		},
		PathInfo: metadata.PathInfo{
			IsDir:  true,
			Source: "/test movie",
			Type:   metadata.UnknownType,
		},
		Role: metadata.MovieDir,
	}
}

func makeMovieDirWithSubdirs() *metadata.Entry {
	return &metadata.Entry{
		Children: []*metadata.Entry{
			makeMovieFile(),
			makeBonusDir(),
			makeSubtitleDir(),
		},
		MediaInfo: metadata.MediaInfo{
			Title: []string{"TEST", "MOVIE"},
			Year:  intPtr(2025),
		},
		PathInfo: metadata.PathInfo{
			IsDir:  true,
			Source: "/test movie",
			Type:   metadata.UnknownType,
		},
		Role: metadata.MovieDir,
	}
}
