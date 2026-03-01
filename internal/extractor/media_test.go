package extractor

import (
	"testing"
	"strings"
	"log/slog"
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"reflect"
)

func intPtr(i int) *int {
	return &i
}

func TestExtractMedia(t *testing.T) {
	logger := slog.Default()
	tests := []struct{
		name		string
		input		string
		expected	metadata.MediaInfo	
	}{
		{
			name:		"successful path",
			input:		"/parent/child/example.series.2025.s01.e001.1080p.x.265.bd.rip.atmos.eng.mp4",
			expected:	metadata.MediaInfo{
				Title: []string{
					"EXAMPLE",
					"SERIES",
				},
				Year: intPtr(2025),
				Season: intPtr(1),
				Episode: intPtr(1),
				Resolution: "1080p",
				Codec: "x265",
				Source: "BDRip",
				Audio: "Atmos",
				Language: []string{"English"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			media := ExtractMedia(test.input, logger)
			if !reflect.DeepEqual(media, test.expected) {
				t.Errorf("ExtractMedia = %+v, want %+v", media, test.expected)
			}
		})
	}
}

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name			string
		input			[]string
		expectedTitle	[]string
		expectedIdx		int
	}{
		{
			name: 		"format <title>.<year (optional)>.<misc pattern>",
			input:		[]string{
				"MOVIE",
				"TITLE",
				"2020",
				"UNRATED",
				"1080P",
			},
			expectedTitle: []string{
				"MOVIE",
				"TITLE",
			},
		},
		{
			name: 		"format <title>.<year (optional)>.<resolution, codec, source, or audio>",
			input:		[]string{
				"MOVIE",
				"TITLE",
				"2020",
				"1080P",
			},
			expectedTitle: []string{
				"MOVIE",
				"TITLE",
			},
		},
		{
			name: 		"format <title>..<resolution, codec, source, or audio>",
			input:		[]string{
				"MY",
				"MOVIE",
				"TITLE",
				"1080P",
				"MP4",
			},
			expectedTitle: []string{
				"MY",
				"MOVIE",
				"TITLE",
			},
		},
		{
			name: 		"format <title>.<year (optional)>.<season or ep>",
			input:		[]string{
				"MOVIE",
				"TITLE",
				"2020",
				"S4",
			},
			expectedTitle: []string{
				"MOVIE",
				"TITLE",
			},
		},
		{
			name: 		"format <title>.<year (optional)>.<file ext>",
			input:		[]string{
				"MOVIE",
				"TITLE",
				"2020",
				"MP4",
			},
			expectedTitle: []string{
				"MOVIE",
				"TITLE",
			},
		},
		{
			name: 		"format <title>.<year (optional)>",
			input:		[]string{
				"MOVIE",
				"TITLE",
				"2020",
			},
			expectedTitle: []string{
				"MOVIE",
				"TITLE",
			},
		},
		/*
			* Tests for titles containing years
		*/
		{
			name: 		"format <title>.<year in title>.<title>.<year>.<terminator>",
			input:		[]string{
				"MOVIE",
				"1999",
				"TITLE",
				"2020",
				"1080P",
			},
			expectedTitle: []string{
				"MOVIE",
				"1999",
				"TITLE",
			},
		},
		{
			name: 		"format <title>.<year in title>.<year>.<terminator>",
			input:		[]string{
				"MOVIE",
				"TITLE",
				"1999",
				"2020",
				"1080P",
			},
			expectedTitle: []string{
				"MOVIE",
				"TITLE",
				"1999",
			},
		},
		{
			name: 		"format <title>.<year in title>.<title>.<terminator>",
			input:		[]string{
				"MOVIE",
				"1999",
				"TITLE",
				"1080P",
			},
			expectedTitle: []string{
				"MOVIE",
				"1999",
				"TITLE",
			},
		},
		{
			// IMPORTANT: Extractor will not be able to differentiate year
			name: 		"format <title>.<year in title>.<terminator>",
			input:		[]string{
				"MOVIE",
				"TITLE",
				"1999",
				"1080P",
			},
			expectedTitle: []string{
				"MOVIE",
				"TITLE",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			title := extractTitle(test.input)	
			if strings.Join(title, "") != strings.Join(test.expectedTitle, "") && len(title) != len(test.expectedTitle) {
				t.Errorf("extractTitle title = %v, want %v", title, test.expectedTitle)
			}


		})
	}
}

