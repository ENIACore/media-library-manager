package enricher

import (
	"log/slog"
	"testing"

	"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/ENIACore/media_library_manager/internal/testutil"
)

var logger = slog.Default()

func TestEnrich(t *testing.T) {
	tests := []struct {
		name        string
		entry       func() *metadata.Entry
		expectError bool
	}{
		{
			name: "movie file at root",
			entry: func() *metadata.Entry {
				return testutil.CreateTestMovieFile(nil)
			},
			expectError: false,
		},
		{
			name: "episode file at root",
			entry: func() *metadata.Entry {
				return testutil.CreateTestEpFile(nil)
			},
			expectError: false,
		},
		{
			name: "movie directory",
			entry: func() *metadata.Entry {
				movieFile := testutil.CreateTestMovieFile(nil)
				bonusFile := testutil.CreateTestBonusFile(nil)
				return testutil.CreateTestMovieDir(nil, movieFile, bonusFile)
			},
			expectError: false,
		},
		{
			name: "season directory",
			entry: func() *metadata.Entry {
				ep1 := testutil.CreateTestEpFile(nil)
				ep2 := testutil.CreateTestEpFile(nil)
				return testutil.CreateTestSeasonDir(nil, ep1, ep2)
			},
			expectError: false,
		},
		{
			name: "series directory",
			entry: func() *metadata.Entry {
				ep := testutil.CreateTestEpFile(nil)
				seasonDir := testutil.CreateTestSeasonDir(nil, ep)
				return testutil.CreateTestSeriesDir(nil, seasonDir)
			},
			expectError: false,
		},
		{
			name: "subtitle file at root",
			entry: func() *metadata.Entry {
				return testutil.CreateTestSubFile(nil)
			},
			expectError: false,
		},
		{
			name: "bonus file at root",
			entry: func() *metadata.Entry {
				return testutil.CreateTestBonusFile(nil)
			},
			expectError: false,
		},
		{
			name: "subtitle dir at root - should error",
			entry: func() *metadata.Entry {
				sub := testutil.CreateTestSubFile(nil)
				return testutil.CreateTestSubDir(nil, sub)
			},
			expectError: true,
		},
		{
			name: "bonus dir at root - should error",
			entry: func() *metadata.Entry {
				bonus := testutil.CreateTestBonusFile(nil)
				return testutil.CreateTestBonusDir(nil, bonus)
			},
			expectError: true,
		},
		{
			name: "DS dir at root - should error",
			entry: func() *metadata.Entry {
				dsFile := &metadata.Entry{Role: metadata.DSFile}
				return &metadata.Entry{
					Role: metadata.DSDir,
					PathInfo: metadata.PathInfo{
						Source: "/test/ds_dir",
						IsDir:  true,
					},
					Children: []*metadata.Entry{dsFile},
				}
			},
			expectError: true,
		},
		{
			name: "BTS dir at root - should error",
			entry: func() *metadata.Entry {
				btsFile := &metadata.Entry{Role: metadata.BTSFile}
				return &metadata.Entry{
					Role: metadata.BTSDir,
					PathInfo: metadata.PathInfo{
						Source: "/test/bts_dir",
						IsDir:  true,
					},
					Children: []*metadata.Entry{btsFile},
				}
			},
			expectError: true,
		},
		{
			name: "unknown role - should error",
			entry: func() *metadata.Entry {
				return &metadata.Entry{
					Role: metadata.UnknownRole,
					PathInfo: metadata.PathInfo{
						Source: "/unknown/entry",
						IsDir:  true,
					},
				}
			},
			expectError: true,
		},
		{
			name: "nil entry - should error",
			entry: func() *metadata.Entry {
				return nil
			},
			expectError: true,
		},
	}

	testDir := testutil.CreateTestDir(t)
	cfg := testutil.CreateTestCfg(testDir)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := test.entry()

			err := Enrich(entry, &cfg, logger)

			if test.expectError && err == nil {
				t.Errorf("Enrich expected error, got nil")
			}
			if !test.expectError && err != nil {
				t.Errorf("Enrich unexpected error: %v", err)
			}
		})
	}
}

