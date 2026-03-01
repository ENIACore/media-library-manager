package patterns

import (
	"sync"
)

var SeasonPatterns = []Pattern{
	// Matches S<1-2 digits>
	`S(\d{1,2})`,
	// Matches S.<1-2 digits>
	`S\.(\d{1,2})`,
	// Matches SEA<1-2 digits> or SEAS<1-2 digits>
	`SEAS?(\d{1,2})`,
	// Matches SEA.<1-2 digits> or SEAS.<1-2 digits>
	`SEAS?\.(\d{1,2})`,
	// Matches SEASON<1-2 digits>
	`SEASON(\d{1,2})`,
	// Matches SEASON.<1-2 digits>
	`SEASON\.(\d{1,2})`,
	// Matches exactly SEASON (full word only)
	`SEASON`,
	// Matches S<1-2 digits>E<1-3 digits>
	`S(\d{1,2})E\d{1,3}`,
}

var EpisodePatterns = []Pattern{
	// Matches E<1-3 digits>
	`E(\d{1,3})`,
	// Matches E.<1-3 digits>
	`E\.(\d{1,3})`,
	// Matches EP<1-3 digits>
	`EP(\d{1,3})`,
	// Matches EP.<1-3 digits>
	`EP\.(\d{1,3})`,
	// Matches EPISODE<1-3 digits>
	`EPISODE(\d{1,3})`,
	// Matches EPISODE.<1-3 digits>
	`EPISODE\.(\d{1,3})`,
	// Matches EPISODE
	`EPISODE`,
	// Matches EP
	`EP`,
	// Matches S<1-2 digits>E<1-3 digits>
	`S\d{1,2}E(\d{1,3})`,
	// Matches <1-2 digits>X<1-3 digits>
	`\d{1,2}X(\d{1,3})`,
	// Matches <1-2 digits>.X.<1-3 digits> (e.g., 1.X.01, 2.X.15)
	//`\d{1,2}\.X\.(\d{1,3})`,
}

var (
	GetSeasonPatterns = sync.OnceValue(func() []*CompiledPattern {
		return compilePatterns(SeasonPatterns)
	})
	GetEpisodePatterns = sync.OnceValue(func() []*CompiledPattern {
		return compilePatterns(EpisodePatterns)
	})
)
