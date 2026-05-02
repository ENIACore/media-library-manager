package enhancer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/ENIACore/media_library_manager/internal/config"
	"github.com/ENIACore/media_library_manager/internal/metadata"
)

const (
	osBaseURL   = "https://api.opensubtitles.com/api/v1"
	osUserAgent = "MediaLibraryManager v1.0"
	httpTimeout = 15 * time.Second
)

type osLoginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type osLoginResponse struct {
	Status int    `json:"status"`
	Token  string `json:"token"`
}

type osSearchResponse struct {
	TotalCount int          `json:"total_count"`
	Data       []osSubtitle `json:"data"`
}

type osSubtitle struct {
	Attributes osSubtitleAttrs `json:"attributes"`
}

type osSubtitleAttrs struct {
	DownloadCount   int      `json:"download_count"`
	HearingImpaired bool     `json:"hearing_impaired"`
	Files           []osFile `json:"files"`
}

type osFile struct {
	FileID   int    `json:"file_id"`
	FileName string `json:"file_name"`
}

type osDownloadBody struct {
	FileID int `json:"file_id"`
}

type osDownloadResponse struct {
	Link      string `json:"link"`
	Remaining int    `json:"remaining"`
}

// FetchSubtitle downloads an English SRT subtitle for entry and writes it to entry.FileInfo.DestPath.
// entry must have TMDBid set (by verifier/enricher) and DestPath set to the target subtitle path (by detector).
func FetchSubtitle(entry *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "FetchSubtitle", "source", entry.Source())

	if cfg.OpenSubtitlesApiKey == "" {
		return fmt.Errorf("enhancer: OpenSubtitles API key not set in config")
	}
	if cfg.OpenSubtitlesUser == "" || cfg.OpenSubtitlesPass == "" {
		return fmt.Errorf("enhancer: OpenSubtitles username and password not set in config")
	}
	if entry.MediaInfo.TMDBid == 0 {
		return fmt.Errorf("enhancer: entry %v has no TMDBid", entry.Source())
	}
	if entry.FileInfo.DestPath == "" {
		return fmt.Errorf("enhancer: entry %v has no subtitle destination path", entry.Source())
	}

	token, err := osLogin(cfg.OpenSubtitlesApiKey, cfg.OpenSubtitlesUser, cfg.OpenSubtitlesPass)
	if err != nil {
		return fmt.Errorf("enhancer: OpenSubtitles login failed: %w", err)
	}

	fileID, err := searchSubtitle(entry, cfg.OpenSubtitlesApiKey)
	if err != nil {
		return fmt.Errorf("enhancer: subtitle search failed for %v: %w", entry.Source(), err)
	}

	if cfg.DryRun {
		lg.Info("dry run: subtitle found, skipping download", "file_id", fileID, "dest", entry.FileInfo.DestPath)
		return nil
	}

	link, remaining, err := requestDownload(fileID, cfg.OpenSubtitlesApiKey, token)
	if err != nil {
		return fmt.Errorf("enhancer: download request failed for %v: %w", entry.Source(), err)
	}

	lg.Info("downloading subtitle", "file_id", fileID, "remaining_downloads", remaining, "dest", entry.FileInfo.DestPath)

	if err := downloadSubtitle(link, entry.FileInfo.DestPath); err != nil {
		return fmt.Errorf("enhancer: subtitle write failed for %v: %w", entry.Source(), err)
	}

	lg.Info("subtitle written", "dest", entry.FileInfo.DestPath)
	return nil
}

func osLogin(apiKey, username, password string) (string, error) {
	body, err := json.Marshal(osLoginBody{Username: username, Password: password})
	if err != nil {
		return "", err
	}

	data, err := osPost("/login", apiKey, "", body)
	if err != nil {
		return "", err
	}

	var resp osLoginResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("failed to parse login response: %w", err)
	}
	if resp.Token == "" {
		return "", fmt.Errorf("login response missing token")
	}
	return resp.Token, nil
}

func searchSubtitle(entry *metadata.Entry, apiKey string) (int, error) {
	params := url.Values{}
	params.Set("languages", "en")

	switch entry.Role {
	case metadata.MovieFile:
		params.Set("tmdb_id", strconv.Itoa(entry.MediaInfo.TMDBid))
		params.Set("type", "movie")
	case metadata.EpisodeFile:
		if entry.MediaInfo.Season == nil {
			return 0, fmt.Errorf("episode entry %v has no season number", entry.Source())
		}
		if entry.MediaInfo.Episode == nil {
			return 0, fmt.Errorf("episode entry %v has no episode number", entry.Source())
		}
		params.Set("parent_tmdb_id", strconv.Itoa(entry.MediaInfo.TMDBid))
		params.Set("type", "episode")
		params.Set("season_number", strconv.Itoa(*entry.MediaInfo.Season))
		params.Set("episode_number", strconv.Itoa(*entry.MediaInfo.Episode))
	default:
		return 0, fmt.Errorf("unsupported entry role %v for subtitle fetch", entry.Role.String())
	}

	data, err := osGet("/subtitles", apiKey, params)
	if err != nil {
		return 0, err
	}

	var resp osSearchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return 0, fmt.Errorf("failed to parse search response: %w", err)
	}
	if len(resp.Data) == 0 {
		return 0, fmt.Errorf("no English subtitles found")
	}

	best := pickBest(resp.Data)
	if best == nil {
		return 0, fmt.Errorf("no subtitles with downloadable files found")
	}

	return best.Attributes.Files[0].FileID, nil
}

// pickBest selects the subtitle with the highest download count, preferring non-hearing-impaired.
func pickBest(subs []osSubtitle) *osSubtitle {
	var best *osSubtitle
	for i := range subs {
		s := &subs[i]
		if len(s.Attributes.Files) == 0 {
			continue
		}
		if best == nil {
			best = s
			continue
		}
		preferS := !s.Attributes.HearingImpaired && best.Attributes.HearingImpaired
		sameHI := s.Attributes.HearingImpaired == best.Attributes.HearingImpaired
		if preferS || (sameHI && s.Attributes.DownloadCount > best.Attributes.DownloadCount) {
			best = s
		}
	}
	return best
}

func requestDownload(fileID int, apiKey, token string) (string, int, error) {
	body, err := json.Marshal(osDownloadBody{FileID: fileID})
	if err != nil {
		return "", 0, err
	}

	data, err := osPost("/download", apiKey, token, body)
	if err != nil {
		return "", 0, err
	}

	var resp osDownloadResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", 0, fmt.Errorf("failed to parse download response: %w", err)
	}
	if resp.Link == "" {
		return "", 0, fmt.Errorf("download response missing link")
	}
	return resp.Link, resp.Remaining, nil
}

func downloadSubtitle(link, destPath string) error {
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("subtitle download returned status %d", resp.StatusCode)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func osGet(endpoint, apiKey string, params url.Values) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, osBaseURL+endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Api-Key", apiKey)
	req.Header.Set("User-Agent", osUserAgent)

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("OpenSubtitles authentication failed (401)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenSubtitles returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func osPost(endpoint, apiKey, token string, body []byte) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, osBaseURL+endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Api-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", osUserAgent)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("OpenSubtitles authentication failed (401)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenSubtitles returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
