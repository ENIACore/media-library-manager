package extractor

import (
	"regexp"
	"strings"
)

// sanitizeName normalizes a filename for pattern matching.
// Converts to uppercase, removes quotes, replaces non-alphanumeric
// characters with dots, and trims leading/trailing spaces and dots.
func sanitizeName(name string) string {
	name = strings.ToUpper(name)
	name = strings.ReplaceAll(name, "'", "")
	name = strings.ReplaceAll(name, "\"", "")

	re := regexp.MustCompile("[^A-Z0-9]+")
	name = string(re.ReplaceAll([]byte(name), []byte(".")))

	name = strings.Trim(name, " .")
	return name
}

// matchSegments joins segments and attempts a full regex match.
// Joins only as many segments as needed based on literal dots in the pattern.
// Returns the match slice (full match + capture groups) or nil.
func matchSegments(segments []string, re *regexp.Regexp) []string {
	numDots := strings.Count(re.String(), `\.`)
	end := min(numDots+1, len(segments))

	str := strings.Join(segments[:end], ".")

	if match := re.FindStringSubmatch(str); match != nil && match[0] == str {
		return match
	}
	return nil
}
