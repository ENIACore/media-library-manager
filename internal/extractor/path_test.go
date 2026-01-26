package extractor

import (
	"log/slog"
	"testing"

	"github.com/ENIACore/media_library_manager/internal/metadata"
)

func TestExtractPath(t *testing.T) {
	logger := slog.Default()
	tests := []struct {
		name     string
		input    string
		expected metadata.PathInfo
	}{
		{
			name:  "valid video file",
			input: "/parent/child/my.movie.mp4",
			expected: metadata.PathInfo{
				Dest:   "",
				Source: "/parent/child/my.movie.mp4",
				Ext:    "MP4",
				Type:   metadata.Video,
				IsDir:  false,
			},
		},
		{
			name:  "valid subtitle file",
			input: "/parent/child/subtitle.srt",
			expected: metadata.PathInfo{
				Dest:   "",
				Source: "/parent/child/subtitle.srt",
				Ext:    "SRT",
				Type:   metadata.Subtitle,
				IsDir:  false,
			},
		},
		{
			name:  "nfo file should be unknown",
			input: "/parent/child/movie.mkv.nfo",
			expected: metadata.PathInfo{
				Dest:   "",
				Source: "/parent/child/movie.mkv.nfo",
				Ext:    "NFO",
				Type:   metadata.UnknownType,
				IsDir:  false,
			},
		},
		{
			name:  "txt file should be unknown",
			input: "/parent/child/readme.txt",
			expected: metadata.PathInfo{
				Dest:   "",
				Source: "/parent/child/readme.txt",
				Ext:    "TXT",
				Type:   metadata.UnknownType,
				IsDir:  false,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pathInfo := ExtractPath(test.input, logger)
			
			if pathInfo.Source != test.expected.Source {
				t.Errorf("Source = %v, want %v", pathInfo.Source, test.expected.Source)
			}
			if pathInfo.Ext != test.expected.Ext {
				t.Errorf("Ext = %v, want %v", pathInfo.Ext, test.expected.Ext)
			}
			if pathInfo.Type != test.expected.Type {
				t.Errorf("Type = %v, want %v", pathInfo.Type, test.expected.Type)
			}
		})
	}
}

func TestExtractType(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType metadata.ContentType
	}{
		{
			name:         "valid video extension",
			input:        "MP4",
			expectedType: metadata.Video,
		},
		{
			name:         "valid subtitle extension",
			input:        "SRT",
			expectedType: metadata.Subtitle,
		},
		{
			name:         "invalid extension",
			input:        "NFO",
			expectedType: metadata.UnknownType,
		},
		{
			name:         "empty extension",
			input:        "",
			expectedType: metadata.UnknownType,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contentType := extractType(test.input)
			if contentType != test.expectedType {
				t.Errorf("extractType content type = %v, want %v", contentType, test.expectedType)
			}
		})
	}
}

func TestParseVideoExt(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "valid ext",
			input:    []string{"MP4"},
			expected: "MP4",
		},
		{
			name:     "invalid ext",
			input:    []string{"MP6"},
			expected: "",
		},
		{
			name:     "mkv ext",
			input:    []string{"MKV"},
			expected: "MKV",
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
		name     string
		input    []string
		expected string
	}{
		{
			name:     "valid ext",
			input:    []string{"SRT"},
			expected: "SRT",
		},
		{
			name:     "invalid ext",
			input:    []string{"SRTT"},
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
		name     string
		input    []string
		expected string
	}{
		{
			name:     "valid ext",
			input:    []string{"MP3"},
			expected: "MP3",
		},
		{
			name:     "invalid ext",
			input:    []string{"MP6"},
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