func TestExtractYear(t *testing.T) {
	tests := []struct{
		name			string
		input			[]string
		expectedYear	*int
	}{
		{
			name: 		"format <title>.<year (optional)>.<misc pattern>",
			input:		[]string{
				"MOVIE",
				"TITLE",
				"2020",
				"UNRATED",
				"1080P",
			},
			expectedYear: intPtr(2020),
		},
		{
			name: "successful <...>.<year (optional)>.<resolution, codec, source, or audio>",
			input:	[]string{
				"2020",
				"1080P",
			},
			expectedYear: intPtr(2020),
		},
		{
			name: "successful <...>.<year (optional)>.<season or ep>",
			input:	[]string{
				"2020",
				"S04",
			},
			expectedYear: intPtr(2020),
		},
		{
			name: "successful <...>.<year (optional)>.<file ext>",
			input:	[]string{
				"2020",
				"MP4",
			},
			expectedYear: intPtr(2020),
		},
		{
			name: "successful <...>.<year (optional)>",
			input:	[]string{
				"2020",
			},
			expectedYear: intPtr(2020),
		},
		{
			name: "missing year",
			input:	[]string{
				"1080P",
				"MP4",
			},
			expectedYear: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			year := extractYear(test.input)
			if !reflect.DeepEqual(year, test.expectedYear) {
    			t.Errorf("extractYear year = %v, want %v", year, test.expectedYear)
			}
		})
	}
}


func TestExtractSeason(t *testing.T) {
	tests := []struct{
		name		string
		input		[]string
		expected	*int
	}{
		{
			name:		"season without number",
			input:		[]string{
				"MY",
				"MOVIE",
				"TITLE",
				"SEASON",
				"EPISODE",
				"1080P",
				"MP4",
			},
			expected: 	intPtr(0),
		},
		{
			name:		"season with number",
			input:		[]string{
				"MY",
				"MOVIE",
				"TITLE",
				"S01",  // Changed from S001
				"EPISODE",
				"1080P",
				"MP4",
			},
			expected: 	intPtr(1),
		},
		{
			name:		"no season",
			input:		[]string{
				"MY",
				"MOVIE",
				"TITLE",
				"EPISODE",
				"1080P",
				"MP4",
			},
			expected: 	nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			season := extractSeason(test.input)

			if !reflect.DeepEqual(season, test.expected) {
    			t.Errorf("extractSeason = %v, want %v", season, test.expected)
			}
		})
	}
}

func TestExtractEpisode(t *testing.T) {
	tests := []struct{
		name		string
		input		[]string
		expected	*int
	}{
		{
			name:		"ep without number",
			input:		[]string{
				"MY",
				"MOVIE",
				"TITLE",
				"SEASON",
				"EPISODE",
				"1080P",
				"MP4",
			},
			expected: 	intPtr(0),
		},
		{
			name:		"ep with number",
			input:		[]string{
				"MY",
				"MOVIE",
				"TITLE",
				"SEASON",
				"EP001",
				"1080P",
				"MP4",
			},
			expected: 	intPtr(1),
		},
		{
			name:		"no ep",
			input:		[]string{
				"MY",
				"MOVIE",
				"TITLE",
				"SEASON",
				"1080P",
				"MP4",
			},
			expected: 	nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ep := extractEpisode(test.input)

			if !reflect.DeepEqual(ep, test.expected) {
    			t.Errorf("extractEpisode = %v, want %v", ep, test.expected)
			}
		})
	}
}

func TestExtractResolution(t *testing.T) {
	tests := []struct{
		name		string
		input		[]string
		expected	string
	}{
		{
			name:		"resolution without capture group",
			input:		[]string{
				"MY",
				"MOVIE",
				"1080P",
				"MP4",
			},
			expected:	"1080p",
		},
		{
			name:		"no resolution",
			input:		[]string{
				"MY",
				"MOVIE",
				"MP4",
			},
			expected:	"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := extractResolution(test.input)	
			if res != test.expected {
				t.Errorf("extractResolution = %v, want %v", res, test.expected)
			}
		})
	}
}

