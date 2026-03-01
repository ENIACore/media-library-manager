package classifier

import (
	"log/slog"
	"testing"

	"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/ENIACore/media_library_manager/internal/testutil"
)

var logger = slog.Default()

func TestClassify(t *testing.T) {
	tests := []struct {
		name         string
		entry        func() *metadata.Entry
		expectedRole metadata.EntryRole
		expectError  bool
	}{
		{
			name: "movie file at root",
			entry: func() *metadata.Entry {
				return testutil.CreateTestMovieFile(nil)
			},
			expectedRole: metadata.MovieFile,
			expectError:  false,
		},
		{
			name: "episode file at root",
			entry: func() *metadata.Entry {
				return testutil.CreateTestEpFile(nil)
			},
			expectedRole: metadata.EpisodeFile,
			expectError:  false,
		},
		{
			name: "subtitle file at root",
			entry: func() *metadata.Entry {
				return testutil.CreateTestSubFile(nil)
			},
			expectedRole: metadata.SubtitleFile,
			expectError:  false,
		},
		{
			name: "bonus file at root",
			entry: func() *metadata.Entry {
				return testutil.CreateTestBonusFile(nil)
			},
			expectedRole: metadata.BonusFile,
			expectError:  false,
		},
		{
			name: "DS file at root",
			entry: func() *metadata.Entry {
				return &metadata.Entry{
					Parent: nil,
					Children: nil,
					MediaInfo: metadata.MediaInfo{
						Title:      []string{"TEST", "TITLE"},
						Year:       testutil.IntPtr(2025),
						Resolution: "1080p",
						Codec:      "x264",
						Source:     "Remux",
						Audio:      "Atmos",
						Language:   []string{"English"},
						DS:         "Deleted.Scenes",
					},
					PathInfo: metadata.PathInfo{
						Source: "test.title.2025.deleted.scenes.1080p.x264.remux.mp4",
						Ext:    "MP4",
						Type:   metadata.Video,
						IsDir:  false,
					},
				}
			},
			expectedRole: metadata.DSFile,
			expectError:  false,
		},
		{
			name: "BTS file at root",
			entry: func() *metadata.Entry {
				return &metadata.Entry{
					Parent: nil,
					Children: nil,
					MediaInfo: metadata.MediaInfo{
						Title:      []string{"TEST", "TITLE"},
						Year:       testutil.IntPtr(2025),
						Resolution: "1080p",
						Codec:      "x264",
						Source:     "Remux",
						Audio:      "Atmos",
						Language:   []string{"English"},
						BTS:        "Behind.The.Scenes",
					},
					PathInfo: metadata.PathInfo{
						Source: "test.title.2025.behind.the.scenes.1080p.x264.remux.mp4",
						Ext:    "MP4",
						Type:   metadata.Video,
						IsDir:  false,
					},
				}
			},
			expectedRole: metadata.BTSFile,
			expectError:  false,
		},
		{
			name: "movie directory",
			entry: func() *metadata.Entry {
				movieFile := testutil.CreateTestMovieFile(nil)
				subFile := testutil.CreateTestSubFile(nil)
				return testutil.CreateTestMovieDir(nil, movieFile, subFile)
			},
			expectedRole: metadata.MovieDir,
			expectError:  false,
		},
		{
			name: "season directory",
			entry: func() *metadata.Entry {
				ep1 := testutil.CreateTestEpFile(nil)
				ep2 := testutil.CreateTestEpFile(nil)
				return testutil.CreateTestSeasonDir(nil, ep1, ep2)
			},
			expectedRole: metadata.SeasonDir,
			expectError:  false,
		},
		{
			name: "series directory",
			entry: func() *metadata.Entry {
				ep1 := testutil.CreateTestEpFile(nil)
				ep2 := testutil.CreateTestEpFile(nil)
				seasonDir := testutil.CreateTestSeasonDir(nil, ep1, ep2)
				return testutil.CreateTestSeriesDir(nil, seasonDir)
			},
			expectedRole: metadata.SeriesDir,
			expectError:  false,
		},
		{
			name: "subtitle directory",
			entry: func() *metadata.Entry {
				sub1 := testutil.CreateTestSubFile(nil)
				sub2 := testutil.CreateTestSubFile(nil)
				return testutil.CreateTestSubDir(nil, sub1, sub2)
			},
			expectedRole: metadata.SubtitleDir,
			expectError:  false,
		},
		{
			name: "bonus directory",
			entry: func() *metadata.Entry {
				bonus := testutil.CreateTestBonusFile(nil)
				sub := testutil.CreateTestSubFile(nil)
				return testutil.CreateTestBonusDir(nil, bonus, sub)
			},
			expectedRole: metadata.BonusDir,
			expectError:  false,
		},
		{
			name: "unknown file type",
			entry: func() *metadata.Entry {
				return &metadata.Entry{
					PathInfo: metadata.PathInfo{
						IsDir:  false,
						Source: "/unknown/file.xyz",
						Type:   metadata.UnknownType,
					},
				}
			},
			expectedRole: metadata.UnknownRole,
			expectError:  true,
		},
		{
			name: "empty directory",
			entry: func() *metadata.Entry {
				return &metadata.Entry{
					PathInfo: metadata.PathInfo{
						IsDir:  true,
						Source: "/empty/dir",
						Type:   metadata.UnknownType,
					},
					Children: []*metadata.Entry{},
				}
			},
			expectedRole: metadata.UnknownRole,
			expectError:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := test.entry()
			resetEntryRoles(entry)

			err := Classify(entry, logger)

			if test.expectError && err == nil {
				t.Errorf("Classify expected error, got nil")
			}
			if !test.expectError && err != nil {
				t.Errorf("Classify unexpected error: %v", err)
			}
			if !test.expectError && entry.Role != test.expectedRole {
				t.Errorf("Role = %v, want %v", entry.Role, test.expectedRole)
			}
		})
	}
}

