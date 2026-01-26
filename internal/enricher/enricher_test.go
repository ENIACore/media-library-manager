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

func TestGetShowTitle(t *testing.T) {
	tests := []struct {
		name     string
		entry    func() *metadata.Entry
		expected []string
	}{
		{
			name: "episode file with title",
			entry: func() *metadata.Entry {
				return testutil.CreateTestEpFile(nil)
			},
			expected: []string{"TEST", "TITLE"},
		},
		{
			name: "season dir with episode children",
			entry: func() *metadata.Entry {
				ep1 := testutil.CreateTestEpFile(nil)
				ep2 := testutil.CreateTestEpFile(nil)
				seasonDir := testutil.CreateTestSeasonDir(nil, ep1, ep2)
				seasonDir.MediaInfo.Title = []string{}
				return seasonDir
			},
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
			expected: []string{"TEST", "TITLE"},
		},
		{
			name: "season dir with own title",
			entry: func() *metadata.Entry {
				ep := testutil.CreateTestEpFile(nil)
				ep.MediaInfo.Title = []string{"EPISODE", "TITLE"}
				seasonDir := testutil.CreateTestSeasonDir(nil, ep)
				seasonDir.MediaInfo.Title = []string{"SEASON", "TITLE"}
				return seasonDir
			},
			expected: []string{"SEASON", "TITLE"},
		},
		{
			name: "series dir with own title",
			entry: func() *metadata.Entry {
				ep := testutil.CreateTestEpFile(nil)
				seasonDir := testutil.CreateTestSeasonDir(nil, ep)
				seriesDir := testutil.CreateTestSeriesDir(nil, seasonDir)
				seriesDir.MediaInfo.Title = []string{"SERIES", "TITLE"}
				return seriesDir
			},
			expected: []string{"SERIES", "TITLE"},
		},
		{
			name: "nil entry",
			entry: func() *metadata.Entry {
				return nil
			},
			expected: nil,
		},
		{
			name: "entry with empty title",
			entry: func() *metadata.Entry {
				ep := testutil.CreateTestEpFile(nil)
				ep.MediaInfo.Title = []string{}
				return ep
			},
			expected: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := test.entry()

			result := getShowTitle(entry)

			if len(result) != len(test.expected) {
				t.Errorf("getShowTitle = %v, want %v", result, test.expected)
				return
			}
			for i := range result {
				if result[i] != test.expected[i] {
					t.Errorf("getShowTitle = %v, want %v", result, test.expected)
					return
				}
			}
		})
	}
}

func TestGetShowYear(t *testing.T) {
	tests := []struct {
		name     string
		entry    func() *metadata.Entry
		expected *int
	}{
		{
			name: "episode file with year",
			entry: func() *metadata.Entry {
				return testutil.CreateTestEpFile(nil)
			},
			expected: testutil.IntPtr(2025),
		},
		{
			name: "season dir with episode children",
			entry: func() *metadata.Entry {
				ep1 := testutil.CreateTestEpFile(nil)
				ep2 := testutil.CreateTestEpFile(nil)
				seasonDir := testutil.CreateTestSeasonDir(nil, ep1, ep2)
				seasonDir.MediaInfo.Year = nil
				return seasonDir
			},
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
			expected: testutil.IntPtr(2025),
		},
		{
			name: "season dir with own year",
			entry: func() *metadata.Entry {
				ep := testutil.CreateTestEpFile(nil)
				ep.MediaInfo.Year = testutil.IntPtr(2020)
				seasonDir := testutil.CreateTestSeasonDir(nil, ep)
				seasonDir.MediaInfo.Year = testutil.IntPtr(2021)
				return seasonDir
			},
			expected: testutil.IntPtr(2021),
		},
		{
			name: "series dir with own year",
			entry: func() *metadata.Entry {
				ep := testutil.CreateTestEpFile(nil)
				seasonDir := testutil.CreateTestSeasonDir(nil, ep)
				seriesDir := testutil.CreateTestSeriesDir(nil, seasonDir)
				seriesDir.MediaInfo.Year = testutil.IntPtr(2022)
				return seriesDir
			},
			expected: testutil.IntPtr(2022),
		},
		{
			name: "nil entry",
			entry: func() *metadata.Entry {
				return nil
			},
			expected: nil,
		},
		{
			name: "entry with nil year",
			entry: func() *metadata.Entry {
				ep := testutil.CreateTestEpFile(nil)
				ep.MediaInfo.Year = nil
				return ep
			},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := test.entry()

			result := getShowYear(entry)

			if test.expected == nil && result != nil {
				t.Errorf("getShowYear = %v, want nil", *result)
			} else if test.expected != nil && result == nil {
				t.Errorf("getShowYear = nil, want %v", *test.expected)
			} else if test.expected != nil && result != nil && *result != *test.expected {
				t.Errorf("getShowYear = %v, want %v", *result, *test.expected)
			}
		})
	}
}

