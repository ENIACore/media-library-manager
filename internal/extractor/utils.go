package extractor

import (
	"regexp"
	"strings"
)

// sanitizeName normalizes a filename for pattern matching.
//	- Converts to uppercase
//	- Replaces all non-alphanumeric characters with '.'
//	- Trims leading and trailing '.'
// Result of transformations are alphanumeric words seperated by exactly one '.' rune.
func sanitizeName(name string) string {
	name = strings.ToUpper(name)
	name = strings.ReplaceAll(name, "'", "")
	name = strings.ReplaceAll(name, "\"", "")

	re := regexp.MustCompile("[^A-Z0-9]+")
	name = string(re.ReplaceAll([]byte(name), []byte(".")))

	name = strings.Trim(name, " .")
	return name
}

// matchSegments matches regex pattern across multiple strings
// Enables multi-word pattern matching across sanitized file or directory name
// Returns [full match, capture groups...]
func matchSegments(segments []string, re *regexp.Regexp) []string {
	numDots := strings.Count(re.String(), `\.`)
	end := min(numDots+1, len(segments))

	str := strings.Join(segments[:end], ".")

	if match := re.FindStringSubmatch(str); match != nil && match[0] == str {
		return match
	}
	return nil
}
