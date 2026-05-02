// Package verifier queries the TMDb API to confirm the title and year extracted
// from a media file or directory are correct, and populates the TMDBid field on
// the root [metadata.Entry] on a successful match.
//
// Verifier runs after the classifier package so that [metadata.EntryRole] is
// already populated on the root entry.
package verifier

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ENIACore/media_library_manager/internal/config"
	"github.com/ENIACore/media_library_manager/internal/metadata"
)

const (
	tmdbBaseURL    = "https://api.themoviedb.org/3"
	searchMovie    = "/search/movie"
	searchTV       = "/search/tv"
	httpTimeout    = 10 * time.Second
)

// tmdbMovieResult is the subset of a TMDb /search/movie result we care about.
type tmdbMovieResult struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	ReleaseDate string `json:"release_date"` // "YYYY-MM-DD"
}

// tmdbTVResult is the subset of a TMDb /search/tv result we care about.
type tmdbTVResult struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	FirstAirDate string `json:"first_air_date"` // "YYYY-MM-DD"
}

type tmdbMovieResponse struct {
	Results []tmdbMovieResult `json:"results"`
}

type tmdbTVResponse struct {
	Results []tmdbTVResult `json:"results"`
}

// Verify queries TMDb to confirm the title and optional year on the root entry
// are correct. On success it sets root.MediaInfo.TMDBid. For series roots it
// also updates root.MediaInfo.Year to the original air year returned by TMDb.
//
// An error is returned when:
//   - the root entry role is not verifiable (not a movie or series root)
//   - the TMDb API key is missing from cfg
//   - no match is found for the extracted title / year
//   - the HTTP request fails
func Verify(root *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "Verify", "source", root.Source())

	if cfg.TMDBApiKey == "" {
		return fmt.Errorf("verifier: TMDB_API_KEY is not set in config")
	}

	title := root.MediaInfo.Title
	if len(title) == 0 {
		return fmt.Errorf("verifier: entry %v has no title to verify", root.Source())
	}

	switch root.Role {
	case metadata.MovieDir, metadata.MovieFile:
		return verifyMovie(root, cfg, lg)
	case metadata.SeriesDir:
		return verifySeries(root, cfg, lg)
	default:
		return fmt.Errorf("verifier: role %v for entry %v is not verifiable at root level", root.Role.String(), root.Source())
	}
}

// verifyMovie searches TMDb /search/movie, confirms the top result matches the
// extracted title and year, then sets TMDBid.
func verifyMovie(root *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	query := joinTitle(root.MediaInfo.Title)
	params := url.Values{}
	params.Set("query", query)
	if root.MediaInfo.Year != nil {
		params.Set("year", strconv.Itoa(*root.MediaInfo.Year))
	}

	body, err := tmdbGet(cfg.TMDBApiKey, searchMovie, params)
	if err != nil {
		return fmt.Errorf("verifier: TMDb request failed for movie %q: %w", query, err)
	}

	var resp tmdbMovieResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("verifier: failed to parse TMDb movie response: %w", err)
	}
	if len(resp.Results) == 0 {
		return fmt.Errorf("verifier: no TMDb results found for movie %q (year=%v)", query, root.MediaInfo.YearString())
	}

	best := resp.Results[0]
	resultYear := yearFromDate(best.ReleaseDate)

	if !titlesMatch(query, best.Title) {
		return fmt.Errorf("verifier: TMDb title mismatch — extracted %q, got %q", query, best.Title)
	}
	if root.MediaInfo.Year != nil && resultYear != 0 && *root.MediaInfo.Year != resultYear {
		return fmt.Errorf("verifier: TMDb year mismatch for %q — extracted %d, got %d", query, *root.MediaInfo.Year, resultYear)
	}

	root.MediaInfo.TMDBid = strconv.Itoa(best.ID)
	logger.Info("movie verified", "title", best.Title, "year", resultYear, "tmdb_id", root.MediaInfo.TMDBid)
	return nil
}