func TestExtractCodec(t *testing.T) {
	tests := []struct{
		name		string
		input		[]string
		expected	string
	}{
		{
			name:		"codec without capture group",
			input:		[]string{
				"MY",
				"MOVIE",
				"X265",
				"MP4",
			},
			expected:	"x265",
		},
		{
			name:		"no codec",
			input:		[]string{
				"MY",
				"MOVIE",
				"MP4",
			},
			expected:	"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := extractCodec(test.input)	
			if res != test.expected {
				t.Errorf("extractCodec = %v, want %v", res, test.expected)
			}
		})
	}
}

func TestExtractSource(t *testing.T) {
	tests := []struct{
		name		string
		input		[]string
		expected	string
	}{
		{
			name:		"source without capture group",
			input:		[]string{
				"MOVIE",
				"TITLE",
				"BD",
				"RIP",
				"MP4",
			},
			expected:	"BDRip",
		},
		{
			name:		"no source",
			input:		[]string{
				"MY",
				"MOVIE",
				"MP4",
			},
			expected:	"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := extractSource(test.input)	
			if res != test.expected {
				t.Errorf("extractSource = %v, want %v", res, test.expected)
			}
		})
	}
}

func TestExtractAudio(t *testing.T) {
	tests := []struct{
		name		string
		input		[]string
		expected	string
	}{
		{
			name:		"audio without capture group",
			input:		[]string{
				"MOVIE",
				"TITLE",
				"DOLBY",
				"ATMOS",
				"MP4",
			},
			expected:	"Atmos",
		},
		{
			name:		"no audio",
			input:		[]string{
				"MY",
				"MOVIE",
				"MP4",
			},
			expected:	"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := extractAudio(test.input)	
			if res != test.expected {
				t.Errorf("extractAudio = %v, want %v", res, test.expected)
			}
		})
	}
}

func TestExtractLanguage(t *testing.T) {
	tests := []struct{
		name		string
		input		[]string
		expected	[]string
	}{
		{
			name:		"valid language",
			input:		[]string{
				"MOVIE",
				"TITLE",
				"2020",
				"1080P",
				"ENG",
				"SRT",
			},
			expected:	[]string{"English"},
		},
		{
			name:		"missing language",
			input:		[]string{
				"MOVIE",
				"TITLE",
				"2020",
				"1080P",
				"SRT",
			},
			expected:	[]string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			language := extractLanguage(test.input)
			if !reflect.DeepEqual(language, test.expected) {
				t.Errorf("extractLanguage = %v, want %v", language, test.expected)
			}
		})
	}
}

func TestParseResolution(t *testing.T) {
	tests := []struct {
		name		string
		input		[]string
		expected	string
	}{
		{
			name:		"valid resolution",
			input:		[]string{
				"2160I",
			},
			expected:	"4K",
		},
		{
			name:		"invalid resolution",
			input:		[]string{
				"2160X",
			},
			expected:	"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := parseResolution(test.input)
			if	res != test.expected {
				t.Errorf("parseResolution = %v, want %v", res, test.expected)
			}
		})
	}
}
func TestParseCodec(t *testing.T) {
	tests := []struct {
		name		string
		input		[]string
		expected	string
	}{
		{
			name:		"codec without seperators",
			input:		[]string{
				"AOV1",
			},
			expected:	"AV1",
		},
		{
			name:		"codec with seperators",
			input:		[]string{
				"SVT",
				"AV1",
			},
			expected:	"AV1",
		},
		{
			name:		"invalid codec with seperators",
			input:		[]string{
				"INCORRECT",
				"SVT",
				"AV1",
			},
			expected:	"",
		},
		{
			name:		"invalid codec without seperators",
			input:		[]string{
				"INCORRECTSVT",
			},
			expected:	"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := parseCodec(test.input)
			if	res != test.expected {
				t.Errorf("parseCodec = %v, want %v", res, test.expected)
			}
		})
	}
}


func TestParseSource(t *testing.T) {
	tests := []struct {
		name		string
		input		[]string
		expected	string
	}{
		{
			name:		"source without seperators",
			input:		[]string{
				"BLURAY",
			},
			expected:	"BluRay",
		},
		{
			name:		"source with seperators",
			input:		[]string{
				"BD",
				"RIP",
			},
			expected:	"BDRip",
		},
		{
			name:		"invalid source with seperators",
			input:		[]string{
				"INCORRECT",
				"BD",
				"RIP",
			},
			expected:	"",
		},
		{
			name:		"invalid source without seperators",
			input:		[]string{
				"INCORRECTBLURAY",
			},
			expected:	"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := parseSource(test.input)
			if	res != test.expected {
				t.Errorf("parseSource = %v, want %v", res, test.expected)
			}
		})
	}
}

