package enhancer

import (
	"bytes"
	"encoding/json"
	"errors"
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

// ErrUnauthorized is returned when the server responds with 401.
var ErrUnauthorized = errors.New("OpenSubtitles authentication failed (401)")

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

// Session holds everything needed for authenticated OpenSubtitles API calls.
// Obtain one via Login and reuse it across FetchSubtitle calls.
type Session struct {
	Token     string
	BaseURL   string
	apiKey    string
	userAgent string
	client    *http.Client
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

	client := &http.Client{Timeout: httpTimeout}
	resp, err := osLogin(client, osDefaultBaseURL+"/login", cfg.OpenSubtitlesApiKey, cfg.OpenSubtitlesUserAgent, cfg.OpenSubtitlesUser, cfg.OpenSubtitlesPass)
	if err != nil {
		return nil, fmt.Errorf("enhancer: OpenSubtitles login failed: %w", err)
	}

	baseURL := buildBaseURL(resp.BaseURL)

	lg.Info("OpenSubtitles login successful",
		"allowed_downloads", resp.User.AllowedDownloads,
		"allowed_translations", resp.User.AllowedTranslations,
		"level", resp.User.Level,
		"vip", resp.User.VIP,
		"base_url", resp.BaseURL,
		"resolved_base_url", baseURL,
	)

	return &Session{
		Token:     resp.Token,
		BaseURL:   baseURL,
		apiKey:    cfg.OpenSubtitlesApiKey,
		userAgent: cfg.OpenSubtitlesUserAgent,
		client:    client,
	}, nil
}

// VerifySession checks whether the session's JWT is still accepted by the API.
// Returns true if valid, false on 401, error on other failures.
func VerifySession(session *Session, logger *slog.Logger) (bool, error) {
	lg := logger.With("func", "VerifySession")
	_, err := session.get("/infos/user", url.Values{})
	if err != nil {
		if errors.Is(err, ErrUnauthorized) {
			lg.Debug("cached session is no longer valid")
			return false, nil
		}
		return false, err
	}
	lg.Debug("cached session verified")
	return true, nil
}

// refreshSession invalidates the disk cache, re-authenticates, and updates the token
// and base URL in-place. The existing client is preserved so injected test clients survive.
func refreshSession(session *Session, cfg *config.Config, logger *slog.Logger) error {
	InvalidateCache(cfg)
	newSess, err := Login(cfg, logger)
	if err != nil {
		return err
	}
	session.Token = newSess.Token
	session.BaseURL = newSess.BaseURL
	if err := SaveSession(session, cfg); err != nil {
		logger.Warn("failed to persist refreshed session", "error", err)
	}
	return nil
}

// FetchSubtitle downloads an English SRT subtitle for entry and writes it to entry.FileInfo.DestPath.
// entry must have TMDBid set and DestPath set to the target subtitle path.
// session must be obtained via Login and reused across calls.
// A 401 from /subtitles or /download triggers a one-time re-login; session is updated in-place.
func FetchSubtitle(entry *metadata.Entry, session *Session, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "FetchSubtitle", "source", entry.Source())

	if entry.MediaInfo.TMDBid == 0 {
		return fmt.Errorf("enhancer: entry %v has no TMDBid", entry.Source())
	}
	if entry.FileInfo.DestPath == "" {
		return fmt.Errorf("enhancer: entry %v has no subtitle destination path", entry.Source())
	}

	reauthed := false

	fileID, err := session.searchSubtitle(entry)
	if errors.Is(err, ErrUnauthorized) && !reauthed {
		lg.Warn("session expired during subtitle search, re-authenticating")
		reauthed = true
		if rerr := refreshSession(session, cfg, logger); rerr != nil {
			return fmt.Errorf("enhancer: re-authentication failed: %w", rerr)
		}
		fileID, err = session.searchSubtitle(entry)
	}
	if err != nil {
		return fmt.Errorf("enhancer: subtitle search failed for %v: %w", entry.Source(), err)
	}

	link, remaining, err := session.requestDownload(fileID)
	if errors.Is(err, ErrUnauthorized) && !reauthed {
		lg.Warn("session expired during download request, re-authenticating")
		reauthed = true
		if rerr := refreshSession(session, cfg, logger); rerr != nil {
			return fmt.Errorf("enhancer: re-authentication failed: %w", rerr)
		}
		link, remaining, err = session.requestDownload(fileID)
	}
	if err != nil {
		return fmt.Errorf("enhancer: download request failed for %v: %w", entry.Source(), err)
	}

	lg.Info("downloading subtitle", "file_id", fileID, "remaining_downloads", remaining, "dest", entry.FileInfo.DestPath)

	if cfg.DryRun {
		lg.Info("dry run: skipping subtitle write", "dest", entry.FileInfo.DestPath)
		return nil
	}

	if err := downloadSubtitle(link, entry.FileInfo.DestPath, session.client); err != nil {
		return fmt.Errorf("enhancer: subtitle write failed for %v: %w", entry.Source(), err)
	}

	lg.Info("subtitle written", "dest", entry.FileInfo.DestPath)
	return nil
}

