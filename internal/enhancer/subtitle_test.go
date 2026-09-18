package enhancer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ENIACore/media_library_manager/internal/config"
	"github.com/ENIACore/media_library_manager/internal/metadata"
)

func noopLogger() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

// testSession creates a Session wired to ts with ts's client.
func testSession(ts *httptest.Server) *Session {
	return &Session{
		Token:     "test-token",
		BaseURL:   ts.URL,
		apiKey:    "test-api-key",
		userAgent: "test-agent/1.0",
		client:    ts.Client(),
	}
}

func intPtr(n int) *int { return &n }

// --- pickBest ---

func TestPickBest(t *testing.T) {
	tests := []struct {
		name    string
		subs    []osSubtitle
		wantNil bool
		wantID  int
	}{
		{
			name:    "nil input",
			wantNil: true,
		},
		{
			name:    "all entries missing files",
			subs:    []osSubtitle{{Attributes: osSubtitleAttrs{Files: nil}}},
			wantNil: true,
		},
		{
			name: "single entry",
			subs: []osSubtitle{
				{Attributes: osSubtitleAttrs{Files: []osFile{{FileID: 7}}}},
			},
			wantID: 7,
		},
		{
			name: "non-HI beats HI regardless of download count",
			subs: []osSubtitle{
				{Attributes: osSubtitleAttrs{HearingImpaired: true, DownloadCount: 99999, Files: []osFile{{FileID: 1}}}},
				{Attributes: osSubtitleAttrs{HearingImpaired: false, DownloadCount: 1, Files: []osFile{{FileID: 2}}}},
			},
			wantID: 2,
		},
		{
			name: "among non-HI picks highest download count",
			subs: []osSubtitle{
				{Attributes: osSubtitleAttrs{DownloadCount: 100, Files: []osFile{{FileID: 1}}}},
				{Attributes: osSubtitleAttrs{DownloadCount: 500, Files: []osFile{{FileID: 2}}}},
				{Attributes: osSubtitleAttrs{DownloadCount: 300, Files: []osFile{{FileID: 3}}}},
			},
			wantID: 2,
		},
		{
			name: "skips entries with no files",
			subs: []osSubtitle{
				{Attributes: osSubtitleAttrs{DownloadCount: 999, Files: nil}},
				{Attributes: osSubtitleAttrs{DownloadCount: 1, Files: []osFile{{FileID: 5}}}},
			},
			wantID: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pickBest(tt.subs)
			if tt.wantNil {
				if got != nil {
					t.Errorf("pickBest() = %+v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("pickBest() = nil, want non-nil")
			}
			if got.Attributes.Files[0].FileID != tt.wantID {
				t.Errorf("FileID = %d, want %d", got.Attributes.Files[0].FileID, tt.wantID)
			}
		})
	}
}

// --- parseRetryAfter ---

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		header string
		want   time.Duration
	}{
		{"30", 30 * time.Second},
		{"120", 120 * time.Second},
		{"", 60 * time.Second},    // missing → default
		{"abc", 60 * time.Second}, // unparseable → default
		{"-1", 60 * time.Second},  // negative → default
		{"0", 60 * time.Second},   // zero → default
	}
	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			if got := parseRetryAfter(tt.header); got != tt.want {
				t.Errorf("parseRetryAfter(%q) = %v, want %v", tt.header, got, tt.want)
			}
		})
	}
}

// --- osLogin ---

func TestOsLogin_success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body osLoginBody
		json.NewDecoder(r.Body).Decode(&body)
		if body.Username != "user" || body.Password != "pass" {
			t.Errorf("unexpected credentials: %+v", body)
		}
		json.NewEncoder(w).Encode(osLoginResponse{Token: "jwt-abc", BaseURL: ""})
	}))
	defer ts.Close()

	resp, err := osLogin(ts.Client(), ts.URL+"/login", "key", "agent", "user", "pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token != "jwt-abc" {
		t.Errorf("Token = %q, want jwt-abc", resp.Token)
	}
}

func TestOsLogin_unauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	_, err := osLogin(ts.Client(), ts.URL+"/login", "key", "agent", "user", "pass")
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("error = %v, want ErrUnauthorized", err)
	}
}