func TestClassifySubtitleFile(t *testing.T) {
	tests := []struct {
		name     string
		entry    func() *metadata.Entry
		expected bool
	}{
		{
			name: "valid subtitle file",
			entry: func() *metadata.Entry {
				return testutil.CreateTestSubFile(nil)
			},
			expected: true,
		},
		{
			name: "movie file",
			entry: func() *metadata.Entry {
				return testutil.CreateTestMovieFile(nil)
			},
			expected: false,
		},
		{
			name: "episode file",
			entry: func() *metadata.Entry {
				return testutil.CreateTestEpFile(nil)
			},
			expected: false,
		},
		{
			name: "bonus file",
			entry: func() *metadata.Entry {
				return testutil.CreateTestBonusFile(nil)
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := test.entry()
			resetEntryRoles(entry)

			result := classifySubtitleFile(entry)

			if result != test.expected {
				t.Errorf("classifySubtitleFile = %v, want %v", result, test.expected)
			}
			if result && entry.Role != metadata.SubtitleFile {
				t.Errorf("Role = %v, want %v", entry.Role, metadata.SubtitleFile)
			}
		})
	}
}

func TestClassifyDSFile(t *testing.T) {
	tests := []struct {
		name     string
		entry    func() *metadata.Entry
		expected bool
	}{
		{
			name: "valid DS file with DS metadata",
			entry: func() *metadata.Entry {
				return &metadata.Entry{
					Parent: nil,
					Children: nil,
					MediaInfo: metadata.MediaInfo{
						Title:      []string{"TEST", "TITLE"},
						Year:       testutil.IntPtr(2025),
						Episode:    nil,
						Season:     nil,
						Resolution: "1080p",
						Codec:      "x264",
						Source:     "Remux",
						Audio:      "Atmos",
						Language:   []string{"English"},
						DS:         "Deleted.Scenes",
						BTS:        "",
						Bonus:      "",
					},
					PathInfo: metadata.PathInfo{
						Dest:   "",
						Source: "test.title.2025.deleted.scenes.1080p.x264.remux.atmos.english.mp4",
						Ext:    "MP4",
						Type:   metadata.Video,
						IsDir:  false,
					},
					Role: metadata.UnknownRole,
				}
			},
			expected: true,
		},
		{
			name: "DS file with parent DS metadata",
			entry: func() *metadata.Entry {
				parent := &metadata.Entry{
					MediaInfo: metadata.MediaInfo{
						DS: "Deleted.Scenes",
					},
				}
				return &metadata.Entry{
					Parent: parent,
					Children: nil,
					MediaInfo: metadata.MediaInfo{
						Title:      []string{"TEST", "TITLE"},
						Year:       testutil.IntPtr(2025),
						Episode:    nil,
						Season:     nil,
						Resolution: "1080p",
						Codec:      "x264",
						Source:     "Remux",
						Audio:      "Atmos",
						Language:   []string{"English"},
						DS:         "",
						BTS:        "",
						Bonus:      "",
					},
					PathInfo: metadata.PathInfo{
						Dest:   "",
						Source: "test.title.2025.1080p.x264.remux.atmos.english.mp4",
						Ext:    "MP4",
						Type:   metadata.Video,
						IsDir:  false,
					},
					Role: metadata.UnknownRole,
				}
			},
			expected: true,
		},
		{
			name: "movie file without DS metadata",
			entry: func() *metadata.Entry {
				return testutil.CreateTestMovieFile(nil)
			},
			expected: false,
		},
		{
			name: "subtitle file",
			entry: func() *metadata.Entry {
				return testutil.CreateTestSubFile(nil)
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := test.entry()
			resetEntryRoles(entry)

			result := classifyDSFile(entry)

			if result != test.expected {
				t.Errorf("classifyDSFile = %v, want %v", result, test.expected)
			}
			if result && entry.Role != metadata.DSFile {
				t.Errorf("Role = %v, want %v", entry.Role, metadata.DSFile)
			}
		})
	}
}