func TestEnrichMovieFile(t *testing.T) {
	t.Run("propagates title and year from MovieDir to MovieFile", func(t *testing.T) {
		movieFile := testutil.CreateTestMovieFile(nil)
		movieFile.MediaInfo.Title = []string{}
		movieFile.MediaInfo.Year = nil

		movieDir := testutil.CreateTestMovieDir(nil, movieFile)
		movieDir.MediaInfo.Title = []string{"MOVIE", "TITLE"}
		movieDir.MediaInfo.Year = testutil.IntPtr(2024)

		enrichMovieFile(movieDir, metadata.MediaInfo{})

		if len(movieFile.MediaInfo.Title) != 2 || movieFile.MediaInfo.Title[0] != "MOVIE" || movieFile.MediaInfo.Title[1] != "TITLE" {
			t.Errorf("MovieFile Title = %v, want [MOVIE TITLE]", movieFile.MediaInfo.Title)
		}
		if movieFile.MediaInfo.Year == nil || *movieFile.MediaInfo.Year != 2024 {
			t.Errorf("MovieFile Year = %v, want 2024", movieFile.MediaInfo.Year)
		}
	})

	t.Run("preserves MovieFile's own title and year", func(t *testing.T) {
		movieFile := testutil.CreateTestMovieFile(nil)
		movieFile.MediaInfo.Title = []string{"FILE", "TITLE"}
		movieFile.MediaInfo.Year = testutil.IntPtr(2020)

		movieDir := testutil.CreateTestMovieDir(nil, movieFile)
		movieDir.MediaInfo.Title = []string{"DIR", "TITLE"}
		movieDir.MediaInfo.Year = testutil.IntPtr(2024)

		enrichMovieFile(movieDir, metadata.MediaInfo{})

		if len(movieFile.MediaInfo.Title) != 2 || movieFile.MediaInfo.Title[0] != "FILE" || movieFile.MediaInfo.Title[1] != "TITLE" {
			t.Errorf("MovieFile Title = %v, want [FILE TITLE]", movieFile.MediaInfo.Title)
		}
		if movieFile.MediaInfo.Year == nil || *movieFile.MediaInfo.Year != 2020 {
			t.Errorf("MovieFile Year = %v, want 2020", movieFile.MediaInfo.Year)
		}
	})

	t.Run("returns early for SeriesDir", func(t *testing.T) {
		ep := testutil.CreateTestEpFile(nil)
		seasonDir := testutil.CreateTestSeasonDir(nil, ep)
		seriesDir := testutil.CreateTestSeriesDir(nil, seasonDir)
		seriesDir.MediaInfo.Title = []string{"SERIES", "TITLE"}

		enrichMovieFile(seriesDir, metadata.MediaInfo{})

		if len(seriesDir.MediaInfo.Title) != 2 || seriesDir.MediaInfo.Title[0] != "SERIES" {
			t.Errorf("SeriesDir Title = %v, want [SERIES TITLE]", seriesDir.MediaInfo.Title)
		}
	})
}

