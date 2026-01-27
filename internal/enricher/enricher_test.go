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
			name: "subtitle file at root - should error",
			entry: func() *metadata.Entry {
				return testutil.CreateTestSubFile(nil)
			},
			expectError: true,
		},
		{
			name: "bonus file at root - should error",
			entry: func() *metadata.Entry {
				return testutil.CreateTestBonusFile(nil)
			},
			expectError: true,
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
			name: "unknown role - should error",
			entry: func() *metadata.Entry {
				return &metadata.Entry{
					Role: metadata.UnknownRole,
					PathInfo: metadata.PathInfo{
						Source: "/unknown/entry",
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

func TestGetTitle(t *testing.T) {
	tests := []struct {
		name     string
		entry    func() *metadata.Entry
		roles    []metadata.EntryRole
		expected []string
	}{
		{
			name: "movie file with title",
			entry: func() *metadata.Entry {
				return testutil.CreateTestMovieFile(nil)
			},
			roles:    []metadata.EntryRole{metadata.MovieFile},
			expected: []string{"TEST", "TITLE"},
		},
		{
			name: "episode file with title",
			entry: func() *metadata.Entry {
				return testutil.CreateTestEpFile(nil)
			},
			roles:    []metadata.EntryRole{metadata.EpisodeFile},
			expected: []string{"TEST", "TITLE"},
		},
		{
			name: "movie dir with movie file child",
			entry: func() *metadata.Entry {
				movieFile := testutil.CreateTestMovieFile(nil)
				movieDir := testutil.CreateTestMovieDir(nil, movieFile)
				movieDir.MediaInfo.Title = []string{}
				return movieDir
			},
			roles:    []metadata.EntryRole{metadata.MovieDir, metadata.MovieFile},
			expected: []string{"TEST", "TITLE"},
		},
		{
			name: "series dir with season and episode children",
			entry: func() *metadata.Entry {
				ep := testutil.CreateTestEpFile(nil)
				seasonDir := testutil.CreateTestSeasonDir(nil, ep)
				seriesDir := testutil.CreateTestSeriesDir(nil, seasonDir)
				seriesDir.MediaInfo.Title = []string{}
				seasonDir.MediaInfo.Title = []string{}
				return seriesDir
			},
			roles:    []metadata.EntryRole{metadata.SeriesDir, metadata.SeasonDir, metadata.EpisodeFile},
			expected: []string{"TEST", "TITLE"},
		},
		{
			name: "parent takes precedence over child",
			entry: func() *metadata.Entry {
				movieFile := testutil.CreateTestMovieFile(nil)
				movieFile.MediaInfo.Title = []string{"CHILD", "TITLE"}
				movieDir := testutil.CreateTestMovieDir(nil, movieFile)
				movieDir.MediaInfo.Title = []string{"PARENT", "TITLE"}
				return movieDir
			},
			roles:    []metadata.EntryRole{metadata.MovieDir, metadata.MovieFile},
			expected: []string{"PARENT", "TITLE"},
		},
		{
			name: "nil entry",
			entry: func() *metadata.Entry {
				return nil
			},
			roles:    []metadata.EntryRole{metadata.MovieFile},
			expected: nil,
		},
		{
			name: "entry with empty title",
			entry: func() *metadata.Entry {
				entry := testutil.CreateTestMovieFile(nil)
				entry.MediaInfo.Title = []string{}
				return entry
			},
			roles:    []metadata.EntryRole{metadata.MovieFile},
			expected: nil,
		},
		{
			name: "no matching roles",
			entry: func() *metadata.Entry {
				return testutil.CreateTestMovieFile(nil)
			},
			roles:    []metadata.EntryRole{metadata.EpisodeFile},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := test.entry()

			result := getTitle(entry, test.roles)

			if len(result) != len(test.expected) {
				t.Errorf("getTitle = %v, want %v", result, test.expected)
				return
			}
			for i := range result {
				if result[i] != test.expected[i] {
					t.Errorf("getTitle = %v, want %v", result, test.expected)
					return
				}
			}
		})
	}
}

func TestGetYear(t *testing.T) {
	tests := []struct {
		name     string
		entry    func() *metadata.Entry
		roles    []metadata.EntryRole
		expected *int
	}{
		{
			name: "movie file with year",
			entry: func() *metadata.Entry {
				return testutil.CreateTestMovieFile(nil)
			},
			roles:    []metadata.EntryRole{metadata.MovieFile},
			expected: testutil.IntPtr(2025),
		},
		{
			name: "episode file with year",
			entry: func() *metadata.Entry {
				return testutil.CreateTestEpFile(nil)
			},
			roles:    []metadata.EntryRole{metadata.EpisodeFile},
			expected: testutil.IntPtr(2025),
		},
		{
			name: "movie dir with movie file child",
			entry: func() *metadata.Entry {
				movieFile := testutil.CreateTestMovieFile(nil)
				movieDir := testutil.CreateTestMovieDir(nil, movieFile)
				movieDir.MediaInfo.Year = nil
				return movieDir
			},
			roles:    []metadata.EntryRole{metadata.MovieDir, metadata.MovieFile},
			expected: testutil.IntPtr(2025),
		},
		{
			name: "series dir with season and episode children",
			entry: func() *metadata.Entry {
				ep := testutil.CreateTestEpFile(nil)
				seasonDir := testutil.CreateTestSeasonDir(nil, ep)
				seriesDir := testutil.CreateTestSeriesDir(nil, seasonDir)
				seriesDir.MediaInfo.Year = nil
				seasonDir.MediaInfo.Year = nil
				return seriesDir
			},
			roles:    []metadata.EntryRole{metadata.SeriesDir, metadata.SeasonDir, metadata.EpisodeFile},
			expected: testutil.IntPtr(2025),
		},
		{
			name: "parent takes precedence over child",
			entry: func() *metadata.Entry {
				movieFile := testutil.CreateTestMovieFile(nil)
				movieFile.MediaInfo.Year = testutil.IntPtr(2020)
				movieDir := testutil.CreateTestMovieDir(nil, movieFile)
				movieDir.MediaInfo.Year = testutil.IntPtr(2021)
				return movieDir
			},
			roles:    []metadata.EntryRole{metadata.MovieDir, metadata.MovieFile},
			expected: testutil.IntPtr(2021),
		},
		{
			name: "nil entry",
			entry: func() *metadata.Entry {
				return nil
			},
			roles:    []metadata.EntryRole{metadata.MovieFile},
			expected: nil,
		},
		{
			name: "entry with nil year",
			entry: func() *metadata.Entry {
				entry := testutil.CreateTestMovieFile(nil)
				entry.MediaInfo.Year = nil
				return entry
			},
			roles:    []metadata.EntryRole{metadata.MovieFile},
			expected: nil,
		},
		{
			name: "no matching roles",
			entry: func() *metadata.Entry {
				return testutil.CreateTestMovieFile(nil)
			},
			roles:    []metadata.EntryRole{metadata.EpisodeFile},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := test.entry()

			result := getYear(entry, test.roles)

			if test.expected == nil && result != nil {
				t.Errorf("getYear = %v, want nil", *result)
			} else if test.expected != nil && result == nil {
				t.Errorf("getYear = nil, want %v", *test.expected)
			} else if test.expected != nil && result != nil && *result != *test.expected {
				t.Errorf("getYear = %v, want %v", *result, *test.expected)
			}
		})
	}
}

func TestPropogateDown(t *testing.T) {
	t.Run("propogate to single entry", func(t *testing.T) {
		entry := testutil.CreateTestMovieFile(nil)
		entry.MediaInfo.Title = []string{"OLD", "TITLE"}
		entry.MediaInfo.Year = testutil.IntPtr(2020)

		src := &metadata.MediaInfo{
			Title: []string{"NEW", "TITLE"},
			Year:  testutil.IntPtr(2025),
		}

		propogateDown(entry, src, []metadata.EntryRole{metadata.MovieFile})

		if len(entry.MediaInfo.Title) != len(src.Title) {
			t.Errorf("Title = %v, want %v", entry.MediaInfo.Title, src.Title)
		}
		for i := range entry.MediaInfo.Title {
			if entry.MediaInfo.Title[i] != src.Title[i] {
				t.Errorf("Title = %v, want %v", entry.MediaInfo.Title, src.Title)
			}
		}
		if *entry.MediaInfo.Year != *src.Year {
			t.Errorf("Year = %v, want %v", *entry.MediaInfo.Year, *src.Year)
		}
	})

	t.Run("propogate to entry with children", func(t *testing.T) {
		movieFile := testutil.CreateTestMovieFile(nil)
		bonusFile := testutil.CreateTestBonusFile(nil)
		movieDir := testutil.CreateTestMovieDir(nil, movieFile, bonusFile)

		movieFile.MediaInfo.Title = []string{"OLD", "TITLE"}
		bonusFile.MediaInfo.Title = []string{"OLD", "TITLE"}

		src := &metadata.MediaInfo{
			Title: []string{"NEW", "TITLE"},
			Year:  testutil.IntPtr(2025),
		}

		propogateDown(movieDir, src, []metadata.EntryRole{metadata.MovieDir, metadata.MovieFile, metadata.BonusFile})

		if len(movieDir.MediaInfo.Title) != len(src.Title) {
			t.Errorf("Parent Title = %v, want %v", movieDir.MediaInfo.Title, src.Title)
		}
		if len(movieFile.MediaInfo.Title) != len(src.Title) {
			t.Errorf("Child 1 Title = %v, want %v", movieFile.MediaInfo.Title, src.Title)
		}
		if len(bonusFile.MediaInfo.Title) != len(src.Title) {
			t.Errorf("Child 2 Title = %v, want %v", bonusFile.MediaInfo.Title, src.Title)
		}
	})

	t.Run("propogate to nested structure", func(t *testing.T) {
		ep := testutil.CreateTestEpFile(nil)
		seasonDir := testutil.CreateTestSeasonDir(nil, ep)
		seriesDir := testutil.CreateTestSeriesDir(nil, seasonDir)

		src := &metadata.MediaInfo{
			Title: []string{"ENRICHED", "TITLE"},
			Year:  testutil.IntPtr(2026),
		}

		propogateDown(seriesDir, src, []metadata.EntryRole{metadata.SeriesDir, metadata.SeasonDir, metadata.EpisodeFile})

		if len(seriesDir.MediaInfo.Title) != len(src.Title) {
			t.Errorf("Series Title = %v, want %v", seriesDir.MediaInfo.Title, src.Title)
		}
		if len(seasonDir.MediaInfo.Title) != len(src.Title) {
			t.Errorf("Season Title = %v, want %v", seasonDir.MediaInfo.Title, src.Title)
		}
		if len(ep.MediaInfo.Title) != len(src.Title) {
			t.Errorf("Episode Title = %v, want %v", ep.MediaInfo.Title, src.Title)
		}
		if *ep.MediaInfo.Year != *src.Year {
			t.Errorf("Episode Year = %v, want %v", *ep.MediaInfo.Year, *src.Year)
		}
	})

	t.Run("only propogate to matching roles", func(t *testing.T) {
		movieFile := testutil.CreateTestMovieFile(nil)
		bonusFile := testutil.CreateTestBonusFile(nil)
		movieDir := testutil.CreateTestMovieDir(nil, movieFile, bonusFile)

		originalBonusTitle := append([]string(nil), bonusFile.MediaInfo.Title...)

		src := &metadata.MediaInfo{
			Title: []string{"NEW", "TITLE"},
			Year:  testutil.IntPtr(2025),
		}

		propogateDown(movieDir, src, []metadata.EntryRole{metadata.MovieDir, metadata.MovieFile})

		if len(movieFile.MediaInfo.Title) != len(src.Title) {
			t.Errorf("Movie Title = %v, want %v", movieFile.MediaInfo.Title, src.Title)
		}
		// Bonus file should not be changed
		if len(bonusFile.MediaInfo.Title) != len(originalBonusTitle) {
			t.Errorf("Bonus Title changed = %v, want unchanged %v", bonusFile.MediaInfo.Title, originalBonusTitle)
		}
	})

	t.Run("propogate with nil roles propogates to all", func(t *testing.T) {
		movieFile := testutil.CreateTestMovieFile(nil)
		bonusFile := testutil.CreateTestBonusFile(nil)
		movieDir := testutil.CreateTestMovieDir(nil, movieFile, bonusFile)

		src := &metadata.MediaInfo{
			Title: []string{"NEW", "TITLE"},
			Year:  testutil.IntPtr(2025),
		}

		propogateDown(movieDir, src, nil)

		if len(movieDir.MediaInfo.Title) != len(src.Title) {
			t.Errorf("Dir Title = %v, want %v", movieDir.MediaInfo.Title, src.Title)
		}
		if len(movieFile.MediaInfo.Title) != len(src.Title) {
			t.Errorf("Movie Title = %v, want %v", movieFile.MediaInfo.Title, src.Title)
		}
		if len(bonusFile.MediaInfo.Title) != len(src.Title) {
			t.Errorf("Bonus Title = %v, want %v", bonusFile.MediaInfo.Title, src.Title)
		}
	})
}

func TestEnrichIntermediaryEntries(t *testing.T) {
	t.Run("enrich bonus dir at height 2", func(t *testing.T) {
		bonus := testutil.CreateTestBonusFile(nil)
		bonus.MediaInfo.Season = testutil.IntPtr(5)
		bonus.MediaInfo.Episode = testutil.IntPtr(10)
		bonusDir := testutil.CreateTestBonusDir(nil, bonus)

		enrichIntermediaryEntries(bonusDir)

		if bonus.MediaInfo.Season == nil || *bonus.MediaInfo.Season != 5 {
			t.Errorf("Bonus Season = %v, want 5", bonus.MediaInfo.Season)
		}
		if bonus.MediaInfo.Episode == nil || *bonus.MediaInfo.Episode != 10 {
			t.Errorf("Bonus Episode = %v, want 10", bonus.MediaInfo.Episode)
		}
	})

	t.Run("enrich subtitle dir at height 2", func(t *testing.T) {
		sub := testutil.CreateTestSubFile(nil)
		sub.MediaInfo.Season = testutil.IntPtr(3)
		sub.MediaInfo.Episode = testutil.IntPtr(7)
		subDir := testutil.CreateTestSubDir(nil, sub)

		enrichIntermediaryEntries(subDir)

		if sub.MediaInfo.Season == nil || *sub.MediaInfo.Season != 3 {
			t.Errorf("Sub Season = %v, want 3", sub.MediaInfo.Season)
		}
		if sub.MediaInfo.Episode == nil || *sub.MediaInfo.Episode != 7 {
			t.Errorf("Sub Episode = %v, want 7", sub.MediaInfo.Episode)
		}
	})

	t.Run("does not enrich at height != 2", func(t *testing.T) {
		bonus := testutil.CreateTestBonusFile(nil)
		bonus.MediaInfo.Season = nil
		bonus.MediaInfo.Episode = nil

		enrichIntermediaryEntries(bonus)

		if bonus.MediaInfo.Season != nil {
			t.Errorf("Bonus Season should remain nil")
		}
		if bonus.MediaInfo.Episode != nil {
			t.Errorf("Bonus Episode should remain nil")
		}
	})
}

func TestEnrichEndToEnd(t *testing.T) {
	testDir := testutil.CreateTestDir(t)
	cfg := testutil.CreateTestCfg(testDir)

	t.Run("movie directory enrichment", func(t *testing.T) {
		movieFile := testutil.CreateTestMovieFile(nil)
		bonusFile := testutil.CreateTestBonusFile(nil)
		subFile := testutil.CreateTestSubFile(nil)
		movieDir := testutil.CreateTestMovieDir(nil, movieFile, bonusFile, subFile)

		movieFile.MediaInfo.Title = []string{"MOVIE", "TITLE"}
		movieFile.MediaInfo.Year = testutil.IntPtr(2024)
		bonusFile.MediaInfo.Title = []string{}
		bonusFile.MediaInfo.Year = nil
		subFile.MediaInfo.Title = []string{}
		subFile.MediaInfo.Year = nil
		movieDir.MediaInfo.Title = []string{}
		movieDir.MediaInfo.Year = nil

		err := Enrich(movieDir, &cfg, logger)

		if err != nil {
			t.Errorf("Enrich unexpected error: %v", err)
		}

		expectedTitle := []string{"MOVIE", "TITLE"}
		expectedYear := testutil.IntPtr(2024)

		if len(movieDir.MediaInfo.Title) != len(expectedTitle) {
			t.Errorf("MovieDir Title = %v, want %v", movieDir.MediaInfo.Title, expectedTitle)
		}
		if *movieDir.MediaInfo.Year != *expectedYear {
			t.Errorf("MovieDir Year = %v, want %v", *movieDir.MediaInfo.Year, *expectedYear)
		}
		if len(bonusFile.MediaInfo.Title) != len(expectedTitle) {
			t.Errorf("Bonus Title = %v, want %v", bonusFile.MediaInfo.Title, expectedTitle)
		}
		if len(subFile.MediaInfo.Title) != len(expectedTitle) {
			t.Errorf("Sub Title = %v, want %v", subFile.MediaInfo.Title, expectedTitle)
		}
	})

	t.Run("series directory enrichment", func(t *testing.T) {
		ep1 := testutil.CreateTestEpFile(nil)
		ep2 := testutil.CreateTestEpFile(nil)
		ep1.MediaInfo.Title = []string{"SHOW", "TITLE"}
		ep1.MediaInfo.Year = testutil.IntPtr(2023)
		ep2.MediaInfo.Title = []string{"SHOW", "TITLE"}
		ep2.MediaInfo.Year = testutil.IntPtr(2023)

		seasonDir := testutil.CreateTestSeasonDir(nil, ep1, ep2)
		seasonDir.MediaInfo.Title = []string{}
		seasonDir.MediaInfo.Year = nil

		subFile := testutil.CreateTestSubFile(nil)
		subFile.MediaInfo.Title = []string{}
		subFile.MediaInfo.Year = nil

		seriesDir := testutil.CreateTestSeriesDir(nil, seasonDir)
		seriesDir.MediaInfo.Title = []string{}
		seriesDir.MediaInfo.Year = nil

		err := Enrich(seriesDir, &cfg, logger)

		if err != nil {
			t.Errorf("Enrich unexpected error: %v", err)
		}

		expectedTitle := []string{"SHOW", "TITLE"}
		expectedYear := testutil.IntPtr(2023)

		if len(seriesDir.MediaInfo.Title) != len(expectedTitle) {
			t.Errorf("SeriesDir Title = %v, want %v", seriesDir.MediaInfo.Title, expectedTitle)
		}
		if *seriesDir.MediaInfo.Year != *expectedYear {
			t.Errorf("SeriesDir Year = %v, want %v", *seriesDir.MediaInfo.Year, *expectedYear)
		}
		if len(seasonDir.MediaInfo.Title) != len(expectedTitle) {
			t.Errorf("SeasonDir Title = %v, want %v", seasonDir.MediaInfo.Title, expectedTitle)
		}
		if *seasonDir.MediaInfo.Year != *expectedYear {
			t.Errorf("SeasonDir Year = %v, want %v", *seasonDir.MediaInfo.Year, *expectedYear)
		}
	})
}