func TestOsLogin_missingToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(osLoginResponse{Token: ""})
	}))
	defer ts.Close()

	_, err := osLogin(ts.Client(), ts.URL+"/login", "key", "agent", "user", "pass")
	if err == nil {
		t.Fatal("expected error for missing token, got nil")
	}
}

// --- searchSubtitle ---

func TestSearchSubtitle_movie(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("tmdb_id") != "12345" {
			t.Errorf("tmdb_id = %q, want 12345", q.Get("tmdb_id"))
		}
		if q.Get("type") != "movie" {
			t.Errorf("type = %q, want movie", q.Get("type"))
		}
		if q.Get("languages") != "en" {
			t.Errorf("languages = %q, want en", q.Get("languages"))
		}
		json.NewEncoder(w).Encode(osSearchResponse{
			Data: []osSubtitle{{
				Attributes: osSubtitleAttrs{
					DownloadCount: 500,
					Files:         []osFile{{FileID: 42}},
				},
			}},
		})
	}))
	defer ts.Close()

	entry := &metadata.Entry{
		Role:      metadata.MovieFile,
		MediaInfo: metadata.MediaInfo{TMDBid: 12345},
	}

	fileID, err := testSession(ts).searchSubtitle(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fileID != 42 {
		t.Errorf("fileID = %d, want 42", fileID)
	}
}

func TestSearchSubtitle_episode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("parent_tmdb_id") != "99" {
			t.Errorf("parent_tmdb_id = %q, want 99", q.Get("parent_tmdb_id"))
		}
		if q.Get("type") != "episode" {
			t.Errorf("type = %q, want episode", q.Get("type"))
		}
		if q.Get("season_number") != "2" {
			t.Errorf("season_number = %q, want 2", q.Get("season_number"))
		}
		if q.Get("episode_number") != "5" {
			t.Errorf("episode_number = %q, want 5", q.Get("episode_number"))
		}
		json.NewEncoder(w).Encode(osSearchResponse{
			Data: []osSubtitle{{
				Attributes: osSubtitleAttrs{Files: []osFile{{FileID: 7}}},
			}},
		})
	}))
	defer ts.Close()

	entry := &metadata.Entry{
		Role: metadata.EpisodeFile,
		MediaInfo: metadata.MediaInfo{
			TMDBid:  99,
			Season:  intPtr(2),
			Episode: intPtr(5),
		},
	}

	fileID, err := testSession(ts).searchSubtitle(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fileID != 7 {
		t.Errorf("fileID = %d, want 7", fileID)
	}
}

func TestSearchSubtitle_noResults(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(osSearchResponse{Data: nil})
	}))
	defer ts.Close()

	entry := &metadata.Entry{Role: metadata.MovieFile, MediaInfo: metadata.MediaInfo{TMDBid: 1}}
	_, err := testSession(ts).searchSubtitle(entry)
	if err == nil {
		t.Fatal("expected error for empty results, got nil")
	}
}

func TestSearchSubtitle_unauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	entry := &metadata.Entry{Role: metadata.MovieFile, MediaInfo: metadata.MediaInfo{TMDBid: 1}}
	_, err := testSession(ts).searchSubtitle(entry)
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("error = %v, want ErrUnauthorized", err)
	}
}

// --- requestDownload ---

func TestRequestDownload(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body osDownloadBody
		json.NewDecoder(r.Body).Decode(&body)
		if body.FileID != 42 {
			t.Errorf("FileID = %d, want 42", body.FileID)
		}
		json.NewEncoder(w).Encode(osDownloadResponse{
			Link:      "https://example.com/file.srt",
			Remaining: 19,
		})
	}))
	defer ts.Close()

	link, remaining, err := testSession(ts).requestDownload(42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link != "https://example.com/file.srt" {
		t.Errorf("link = %q, want https://example.com/file.srt", link)
	}
	if remaining != 19 {
		t.Errorf("remaining = %d, want 19", remaining)
	}
}

func TestRequestDownload_missingLink(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(osDownloadResponse{Link: "", Remaining: 20})
	}))
	defer ts.Close()

	_, _, err := testSession(ts).requestDownload(1)
	if err == nil {
		t.Fatal("expected error for missing link, got nil")
	}
}

