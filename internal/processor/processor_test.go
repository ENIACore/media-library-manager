package processor

import (
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/ENIACore/media_library_manager/internal/testutil"
)

var logger = slog.Default()

func TestResolve(t *testing.T) {
	/*
		Testing error cases for entries that cannot be alone at root
	*/
	t.Run("error cases for root level entries", func(t *testing.T) {
		tests := []struct {
			name  string
			entry func() *metadata.Entry
		}{
			{
				name: "subtitle file at root",
				entry: func() *metadata.Entry {
					return testutil.CreateTestSubFile(nil)
				},
			},
			{
				name: "bonus file at root",
				entry: func() *metadata.Entry {
					return testutil.CreateTestBonusFile(nil)
				},
			},
			{
				name: "subtitle dir at root",
				entry: func() *metadata.Entry {
					sub := testutil.CreateTestSubFile(nil)
					return testutil.CreateTestSubDir(nil, sub)
				},
			},
			{
				name: "bonus dir at root",
				entry: func() *metadata.Entry {
					bonus := testutil.CreateTestBonusFile(nil)
					return testutil.CreateTestBonusDir(nil, bonus)
				},
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				testDir := testutil.CreateTestDir(t)
				cfg := testutil.CreateTestCfg(testDir)
				entry := test.entry()

				err := Resolve(entry, &cfg, logger)

				if err == nil {
					t.Errorf("Resolve expected error for %v, got nil", test.name)
				}
			})
		}
	})

	/*
		Testing movie file at root
	*/
	t.Run("movie file at root", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		entry := testutil.CreateTestMovieFile(nil)

		err := Resolve(entry, &cfg, logger)

		if err != nil {
			t.Errorf("Resolve unexpected error: %v", err)
		}
		expected := filepath.Join(cfg.MoviePath, "TEST.TITLE.2025", "TEST.TITLE.2025.1080p.x264.Remux.Atmos.English.mp4")
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})

	/*
		Testing episode file at root
	*/
	t.Run("episode file at root", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		entry := testutil.CreateTestEpFile(nil)

		err := Resolve(entry, &cfg, logger)

		if err != nil {
			t.Errorf("Resolve unexpected error: %v", err)
		}
		expected := filepath.Join(cfg.ShowPath, "TEST.TITLE.2025", "S01", "TEST.TITLE.2025.S01E01.1080p.x264.Remux.Atmos.English.mp4")
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})

	/*
		Testing movie directory
	*/
	t.Run("movie directory", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		movieFile := testutil.CreateTestMovieFile(nil)
		bonusFile := testutil.CreateTestBonusFile(nil)
		subFile := testutil.CreateTestSubFile(nil)
		entry := testutil.CreateTestMovieDir(nil, movieFile, bonusFile, subFile)

		err := Resolve(entry, &cfg, logger)

		if err != nil {
			t.Errorf("Resolve unexpected error: %v", err)
		}

		expectedDir := filepath.Join(cfg.MoviePath, "TEST.TITLE.2025")
		if entry.PathInfo.Dest != expectedDir {
			t.Errorf("Dir Dest = %v, want %v", entry.PathInfo.Dest, expectedDir)
		}
	})

	/*
		Testing season directory
	*/
	t.Run("season directory", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		ep1 := testutil.CreateTestEpFile(nil)
		ep2 := testutil.CreateTestEpFile(nil)
		sub := testutil.CreateTestSubFile(nil)
		entry := testutil.CreateTestSeasonDir(nil, ep1, ep2, sub)
		entry.MediaInfo.Title = ep1.MediaInfo.Title
		entry.MediaInfo.Year = ep1.MediaInfo.Year

		err := Resolve(entry, &cfg, logger)

		if err != nil {
			t.Errorf("Resolve unexpected error: %v", err)
		}

		expectedDir := filepath.Join(cfg.ShowPath, "TEST.TITLE.2025", "S01")
		if entry.PathInfo.Dest != expectedDir {
			t.Errorf("Dir Dest = %v, want %v", entry.PathInfo.Dest, expectedDir)
		}
	})

	/*
		Testing series directory
	*/
	t.Run("series directory", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		ep := testutil.CreateTestEpFile(nil)
		seasonDir := testutil.CreateTestSeasonDir(nil, ep)
		bonus := testutil.CreateTestBonusFile(nil)
		bonusDir := testutil.CreateTestBonusDir(nil, bonus)
		sub := testutil.CreateTestSubFile(nil)
		subDir := testutil.CreateTestSubDir(nil, sub)
		entry := testutil.CreateTestSeriesDir(nil, seasonDir, bonusDir, subDir)

		err := Resolve(entry, &cfg, logger)

		if err != nil {
			t.Errorf("Resolve unexpected error: %v", err)
		}

		expectedDir := filepath.Join(cfg.ShowPath, "TEST.TITLE")
		if entry.PathInfo.Dest != expectedDir {
			t.Errorf("Dir Dest = %v, want %v", entry.PathInfo.Dest, expectedDir)
		}

		expectedSeasonDir := filepath.Join(cfg.ShowPath, "TEST.TITLE", "S01")
		if entry.Children[0].PathInfo.Dest != expectedSeasonDir {
			t.Errorf("Season Dir Dest = %v, want %v", entry.Children[0].PathInfo.Dest, expectedSeasonDir)
		}

		expectedBonusDir := filepath.Join(cfg.ShowPath, "TEST.TITLE", "Extras")
		if entry.Children[1].PathInfo.Dest != expectedBonusDir {
			t.Errorf("Bonus Dir Dest = %v, want %v", entry.Children[1].PathInfo.Dest, expectedBonusDir)
		}

		expectedSubtitleDir := filepath.Join(cfg.ShowPath, "TEST.TITLE", "Subtitles")
		if entry.Children[2].PathInfo.Dest != expectedSubtitleDir {
			t.Errorf("Subtitle Dir Dest = %v, want %v", entry.Children[2].PathInfo.Dest, expectedSubtitleDir)
		}
	})

	/*
		Testing unknown role error
	*/
	t.Run("unknown role", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		entry := &metadata.Entry{
			Role: metadata.UnknownRole,
			PathInfo: metadata.PathInfo{
				Source: "/unknown/entry",
			},
		}

		err := Resolve(entry, &cfg, logger)

		if err == nil {
			t.Errorf("Resolve expected error for unknown role, got nil")
		}
	})
}