func TestEnrichEpisodeFiles(t *testing.T) {
	t.Run("propagates season/episode/title/year from SeriesDir to EpisodeFile", func(t *testing.T) {
		ep := testutil.CreateTestEpFile(nil)
		ep.MediaInfo.Season = nil
		ep.MediaInfo.Episode = nil
		ep.MediaInfo.Title = []string{}
		ep.MediaInfo.Year = nil

		// Create intermediate directory with season/episode info
		intermediateDir := &metadata.Entry{
			Role: metadata.UnknownRole,
			PathInfo: metadata.PathInfo{
				Source: "/test/series/season01",
				IsDir:  true,
			},
			MediaInfo: metadata.MediaInfo{
				Season:  testutil.IntPtr(2),
				Episode: []int{5},
			},
			Children: []*metadata.Entry{ep},
		}

		seasonDir := testutil.CreateTestSeasonDir(nil, intermediateDir)
		seasonDir.MediaInfo.Title = []string{"SERIES", "TITLE"}
		seasonDir.MediaInfo.Year = testutil.IntPtr(2023)

		seriesDir := testutil.CreateTestSeriesDir(nil, seasonDir)

		enrichEpisodeFiles(seriesDir, metadata.MediaInfo{})

		if ep.MediaInfo.Season == nil || *ep.MediaInfo.Season != 2 {
			t.Errorf("Episode Season = %v, want 2", ep.MediaInfo.Season)
		}
		if len(ep.MediaInfo.Episode) == 0 || ep.MediaInfo.Episode[0] != 5 {
			t.Errorf("Episode Episode = %v, want [5]", ep.MediaInfo.Episode)
		}
		if len(ep.MediaInfo.Title) != 2 || ep.MediaInfo.Title[0] != "SERIES" {
			t.Errorf("Episode Title = %v, want [SERIES TITLE]", ep.MediaInfo.Title)
		}
		if ep.MediaInfo.Year == nil || *ep.MediaInfo.Year != 2023 {
			t.Errorf("Episode Year = %v, want 2023", ep.MediaInfo.Year)
		}
	})

	t.Run("preserves EpisodeFile's own season and episode", func(t *testing.T) {
		ep := testutil.CreateTestEpFile(nil)
		ep.MediaInfo.Season = testutil.IntPtr(1)
		ep.MediaInfo.Episode = []int{3}

		seasonDir := testutil.CreateTestSeasonDir(nil, ep)
		seasonDir.MediaInfo.Season = testutil.IntPtr(2)
		seasonDir.MediaInfo.Episode = []int{5}

		enrichEpisodeFiles(seasonDir, metadata.MediaInfo{})

		if ep.MediaInfo.Season == nil || *ep.MediaInfo.Season != 1 {
			t.Errorf("Episode Season = %v, want 1", ep.MediaInfo.Season)
		}
		if len(ep.MediaInfo.Episode) == 0 || ep.MediaInfo.Episode[0] != 3 {
			t.Errorf("Episode Episode = %v, want [3]", ep.MediaInfo.Episode)
		}
	})

	t.Run("returns early for MovieDir", func(t *testing.T) {
		movieFile := testutil.CreateTestMovieFile(nil)
		movieDir := testutil.CreateTestMovieDir(nil, movieFile)

		enrichEpisodeFiles(movieDir, metadata.MediaInfo{})

		if len(movieDir.MediaInfo.Title) != 2 || movieDir.MediaInfo.Title[0] != "TEST" {
			t.Errorf("MovieDir Title = %v, want [TEST TITLE]", movieDir.MediaInfo.Title)
		}
	})
}

