package main

import (
	"fmt"
	"log/slog"
	"github.com/ENIACore/media_library_manager/internal/extractor"
)

func main() {
	lg := slog.Default()
	subtitlePath := "/Users/chaselamkin/Desktop/testing-dir/input/The Rip (2026) [1080p] [WEBRip] [5.1] [YTS.BZ]/Subs/Brazilian.por.srt"
	info := extractor.ExtractMedia(subtitlePath, lg)	
	fmt.Printf("lg Media info is %+v", info)
}