func TestResolveMovieFile(t *testing.T) {
	t.Run("with base path", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		entry := testutil.CreateTestMovieFile(nil)
		basePath := filepath.Join(cfg.MoviePath, "TEST.TITLE.2025")

		err := resolveMovieFile(basePath, entry, &cfg, logger)

		if err != nil {
			t.Errorf("resolveMovieFile unexpected error: %v", err)
		}
		expected := filepath.Join(basePath, "TEST.TITLE.2025.1080p.x264.Remux.Atmos.English.mp4")
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})

	t.Run("without base path", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		entry := testutil.CreateTestMovieFile(nil)

		err := resolveMovieFile("", entry, &cfg, logger)

		if err != nil {
			t.Errorf("resolveMovieFile unexpected error: %v", err)
		}
		expected := filepath.Join(cfg.MoviePath, "TEST.TITLE.2025", "TEST.TITLE.2025.1080p.x264.Remux.Atmos.English.mp4")
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})

	t.Run("wrong role", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		entry := testutil.CreateTestSubFile(nil)
		basePath := filepath.Join(cfg.MoviePath, "TEST.TITLE.2025")

		err := resolveMovieFile(basePath, entry, &cfg, logger)

		if err == nil {
			t.Errorf("resolveMovieFile expected error for subtitle file")
		}
	})
}