func TestRequestDownload_unauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	_, _, err := testSession(ts).requestDownload(1)
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("error = %v, want ErrUnauthorized", err)
	}
}

// --- downloadSubtitle ---

const testSRTContent = "1\n00:00:01,000 --> 00:00:04,000\nHello, world.\n"

func TestDownloadSubtitle_writesFile(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, testSRTContent)
	}))
	defer ts.Close()

	dest := filepath.Join(t.TempDir(), "test.English.srt")
	if err := downloadSubtitle(ts.URL+"/file.srt", dest, ts.Client()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("could not read written file: %v", err)
	}
	if string(got) != testSRTContent {
		t.Errorf("content = %q, want %q", got, testSRTContent)
	}
}

func TestDownloadSubtitle_atomicOnFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	dest := filepath.Join(t.TempDir(), "test.English.srt")
	err := downloadSubtitle(ts.URL+"/file.srt", dest, ts.Client())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, statErr := os.Stat(dest); !errors.Is(statErr, os.ErrNotExist) {
		t.Error("dest file should not exist after failed download")
	}
	if _, statErr := os.Stat(dest + ".tmp"); !errors.Is(statErr, os.ErrNotExist) {
		t.Error("temp file should be cleaned up after failed download")
	}
}

// --- FetchSubtitle end-to-end ---

func TestFetchSubtitle_movie(t *testing.T) {
	// serverURL is captured by the /download handler after ts is created.
	var serverURL string

	mux := http.NewServeMux()
	mux.HandleFunc("/subtitles", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(osSearchResponse{
			Data: []osSubtitle{{
				Attributes: osSubtitleAttrs{Files: []osFile{{FileID: 1}}},
			}},
		})
	})
	mux.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(osDownloadResponse{
			Link:      serverURL + "/srt",
			Remaining: 20,
		})
	})
	mux.HandleFunc("/srt", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, testSRTContent)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()
	serverURL = ts.URL

	dest := filepath.Join(t.TempDir(), "movie.English.srt")
	entry := &metadata.Entry{
		Role:      metadata.MovieFile,
		MediaInfo: metadata.MediaInfo{TMDBid: 123},
		FileInfo:  metadata.FileInfo{DestPath: dest},
	}

	if err := FetchSubtitle(entry, testSession(ts), &config.Config{DryRun: false}, noopLogger()); err != nil {
		t.Fatalf("FetchSubtitle() error: %v", err)
	}

	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("subtitle file not written: %v", err)
	}
	if string(got) != testSRTContent {
		t.Errorf("content = %q, want %q", got, testSRTContent)
	}
}

func TestFetchSubtitle_dryRun(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/subtitles", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(osSearchResponse{
			Data: []osSubtitle{{Attributes: osSubtitleAttrs{Files: []osFile{{FileID: 1}}}}},
		})
	})
	mux.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(osDownloadResponse{Link: "http://example.com/file.srt", Remaining: 5})
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	dest := filepath.Join(t.TempDir(), "movie.English.srt")
	entry := &metadata.Entry{
		Role:      metadata.MovieFile,
		MediaInfo: metadata.MediaInfo{TMDBid: 1},
		FileInfo:  metadata.FileInfo{DestPath: dest},
	}

	if err := FetchSubtitle(entry, testSession(ts), &config.Config{DryRun: true}, noopLogger()); err != nil {
		t.Fatalf("FetchSubtitle() error: %v", err)
	}
	if _, err := os.Stat(dest); !errors.Is(err, os.ErrNotExist) {
		t.Error("dry run should not write the subtitle file")
	}
}

func TestFetchSubtitle_missingTMDBid(t *testing.T) {
	entry := &metadata.Entry{
		Role:     metadata.MovieFile,
		FileInfo: metadata.FileInfo{DestPath: "/tmp/test.srt"},
	}
	err := FetchSubtitle(entry, &Session{}, &config.Config{}, noopLogger())
	if err == nil {
		t.Fatal("expected error for missing TMDBid, got nil")
	}
}
