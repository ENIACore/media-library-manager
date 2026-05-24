package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ENIACore/media_library_manager/internal/config"
)

// formatTimestamp formats a timestamp for log directory naming.
// Returns a string in the format YYYY-MM-DD_HH:MM:SS.
func formatTimestamp(now time.Time) string {
	return now.Format("2006-01-02_15:04:05")
}

// stripTime removes the default time attribute from log records to keep output uncluttered.
func stripTime(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey && len(groups) == 0 {
		return slog.Attr{}
	}
	return a
}

// handlerOpts builds the standard handler options for the given level, with timestamps removed.
func handlerOpts(level slog.Level) *slog.HandlerOptions {
	return &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: stripTime,
	}
}

// getFile opens or creates a log file for writing.
// Creates the directory if it doesn't exist. Returns os.Stdout
// if directory or file creation fails, logging a warning to stderr.
func getFile(dirpath string, filename string) io.Writer {
	err := os.MkdirAll(dirpath, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create log directory: %v, using Stdout instead of log file %v", err, filename)
		return os.Stdout
	}

	logpath := filepath.Join(dirpath, filename)
	file, err := os.OpenFile(logpath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create log file: %v, using Stdout instead", filename)
		return os.Stdout
	}
	return file
}

// multiHandler implements slog.Handler interface by delegating to multiple handlers.
// This allows writing logs to multiple files simultaneously based on severity level.
type multiHandler struct {
	handlers []slog.Handler
}

func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return &multiHandler{handlers: handlers}
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return &multiHandler{handlers: handlers}
}

var getSessionTimestamp = sync.OnceValue(func() string {
	return formatTimestamp(time.Now())
})

// NewLogger creates and returns a structured logger configured to write to multiple log files.
// Creates separate log files for debug, info, and warn levels in a timestamped session directory.
// The session timestamp is generated once and reused for all loggers in the process.
func NewLogger(cfg *config.Config) *slog.Logger {
	if cfg.LogStdout {
		return slog.New(slog.NewTextHandler(os.Stdout, handlerOpts(slog.LevelDebug)))
	}

	basepath := filepath.Join(cfg.ManagerPath, "logs", getSessionTimestamp())

	debugFile := getFile(basepath, "debug.log")
	infoFile := getFile(basepath, "info.log")
	warnFile := getFile(basepath, "warn.log")

	handler := &multiHandler{
		handlers: []slog.Handler{
			slog.NewTextHandler(debugFile, handlerOpts(slog.LevelDebug)),
			slog.NewTextHandler(infoFile, handlerOpts(slog.LevelInfo)),
			slog.NewTextHandler(warnFile, handlerOpts(slog.LevelWarn)),
		},
	}

	return slog.New(handler)
}
