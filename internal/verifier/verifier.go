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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ENIACore/media_library_manager/internal/classifier"
	"github.com/ENIACore/media_library_manager/internal/config"
	"github.com/ENIACore/media_library_manager/internal/metadata"
)

const (
	tmdbBaseURL  = "https://api.themoviedb.org/3"
	searchMulti  = "/search/multi"
	movieDetail  = "/movie/"
	tvDetail     = "/tv/"
	httpTimeout  = 10 * time.Second
	maxResults   = 5
)

type tmdbMultiResult struct {
	ID           int    `json:"id"`
	MediaType    string `json:"media_type"`
	Title        string `json:"title"`          // movie
	Name         string `json:"name"`           // tv
	ReleaseDate  string `json:"release_date"`   // movie
	FirstAirDate string `json:"first_air_date"` // tv
}

func (r tmdbMultiResult) canonicalTitle() string {
	if r.MediaType == "tv" {
		return r.Name
	}
	return r.Title
}

func (r tmdbMultiResult) canonicalDate() string {
	if r.MediaType == "tv" {
		return r.FirstAirDate
	}
	return r.ReleaseDate
}

type tmdbMultiResponse struct {
	Results []tmdbMultiResult `json:"results"`
}

type tmdbCandidate struct {
	id        int
	title     string
	date      string
	mediaType string
}

