package logger

import (
	"io"
	"fmt"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
	"github.com/ENIACore/media_library_manager/internal/config"
)

func formatTimestamp(now time.Time) string {
	return now.Format("2006-01-02_15:04:05")
}

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

// Implements slog.Handler interface
type multiHandler struct {
	handlers []slog.Handler
}

// Returns true if ANY handler accepts log
func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Re-runs Enabled to only call Handle on specific handlers
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

// Adds WithAttrs to ALL handlers
func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return &multiHandler{handlers: handlers}
}

// Adds WithGroup to ALL handlers
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

func NewLogger(cfg *config.Config) *slog.Logger {
	basepath := filepath.Join(cfg.ManagerPath, "logs", getSessionTimestamp())


	debugFile := getFile(basepath, "DEBUG.log")
	infoFile := getFile(basepath, "INFO.log")
	warnFile := getFile(basepath, "WARN.log")

	handler := &multiHandler{
		handlers: []slog.Handler{
			slog.NewTextHandler(debugFile, &slog.HandlerOptions{Level: slog.LevelDebug}),
			slog.NewTextHandler(infoFile, &slog.HandlerOptions{Level: slog.LevelInfo}),
			slog.NewTextHandler(warnFile, &slog.HandlerOptions{Level: slog.LevelWarn}),
		},
	}

	return slog.New(handler).With("dry-run", cfg.DryRun)
}