func TestClassifyBTSFile(t *testing.T) {
	tests := []struct {
		name     string
		entry    func() *metadata.Entry
		expected bool
	}{
		{
			name: "valid BTS file with BTS metadata",
			entry: func() *metadata.Entry {
				return &metadata.Entry{
					Parent: nil,
					Children: nil,
					MediaInfo: metadata.MediaInfo{
						Title:      []string{"TEST", "TITLE"},
						Year:       testutil.IntPtr(2025),
						Episode:    nil,
						Season:     nil,
						Resolution: "1080p",
						Codec:      "x264",
						Source:     "Remux",
						Audio:      "Atmos",
						Language:   []string{"English"},
						DS:         "",
						BTS:        "Behind.The.Scenes",
						Bonus:      "",
					},
					PathInfo: metadata.PathInfo{
						Dest:   "",
						Source: "test.title.2025.behind.the.scenes.1080p.x264.remux.atmos.english.mp4",
						Ext:    "MP4",
						Type:   metadata.Video,
						IsDir:  false,
					},
					Role: metadata.UnknownRole,
				}
			},
			expected: true,
		},
		{
			name: "BTS file with parent BTS metadata",
			entry: func() *metadata.Entry {
				parent := &metadata.Entry{
					MediaInfo: metadata.MediaInfo{
						BTS: "Behind.The.Scenes",
					},
				}
				return &metadata.Entry{
					Parent: parent,
					Children: nil,
					MediaInfo: metadata.MediaInfo{
						Title:      []string{"TEST", "TITLE"},
						Year:       testutil.IntPtr(2025),
						Episode:    nil,
						Season:     nil,
						Resolution: "1080p",
						Codec:      "x264",
						Source:     "Remux",
						Audio:      "Atmos",
						Language:   []string{"English"},
						DS:         "",
						BTS:        "",
						Bonus:      "",
					},
					PathInfo: metadata.PathInfo{
						Dest:   "",
						Source: "test.title.2025.1080p.x264.remux.atmos.english.mp4",
						Ext:    "MP4",
						Type:   metadata.Video,
						IsDir:  false,
					},
					Role: metadata.UnknownRole,
				}
			},
			expected: true,
		},
		{
			name: "movie file without BTS metadata",
			entry: func() *metadata.Entry {
				return testutil.CreateTestMovieFile(nil)
			},
			expected: false,
		},
		{
			name: "subtitle file",
			entry: func() *metadata.Entry {
				return testutil.CreateTestSubFile(nil)
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := test.entry()
			resetEntryRoles(entry)

			result := classifyBTSFile(entry)

			if result != test.expected {
				t.Errorf("classifyBTSFile = %v, want %v", result, test.expected)
			}
			if result && entry.Role != metadata.BTSFile {
				t.Errorf("Role = %v, want %v", entry.Role, metadata.BTSFile)
			}
		})
	}
}

