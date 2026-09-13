package logger_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"

	"ArifulProtik/TownHall/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWithWriter_ProductionEmitsJSONWithSource(t *testing.T) {
	var buf bytes.Buffer
	l := logger.NewWithWriter("production", "info", &buf)
	require.NotNil(t, l)

	l.Info("hello")

	var m map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m), "production output must be JSON")
	assert.Equal(t, "hello", m["msg"])
	src, ok := m["source"].(map[string]any)
	require.True(t, ok, "expected source with file:line, got: %s", buf.String())
	assert.Contains(t, src["file"], "logger_test.go")
	assert.NotZero(t, src["line"])
}

func TestNewWithWriter_SourceIsShortPath(t *testing.T) {
	var buf bytes.Buffer
	l := logger.NewWithWriter("production", "info", &buf)

	l.Info("hello")

	var m map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	src, ok := m["source"].(map[string]any)
	require.True(t, ok, "expected source, got: %s", buf.String())
	file, _ := src["file"].(string)
	assert.Equal(t, "logger/logger_test.go", file)
	assert.NotContains(t, file, "workspace", "source must be trimmed to last two segments")
}

func TestNewWithWriter_DevelopmentIsPrettySingleLine(t *testing.T) {
	var buf bytes.Buffer
	l := logger.NewWithWriter("development", "info", &buf)

	l.Info("POST /ping → 200 · 3ms", slog.String("request_id", "req-1"))

	out := buf.String()
	assert.NotContains(t, out, "\x1b[", "non-TTY output must not contain ANSI codes")
	assert.Regexp(t, `^\d{2}:\d{2}:\d{2}\.\d{3} INFO `, out, "line starts with short time and level")
	assert.Contains(t, out, "POST /ping → 200 · 3ms")
	assert.Contains(t, out, "request_id=req-1")
	assert.Contains(t, out, "source=")
	assert.NotContains(t, out, "\n\n")
	assert.Equal(t, 1, bytes.Count([]byte(out), []byte("\n")), "one record per line")
}

func TestParseLevel(t *testing.T) {
	assert.Equal(t, slog.LevelDebug, logger.ParseLevel("debug"))
	assert.Equal(t, slog.LevelInfo, logger.ParseLevel("info"))
	assert.Equal(t, slog.LevelWarn, logger.ParseLevel("warn"))
	assert.Equal(t, slog.LevelError, logger.ParseLevel("error"))
	assert.Equal(t, slog.LevelInfo, logger.ParseLevel("bogus"), "invalid level falls back to info")
	assert.Equal(t, slog.LevelInfo, logger.ParseLevel(""))
}

func TestLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := logger.NewWithWriter("production", "error", &buf)

	l.Warn("suppressed")
	assert.Empty(t, buf.String())

	l.Error("emitted")
	assert.Contains(t, buf.String(), "emitted")
}

func TestWithRequestID(t *testing.T) {
	var buf bytes.Buffer
	l := logger.NewWithWriter("production", "info", &buf)

	logger.WithRequestID(l, "req-123").Info("hello")

	var m map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	assert.Equal(t, "req-123", m["request_id"])
}

func TestWithRequestID_EmptyReturnsBareLogger(t *testing.T) {
	var buf bytes.Buffer
	l := logger.NewWithWriter("production", "info", &buf)

	logger.WithRequestID(l, "").Info("hello")

	var m map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	_, ok := m["request_id"]
	assert.False(t, ok, "empty request id must not add the attr")
}

func TestRequestIDContextRoundTrip(t *testing.T) {
	ctx := logger.ContextWithRequestID(t.Context(), "req-456")
	assert.Equal(t, "req-456", logger.RequestIDFromContext(ctx))
	assert.Empty(t, logger.RequestIDFromContext(t.Context()))
}

func TestLogWithContextCarriesRequestID(t *testing.T) {
	var buf bytes.Buffer
	l := logger.NewWithWriter("production", "info", &buf)

	ctx := logger.ContextWithRequestID(t.Context(), "req-789")
	logger.WithContext(ctx, l).Warn("duplicate")

	var m map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	assert.Equal(t, "req-789", m["request_id"])
}
