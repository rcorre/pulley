// Package log configures a file-based slog logger for pulley.
// Call Init once at startup; all packages use slog.Debug/Info/Error directly.
package log

import (
	"io"
	"log/slog"
	"os"
)

type nopCloser struct{}

func (nopCloser) Close() error { return nil }

// Init configures the default slog logger to write to path at debug level.
// If path is empty, all output is discarded (avoids polluting the TUI with
// log lines on stderr).
// Returns a closer the caller must call on exit.
func Init(path string) (io.Closer, error) {
	if path == "" {
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
		return nopCloser{}, nil
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600) //nolint:gosec
	if err != nil {
		return nil, err
	}
	h := slog.NewTextHandler(f, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(h))
	return f, nil
}