func TestClassifyBonusFile(t *testing.T) {
	tests := []struct {
		name     string
		entry    func() *metadata.Entry
		expected bool
	}{
		{
			name: "valid bonus file",
			entry: func() *metadata.Entry {
				return testutil.CreateTestBonusFile(nil)
			},
			expected: true,
		},
		{
			name: "movie file",
			entry: func() *metadata.Entry {
				return testutil.CreateTestMovieFile(nil)
			},
			expected: false,
		},
		{
			name: "episode file",
			entry: func() *metadata.Entry {
				return testutil.CreateTestEpFile(nil)
			},
			expected: false,
		},
		{
			name: "subtitle file",
			entry: func() *metadata.Entry {
				return testutil.CreateTestSubFile(nil)
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := test.entry()
			resetEntryRoles(entry)

			result := classifyBonusFile(entry)

			if result != test.expected {
				t.Errorf("classifyBonusFile = %v, want %v", result, test.expected)
			}
			if result && entry.Role != metadata.BonusFile {
				t.Errorf("Role = %v, want %v", entry.Role, metadata.BonusFile)
			}
		})
	}
}

func TestClassifyEpisodeFile(t *testing.T) {
	tests := []struct {
		name     string
		entry    func() *metadata.Entry
		expected bool
	}{
		{
			name: "valid episode file",
			entry: func() *metadata.Entry {
				return testutil.CreateTestEpFile(nil)
			},
			expected: true,
		},
		{
			name: "episode file with parent season",
			entry: func() *metadata.Entry {
				seasonDir := testutil.CreateTestSeasonDir(nil)
				ep := &metadata.Entry{
					Parent: seasonDir,
					PathInfo: metadata.PathInfo{
						IsDir:  false,
						Source: "/show/S01/episode.mp4",
						Type:   metadata.Video,
					},
					MediaInfo: metadata.MediaInfo{},
				}
				return ep
			},
			expected: true,
		},
		{
			name: "movie file",
			entry: func() *metadata.Entry {
				return testutil.CreateTestMovieFile(nil)
			},
			expected: false,
		},
		{
			name: "bonus file",
			entry: func() *metadata.Entry {
				return testutil.CreateTestBonusFile(nil)
			},
			expected: false,
		},
		{
			name: "subtitle file",
			entry: func() *metadata.Entry {
				return testutil.CreateTestSubFile(nil)
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := test.entry()
			resetEntryRoles(entry)

			result := classifyEpisodeFile(entry)

			if result != test.expected {
				t.Errorf("classifyEpisodeFile = %v, want %v", result, test.expected)
			}
			if result && entry.Role != metadata.EpisodeFile {
				t.Errorf("Role = %v, want %v", entry.Role, metadata.EpisodeFile)
			}
		})
	}
}

