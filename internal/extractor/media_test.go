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
				Source: "BluRay",
				Audio: "Atmos",
				Language: "English",
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
				"S001",
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
			expected:	"BluRay",
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
		expected	string
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
			expected:	"English",
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
			expected:	"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			language := extractLanguage(test.input)
			if language != test.expected {
				t.Errorf("extractLanguage = %v, want %v", language, test.expected)
			}
		})
	}
}

func TestExtractBonus(t *testing.T) {
	tests := []struct{
		name		string
		input		[]string
		expected	string
	}{
		{
			name:		"valid bonus",
			input:		[]string{
				"MY",
				"MOVIE",
				"2025",
				"BEHIND",	
				"THE",	
				"SCENES",	
			},
			expected:	"Behind.The.Scenes",
		},
		{
			name:		"invalid bonus",
			input:		[]string{
				"MY",
				"MOVIE",
				"2025",
				"XBEHIND",	
				"XTHE",	
				"XSCENES",	
			},
			expected:	"",
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
			expected:	"BluRay",
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
func TestParseBonus(t *testing.T) {
	tests := []struct {
		name		string
		input		[]string
		expected	string
	}{
		{
			name:		"valid bonus",
			input:		[]string{
				"BEHIND",
				"THE",
				"SCENES",
			},
			expected:	"Behind.The.Scenes",
		},
		{
			name:		"invalid bonus",
			input:		[]string{
				"XBEHIND",
				"XTHE",
				"XSCENES",
			},
			expected:	"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := parseBonus(test.input)
			if	res != test.expected {
				t.Errorf("parseBonus = %v, want %v", res, test.expected)
			}
		})
	}
}