func TestEnrichSubtitleFiles(t *testing.T) {
	t.Run("propagates season/episode/title/year to SubtitleFile", func(t *testing.T) {
		subFile := testutil.CreateTestSubFile(nil)
		subFile.MediaInfo.Season = nil
		subFile.MediaInfo.Episode = nil
		subFile.MediaInfo.Title = []string{}
		subFile.MediaInfo.Year = nil

		// Create intermediate directory with season/episode info
		intermediateDir := &metadata.Entry{
			Role: metadata.UnknownRole,
			PathInfo: metadata.PathInfo{
				Source: "/test/season/subs",
				IsDir:  true,
			},
			MediaInfo: metadata.MediaInfo{
				Season:  testutil.IntPtr(3),
				Episode: []int{7},
			},
			Children: []*metadata.Entry{subFile},
		}

		seasonDir := testutil.CreateTestSeasonDir(nil, intermediateDir)
		seasonDir.MediaInfo.Title = []string{"SERIES", "TITLE"}
		seasonDir.MediaInfo.Year = testutil.IntPtr(2022)

		enrichSubtitleFiles(seasonDir, metadata.MediaInfo{})

		if subFile.MediaInfo.Season == nil || *subFile.MediaInfo.Season != 3 {
			t.Errorf("Subtitle Season = %v, want 3", subFile.MediaInfo.Season)
		}
		if len(subFile.MediaInfo.Episode) == 0 || subFile.MediaInfo.Episode[0] != 7 {
			t.Errorf("Subtitle Episode = %v, want [7]", subFile.MediaInfo.Episode)
		}
		if len(subFile.MediaInfo.Title) != 2 || subFile.MediaInfo.Title[0] != "SERIES" {
			t.Errorf("Subtitle Title = %v, want [SERIES TITLE]", subFile.MediaInfo.Title)
		}
		if subFile.MediaInfo.Year == nil || *subFile.MediaInfo.Year != 2022 {
			t.Errorf("Subtitle Year = %v, want 2022", subFile.MediaInfo.Year)
		}
	})

	t.Run("preserves SubtitleFile's own season and episode", func(t *testing.T) {
		subFile := testutil.CreateTestSubFile(nil)
		subFile.MediaInfo.Season = testutil.IntPtr(1)
		subFile.MediaInfo.Episode = []int{2}

		seasonDir := testutil.CreateTestSeasonDir(nil, subFile)
		seasonDir.MediaInfo.Season = testutil.IntPtr(5)
		seasonDir.MediaInfo.Episode = []int{10}

		enrichSubtitleFiles(seasonDir, metadata.MediaInfo{})

		if subFile.MediaInfo.Season == nil || *subFile.MediaInfo.Season != 1 {
			t.Errorf("Subtitle Season = %v, want 1", subFile.MediaInfo.Season)
		}
		if len(subFile.MediaInfo.Episode) == 0 || subFile.MediaInfo.Episode[0] != 2 {
			t.Errorf("Subtitle Episode = %v, want [2]", subFile.MediaInfo.Episode)
		}
	})
}

func TestEnrichExtrasFiles(t *testing.T) {
	t.Run("propagates season and episode to BonusFile", func(t *testing.T) {
		bonusFile := testutil.CreateTestBonusFile(nil)
		bonusFile.MediaInfo.Season = nil
		bonusFile.MediaInfo.Episode = nil

		// Create intermediate directory with season/episode info
		intermediateDir := &metadata.Entry{
			Role: metadata.UnknownRole,
			PathInfo: metadata.PathInfo{
				Source: "/test/season/bonus",
				IsDir:  true,
			},
			MediaInfo: metadata.MediaInfo{
				Season:  testutil.IntPtr(4),
				Episode: []int{8},
			},
			Children: []*metadata.Entry{bonusFile},
		}

		seasonDir := testutil.CreateTestSeasonDir(nil, intermediateDir)

		enrichExtrasFiles(seasonDir, metadata.MediaInfo{})

		if bonusFile.MediaInfo.Season == nil || *bonusFile.MediaInfo.Season != 4 {
			t.Errorf("Bonus Season = %v, want 4", bonusFile.MediaInfo.Season)
		}
		if len(bonusFile.MediaInfo.Episode) == 0 || bonusFile.MediaInfo.Episode[0] != 8 {
			t.Errorf("Bonus Episode = %v, want [8]", bonusFile.MediaInfo.Episode)
		}
	})

	t.Run("preserves BonusFile's own season and episode", func(t *testing.T) {
		bonusFile := testutil.CreateTestBonusFile(nil)
		bonusFile.MediaInfo.Season = testutil.IntPtr(1)
		bonusFile.MediaInfo.Episode = []int{2}

		seasonDir := testutil.CreateTestSeasonDir(nil, bonusFile)
		seasonDir.MediaInfo.Season = testutil.IntPtr(5)
		seasonDir.MediaInfo.Episode = []int{10}

		enrichExtrasFiles(seasonDir, metadata.MediaInfo{})

		if bonusFile.MediaInfo.Season == nil || *bonusFile.MediaInfo.Season != 1 {
			t.Errorf("Bonus Season = %v, want 1", bonusFile.MediaInfo.Season)
		}
		if len(bonusFile.MediaInfo.Episode) == 0 || bonusFile.MediaInfo.Episode[0] != 2 {
			t.Errorf("Bonus Episode = %v, want [2]", bonusFile.MediaInfo.Episode)
		}
	})

	t.Run("does not propagate title and year from MovieDir to extras", func(t *testing.T) {
		bonusFile := testutil.CreateTestBonusFile(nil)
		bonusFile.MediaInfo.Title = []string{"BONUS", "TITLE"}

		movieDir := testutil.CreateTestMovieDir(nil, bonusFile)
		movieDir.MediaInfo.Title = []string{"MOVIE", "TITLE"}

		enrichExtrasFiles(movieDir, metadata.MediaInfo{})

		if len(bonusFile.MediaInfo.Title) != 2 || bonusFile.MediaInfo.Title[0] != "BONUS" {
			t.Errorf("Bonus Title = %v, want [BONUS TITLE]", bonusFile.MediaInfo.Title)
		}
	})
}