// osLogin POSTs credentials to loginURL and returns the parsed response.
// loginURL is always osDefaultBaseURL+"/login" in production; tests pass a httptest URL.
func osLogin(client *http.Client, loginURL, apiKey, userAgent, username, password string) (*osLoginResponse, error) {
	body, err := json.Marshal(osLoginBody{Username: username, Password: password})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, loginURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Api-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenSubtitles returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var loginResp osLoginResponse
	if err := json.Unmarshal(data, &loginResp); err != nil {
		return nil, fmt.Errorf("failed to parse login response: %w", err)
	}
	if loginResp.Token == "" {
		return nil, fmt.Errorf("login response missing token")
	}
	return &loginResp, nil
}

func (s *Session) searchSubtitle(entry *metadata.Entry) (int, error) {
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

	data, err := s.get("/subtitles", params)
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

func (s *Session) requestDownload(fileID int) (string, int, error) {
	body, err := json.Marshal(osDownloadBody{FileID: fileID})
	if err != nil {
		return "", 0, err
	}

	data, err := s.post("/download", body)
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

// downloadSubtitle fetches link and writes the response to destPath atomically via a temp file.
// On any failure the temp file is cleaned up and destPath is left untouched.
func downloadSubtitle(link, destPath string, client *http.Client) error {
	resp, err := client.Get(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("subtitle download returned status %d", resp.StatusCode)
	}

	tmp := destPath + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
		os.Remove(tmp) // no-op after successful rename
	}()

	if _, err = io.Copy(f, resp.Body); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, destPath)
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

func (s *Session) get(endpoint string, params url.Values) ([]byte, error) {
	for attempt := 0; ; attempt++ {
		req, err := http.NewRequest(http.MethodGet, s.BaseURL+endpoint+"?"+params.Encode(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Api-Key", s.apiKey)
		req.Header.Set("User-Agent", s.userAgent)
		req.Header.Set("Accept", "application/json")
		if s.Token != "" {
			req.Header.Set("Authorization", "Bearer "+s.Token)
		}

		resp, err := s.client.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusTooManyRequests && attempt == 0 {
			wait := parseRetryAfter(resp.Header.Get("Retry-After"))
			resp.Body.Close()
			slog.Default().Warn("OpenSubtitles rate limited, retrying", "wait", wait)
			time.Sleep(wait)
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return nil, ErrUnauthorized
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("OpenSubtitles returned status %d", resp.StatusCode)
		}

		return io.ReadAll(resp.Body)
	}
}

func (s *Session) post(endpoint string, body []byte) ([]byte, error) {
	for attempt := 0; ; attempt++ {
		req, err := http.NewRequest(http.MethodPost, s.BaseURL+endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Api-Key", s.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", s.userAgent)
		if s.Token != "" {
			req.Header.Set("Authorization", "Bearer "+s.Token)
		}

		resp, err := s.client.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusTooManyRequests && attempt == 0 {
			wait := parseRetryAfter(resp.Header.Get("Retry-After"))
			resp.Body.Close()
			slog.Default().Warn("OpenSubtitles rate limited, retrying", "wait", wait)
			time.Sleep(wait)
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return nil, ErrUnauthorized
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("OpenSubtitles returned status %d", resp.StatusCode)
		}

		return io.ReadAll(resp.Body)
	}
}

// buildBaseURL converts the host returned in the login response to a full API URL.
func buildBaseURL(host string) string {
	if host == "" {
		return osDefaultBaseURL
	}
	return "https://" + host + "/api/v1"
}

// parseRetryAfter parses the Retry-After header (integer seconds).
// Falls back to 60s if absent, negative, or unparseable.
func parseRetryAfter(header string) time.Duration {
	if secs, err := strconv.Atoi(header); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	return 60 * time.Second
}
