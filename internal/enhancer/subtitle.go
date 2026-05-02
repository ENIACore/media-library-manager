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
	osDefaultBaseURL = "https://api.opensubtitles.com/api/v1"
	httpTimeout      = 15 * time.Second
)

type osLoginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type osLoginResponse struct {
	User    osLoginUser `json:"user"`
	BaseURL string      `json:"base_url"`
	Token   string      `json:"token"`
	Status  int         `json:"status"`
}

type osLoginUser struct {
	AllowedTranslations int    `json:"allowed_translations"`
	AllowedDownloads    int    `json:"allowed_downloads"`
	Level               string `json:"level"`
	UserID              int    `json:"user_id"`
	ExtInstalled        bool   `json:"ext_installed"`
	VIP                 bool   `json:"vip"`
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

// Session holds the authenticated OpenSubtitles session state.
// The token is a JWT that lasts 24 hours; reuse the session across calls and
// re-login when it expires. BaseURL may differ from the default for VIP users.
type Session struct {
	Token   string
	BaseURL string
}

// Login authenticates with OpenSubtitles and returns a Session for use with FetchSubtitle.
func Login(cfg *config.Config, logger *slog.Logger) (*Session, error) {
	lg := logger.With("func", "Login")

	if cfg.OpenSubtitlesApiKey == "" {
		return nil, fmt.Errorf("enhancer: OpenSubtitles API key not set in config")
	}
	if cfg.OpenSubtitlesUser == "" || cfg.OpenSubtitlesPass == "" {
		return nil, fmt.Errorf("enhancer: OpenSubtitles username and password not set in config")
	}
	if cfg.OpenSubtitlesUserAgent == "" {
		return nil, fmt.Errorf("enhancer: OpenSubtitles user agent not set in config")
	}

	resp, err := osLogin(cfg.OpenSubtitlesApiKey, cfg.OpenSubtitlesUserAgent, cfg.OpenSubtitlesUser, cfg.OpenSubtitlesPass)
	if err != nil {
		return nil, fmt.Errorf("enhancer: OpenSubtitles login failed: %w", err)
	}

	baseURL := buildBaseURL(resp.BaseURL)

	lg.Info("OpenSubtitles login successful",
		"jwt", resp.Token,
		"allowed_downloads", resp.User.AllowedDownloads,
		"allowed_translations", resp.User.AllowedTranslations,
		"level", resp.User.Level,
		"vip", resp.User.VIP,
		"base_url", resp.BaseURL,
		"resolved_base_url", baseURL,
	)

	return &Session{
		Token:   resp.Token,
		BaseURL: baseURL,
	}, nil
}

// buildBaseURL converts the host returned in the login response into a full URL.
// Falls back to the documented default if the response is missing the field.
func buildBaseURL(host string) string {
	if host == "" {
		return osDefaultBaseURL
	}
	return "https://" + host + "/api/v1"
}

// FetchSubtitle downloads an English SRT subtitle for entry and writes it to entry.FileInfo.DestPath.
// entry must have TMDBid set (by verifier/enricher) and DestPath set to the target subtitle path (by detector).
// session must be obtained once via Login and reused across calls.
func FetchSubtitle(entry *metadata.Entry, session *Session, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "FetchSubtitle", "source", entry.Source())

	if entry.MediaInfo.TMDBid == 0 {
		return fmt.Errorf("enhancer: entry %v has no TMDBid", entry.Source())
	}
	if entry.FileInfo.DestPath == "" {
		return fmt.Errorf("enhancer: entry %v has no subtitle destination path", entry.Source())
	}

	fileID, err := searchSubtitle(entry, cfg.OpenSubtitlesApiKey, cfg.OpenSubtitlesUserAgent, session)
	if err != nil {
		return fmt.Errorf("enhancer: subtitle search failed for %v: %w", entry.Source(), err)
	}

	if cfg.DryRun {
		lg.Info("dry run: subtitle found, skipping download", "file_id", fileID, "dest", entry.FileInfo.DestPath)
		return nil
	}

	link, remaining, err := requestDownload(fileID, cfg.OpenSubtitlesApiKey, cfg.OpenSubtitlesUserAgent, session)
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

func osLogin(apiKey, userAgent, username, password string) (*osLoginResponse, error) {
	body, err := json.Marshal(osLoginBody{Username: username, Password: password})
	if err != nil {
		return nil, err
	}

	// Login itself always goes to the default base URL; the response tells us where to go next.
	data, err := osPost(osDefaultBaseURL, "/login", apiKey, userAgent, "", body)
	if err != nil {
		return nil, err
	}

	var resp osLoginResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse login response: %w", err)
	}
	if resp.Token == "" {
		return nil, fmt.Errorf("login response missing token")
	}
	return &resp, nil
}

func searchSubtitle(entry *metadata.Entry, apiKey, userAgent string, session *Session) (int, error) {
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

	data, err := osGet(session.BaseURL, "/subtitles", apiKey, userAgent, session.Token, params)
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

func requestDownload(fileID int, apiKey, userAgent string, session *Session) (string, int, error) {
	body, err := json.Marshal(osDownloadBody{FileID: fileID})
	if err != nil {
		return "", 0, err
	}

	data, err := osPost(session.BaseURL, "/download", apiKey, userAgent, session.Token, body)
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

func osGet(baseURL, endpoint, apiKey, userAgent, token string, params url.Values) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, baseURL+endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Api-Key", apiKey)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
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

func osPost(baseURL, endpoint, apiKey, userAgent, token string, body []byte) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, baseURL+endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Api-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
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