func TestEnrichEndToEnd(t *testing.T) {
	testDir := testutil.CreateTestDir(t)
	cfg := testutil.CreateTestCfg(testDir)

	t.Run("movie directory enrichment - dir to file propagation", func(t *testing.T) {
		movieFile := testutil.CreateTestMovieFile(nil)
		movieFile.MediaInfo.Title = []string{}
		movieFile.MediaInfo.Year = nil

		bonusFile := testutil.CreateTestBonusFile(nil)
		bonusFile.MediaInfo.Title = []string{}
		bonusFile.MediaInfo.Year = nil

		subFile := testutil.CreateTestSubFile(nil)
		subFile.MediaInfo.Title = []string{}
		subFile.MediaInfo.Year = nil

		movieDir := testutil.CreateTestMovieDir(nil, movieFile, bonusFile, subFile)
		movieDir.MediaInfo.Title = []string{"MOVIE", "TITLE"}
		movieDir.MediaInfo.Year = testutil.IntPtr(2024)

		err := Enrich(movieDir, &cfg, logger)

		if err != nil {
			t.Errorf("Enrich unexpected error: %v", err)
		}

		expectedTitle := []string{"MOVIE", "TITLE"}
		expectedYear := 2024

		if len(movieFile.MediaInfo.Title) != 2 || movieFile.MediaInfo.Title[0] != "MOVIE" {
			t.Errorf("MovieFile Title = %v, want %v", movieFile.MediaInfo.Title, expectedTitle)
		}
		if movieFile.MediaInfo.Year == nil || *movieFile.MediaInfo.Year != expectedYear {
			t.Errorf("MovieFile Year = %v, want %v", movieFile.MediaInfo.Year, expectedYear)
		}
		if len(subFile.MediaInfo.Title) != 2 || subFile.MediaInfo.Title[0] != "MOVIE" {
			t.Errorf("SubFile Title = %v, want %v", subFile.MediaInfo.Title, expectedTitle)
		}
		if subFile.MediaInfo.Year == nil || *subFile.MediaInfo.Year != expectedYear {
			t.Errorf("SubFile Year = %v, want %v", subFile.MediaInfo.Year, expectedYear)
		}
	})

	t.Run("series directory enrichment - propagates to episodes", func(t *testing.T) {
		ep1 := testutil.CreateTestEpFile(nil)
		ep1.MediaInfo.Season = nil
		ep1.MediaInfo.Episode = nil
		ep1.MediaInfo.Title = []string{}
		ep1.MediaInfo.Year = nil

		ep2 := testutil.CreateTestEpFile(nil)
		ep2.MediaInfo.Season = nil
		ep2.MediaInfo.Episode = nil
		ep2.MediaInfo.Title = []string{}
		ep2.MediaInfo.Year = nil

		// Create intermediate directory with season info
		intermediateDir := &metadata.Entry{
			Role: metadata.UnknownRole,
			PathInfo: metadata.PathInfo{
				Source: "/test/series/season01",
				IsDir:  true,
			},
			MediaInfo: metadata.MediaInfo{
				Season: testutil.IntPtr(2),
			},
			Children: []*metadata.Entry{ep1, ep2},
		}

		seasonDir := testutil.CreateTestSeasonDir(nil, intermediateDir)
		seasonDir.MediaInfo.Title = []string{"SHOW", "TITLE"}
		seasonDir.MediaInfo.Year = testutil.IntPtr(2023)

		seriesDir := testutil.CreateTestSeriesDir(nil, seasonDir)

		err := Enrich(seriesDir, &cfg, logger)

		if err != nil {
			t.Errorf("Enrich unexpected error: %v", err)
		}

		expectedTitle := []string{"SHOW", "TITLE"}
		expectedYear := 2023
		expectedSeason := 2

		if len(ep1.MediaInfo.Title) != 2 || ep1.MediaInfo.Title[0] != "SHOW" {
			t.Errorf("Episode1 Title = %v, want %v", ep1.MediaInfo.Title, expectedTitle)
		}
		if ep1.MediaInfo.Year == nil || *ep1.MediaInfo.Year != expectedYear {
			t.Errorf("Episode1 Year = %v, want %v", ep1.MediaInfo.Year, expectedYear)
		}
		if ep1.MediaInfo.Season == nil || *ep1.MediaInfo.Season != expectedSeason {
			t.Errorf("Episode1 Season = %v, want %v", ep1.MediaInfo.Season, expectedSeason)
		}

		if len(ep2.MediaInfo.Title) != 2 || ep2.MediaInfo.Title[0] != "SHOW" {
			t.Errorf("Episode2 Title = %v, want %v", ep2.MediaInfo.Title, expectedTitle)
		}
		if ep2.MediaInfo.Year == nil || *ep2.MediaInfo.Year != expectedYear {
			t.Errorf("Episode2 Year = %v, want %v", ep2.MediaInfo.Year, expectedYear)
		}
		if ep2.MediaInfo.Season == nil || *ep2.MediaInfo.Season != expectedSeason {
			t.Errorf("Episode2 Season = %v, want %v", ep2.MediaInfo.Season, expectedSeason)
		}
	})

	t.Run("preserves existing metadata on files", func(t *testing.T) {
		movieFile := testutil.CreateTestMovieFile(nil)
		movieFile.MediaInfo.Title = []string{"FILE", "TITLE"}
		movieFile.MediaInfo.Year = testutil.IntPtr(2020)

		movieDir := testutil.CreateTestMovieDir(nil, movieFile)
		movieDir.MediaInfo.Title = []string{"DIR", "TITLE"}
		movieDir.MediaInfo.Year = testutil.IntPtr(2024)

		err := Enrich(movieDir, &cfg, logger)

		if err != nil {
			t.Errorf("Enrich unexpected error: %v", err)
		}

		if len(movieFile.MediaInfo.Title) != 2 || movieFile.MediaInfo.Title[0] != "FILE" {
			t.Errorf("MovieFile Title = %v, want [FILE TITLE]", movieFile.MediaInfo.Title)
		}
		if movieFile.MediaInfo.Year == nil || *movieFile.MediaInfo.Year != 2020 {
			t.Errorf("MovieFile Year = %v, want 2020", movieFile.MediaInfo.Year)
		}
	})
}
