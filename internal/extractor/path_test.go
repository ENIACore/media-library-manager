package extractor

import (
	"testing"
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"log/slog"
	"reflect"
)

func TestExtractPath(t *testing.T) {
	logger := slog.Default()
	tests := []struct {
		name			string
		input			string
		expected		metadata.PathInfo	
	}{
		{
			name:			"valid input",
			input:			"/parent/child/my.movie.mp4",
			expected:		metadata.PathInfo{
				Dest: "",
				Source:	"/parent/child/my.movie.mp4",
				Ext: "MP4",
				Type: metadata.Video,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pathInfo := ExtractPath(test.input, logger)

			if !reflect.DeepEqual(pathInfo, test.expected) {
				t.Errorf("ExtractPath = %+v, want %+v", pathInfo, test.expected)
			}
		})
	}
}

func TestExtractType(t *testing.T) {
	tests := []struct {
		name			string
		input			[]string
		expectedType	metadata.ContentType
		expectedExt		string
	}{
		{
			name:			"valid video extension",
			input:			[]string{
				"MY",
				"VIDEO",
				"MP4",
			},
			expectedType: metadata.Video,
			expectedExt: "MP4",
		},
		{
			name:			"valid subtitle extension",
			input:			[]string{
				"MY",
				"SUBTITLE",
				"SRT",
			},
			expectedType: metadata.Subtitle,
			expectedExt: "SRT",
		},
		/* Add audio later when supported
		{
			name:			"valid video extension",
			input:			[]string{
				"MY",
				"VIDEO",
				"MP4",
			},
			expectedType: metadata.Video,
			expectedExt: "MP4",
		},
		*/
		{
			name:			"invalid extension",
			input:			[]string{
				"MY",
				"VIDEO",
				"MP45",
			},
			expectedType: metadata.ContentType(metadata.UnknownType),
			expectedExt: "",
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contentType, ext := extractType(test.input)

			if contentType != test.expectedType {
				t.Errorf("extractType content type = %v, want %v", contentType, test.expectedType)
			}
			if ext != test.expectedExt {
				t.Errorf("extractType ext = %v, want %v", ext, test.expectedExt)
			}
		})
	}
}

func TestParseVideoExt(t *testing.T) {
	tests := []struct {
		name			string
		input			[]string
		expected		string
	}{
		{
			name:		"valid ext",
			input:		[]string{
				"MP4",
			},
			expected: "MP4",
		},
		{
			name:		"invalid ext",
			input:		[]string{
				"MP6",
			},
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ext := parseVideoExt(test.input)
			if ext != test.expected {
				t.Errorf("parseVideoExt = %v, want %v", ext, test.expected)
			}
		})
	}
}

func TestParseSubtitleExt(t *testing.T) {
	tests := []struct {
		name			string
		input			[]string
		expected		string
	}{
		{
			name:		"valid ext",
			input:		[]string{
				"SRT",
			},
			expected: "SRT",
		},
		{
			name:		"invalid ext",
			input:		[]string{
				"SRTT",
			},
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ext := parseSubtitleExt(test.input)
			if ext != test.expected {
				t.Errorf("parseSubtitleExt = %v, want %v", ext, test.expected)
			}
		})
	}
}


func TestParseAudioExt(t *testing.T) {
	tests := []struct {
		name			string
		input			[]string
		expected		string
	}{
		{
			name:		"valid ext",
			input:		[]string{
				"MP3",
			},
			expected: "MP3",
		},
		{
			name:		"invalid ext",
			input:		[]string{
				"MP6",
			},
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ext := parseAudioExt(test.input)
			if ext != test.expected {
				t.Errorf("parseAudioExt = %v, want %v", ext, test.expected)
			}
		})
	}
}