func TestResolveEpisodeFile(t *testing.T) {
	t.Run("with base path", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		entry := testutil.CreateTestEpFile(nil)
		basePath := filepath.Join(cfg.ShowPath, "TEST.TITLE.2025", "S01")

		err := resolveEpisodeFile(basePath, entry, &cfg, logger)

		if err != nil {
			t.Errorf("resolveEpisodeFile unexpected error: %v", err)
		}
		expected := filepath.Join(basePath, "TEST.TITLE.2025.S01E01.1080p.x264.Remux.Atmos.English.mp4")
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})

	t.Run("without base path", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		entry := testutil.CreateTestEpFile(nil)

		err := resolveEpisodeFile("", entry, &cfg, logger)

		if err != nil {
			t.Errorf("resolveEpisodeFile unexpected error: %v", err)
		}
		expected := filepath.Join(cfg.ShowPath, "TEST.TITLE.2025", "S01", "TEST.TITLE.2025.S01E01.1080p.x264.Remux.Atmos.English.mp4")
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})

	t.Run("wrong role", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		entry := testutil.CreateTestMovieFile(nil)
		basePath := filepath.Join(cfg.ShowPath, "TEST.TITLE.2025", "S01")

		err := resolveEpisodeFile(basePath, entry, &cfg, logger)

		if err == nil {
			t.Errorf("resolveEpisodeFile expected error for movie file")
		}
	})
}

func TestResolveSubtitleFile(t *testing.T) {
	t.Run("valid subtitle file", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		entry := testutil.CreateTestSubFile(nil)
		basePath := filepath.Join(cfg.MoviePath, "TEST.TITLE.2025")

		err := resolveSubtitleFile(basePath, entry, &cfg, logger)

		if err != nil {
			t.Errorf("resolveSubtitleFile unexpected error: %v", err)
		}
		// testutil.CreateTestSubFile has empty Title array, so filename starts with year
		expected := filepath.Join(basePath, ".2025.English.srt")
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})

	t.Run("wrong role", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		entry := testutil.CreateTestMovieFile(nil)
		basePath := filepath.Join(cfg.MoviePath, "TEST.TITLE.2025")

		err := resolveSubtitleFile(basePath, entry, &cfg, logger)

		if err == nil {
			t.Errorf("resolveSubtitleFile expected error for movie file")
		}
	})

	t.Run("missing base path", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		entry := testutil.CreateTestSubFile(nil)

		err := resolveSubtitleFile("", entry, &cfg, logger)

		if err == nil {
			t.Errorf("resolveSubtitleFile expected error for missing base path")
		}
	})
}

func TestResolveSubtitleDir(t *testing.T) {
	t.Run("valid subtitle dir", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		sub1 := testutil.CreateTestSubFile(nil)
		sub2 := testutil.CreateTestSubFile(nil)
		entry := testutil.CreateTestSubDir(nil, sub1, sub2)
		basePath := filepath.Join(cfg.MoviePath, "TEST.TITLE.2025")

		err := resolveSubtitleDir(basePath, entry, &cfg, logger)

		if err != nil {
			t.Errorf("resolveSubtitleDir unexpected error: %v", err)
		}

		expected := filepath.Join(basePath, "Subtitles")
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dir Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})

	t.Run("wrong role", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		movie := testutil.CreateTestMovieFile(nil)
		entry := testutil.CreateTestMovieDir(nil, movie)
		basePath := filepath.Join(cfg.MoviePath, "TEST.TITLE.2025")

		err := resolveSubtitleDir(basePath, entry, &cfg, logger)

		if err == nil {
			t.Errorf("resolveSubtitleDir expected error for movie dir")
		}
	})

	t.Run("missing base path", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		sub := testutil.CreateTestSubFile(nil)
		entry := testutil.CreateTestSubDir(nil, sub)

		err := resolveSubtitleDir("", entry, &cfg, logger)

		if err == nil {
			t.Errorf("resolveSubtitleDir expected error for missing base path")
		}
	})
}

func TestResolveBonusFile(t *testing.T) {
	t.Run("valid bonus file", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		entry := testutil.CreateTestBonusFile(nil)
		basePath := filepath.Join(cfg.MoviePath, "TEST.TITLE.2025", "Extras")

		err := resolveBonusFile(basePath, entry, &cfg, logger)

		if err != nil {
			t.Errorf("resolveBonusFile unexpected error: %v", err)
		}
		expected := filepath.Join(basePath, "TEST.TITLE.2025.1080p.x264.Remux.Atmos.English.Behind.The.Scenes.mp4")
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})

	t.Run("wrong role", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		entry := testutil.CreateTestMovieFile(nil)
		basePath := filepath.Join(cfg.MoviePath, "TEST.TITLE.2025", "Extras")

		err := resolveBonusFile(basePath, entry, &cfg, logger)

		if err == nil {
			t.Errorf("resolveBonusFile expected error for movie file")
		}
	})

	t.Run("missing base path", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		entry := testutil.CreateTestBonusFile(nil)

		err := resolveBonusFile("", entry, &cfg, logger)

		if err == nil {
			t.Errorf("resolveBonusFile expected error for missing base path")
		}
	})
}