func TestClassifyMovieFile(t *testing.T) {
	tests := []struct {
		name     string
		entry    func() *metadata.Entry
		expected bool
	}{
		{
			name: "valid movie file",
			entry: func() *metadata.Entry {
				return testutil.CreateTestMovieFile(nil)
			},
			expected: true,
		},
		{
			name: "episode file",
			entry: func() *metadata.Entry {
				return testutil.CreateTestEpFile(nil)
			},
			expected: false,
		},
		{
			name: "bonus file",
			entry: func() *metadata.Entry {
				return testutil.CreateTestBonusFile(nil)
			},
			expected: false,
		},
		{
			name: "subtitle file",
			entry: func() *metadata.Entry {
				return testutil.CreateTestSubFile(nil)
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := test.entry()
			resetEntryRoles(entry)

			result := classifyMovieFile(entry)

			if result != test.expected {
				t.Errorf("classifyMovieFile = %v, want %v", result, test.expected)
			}
			if result && entry.Role != metadata.MovieFile {
				t.Errorf("Role = %v, want %v", entry.Role, metadata.MovieFile)
			}
		})
	}
}

func TestClassifySubtitleDir(t *testing.T) {
	tests := []struct {
		name     string
		entry    func() *metadata.Entry
		expected bool
	}{
		{
			name: "valid subtitle directory",
			entry: func() *metadata.Entry {
				sub1 := testutil.CreateTestSubFile(nil)
				sub2 := testutil.CreateTestSubFile(nil)
				sub3 := testutil.CreateTestSubFile(nil)
				return testutil.CreateTestSubDir(nil, sub1, sub2, sub3)
			},
			expected: true,
		},
		{
			name: "directory with non-subtitle file",
			entry: func() *metadata.Entry {
				sub := testutil.CreateTestSubFile(nil)
				movie := testutil.CreateTestMovieFile(nil)
				return testutil.CreateTestSubDir(nil, sub, movie)
			},
			expected: false,
		},
		{
			name: "empty directory",
			entry: func() *metadata.Entry {
				return testutil.CreateTestSubDir(nil)
			},
			expected: false,
		},
		{
			name: "file instead of directory",
			entry: func() *metadata.Entry {
				return testutil.CreateTestSubFile(nil)
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := test.entry()
			resetEntryRoles(entry)

			result := classifySubtitleDir(entry, logger)

			if result != test.expected {
				t.Errorf("classifySubtitleDir = %v, want %v", result, test.expected)
			}
			if result && entry.Role != metadata.SubtitleDir {
				t.Errorf("Role = %v, want %v", entry.Role, metadata.SubtitleDir)
			}
		})
	}
}

func TestClassifyExtrasDir(t *testing.T) {
	tests := []struct {
		name     string
		entry    func() *metadata.Entry
		expected bool
	}{
		{
			name: "valid bonus directory",
			entry: func() *metadata.Entry {
				bonus := testutil.CreateTestBonusFile(nil)
				sub := testutil.CreateTestSubFile(nil)
				return testutil.CreateTestBonusDir(nil, bonus, sub)
			},
			expected: true,
		},
		{
			name: "bonus directory with only bonus files",
			entry: func() *metadata.Entry {
				bonus1 := testutil.CreateTestBonusFile(nil)
				bonus2 := testutil.CreateTestBonusFile(nil)
				return testutil.CreateTestBonusDir(nil, bonus1, bonus2)
			},
			expected: true,
		},
		{
			name: "directory with only subtitles",
			entry: func() *metadata.Entry {
				sub1 := testutil.CreateTestSubFile(nil)
				sub2 := testutil.CreateTestSubFile(nil)
				return testutil.CreateTestBonusDir(nil, sub1, sub2)
			},
			expected: true,
		},
		{
			name: "directory with movie file",
			entry: func() *metadata.Entry {
				bonus := testutil.CreateTestBonusFile(nil)
				movie := testutil.CreateTestMovieFile(nil)
				return testutil.CreateTestBonusDir(nil, bonus, movie)
			},
			expected: true,
		},
		{
			name: "file instead of directory",
			entry: func() *metadata.Entry {
				return testutil.CreateTestBonusFile(nil)
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := test.entry()
			resetEntryRoles(entry)

			result := classifyExtrasDir(entry, metadata.UnknownRole, logger)

			if result != test.expected {
				t.Errorf("classifyExtrasDir = %v, want %v", result, test.expected)
			}
			if result && entry.Role != metadata.BonusDir {
				t.Errorf("Role = %v, want %v", entry.Role, metadata.BonusDir)
			}
		})
	}
}