func TestGetMovieTitle(t *testing.T) {
	tests := []struct {
		name     string
		entry    func() *metadata.Entry
		expected []string
	}{
		{
			name: "movie file with title",
			entry: func() *metadata.Entry {
				return testutil.CreateTestMovieFile(nil)
			},
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
			expected: []string{"TEST", "TITLE"},
		},
		{
			name: "movie dir with own title",
			entry: func() *metadata.Entry {
				movieFile := testutil.CreateTestMovieFile(nil)
				movieFile.MediaInfo.Title = []string{"CHILD", "TITLE"}
				movieDir := testutil.CreateTestMovieDir(nil, movieFile)
				movieDir.MediaInfo.Title = []string{"PARENT", "TITLE"}
				return movieDir
			},
			expected: []string{"PARENT", "TITLE"},
		},
		{
			name: "nil entry",
			entry: func() *metadata.Entry {
				return nil
			},
			expected: nil,
		},
		{
			name: "entry with empty title",
			entry: func() *metadata.Entry {
				movie := testutil.CreateTestMovieFile(nil)
				movie.MediaInfo.Title = []string{}
				return movie
			},
			expected: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := test.entry()

			result := getMovieTitle(entry)

			if len(result) != len(test.expected) {
				t.Errorf("getMovieTitle = %v, want %v", result, test.expected)
				return
			}
			for i := range result {
				if result[i] != test.expected[i] {
					t.Errorf("getMovieTitle = %v, want %v", result, test.expected)
					return
				}
			}
		})
	}
}

func TestGetMovieYear(t *testing.T) {
	tests := []struct {
		name     string
		entry    func() *metadata.Entry
		expected *int
	}{
		{
			name: "movie file with year",
			entry: func() *metadata.Entry {
				return testutil.CreateTestMovieFile(nil)
			},
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
			expected: testutil.IntPtr(2025),
		},
		{
			name: "movie dir with own year",
			entry: func() *metadata.Entry {
				movieFile := testutil.CreateTestMovieFile(nil)
				movieFile.MediaInfo.Year = testutil.IntPtr(2020)
				movieDir := testutil.CreateTestMovieDir(nil, movieFile)
				movieDir.MediaInfo.Year = testutil.IntPtr(2021)
				return movieDir
			},
			expected: testutil.IntPtr(2021),
		},
		{
			name: "nil entry",
			entry: func() *metadata.Entry {
				return nil
			},
			expected: nil,
		},
		{
			name: "entry with nil year",
			entry: func() *metadata.Entry {
				movie := testutil.CreateTestMovieFile(nil)
				movie.MediaInfo.Year = nil
				return movie
			},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := test.entry()

			result := getMovieYear(entry)

			if test.expected == nil && result != nil {
				t.Errorf("getMovieYear = %v, want nil", *result)
			} else if test.expected != nil && result == nil {
				t.Errorf("getMovieYear = nil, want %v", *test.expected)
			} else if test.expected != nil && result != nil && *result != *test.expected {
				t.Errorf("getMovieYear = %v, want %v", *result, *test.expected)
			}
		})
	}
}