func TestResolveBonusDir(t *testing.T) {
	t.Run("valid bonus dir", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		bonus := testutil.CreateTestBonusFile(nil)
		sub := testutil.CreateTestSubFile(nil)
		entry := testutil.CreateTestBonusDir(nil, bonus, sub)
		basePath := filepath.Join(cfg.MoviePath, "TEST.TITLE.2025")

		err := resolveBonusDir(basePath, entry, &cfg, logger)

		if err != nil {
			t.Errorf("resolveBonusDir unexpected error: %v", err)
		}

		expected := filepath.Join(basePath, "Extras")
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dir Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})

	t.Run("wrong role", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		movie := testutil.CreateTestMovieFile(nil)
		entry := testutil.CreateTestMovieDir(nil, movie)
		basePath := filepath.Join(cfg.MoviePath, "TEST.TITLE.2025")

		err := resolveBonusDir(basePath, entry, &cfg, logger)

		if err == nil {
			t.Errorf("resolveBonusDir expected error for movie dir")
		}
	})

	t.Run("missing base path", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		bonus := testutil.CreateTestBonusFile(nil)
		entry := testutil.CreateTestBonusDir(nil, bonus)

		err := resolveBonusDir("", entry, &cfg, logger)

		if err == nil {
			t.Errorf("resolveBonusDir expected error for missing base path")
		}
	})
}

func TestResolveSeasonDir(t *testing.T) {
	t.Run("with base path", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		ep := testutil.CreateTestEpFile(nil)
		entry := testutil.CreateTestSeasonDir(nil, ep)
		basePath := filepath.Join(cfg.ShowPath, "TEST.TITLE.2025")

		err := resolveSeasonDir(basePath, entry, &cfg, logger)

		if err != nil {
			t.Errorf("resolveSeasonDir unexpected error: %v", err)
		}

		expected := filepath.Join(basePath, "S01")
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dir Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})

	//TODO
	t.Run("without base path", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		ep := testutil.CreateTestEpFile(nil)
		entry := testutil.CreateTestSeasonDir(nil, ep)
		// Mimic behavior of enricher
		entry.MediaInfo.Title = ep.MediaInfo.Title
		entry.MediaInfo.Year = ep.MediaInfo.Year

		err := resolveSeasonDir("", entry, &cfg, logger)

		if err != nil {
			t.Errorf("resolveSeasonDir unexpected error: %v", err)
		}

		expected := filepath.Join(cfg.ShowPath, "TEST.TITLE.2025", "S01")
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dir Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})

	t.Run("with subtitle dir child", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		ep := testutil.CreateTestEpFile(nil)
		sub := testutil.CreateTestSubFile(nil)
		subDir := testutil.CreateTestSubDir(nil, sub)
		entry := testutil.CreateTestSeasonDir(nil, ep, subDir)
		// Mimic behavior of enricher
		entry.MediaInfo.Title = ep.MediaInfo.Title
		entry.MediaInfo.Year = ep.MediaInfo.Year

		err := resolveSeasonDir("", entry, &cfg, logger)

		if err != nil {
			t.Errorf("resolveSeasonDir unexpected error: %v", err)
		}

		expectedSubDir := filepath.Join(cfg.ShowPath, "TEST.TITLE.2025", "S01", "Subtitles")
		if entry.Children[1].PathInfo.Dest != expectedSubDir {
			t.Errorf("Subtitle Dir Dest = %v, want %v", entry.Children[1].PathInfo.Dest, expectedSubDir)
		}
	})

	t.Run("wrong role", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		movie := testutil.CreateTestMovieFile(nil)
		entry := testutil.CreateTestMovieDir(nil, movie)

		err := resolveSeasonDir("", entry, &cfg, logger)

		if err == nil {
			t.Errorf("resolveSeasonDir expected error for movie dir")
		}
	})
}

