// Package verifier queries the TMDb API to confirm the title and year extracted
// from a media file or directory are correct, and populates the TMDBid field on
// the root [metadata.Entry] on a successful match.
//
// Verifier runs after the classifier package so that [metadata.EntryRole] is
// already populated on the root entry.
package verifier

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ENIACore/media_library_manager/internal/config"
	"github.com/ENIACore/media_library_manager/internal/metadata"
)

const (
	tmdbBaseURL = "https://api.themoviedb.org/3"
	searchMovie = "/search/movie"
	searchTV    = "/search/tv"
	httpTimeout = 10 * time.Second
	maxResults  = 5
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

// tmdbCandidate is a normalized view of a TMDb result used for interactive
// selection, regardless of whether it came from /search/movie or /search/tv.
type tmdbCandidate struct {
	id    int
	title string
	date  string
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

	if len(root.MediaInfo.Title) == 0 {
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

// verifyMovie searches TMDb /search/movie, selects a result (interactively or
// automatically), confirms title/year match, then sets TMDBid and canonical title.
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

	results := resp.Results
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	candidates := make([]tmdbCandidate, len(results))
	for i, r := range results {
		candidates[i] = tmdbCandidate{id: r.ID, title: r.Title, date: r.ReleaseDate}
	}

	best, err := selectResult(query, "movie", candidates, root.MediaInfo.Year, cfg)
	if err != nil {
		return err
	}

	root.MediaInfo.Title = titleToSlice(best.title)
	root.MediaInfo.TMDBid = "tmdb-" + strconv.Itoa(best.id)
	year := yearFromDate(best.date)
	root.MediaInfo.Year = &year
	logger.Info("movie verified", "title", best.title, "year", year, "tmdb_id", root.MediaInfo.TMDBid)
	return nil
}

// verifySeries searches TMDb /search/tv, selects a result (interactively or
// automatically), confirms title/year match, sets TMDBid, and updates Year to
// the original air year.
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

	results := resp.Results
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	candidates := make([]tmdbCandidate, len(results))
	for i, r := range results {
		candidates[i] = tmdbCandidate{id: r.ID, title: r.Name, date: r.FirstAirDate}
	}

	best, err := selectResult(query, "series", candidates, root.MediaInfo.Year, cfg)
	if err != nil {
		return err
	}

	airYear := yearFromDate(best.date)
	root.MediaInfo.Title = titleToSlice(best.title)
	root.MediaInfo.TMDBid = "tmdb-" + strconv.Itoa(best.id)
	if airYear != 0 {
		root.MediaInfo.Year = &airYear
	}
	logger.Info("series verified", "title", best.title, "first_air_year", airYear, "tmdb_id", root.MediaInfo.TMDBid)
	return nil
}

// selectResult presents candidates to the user interactively and returns the
// chosen one, or returns candidates[0] automatically with title/year validation
// when cfg.Interactive is false.
func selectResult(query, mediaType string, candidates []tmdbCandidate, extractedYear *int, cfg *config.Config) (tmdbCandidate, error) {
	if cfg.Interactive {
		fmt.Printf("\nTMDb results for %s %q:\n", mediaType, query)
		for i, c := range candidates {
			fmt.Printf("  [%d] %s (%s)\n", i+1, c.title, c.date[:safeYearLen(c.date)])
		}
		idx, err := promptSelection(len(candidates))
		if err != nil {
			return tmdbCandidate{}, fmt.Errorf("verifier: selection failed: %w", err)
		}
		return candidates[idx], nil
	}

	best := candidates[0]
	if !titlesMatch(query, best.title) {
		return tmdbCandidate{}, fmt.Errorf("verifier: TMDb title mismatch — extracted %q, got %q", query, best.title)
	}
	resultYear := yearFromDate(best.date)
	if extractedYear != nil && resultYear != 0 && *extractedYear != resultYear {
		return tmdbCandidate{}, fmt.Errorf("verifier: TMDb year mismatch for %q — extracted %d, got %d", query, *extractedYear, resultYear)
	}
	return best, nil
}

// promptSelection prints a prompt and reads a 1-based integer selection from
// stdin. Returns the 0-based index of the chosen result.
func promptSelection(count int) (int, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Select [1-%d]: ", count)
		line, err := reader.ReadString('\n')
		if err != nil {
			return 0, fmt.Errorf("failed to read input: %w", err)
		}
		line = strings.TrimSpace(line)
		n, err := strconv.Atoi(line)
		if err != nil || n < 1 || n > count {
			fmt.Printf("Invalid selection, enter a number between 1 and %d\n", count)
			continue
		}
		return n - 1, nil
	}
}

// tmdbGet performs a GET request to the given TMDb endpoint.
// Uses Bearer auth for v4 JWT tokens, api_key query param for v3 hex keys.
func tmdbGet(apiKey string, endpoint string, params url.Values) ([]byte, error) {
	reqURL := tmdbBaseURL + endpoint + "?" + params.Encode()

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

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

// titleToSlice splits a TMDb title string into an uppercase []string,
// consistent with how the rest of the pipeline stores titles.
// e.g. "The Dark Knight" → ["THE", "DARK", "KNIGHT"]
func titleToSlice(title string) []string {
	parts := strings.Fields(title)
	for i, p := range parts {
		parts[i] = strings.ToUpper(p)
	}
	return parts
}

// joinTitle converts a Title []string to a lowercase space-separated search string.
// e.g. ["DARK", "KNIGHT"] → "dark knight"
func joinTitle(parts []string) string {
	words := make([]string, len(parts))
	for i, p := range parts {
		words[i] = strings.ToLower(p)
	}
	return strings.Join(words, " ")
}

// titlesMatch normalizes both strings by lowercasing and stripping all
// non-alphanumeric characters before comparing, so punctuation and
// articles don't cause false mismatches.
func titlesMatch(extracted, result string) bool {
	nonAlphaNum := regexp.MustCompile(`[^a-z0-9]`)
	norm := func(s string) string {
		return nonAlphaNum.ReplaceAllString(strings.ToLower(s), "")
	}
	return norm(extracted) == norm(result)
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

// safeYearLen returns 4 if the date string is at least 4 chars, otherwise its length.
// Used to safely slice the year out of a date string for display.
func safeYearLen(date string) int {
	if len(date) >= 4 {
		return 4
	}
	return len(date)
}

// looksLikeJWT returns true if the key appears to be a v4 JWT read access token.
func looksLikeJWT(key string) bool {
	return strings.HasPrefix(key, "eyJ")
}
