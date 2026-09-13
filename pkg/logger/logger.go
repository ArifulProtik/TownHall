// Package logger builds the process-wide slog logger with request-ID helpers.
package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// RequestIDKey is the echo-context key and slog attr name for request IDs.
const RequestIDKey = "request_id"

// New returns a process-wide logger writing to stdout:
// JSON in production (collector-friendly), human-readable text otherwise.
// Every record carries file:line via AddSource.
func New(env, level string) *slog.Logger {
	return NewWithWriter(env, level, os.Stdout)
}

// NewWithWriter is New with an injectable destination (tests, future file output).
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

// shortSource trims source.file to the last two path segments
// (e.g. auth/auth_service.go) so log lines stay readable.
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

// WithRequestID returns a logger carrying the request ID, or l unchanged when empty.
func WithRequestID(l *slog.Logger, requestID string) *slog.Logger {
	if requestID == "" {
		return l
	}
	return l.With(slog.String(RequestIDKey, requestID))
}

type ctxKey struct{}

// ContextWithRequestID stashes the request ID in a std context (echo context
// doesn't cross into services; this does without changing signatures).
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ctxKey{}, requestID)
}

// RequestIDFromContext extracts the request ID, or "" when absent.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}

// WithContext returns a logger carrying the request ID from ctx, if any.
func WithContext(ctx context.Context, l *slog.Logger) *slog.Logger {
	return WithRequestID(l, RequestIDFromContext(ctx))
}