func TestParseAudio(t *testing.T) {
	tests := []struct {
		name		string
		input		[]string
		expected	string
	}{
		{
			name:		"audio without seperators",
			input:		[]string{
				"DTSX",
			},
			expected:	"DTS-X",
		},
		{
			name:		"audio with seperators",
			input:		[]string{
				"DTS",
				"X",
			},
			expected:	"DTS-X",
		},
		{
			name:		"invalid audio without seperators",
			input:		[]string{
				"INCORRECTDTSX",
			},
			expected:	"",
		},
		{
			name:		"invalid audio with seperators",
			input:		[]string{
				"INCORRECT",
				"DTSX",
			},
			expected:	"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := parseAudio(test.input)
			if	res != test.expected {
				t.Errorf("parseAudio = %v, want %v", res, test.expected)
			}
		})
	}
}


func TestParseYear(t *testing.T) {
	tests := []struct {
		name			string
		input			string
		expectedValue	*int
	}{
		{
			name:			"valid year",		
			input:			"2000",
			expectedValue:	intPtr(2000),
		},
		{
			name:			"too early year",			
			input: 			"1900",
			expectedValue: 	nil,
		},
		{
			name:			"too late year",		
			input: 			"3000",
			expectedValue: 	nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			year := parseYear(test.input)
			if !reflect.DeepEqual(year, test.expectedValue) {
    			t.Errorf("parseYear = %v, want %v", year, test.expectedValue)
			}
		})
	}
}

func TestParseSeason(t *testing.T) {
	tests := []struct {
		name			string
		input			[]string
		expected		*int
	}{
		{
			name:		"valid season with number",
			input:		[]string{
				"S04",
				"1080P",
				"X265",
				"MP4",
			},
			expected: intPtr(4),
		},
		{
			name:		"valid season without number",
			input:		[]string{
				"SEASON",
				"1080P",
				"X265",
				"MP4",
			},
			expected: intPtr(0),
		},
		{
			name:		"invalid season",
			input:		[]string{
				"INCORRECTS04",
				"1080P",
				"X265",
				"MP4",
			},
			expected: nil,
		},
		{
			name:		"season at second segment",
			input:		[]string{
				"INCORRECT",
				"S04",
				"1080P",
				"X265",
				"MP4",
			},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			season := parseSeason(test.input)
			if !reflect.DeepEqual(season, test.expected) {
    			t.Errorf("parseSeason = %v, want %v", season, test.expected)
			}
		})
	}
}

func TestParseEpisode(t *testing.T) {
	tests := []struct {
		name			string
		input			[]string
		expected		*int
	}{
		{
			name:		"valid episode with number",
			input:		[]string{
				"E04",
				"1080P",
				"X265",
				"MP4",
			},
			expected: intPtr(4),
		},
		{
			name:		"valid episode without number",
			input:		[]string{
				"EPISODE",
				"1080P",
				"X265",
				"MP4",
			},
			expected: intPtr(0),
		},
		{
			name:		"invalid episode",
			input:		[]string{
				"INCORRECTE04",
				"1080P",
				"X265",
				"MP4",
			},
			expected: nil,
		},
		{
			name:		"episode at second segment",
			input:		[]string{
				"INCORRECT",
				"E04",
				"1080P",
				"X265",
				"MP4",
			},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ep := parseEpisode(test.input)
			if !reflect.DeepEqual(ep, test.expected) {
    			t.Errorf("parseEpisode = %v, want %v", ep, test.expected)
			}
		})
	}
}


func TestParseLanguage(t *testing.T) {
	tests := []struct {
		name			string
		input			[]string
		expected		string
	}{
		{
			name:		"valid language",
			input:		[]string{
				"ENG",
				"SRT",
			},
			expected: "English",
		},
		{
			name:		"no language",
			input:		[]string{
				"1080P",
				"SRT",
			},
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			language := parseLanguage(test.input)
			if language != test.expected {
				t.Errorf("parseLanguage = %v, want %v", language, test.expected)
			}
		})
	}
}

func TestSanitizePrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name: "www.domain.org prefix",
			input: []string{
				"WWW",
				"UINDEX",
				"ORG",
				"GOON",
				"LAST",
				"OF",
				"THE",
				"ENFORCERS",
			},
			expected: []string{
				"GOON",
				"LAST",
				"OF",
				"THE",
				"ENFORCERS",
			},
		},
		{
			name: "domain.org without www",
			input: []string{
				"UINDEX",
				"ORG",
				"THE",
				"AMAZING",
				"DIGITAL",
				"CIRCUS",
			},
			expected: []string{
				"THE",
				"AMAZING",
				"DIGITAL",
				"CIRCUS",
			},
		},
		{
			name: "domain.com prefix",
			input: []string{
				"EXAMPLE",
				"COM",
				"MOVIE",
				"TITLE",
				"2025",
			},
			expected: []string{
				"MOVIE",
				"TITLE",
				"2025",
			},
		},
		{
			name: "no website prefix",
			input: []string{
				"GOON",
				"LAST",
				"OF",
				"THE",
				"ENFORCERS",
			},
			expected: []string{
				"GOON",
				"LAST",
				"OF",
				"THE",
				"ENFORCERS",
			},
		},
		{
			name: "false positive - organization in title",
			input: []string{
				"THE",
				"ORGANIZATION",
				"MOVIE",
			},
			expected: []string{
				"THE",
				"ORGANIZATION",
				"MOVIE",
			},
		},
		{
			name: "empty segments",
			input:    []string{},
			expected: []string{},
		},
		{
			name: "single segment",
			input: []string{
				"MOVIE",
			},
			expected: []string{
				"MOVIE",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := sanitizePrefix(test.input)
			if len(result) != len(test.expected) {
				t.Errorf("sanitizePrefix() length = %v, want %v", len(result), len(test.expected))
				return
			}
			for i := range result {
				if result[i] != test.expected[i] {
					t.Errorf("sanitizePrefix()[%d] = %v, want %v", i, result[i], test.expected[i])
				}
			}
		})
	}
}

func TestParseWebsite(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name: "www.domain.org",
			input: []string{
				"WWW",
				"UINDEX",
				"ORG",
			},
			expected: "WWW.UINDEX.ORG",
		},
		{
			name: "domain.com",
			input: []string{
				"EXAMPLE",
				"COM",
			},
			expected: "EXAMPLE.COM",
		},
		{
			name: "domain.net",
			input: []string{
				"SITE",
				"NET",
			},
			expected: "SITE.NET",
		},
		{
			name: "www.domain.io",
			input: []string{
				"WWW",
				"MYSITE",
				"IO",
			},
			expected: "WWW.MYSITE.IO",
		},
		{
			name: "no website pattern",
			input: []string{
				"MOVIE",
				"TITLE",
			},
			expected: "",
		},
		{
			name: "invalid TLD",
			input: []string{
				"EXAMPLE",
				"INVALID",
			},
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := parseWebsite(test.input)
			if result != test.expected {
				t.Errorf("parseWebsite() = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestParseEdition(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name: "remastered edition",
			input: []string{
				"REMASTERED",
				"1080P",
			},
			expected: "Remastered",
		},
		{
			name: "extended cut",
			input: []string{
				"EXTENDED",
				"CUT",
				"1080P",
			},
			expected: "Extended",
		},
		{
			name: "unrated edition",
			input: []string{
				"UNRATED",
				"1080P",
			},
			expected: "Unrated",
		},
		{
			name: "directors cut",
			input: []string{
				"DIRECTORS",
				"CUT",
				"1080P",
			},
			expected: "DirectorsCut",
		},
		{
			name: "theatrical edition",
			input: []string{
				"THEATRICAL",
				"1080P",
			},
			expected: "Theatrical",
		},
		{
			name: "imax edition",
			input: []string{
				"IMAX",
				"1080P",
			},
			expected: "IMAX",
		},
		{
			name: "3d edition",
			input: []string{
				"3D",
				"1080P",
			},
			expected: "3D",
		},
		{
			name: "criterion edition",
			input: []string{
				"CRITERION",
				"1080P",
			},
			expected: "Criterion",
		},
		{
			name: "final cut",
			input: []string{
				"FINAL",
				"CUT",
				"1080P",
			},
			expected: "FinalCut",
		},
		{
			name: "no edition",
			input: []string{
				"1080P",
				"X265",
			},
			expected: "",
		},
		{
			name: "invalid edition - not at leftmost",
			input: []string{
				"INCORRECT",
				"REMASTERED",
				"1080P",
			},
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := parseEdition(test.input)
			if result != test.expected {
				t.Errorf("parseEdition() = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestExtractEdition(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name: "edition in middle",
			input: []string{
				"MOVIE",
				"TITLE",
				"2020",
				"REMASTERED",
				"1080P",
				"MP4",
			},
			expected: "Remastered",
		},
		{
			name: "extended cut with separators",
			input: []string{
				"MY",
				"MOVIE",
				"EXTENDED",
				"CUT",
				"1080P",
			},
			expected: "Extended",
		},
		{
			name: "directors cut",
			input: []string{
				"MOVIE",
				"TITLE",
				"DIRECTORS",
				"CUT",
				"BLURAY",
			},
			expected: "DirectorsCut",
		},
		{
			name: "imax edition",
			input: []string{
				"MOVIE",
				"IMAX",
				"1080P",
			},
			expected: "IMAX",
		},
		{
			name: "uncut edition",
			input: []string{
				"MOVIE",
				"UNCUT",
				"1080P",
			},
			expected: "Uncut",
		},
		{
			name: "no edition",
			input: []string{
				"MY",
				"MOVIE",
				"1080P",
				"MP4",
			},
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			edition := extractEdition(test.input)
			if edition != test.expected {
				t.Errorf("extractEdition() = %v, want %v", edition, test.expected)
			}
		})
	}
}

func TestExtractBTS(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name: "valid bts",
			input: []string{
				"MY",
				"MOVIE",
				"2025",
				"BEHIND",
				"THE",
				"SCENES",
			},
			expected: "Behind.The.Scenes",
		},
		{
			name: "invalid bts",
			input: []string{
				"MY",
				"MOVIE",
				"2025",
				"XBEHIND",
				"XTHE",
				"XSCENES",
			},
			expected: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bts := extractBTS(test.input)
			if bts != test.expected {
				t.Errorf("extractBTS = %v, want %v", bts, test.expected)
			}
		})
	}
}