type tmdbDetailResponse struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`          // movie
	Name         string `json:"name"`           // tv
	ReleaseDate  string `json:"release_date"`   // movie
	FirstAirDate string `json:"first_air_date"` // tv
}

// lookupByID fetches a single TMDB record directly by numeric ID, trying preferred
// media type first and falling back to the other. Returns an error if both fail.
func lookupByID(id int, preferred, fallback, apiKey string) (tmdbCandidate, error) {
	fetch := func(mediaType string) (tmdbCandidate, error) {
		endpoint := movieDetail + strconv.Itoa(id)
		if mediaType == "tv" {
			endpoint = tvDetail + strconv.Itoa(id)
		}
		body, err := tmdbGet(apiKey, endpoint, url.Values{})
		if err != nil {
			return tmdbCandidate{}, err
		}
		var detail tmdbDetailResponse
		if err := json.Unmarshal(body, &detail); err != nil {
			return tmdbCandidate{}, fmt.Errorf("failed to parse TMDb detail response: %w", err)
		}
		title, date := detail.Title, detail.ReleaseDate
		if mediaType == "tv" {
			title, date = detail.Name, detail.FirstAirDate
		}
		return tmdbCandidate{id: detail.ID, title: title, date: date, mediaType: mediaType}, nil
	}

	if c, err := fetch(preferred); err == nil {
		return c, nil
	}
	return fetch(fallback)
}

func Verify(root *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "Verify", "source", root.Source())

	if cfg.TMDBApiKey == "" {
		return fmt.Errorf("verifier: TMDB_API_KEY is not set in config")
	}

	if len(root.MediaInfo.Title) == 0 {
		return fmt.Errorf("verifier: entry %v has no title to verify", root.Source())
	}

	switch root.Role {
	case metadata.MovieDir, metadata.MovieFile,
		metadata.SeriesDir, metadata.SeasonDir, metadata.EpisodeFile:
	default:
		return fmt.Errorf("verifier: role %v for entry %v is not verifiable at root level", root.Role.String(), root.Source())
	}

	tmdbType, err := verifyEntry(root, cfg, lg)
	if err != nil {
		return err
	}

	return classifier.Reclassify(root, tmdbType, lg)
}

// verifyEntry searches TMDb using the preferred media type derived from the current
// role and falls back to the other type when no results are found.
// It updates root.MediaInfo (Title, Year, TMDBid) and returns the TMDB media_type.
func verifyEntry(root *metadata.Entry, cfg *config.Config, logger *slog.Logger) (string, error) {
	query := joinTitle(root.MediaInfo.Title)
	preferred := preferredMediaType(root.Role)
	fallback := "tv"
	if preferred == "tv" {
		fallback = "movie"
	}

	// If the filename already contains a TMDB id, try a direct lookup first.
	if root.MediaInfo.TMDBid != 0 {
		if best, err := lookupByID(root.MediaInfo.TMDBid, preferred, fallback, cfg.TMDBApiKey); err == nil {
			root.MediaInfo.Title = titleToSlice(best.title)
			root.MediaInfo.TMDBid = best.id
			if year := yearFromDate(best.date); year != 0 {
				root.MediaInfo.Year = &year
			}
			logger.Info("verified by id", "title", best.title, "year", root.MediaInfo.YearString(), "media_type", best.mediaType, "tmdb_id", root.TMDB())
			return best.mediaType, nil
		}
		logger.Info("tmdb id lookup failed, falling back to title search", "tmdb_id", root.MediaInfo.TMDBid)
	}

	candidates, err := multiSearch(query, preferred, root.MediaInfo.Year, cfg.TMDBApiKey)
	if err != nil {
		return "", fmt.Errorf("verifier: TMDb request failed for %q: %w", query, err)
	}

	// Nothing found for the preferred type — try the other one.
	if len(candidates) == 0 {
		candidates, err = multiSearch(query, fallback, root.MediaInfo.Year, cfg.TMDBApiKey)
		if err != nil {
			return "", fmt.Errorf("verifier: TMDb fallback request failed for %q: %w", query, err)
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("verifier: no TMDb results found for %q (year=%v)", query, root.MediaInfo.YearString())
	}

	best, err := selectResult(query, candidates[0].mediaType, candidates, root.MediaInfo.Year, cfg)
	if err != nil {
		return "", err
	}

	root.MediaInfo.Title = titleToSlice(best.title)
	root.MediaInfo.TMDBid = best.id
	if year := yearFromDate(best.date); year != 0 {
		root.MediaInfo.Year = &year
	}

	logger.Info("verified", "title", best.title, "year", root.MediaInfo.YearString(), "media_type", best.mediaType, "tmdb_id", root.TMDB())
	return best.mediaType, nil
}

// preferredMediaType returns "tv" for series/season/episode roles and "movie" for everything else.
func preferredMediaType(role metadata.EntryRole) string {
	switch role {
	case metadata.SeriesDir, metadata.SeasonDir, metadata.EpisodeFile:
		return "tv"
	default:
		return "movie"
	}
}

// multiSearch calls /search/multi and returns candidates filtered to the requested mediaType ("movie" or "tv").
func multiSearch(query, mediaType string, year *int, apiKey string) ([]tmdbCandidate, error) {
	params := url.Values{}
	params.Set("query", query)
	if year != nil {
		params.Set("year", strconv.Itoa(*year))
	}

	body, err := tmdbGet(apiKey, searchMulti, params)
	if err != nil {
		return nil, err
	}

	var resp tmdbMultiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse TMDb multi response: %w", err)
	}

	var candidates []tmdbCandidate
	for _, r := range resp.Results {
		if r.MediaType != mediaType {
			continue
		}
		candidates = append(candidates, tmdbCandidate{
			id:        r.ID,
			title:     r.canonicalTitle(),
			date:      r.canonicalDate(),
			mediaType: r.MediaType,
		})
		if len(candidates) == maxResults {
			break
		}
	}

	// TMDb ranks by popularity, not year — sort year-matching candidates first.
	if year != nil {
		sort.SliceStable(candidates, func(i, j int) bool {
			yi := yearFromDate(candidates[i].date) == *year
			yj := yearFromDate(candidates[j].date) == *year
			return yi && !yj
		})
	}

	return candidates, nil
}

func selectResult(query, mediaType string, candidates []tmdbCandidate, extractedYear *int, cfg *config.Config) (tmdbCandidate, error) {
	if len(candidates) == 0 {
		return tmdbCandidate{}, fmt.Errorf("verifier: no TMDb matches found for %s %q", mediaType, query)
	}

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
		fmt.Printf("verifier: warning — title mismatch: extracted %q, got %q\n", query, best.title)
	}

	resultYear := yearFromDate(best.date)
	if extractedYear != nil && resultYear != 0 && *extractedYear != resultYear {
		fmt.Printf("verifier: warning — year mismatch: extracted %d, got %d\n", *extractedYear, resultYear)
	}

	return best, nil
}

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

func titleToSlice(title string) []string {
	parts := strings.Fields(title)
	for i, p := range parts {
		parts[i] = strings.ToUpper(p)
	}
	return parts
}

func joinTitle(parts []string) string {
	words := make([]string, len(parts))
	for i, p := range parts {
		words[i] = strings.ToLower(p)
	}
	return strings.Join(words, " ")
}

func titlesMatch(extracted, result string) bool {
	nonAlphaNum := regexp.MustCompile(`[^a-z0-9]`)
	norm := func(s string) string {
		return nonAlphaNum.ReplaceAllString(strings.ToLower(s), "")
	}
	return norm(extracted) == norm(result)
}

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

func safeYearLen(date string) int {
	if len(date) >= 4 {
		return 4
	}
	return len(date)
}

func looksLikeJWT(key string) bool {
	return strings.HasPrefix(key, "eyJ")
}