func TestResolveSeriesDir(t *testing.T) {
	t.Run("full series dir", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		ep := testutil.CreateTestEpFile(nil)
		seasonDir := testutil.CreateTestSeasonDir(nil, ep)
		bonus := testutil.CreateTestBonusFile(nil)
		bonusDir := testutil.CreateTestBonusDir(nil, bonus)
		sub := testutil.CreateTestSubFile(nil)
		subDir := testutil.CreateTestSubDir(nil, sub)
		entry := testutil.CreateTestSeriesDir(nil, seasonDir, bonusDir, subDir)

		err := resolveSeriesDir(entry, &cfg, logger)

		if err != nil {
			t.Errorf("resolveSeriesDir unexpected error: %v", err)
		}

		expected := filepath.Join(cfg.ShowPath, "TEST.TITLE")
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dir Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}

		expectedSeason := filepath.Join(cfg.ShowPath, "TEST.TITLE", "S01")
		if entry.Children[0].PathInfo.Dest != expectedSeason {
			t.Errorf("Season Dir Dest = %v, want %v", entry.Children[0].PathInfo.Dest, expectedSeason)
		}
	})

	t.Run("wrong role", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		movie := testutil.CreateTestMovieFile(nil)
		entry := testutil.CreateTestMovieDir(nil, movie)

		err := resolveSeriesDir(entry, &cfg, logger)

		if err == nil {
			t.Errorf("resolveSeriesDir expected error for movie dir")
		}
	})
}

func TestResolveMovieDir(t *testing.T) {
	t.Run("full movie dir", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		movie := testutil.CreateTestMovieFile(nil)
		bonus := testutil.CreateTestBonusFile(nil)
		sub := testutil.CreateTestSubFile(nil)
		entry := testutil.CreateTestMovieDir(nil, movie, bonus, sub)

		err := resolveMovieDir(entry, &cfg, logger)

		if err != nil {
			t.Errorf("resolveMovieDir unexpected error: %v", err)
		}

		expected := filepath.Join(cfg.MoviePath, "TEST.TITLE.2025")
		if entry.PathInfo.Dest != expected {
			t.Errorf("Dir Dest = %v, want %v", entry.PathInfo.Dest, expected)
		}
	})

	t.Run("movie dir with subdirs", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		movie := testutil.CreateTestMovieFile(nil)
		bonus := testutil.CreateTestBonusFile(nil)
		bonusDir := testutil.CreateTestBonusDir(nil, bonus)
		sub := testutil.CreateTestSubFile(nil)
		subDir := testutil.CreateTestSubDir(nil, sub)
		entry := testutil.CreateTestMovieDir(nil, movie, bonusDir, subDir)

		err := resolveMovieDir(entry, &cfg, logger)

		if err != nil {
			t.Errorf("resolveMovieDir unexpected error: %v", err)
		}

		expectedBonusDir := filepath.Join(cfg.MoviePath, "TEST.TITLE.2025", "Extras")
		if entry.Children[1].PathInfo.Dest != expectedBonusDir {
			t.Errorf("Bonus Dir Dest = %v, want %v", entry.Children[1].PathInfo.Dest, expectedBonusDir)
		}

		expectedSubtitleDir := filepath.Join(cfg.MoviePath, "TEST.TITLE.2025", "Subtitles")
		if entry.Children[2].PathInfo.Dest != expectedSubtitleDir {
			t.Errorf("Subtitle Dir Dest = %v, want %v", entry.Children[2].PathInfo.Dest, expectedSubtitleDir)
		}
	})

	t.Run("wrong role", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		ep := testutil.CreateTestEpFile(nil)
		entry := testutil.CreateTestSeasonDir(nil, ep)

		err := resolveMovieDir(entry, &cfg, logger)

		if err == nil {
			t.Errorf("resolveMovieDir expected error for season dir")
		}
	})
}

