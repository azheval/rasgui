package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type HourlyWriter struct {
	dir      string
	prefix   string
	mu       sync.Mutex
	file     *os.File
	current  string
	mirrored io.Writer
}

func NewHourlyWriter(dir, prefix string, mirrored io.Writer) *HourlyWriter {
	return &HourlyWriter{dir: dir, prefix: prefix, mirrored: mirrored}
}

func (w *HourlyWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	hourKey := time.Now().Format("20060102-15")
	if w.file == nil || w.current != hourKey {
		if err := w.rotate(hourKey); err != nil {
			return 0, err
		}
	}

	if w.mirrored != nil {
		_, _ = w.mirrored.Write(p)
	}
	return w.file.Write(p)
}

func (w *HourlyWriter) rotate(hourKey string) error {
	if err := os.MkdirAll(w.dir, 0o755); err != nil {
		return err
	}
	if w.file != nil {
		_ = w.file.Close()
	}
	name := fmt.Sprintf("%s-%s.log", w.prefix, hourKey)
	file, err := os.OpenFile(filepath.Join(w.dir, name), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	w.file = file
	w.current = hourKey
	return nil
}

func (w *HourlyWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

func BuildLogger(dir, level string) (*slog.Logger, io.Closer, error) {
	writer := NewHourlyWriter(dir, "rasgui", os.Stdout)
	lvl := parseLevel(level)
	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: lvl})
	return slog.New(handler), writer, nil
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