func TestClassifySeasonDir(t *testing.T) {
	tests := []struct {
		name     string
		entry    func() *metadata.Entry
		expected bool
	}{
		{
			name: "valid season directory",
			entry: func() *metadata.Entry {
				ep1 := testutil.CreateTestEpFile(nil)
				ep2 := testutil.CreateTestEpFile(nil)
				sub := testutil.CreateTestSubFile(nil)
				return testutil.CreateTestSeasonDir(nil, ep1, ep2, sub)
			},
			expected: true,
		},
		{
			name: "season directory with subtitle dir",
			entry: func() *metadata.Entry {
				ep := testutil.CreateTestEpFile(nil)
				subFile := testutil.CreateTestSubFile(nil)
				subDir := testutil.CreateTestSubDir(nil, subFile)
				return testutil.CreateTestSeasonDir(nil, ep, subDir)
			},
			expected: true,
		},
		{
			name: "directory without episodes",
			entry: func() *metadata.Entry {
				sub1 := testutil.CreateTestSubFile(nil)
				sub2 := testutil.CreateTestSubFile(nil)
				return testutil.CreateTestSeasonDir(nil, sub1, sub2)
			},
			expected: false,
		},
		{
			name: "directory with movie file",
			entry: func() *metadata.Entry {
				ep := testutil.CreateTestEpFile(nil)
				movie := testutil.CreateTestMovieFile(nil)
				return testutil.CreateTestSeasonDir(nil, ep, movie)
			},
			expected: true,
		},
		{
			name: "file instead of directory",
			entry: func() *metadata.Entry {
				return testutil.CreateTestEpFile(nil)
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := test.entry()
			resetEntryRoles(entry)

			result := classifySeasonDir(entry, logger)

			if result != test.expected {
				t.Errorf("classifySeasonDir = %v, want %v", result, test.expected)
			}
			if result && entry.Role != metadata.SeasonDir {
				t.Errorf("Role = %v, want %v", entry.Role, metadata.SeasonDir)
			}
		})
	}
}

func TestClassifySeriesDir(t *testing.T) {
	tests := []struct {
		name     string
		entry    func() *metadata.Entry
		expected bool
	}{
		{
			name: "valid series directory",
			entry: func() *metadata.Entry {
				ep1 := testutil.CreateTestEpFile(nil)
				ep2 := testutil.CreateTestEpFile(nil)
				seasonDir := testutil.CreateTestSeasonDir(nil, ep1, ep2)
				return testutil.CreateTestSeriesDir(nil, seasonDir)
			},
			expected: true,
		},
		{
			name: "series directory with bonus and subtitle dirs",
			entry: func() *metadata.Entry {
				ep := testutil.CreateTestEpFile(nil)
				seasonDir := testutil.CreateTestSeasonDir(nil, ep)
				bonus := testutil.CreateTestBonusFile(nil)
				bonusDir := testutil.CreateTestBonusDir(nil, bonus)
				sub := testutil.CreateTestSubFile(nil)
				subDir := testutil.CreateTestSubDir(nil, sub)
				return testutil.CreateTestSeriesDir(nil, seasonDir, bonusDir, subDir)
			},
			expected: true,
		},
		{
			name: "directory without season",
			entry: func() *metadata.Entry {
				bonus := testutil.CreateTestBonusFile(nil)
				bonusDir := testutil.CreateTestBonusDir(nil, bonus)
				sub := testutil.CreateTestSubFile(nil)
				subDir := testutil.CreateTestSubDir(nil, sub)
				return testutil.CreateTestSeriesDir(nil, bonusDir, subDir)
			},
			expected: false,
		},
		{
			name: "directory with loose files",
			entry: func() *metadata.Entry {
				ep := testutil.CreateTestEpFile(nil)
				seasonDir := testutil.CreateTestSeasonDir(nil, ep)
				looseEp := testutil.CreateTestEpFile(nil)
				return testutil.CreateTestSeriesDir(nil, seasonDir, looseEp)
			},
			expected: false,
		},
		{
			name: "file instead of directory",
			entry: func() *metadata.Entry {
				return testutil.CreateTestEpFile(nil)
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := test.entry()
			resetEntryRoles(entry)

			result := classifySeriesDir(entry, logger)

			if result != test.expected {
				t.Errorf("classifySeriesDir = %v, want %v", result, test.expected)
			}
			if result && entry.Role != metadata.SeriesDir {
				t.Errorf("Role = %v, want %v", entry.Role, metadata.SeriesDir)
			}
		})
	}
}

