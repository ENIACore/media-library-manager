package extractor

import (
	"path/filepath"
	"regexp"
	"strconv"
)

var tmdbIDRe = regexp.MustCompile(`\[tmdb-(\d+)\]`)

// ExtractTMDBid walks up the directory tree from path, looking for a [tmdb-<id>] tag
// in any directory name. Returns the id, or 0 if not found.
func ExtractTMDBid(path string) int {
	dir := filepath.Dir(path)
	for {
		if m := tmdbIDRe.FindStringSubmatch(filepath.Base(dir)); m != nil {
			if id, err := strconv.Atoi(m[1]); err == nil {
				return id
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return 0
}
