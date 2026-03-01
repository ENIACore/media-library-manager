package extractor

import (
	"regexp"
	"testing"
)

func TestSanitizeName(t *testing.T) {
	tests := []struct{
		name		string
		input		string
		expected	[]string
	}{
		{
			// Tests removing all special characters, removing spaces, removing quotations, capitalization, removing trailing and preceding ".", combining names by a single "."
			name:		"remove all specicial characters",
			input:		"!@#$%^&*()-_+=\"'. mOvIe .!@#$%^&*()-_+=\"' tItLe !@#$%^&*()-_+=\"'.",
			expected:	[]string{"MOVIE", "TITLE"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sanitizedName := SanitizeName(test.input)
			if len(sanitizedName) != len(test.expected) {
				t.Errorf("sanitizeName() = %v, want %v", sanitizedName, test.expected)
				return
			}
			for i := range sanitizedName {
				if sanitizedName[i] != test.expected[i] {
					t.Errorf("sanitizeName() = %v, want %v", sanitizedName, test.expected)
					return
				}
			}
		})
	}
}

func TestMatchSegments(t *testing.T) {
	tests := []struct{
		name		string
		pattern		string
		input		[]string
		expected	[]string
	}{
		{
			name:		"single segment match",
			pattern:	`240[PI]?`, 	
			input:		[]string{
				`240P`,	
			},
			expected:	[]string{
				`240P`,	
			},
		},
		{
			name:		"single segment match with capture group",
			pattern:	`S(\d+)E\d+`,
			input:		[]string{
				`S123E1234567`,	
			},
			expected:	[]string{
				`S123E1234567`,	
				`123`,
			},
		},
		{
			name:		"double segment match",
			pattern:	`7\.1`,
			input:		[]string{
				`7`,
				`1`,
			},
			expected:	[]string{
				`7.1`,
			},
		},
		{
			name:		"double segment match with capture group",
			pattern:	`EPISODE\.(\d+)`,
			input:		[]string{
				`EPISODE`,
				`123456`,	
			},
			expected:	[]string{
				`EPISODE.123456`,	
				`123456`,
			},
		},
		{
			name:		"triple segment match with capture group",
			pattern:	`\d+\.X\.(\d+)`,
			input:		[]string{
				`321`,
				`X`,
				`123`,	
			},
			expected:	[]string{
				`321.X.123`,	
				`123`,	
			},
		},
		{
			name:		"single segment match with extra words",
			pattern:	`240[PI]?`, 	
			input:		[]string{
				`240P`,
				`EXTRA`,
				`WORDS`,
			},
			expected:	[]string{
				`240P`,	
			},
		},
		{
			name:		"double segment match with extra words",
			pattern:	`7\.1`,
			input:		[]string{
				`7`,
				`1`,
				`EXTRA`,
				`WORDS`,
			},
			expected:	[]string{
				`7.1`,
			},
		},
		{
			name:		"single segment match with extra words without seperator",
			pattern:	`240[PI]?`, 	
			input:		[]string{
				`240PEXTRAWORDS`,
			},
			expected: nil,			
		},
		{
			name:		"single segment match with extra words without seperator at beginning",
			pattern:	`240[PI]?`, 	
			input:		[]string{
				`EXTRAWORDS240P`,
			},
			expected: nil,			
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			re := regexp.MustCompile(test.pattern)
			res := matchSegments(test.input, re)

			if test.expected == nil {
				if res != nil {
					t.Errorf("expected nil, got %v", res)
				}
				return
			}
			
			if res == nil {
				t.Errorf("expected %v, got nil", test.expected)
				return
			}
			
			if len(res) != len(test.expected) {
				t.Errorf("expected %d matches, got %d", len(test.expected), len(res))
				return
			}
			
			for i, match := range res {
				if match != test.expected[i] {
					t.Errorf("match[%d]: expected %q, got %q", i, test.expected[i], match)
				}
			}
		})
	}
}