/*
	Helper function tests
*/

func TestBuildFilename(t *testing.T) {
	tests := []struct {
		name     string
		entry    func() *metadata.Entry
		expected string
	}{
		{
			name: "movie file - season/episode cleared by resolver",
			entry: func() *metadata.Entry {
				e := testutil.CreateTestMovieFile(nil)
				// Simulate what resolveMovieFile does
				e.MediaInfo.Season = nil
				e.MediaInfo.Episode = nil
				e.MediaInfo.Bonus = ""
				return e
			},
			expected: "TEST.TITLE.2025.1080p.x264.Remux.Atmos.English.mp4",
		},
		{
			name: "episode file",
			entry: func() *metadata.Entry {
				e := testutil.CreateTestEpFile(nil)
				e.MediaInfo.Bonus = ""
				return e
			},
			expected: "TEST.TITLE.2025.S01E01.1080p.x264.Remux.Atmos.English.mp4",
		},
		{
			name: "subtitle file - metadata cleared by resolver",
			entry: func() *metadata.Entry {
				e := testutil.CreateTestSubFile(nil)
				// Simulate what resolveSubtitleFile does
				e.MediaInfo.Resolution = ""
				e.MediaInfo.Codec = ""
				e.MediaInfo.Source = ""
				e.MediaInfo.Audio = ""
				e.MediaInfo.Bonus = ""
				return e
			},
			expected: ".2025.English.srt",
		},
		{
			name: "bonus file",
			entry: func() *metadata.Entry {
				return testutil.CreateTestBonusFile(nil)
			},
			expected: "TEST.TITLE.2025.1080p.x264.Remux.Atmos.English.Behind.The.Scenes.mp4",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := test.entry()
			result := buildFilename(entry)
			if result != test.expected {
				t.Errorf("buildFilename = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestBuildTitlePath(t *testing.T) {
	tests := []struct {
		name     string
		entry    func() *metadata.Entry
		expected string
	}{
		{
			name: "with year",
			entry: func() *metadata.Entry {
				return testutil.CreateTestMovieFile(nil)
			},
			expected: "TEST.TITLE.2025",
		},
		{
			name: "without year",
			entry: func() *metadata.Entry {
				entry := testutil.CreateTestMovieFile(nil)
				entry.MediaInfo.Year = nil
				return entry
			},
			expected: "TEST.TITLE",
		},
		{
			name: "empty title",
			entry: func() *metadata.Entry {
				entry := testutil.CreateTestMovieFile(nil)
				entry.MediaInfo.Title = []string{}
				return entry
			},
			expected: ".2025",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := test.entry()
			result := buildTitlePath(entry)
			if result != test.expected {
				t.Errorf("buildTitlePath = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestBuildSeasonPath(t *testing.T) {
	tests := []struct {
		name     string
		entry    func() *metadata.Entry
		expected string
	}{
		{
			name: "with season",
			entry: func() *metadata.Entry {
				return testutil.CreateTestSeasonDir(nil)
			},
			expected: "S01",
		},
		{
			name: "different season number",
			entry: func() *metadata.Entry {
				entry := testutil.CreateTestSeasonDir(nil)
				entry.MediaInfo.Season = testutil.IntPtr(5)
				return entry
			},
			expected: "S05",
		},
		{
			name: "without season defaults to S01",
			entry: func() *metadata.Entry {
				entry := testutil.CreateTestSeasonDir(nil)
				entry.MediaInfo.Season = nil
				return entry
			},
			expected: "S01",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := test.entry()
			result := buildSeasonPath(entry)
			if result != test.expected {
				t.Errorf("buildSeasonPath = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "uppercase - first char stays upper, rest unchanged",
			input:    "TEST",
			expected: "TEST",
		},
		{
			name:     "lowercase - first char uppercased",
			input:    "test",
			expected: "Test",
		},
		{
			name:     "mixed - first char uppercased, rest unchanged",
			input:    "tEsT",
			expected: "TEsT",
		},
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
		{
			name:     "single char lowercase",
			input:    "t",
			expected: "T",
		},
		{
			name:     "single char uppercase",
			input:    "T",
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
