package resolver

import (
	"testing"

	"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/ENIACore/media_library_manager/internal/testutil"
)

func TestResolve(t *testing.T) {
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

				err := Resolve(entry, &cfg)

				if err == nil {
					t.Errorf("Resolve expected error for %v, got nil", test.name)
				}
			})
		}
	})

	t.Run("unknown role", func(t *testing.T) {
		testDir := testutil.CreateTestDir(t)
		cfg := testutil.CreateTestCfg(testDir)
		entry := &metadata.Entry{
			Role: metadata.UnknownRole,
			FileInfo: metadata.FileInfo{
				SourcePath: "/unknown/entry",
			},
		}

		err := Resolve(entry, &cfg)

		if err == nil {
			t.Errorf("Resolve expected error for unknown role, got nil")
		}
	})
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "all uppercase - first char stays upper, rest lowercased",
			input:    "TEST",
			expected: "Test",
		},
		{
			name:     "all lowercase - first char uppercased",
			input:    "test",
			expected: "Test",
		},
		{
			name:     "mixed case - first char uppercased, rest lowercased",
			input:    "tEsT",
			expected: "Test",
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
