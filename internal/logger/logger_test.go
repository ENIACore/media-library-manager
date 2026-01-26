package logger

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ENIACore/media_library_manager/internal/config"
)

func TestFormatTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name: "testing regular date object",
			// Date of December 30th 2025 at 12:01:02PM (3 nano seconds)
			input:    time.Date(2025, 12, 30, 12, 1, 2, 3, time.UTC),
			expected: "2025-12-30_12:01:02",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			formattedTimestamp := formatTimestamp(test.input)
			if test.expected != formattedTimestamp {
				t.Errorf("formatTimestamp() = %v want %v", formattedTimestamp, test.expected)
			}
		})
	}
}

func TestGetFile(t *testing.T) {
	tempdir := t.TempDir()

	tests := []struct {
		name     string
		dirpath  string
		filename string
		expected string
	}{
		{
			name:     "test normal log file",
			dirpath:  filepath.Join(tempdir, "logs", getSessionTimestamp()),
			filename: "test.log",
			expected: filepath.Join(tempdir, "logs", getSessionTimestamp(), "test.log"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			writer := getFile(test.dirpath, test.filename)
			file := writer.(*os.File)
			if file.Name() != test.expected {
				t.Errorf("getFile() = %v want, %v", file.Name(), test.expected)
			}

			_, err := os.Stat(test.expected)
			if err != nil {
				t.Errorf("File %v DNE, and returns error %v", file.Name(), err)
			}

			_, err = os.Stat(test.dirpath)
			if err != nil {
				t.Errorf("Dir %v DNE, and returns error %v", file.Name(), err)
			}

			_, err = writer.Write([]byte("test"))
			if err != nil {
				t.Errorf("Failed to write to file %v", err)
			}
		})
	}
}

func TestEnabled(t *testing.T) {
	tests := []struct {
		name     string
		level    slog.Level
		handler  *multiHandler
		expected bool
	}{
		{
			name:  "error log with info and error handlers",
			level: slog.LevelError,
			handler: &multiHandler{
				handlers: []slog.Handler{
					slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
					slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}),
				},
			},
			expected: true,
		},
		{
			name:  "debug log with info and error handlers",
			level: slog.LevelDebug,
			handler: &multiHandler{
				handlers: []slog.Handler{
					slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
					slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}),
				},
			},
			expected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := test.handler.Enabled(context.TODO(), test.level)
			if res != test.expected {
				t.Errorf("Enabled() = %v, want %v", res, test.expected)
			}
		})
	}
}

func TestHandle(t *testing.T) {
	basepath := filepath.Join(t.TempDir(), "logs", getSessionTimestamp())
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

	tests := []struct {
		name  string
		level slog.Level
		msg   string
	}{
		{
			name:  "test error log",
			level: slog.LevelError,
			msg:   "test",
		},
		{
			name:  "test info log",
			level: slog.LevelInfo,
			msg:   "test info message",
		},
		{
			name:  "test debug log",
			level: slog.LevelDebug,
			msg:   "test debug message",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := handler.Handle(context.TODO(), slog.NewRecord(time.Now(), test.level, test.msg, 0))
			if err != nil {
				t.Errorf("Handle() = err not nil, err is %v", err)
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	t.Run("log to stdout", func(t *testing.T) {
		cfg := &config.Config{
			LogStdout:   true,
			ManagerPath: t.TempDir(),
		}

		logger := NewLogger(cfg)
		if logger == nil {
			t.Error("NewLogger() returned nil")
		}

		// Verify logger is functional
		logger.Info("test message")
	})

	t.Run("log to files", func(t *testing.T) {
		tempdir := t.TempDir()
		cfg := &config.Config{
			LogStdout:   false,
			ManagerPath: tempdir,
		}

		logger := NewLogger(cfg)
		if logger == nil {
			t.Error("NewLogger() returned nil")
		}

		// Verify logger is functional
		logger.Info("test message")
		logger.Debug("debug message")
		logger.Warn("warn message")

		// Verify log directory was created
		logsDir := filepath.Join(tempdir, "logs")
		if _, err := os.Stat(logsDir); os.IsNotExist(err) {
			t.Errorf("logs directory was not created at %v", logsDir)
		}
	})
}

func TestWithAttrs(t *testing.T) {
	handler := &multiHandler{
		handlers: []slog.Handler{
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		},
	}

	attrs := []slog.Attr{
		slog.String("key", "value"),
		slog.Int("count", 42),
	}

	newHandler := handler.WithAttrs(attrs)
	if newHandler == nil {
		t.Error("WithAttrs() returned nil")
	}

	multi, ok := newHandler.(*multiHandler)
	if !ok {
		t.Error("WithAttrs() did not return *multiHandler")
	}

	if len(multi.handlers) != len(handler.handlers) {
		t.Errorf("WithAttrs() handler count = %v, want %v", len(multi.handlers), len(handler.handlers))
	}
}

func TestWithGroup(t *testing.T) {
	handler := &multiHandler{
		handlers: []slog.Handler{
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		},
	}

	newHandler := handler.WithGroup("testgroup")
	if newHandler == nil {
		t.Error("WithGroup() returned nil")
	}

	multi, ok := newHandler.(*multiHandler)
	if !ok {
		t.Error("WithGroup() did not return *multiHandler")
	}

	if len(multi.handlers) != len(handler.handlers) {
		t.Errorf("WithGroup() handler count = %v, want %v", len(multi.handlers), len(handler.handlers))
	}
}