func TestClassifyMovieDir(t *testing.T) {
	tests := []struct {
		name     string
		entry    func() *metadata.Entry
		expected bool
	}{
		{
			name: "valid movie directory",
			entry: func() *metadata.Entry {
				movie := testutil.CreateTestMovieFile(nil)
				sub := testutil.CreateTestSubFile(nil)
				return testutil.CreateTestMovieDir(nil, movie, sub)
			},
			expected: true,
		},
		{
			name: "movie directory with bonus and subtitle dirs",
			entry: func() *metadata.Entry {
				movie := testutil.CreateTestMovieFile(nil)
				bonus := testutil.CreateTestBonusFile(nil)
				bonusDir := testutil.CreateTestBonusDir(nil, bonus)
				sub := testutil.CreateTestSubFile(nil)
				subDir := testutil.CreateTestSubDir(nil, sub)
				return testutil.CreateTestMovieDir(nil, movie, bonusDir, subDir)
			},
			expected: true,
		},
		{
			name: "movie directory with bonus files",
			entry: func() *metadata.Entry {
				movie := testutil.CreateTestMovieFile(nil)
				bonus := testutil.CreateTestBonusFile(nil)
				return testutil.CreateTestMovieDir(nil, movie, bonus)
			},
			expected: true,
		},
		{
			name: "directory with multiple movie files",
			entry: func() *metadata.Entry {
				movie1 := testutil.CreateTestMovieFile(nil)
				movie2 := testutil.CreateTestMovieFile(nil)
				return testutil.CreateTestMovieDir(nil, movie1, movie2)
			},
			expected: false,
		},
		{
			name: "directory without movie file",
			entry: func() *metadata.Entry {
				bonus := testutil.CreateTestBonusFile(nil)
				sub := testutil.CreateTestSubFile(nil)
				return testutil.CreateTestMovieDir(nil, bonus, sub)
			},
			expected: false,
		},
		{
			name: "directory with season metadata",
			entry: func() *metadata.Entry {
				movie := testutil.CreateTestMovieFile(nil)
				dir := testutil.CreateTestMovieDir(nil, movie)
				dir.MediaInfo.Season = testutil.IntPtr(1)
				return dir
			},
			expected: false,
		},
		{
			name: "directory with episode metadata",
			entry: func() *metadata.Entry {
				movie := testutil.CreateTestMovieFile(nil)
				dir := testutil.CreateTestMovieDir(nil, movie)
				dir.MediaInfo.Episode = testutil.IntPtr(1)
				return dir
			},
			expected: false,
		},
		{
			name: "file instead of directory",
			entry: func() *metadata.Entry {
				return testutil.CreateTestMovieFile(nil)
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := test.entry()
			resetEntryRoles(entry)

			result := classifyMovieDir(entry, logger)

			if result != test.expected {
				t.Errorf("classifyMovieDir = %v, want %v", result, test.expected)
			}
			if result && entry.Role != metadata.MovieDir {
				t.Errorf("Role = %v, want %v", entry.Role, metadata.MovieDir)
			}
		})
	}
}

