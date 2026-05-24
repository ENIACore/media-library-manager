package extractor

import (
	"regexp"
	"strings"

	"path/filepath"
)

// sanitizePrefix finds common unwanted patterns at the beginning of file names (i.e websites) and removes them from the slice
func sanitizePrefix(segments []string) []string {
	if match := parseWebsite(segments); match != "" {
		numParts := len(strings.Split(match, "."))
		return segments[numParts:]
	}
	return segments
}

func SanitizeName(path string) []string {
	name := filepath.Base(path)

	name = strings.ToUpper(name)
	name = strings.ReplaceAll(name, "'", "")
	name = strings.ReplaceAll(name, "\"", "")

	re := regexp.MustCompile("[^A-Z0-9]+")
	name = string(re.ReplaceAll([]byte(name), []byte(".")))

	name = strings.Trim(name, " .")

	sanitizedName := strings.Split(name, ".")
	sanitizedName = sanitizePrefix(sanitizedName)
	return sanitizedName
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