// verifySeries searches TMDb /search/tv, confirms the top result matches the
// extracted title and year, sets TMDBid, and updates Year to the original air year.
func verifySeries(root *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	query := joinTitle(root.MediaInfo.Title)
	params := url.Values{}
	params.Set("query", query)
	if root.MediaInfo.Year != nil {
		params.Set("first_air_date_year", strconv.Itoa(*root.MediaInfo.Year))
	}

	body, err := tmdbGet(cfg.TMDBApiKey, searchTV, params)
	if err != nil {
		return fmt.Errorf("verifier: TMDb request failed for series %q: %w", query, err)
	}

	var resp tmdbTVResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("verifier: failed to parse TMDb TV response: %w", err)
	}
	if len(resp.Results) == 0 {
		return fmt.Errorf("verifier: no TMDb results found for series %q (year=%v)", query, root.MediaInfo.YearString())
	}

	best := resp.Results[0]
	airYear := yearFromDate(best.FirstAirDate)

	if !titlesMatch(query, best.Name) {
		return fmt.Errorf("verifier: TMDb title mismatch — extracted %q, got %q", query, best.Name)
	}
	if root.MediaInfo.Year != nil && airYear != 0 && *root.MediaInfo.Year != airYear {
		return fmt.Errorf("verifier: TMDb year mismatch for series %q — extracted %d, got %d", query, *root.MediaInfo.Year, airYear)
	}

	root.MediaInfo.TMDBid = strconv.Itoa(best.ID)
	// Always update series year to the canonical original air year from TMDb.
	if airYear != 0 {
		root.MediaInfo.Year = &airYear
	}
	logger.Info("series verified", "title", best.Name, "first_air_year", airYear, "tmdb_id", root.MediaInfo.TMDBid)
	return nil
}

// tmdbGet performs a GET request to the given TMDb endpoint with the provided
// query parameters, authenticating via the v3 Bearer token (API Read Access Token)
// or falling back to the api_key query param if the key looks like a v3 key.
func tmdbGet(apiKey string, endpoint string, params url.Values) ([]byte, error) {
	reqURL := tmdbBaseURL + endpoint + "?" + params.Encode()

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	// TMDb supports both Bearer (v4 read token) and api_key param (v3).
	// A v4 read token is a long JWT; a v3 key is a 32-char hex string.
	// We use Bearer for v4 tokens and query param for v3 keys.
	if looksLikeJWT(apiKey) {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	} else {
		q := req.URL.Query()
		q.Set("api_key", apiKey)
		req.URL.RawQuery = q.Encode()
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("TMDb authentication failed (401): check your API key")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDb returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// joinTitle converts a Title []string to a space-separated search string.
// e.g. ["DARK", "KNIGHT"] → "Dark Knight"
func joinTitle(parts []string) string {
	words := make([]string, len(parts))
	for i, p := range parts {
		words[i] = strings.Title(strings.ToLower(p)) //nolint:staticcheck
	}
	return strings.Join(words, " ")
}

// normalizeTitle lowercases and strips all non-alphanumeric characters for
// loose comparison, so punctuation and articles don't cause false mismatches.
var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]`)

func normalizeTitle(s string) string {
	return nonAlphaNum.ReplaceAllString(strings.ToLower(s), "")
}

func titlesMatch(extracted, result string) bool {
	return normalizeTitle(extracted) == normalizeTitle(result)
}

// yearFromDate parses the year out of a "YYYY-MM-DD" date string.
// Returns 0 if the string is empty or malformed.
func yearFromDate(date string) int {
	if len(date) < 4 {
		return 0
	}
	y, err := strconv.Atoi(date[:4])
	if err != nil {
		return 0
	}
	return y
}

// looksLikeJWT returns true if the key appears to be a JWT (v4 read access token).
func looksLikeJWT(key string) bool {
	return strings.HasPrefix(key, "eyJ")
}

