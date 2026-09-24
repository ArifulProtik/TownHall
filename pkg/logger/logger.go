// Package logger builds the process-wide slog logger.
// JSON in production, human-readable text elsewhere.
package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// RequestIDKey is the echo-context key for request IDs. The access log
// reads it; handlers and services never touch it.
const RequestIDKey = "request_id"

// New returns a process-wide logger writing to stdout.
// Every record carries file:line via AddSource.
func New(env, level string) *slog.Logger {
	return NewWithWriter(env, level, os.Stdout)
}

// NewWithWriter is New with an injectable destination (tests).
func NewWithWriter(env, level string, w io.Writer) *slog.Logger {
	lvl := ParseLevel(level)
	if level != "" && lvl == slog.LevelInfo && !isInfoAlias(level) {
		fmt.Fprintf(os.Stderr, "logger: invalid level %q, falling back to info\n", level)
	}
	opts := &slog.HandlerOptions{Level: lvl, AddSource: true, ReplaceAttr: shortSource}
	if strings.EqualFold(env, "production") {
		return slog.New(slog.NewJSONHandler(w, opts))
	}
	return slog.New(newPrettyHandler(w, lvl))
}

// shortSource trims source paths to two segments (auth/auth_service.go).
func shortSource(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.SourceKey {
		if src, ok := a.Value.Any().(*slog.Source); ok && src != nil {
			src.File = shortFile(src.File)
		}
	}
	return a
}

// shortFile keeps the last two path segments of p.
func shortFile(p string) string {
	parts := strings.Split(p, "/")
	if len(parts) > 2 {
		return strings.Join(parts[len(parts)-2:], "/")
	}
	return p
}

// ParseLevel maps a string to a slog level; unknown or empty means info.
func ParseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
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

func isInfoAlias(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "info":
		return true
	}
	return false
}