func TestSetEntryValues(t *testing.T) {
	t.Run("set values on single entry", func(t *testing.T) {
		entry := testutil.CreateTestMovieFile(nil)
		entry.MediaInfo.Title = []string{"OLD", "TITLE"}
		entry.MediaInfo.Year = testutil.IntPtr(2020)

		newTitle := []string{"NEW", "TITLE"}
		newYear := testutil.IntPtr(2025)

		setEntryValues(entry, newTitle, newYear)

		if len(entry.MediaInfo.Title) != len(newTitle) {
			t.Errorf("Title = %v, want %v", entry.MediaInfo.Title, newTitle)
		}
		for i := range entry.MediaInfo.Title {
			if entry.MediaInfo.Title[i] != newTitle[i] {
				t.Errorf("Title = %v, want %v", entry.MediaInfo.Title, newTitle)
			}
		}
		if *entry.MediaInfo.Year != *newYear {
			t.Errorf("Year = %v, want %v", *entry.MediaInfo.Year, *newYear)
		}
	})

	t.Run("set values on entry with children", func(t *testing.T) {
		movieFile := testutil.CreateTestMovieFile(nil)
		bonusFile := testutil.CreateTestBonusFile(nil)
		movieDir := testutil.CreateTestMovieDir(nil, movieFile, bonusFile)

		movieFile.MediaInfo.Title = []string{"OLD", "TITLE"}
		bonusFile.MediaInfo.Title = []string{"OLD", "TITLE"}

		newTitle := []string{"NEW", "TITLE"}
		newYear := testutil.IntPtr(2025)

		setEntryValues(movieDir, newTitle, newYear)

		if len(movieDir.MediaInfo.Title) != len(newTitle) {
			t.Errorf("Parent Title = %v, want %v", movieDir.MediaInfo.Title, newTitle)
		}
		if len(movieFile.MediaInfo.Title) != len(newTitle) {
			t.Errorf("Child 1 Title = %v, want %v", movieFile.MediaInfo.Title, newTitle)
		}
		if len(bonusFile.MediaInfo.Title) != len(newTitle) {
			t.Errorf("Child 2 Title = %v, want %v", bonusFile.MediaInfo.Title, newTitle)
		}
	})

	t.Run("set values on nested structure", func(t *testing.T) {
		ep := testutil.CreateTestEpFile(nil)
		seasonDir := testutil.CreateTestSeasonDir(nil, ep)
		seriesDir := testutil.CreateTestSeriesDir(nil, seasonDir)

		newTitle := []string{"ENRICHED", "TITLE"}
		newYear := testutil.IntPtr(2026)

		setEntryValues(seriesDir, newTitle, newYear)

		if len(seriesDir.MediaInfo.Title) != len(newTitle) {
			t.Errorf("Series Title = %v, want %v", seriesDir.MediaInfo.Title, newTitle)
		}
		if len(seasonDir.MediaInfo.Title) != len(newTitle) {
			t.Errorf("Season Title = %v, want %v", seasonDir.MediaInfo.Title, newTitle)
		}
		if len(ep.MediaInfo.Title) != len(newTitle) {
			t.Errorf("Episode Title = %v, want %v", ep.MediaInfo.Title, newTitle)
		}
		if *ep.MediaInfo.Year != *newYear {
			t.Errorf("Episode Year = %v, want %v", *ep.MediaInfo.Year, *newYear)
		}
	})

	t.Run("set nil values", func(t *testing.T) {
		entry := testutil.CreateTestMovieFile(nil)

		setEntryValues(entry, nil, nil)

		if entry.MediaInfo.Title != nil {
			t.Errorf("Title = %v, want nil", entry.MediaInfo.Title)
		}
		if entry.MediaInfo.Year != nil {
			t.Errorf("Year = %v, want nil", entry.MediaInfo.Year)
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