func TestParseBTS(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name: "valid bts",
			input: []string{
				"BEHIND",
				"THE",
				"SCENES",
			},
			expected: "Behind.The.Scenes",
		},
		{
			name: "invalid bts",
			input: []string{
				"XBEHIND",
				"XTHE",
				"XSCENES",
			},
			expected: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := parseBTS(test.input)
			if res != test.expected {
				t.Errorf("parseBTS = %v, want %v", res, test.expected)
			}
		})
	}
}

func TestExtractDS(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name: "valid ds",
			input: []string{
				"MY",
				"MOVIE",
				"2025",
				"DELETED",
				"SCENES",
			},
			expected: "Deleted.Scenes",
		},
		{
			name: "invalid ds",
			input: []string{
				"MY",
				"MOVIE",
				"2025",
				"XDELETED",
				"XSCENES",
			},
			expected: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ds := extractDS(test.input)
			if ds != test.expected {
				t.Errorf("extractDS = %v, want %v", ds, test.expected)
			}
		})
	}
}

func TestParseDS(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name: "valid ds",
			input: []string{
				"DELETED",
				"SCENES",
			},
			expected: "Deleted.Scenes",
		},
		{
			name: "invalid ds",
			input: []string{
				"XDELETED",
				"XSCENES",
			},
			expected: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := parseDS(test.input)
			if res != test.expected {
				t.Errorf("parseDS = %v, want %v", res, test.expected)
			}
		})
	}
}

func TestExtractBonus(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name: "valid bonus",
			input: []string{
				"MY",
				"MOVIE",
				"2025",
				"FEATURETTE",
			},
			expected: "Featurette",
		},
		{
			name: "invalid bonus",
			input: []string{
				"MY",
				"MOVIE",
				"2025",
				"XFEATURETTE",
			},
			expected: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bonus := extractBonus(test.input)
			if bonus != test.expected {
				t.Errorf("extractBonus = %v, want %v", bonus, test.expected)
			}
		})
	}
}

func TestParseBonus(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name: "valid bonus",
			input: []string{
				"FEATURETTE",
			},
			expected: "Featurette",
		},
		{
			name: "invalid bonus",
			input: []string{
				"XFEATURETTE",
			},
			expected: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := parseBonus(test.input)
			if res != test.expected {
				t.Errorf("parseBonus = %v, want %v", res, test.expected)
			}
		})
	}
}