func TestClassifyChildrenRoles(t *testing.T) {
	t.Run("movie directory children are classified", func(t *testing.T) {
		movie := testutil.CreateTestMovieFile(nil)
		bonus := testutil.CreateTestBonusFile(nil)
		sub := testutil.CreateTestSubFile(nil)
		dir := testutil.CreateTestMovieDir(nil, movie, bonus, sub)
		resetEntryRoles(dir)

		err := Classify(dir, logger)

		if err != nil {
			t.Errorf("Classify unexpected error: %v", err)
		}
		if dir.Children[0].Role != metadata.MovieFile {
			t.Errorf("Movie child Role = %v, want %v", dir.Children[0].Role, metadata.MovieFile)
		}
		if dir.Children[1].Role != metadata.BonusFile {
			t.Errorf("Bonus child Role = %v, want %v", dir.Children[1].Role, metadata.BonusFile)
		}
		if dir.Children[2].Role != metadata.SubtitleFile {
			t.Errorf("Subtitle child Role = %v, want %v", dir.Children[2].Role, metadata.SubtitleFile)
		}
	})

	t.Run("season directory children are classified", func(t *testing.T) {
		ep1 := testutil.CreateTestEpFile(nil)
		ep2 := testutil.CreateTestEpFile(nil)
		sub := testutil.CreateTestSubFile(nil)
		dir := testutil.CreateTestSeasonDir(nil, ep1, ep2, sub)
		resetEntryRoles(dir)

		err := Classify(dir, logger)

		if err != nil {
			t.Errorf("Classify unexpected error: %v", err)
		}
		if dir.Children[0].Role != metadata.EpisodeFile {
			t.Errorf("Episode 1 Role = %v, want %v", dir.Children[0].Role, metadata.EpisodeFile)
		}
		if dir.Children[1].Role != metadata.EpisodeFile {
			t.Errorf("Episode 2 Role = %v, want %v", dir.Children[1].Role, metadata.EpisodeFile)
		}
		if dir.Children[2].Role != metadata.SubtitleFile {
			t.Errorf("Subtitle Role = %v, want %v", dir.Children[2].Role, metadata.SubtitleFile)
		}
	})

	t.Run("series directory nested children are classified", func(t *testing.T) {
		ep := testutil.CreateTestEpFile(nil)
		seasonDir := testutil.CreateTestSeasonDir(nil, ep)
		bonus := testutil.CreateTestBonusFile(nil)
		bonusDir := testutil.CreateTestBonusDir(nil, bonus)
		sub := testutil.CreateTestSubFile(nil)
		subDir := testutil.CreateTestSubDir(nil, sub)
		dir := testutil.CreateTestSeriesDir(nil, seasonDir, bonusDir, subDir)
		resetEntryRoles(dir)

		err := Classify(dir, logger)

		if err != nil {
			t.Errorf("Classify unexpected error: %v", err)
		}
		if dir.Children[0].Role != metadata.SeasonDir {
			t.Errorf("Season dir Role = %v, want %v", dir.Children[0].Role, metadata.SeasonDir)
		}
		if dir.Children[0].Children[0].Role != metadata.EpisodeFile {
			t.Errorf("Episode Role = %v, want %v", dir.Children[0].Children[0].Role, metadata.EpisodeFile)
		}
		if dir.Children[1].Role != metadata.BonusDir {
			t.Errorf("Bonus dir Role = %v, want %v", dir.Children[1].Role, metadata.BonusDir)
		}
		if dir.Children[1].Children[0].Role != metadata.BonusFile {
			t.Errorf("Bonus file Role = %v, want %v", dir.Children[1].Children[0].Role, metadata.BonusFile)
		}
		if dir.Children[2].Role != metadata.SubtitleDir {
			t.Errorf("Subtitle dir Role = %v, want %v", dir.Children[2].Role, metadata.SubtitleDir)
		}
		if dir.Children[2].Children[0].Role != metadata.SubtitleFile {
			t.Errorf("Subtitle file Role = %v, want %v", dir.Children[2].Children[0].Role, metadata.SubtitleFile)
		}
	})
}

func resetEntryRoles(entry *metadata.Entry) {
	entry.Role = metadata.UnknownRole
	for _, child := range entry.Children {
		resetEntryRoles(child)
	}
}
