package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ANSI styles for TTY output. Colors stay off unless the writer is a
// terminal (see colorize): pipes, tests, and CI always get plain text.
const (
	ansiReset = "\x1b[0m"
	ansiDim   = "\x1b[2m"
	ansiBold  = "\x1b[1m"
	ansiRed   = "\x1b[31m"
	ansiGreen = "\x1b[32m"
	ansiYello = "\x1b[33m"
	ansiCyan  = "\x1b[36m"
)

// prettyHandler renders NestJS-style single lines for human eyes:
// 18:04:12.345 INFO  POST /ping → 200 · 3ms request_id=abc source=pkg/a.go:1
// Level color doubles as status color: access logs already map
// status class to record level (Info <400, Warn 4xx, Error 5xx).
type prettyHandler struct {
	w      io.Writer
	level  slog.Level
	color  bool
	mu     *sync.Mutex
	attrs  []slog.Attr
	groups []string
}

func newPrettyHandler(w io.Writer, level slog.Level) *prettyHandler {
	return &prettyHandler{w: w, level: level, color: colorize(w), mu: &sync.Mutex{}}
}

func (h *prettyHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level
}

func (h *prettyHandler) Handle(_ context.Context, r slog.Record) error {
	var b strings.Builder

	b.WriteString(h.style(ansiDim, r.Time.Format("15:04:05.000")))
	b.WriteString(" ")
	b.WriteString(h.levelString(r.Level))
	b.WriteString(" ")
	b.WriteString(r.Message)

	attrs := make([]slog.Attr, 0, len(h.attrs)+r.NumAttrs())
	attrs = append(attrs, h.attrs...)
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})
	for _, a := range attrs {
		b.WriteString(" ")
		b.WriteString(h.formatAttr(a))
	}

	if r.PC != 0 {
		if fs := frameFileLine(r.PC); fs != "" {
			b.WriteString(" ")
			b.WriteString(h.style(ansiDim, "source="+fs))
		}
	}
	b.WriteString("\n")

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := io.WriteString(h.w, b.String())
	return err
}

func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	cp := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	cp = append(cp, h.attrs...)
	cp = append(cp, attrs...)
	return &prettyHandler{w: h.w, level: h.level, color: h.color, mu: h.mu, attrs: cp, groups: h.groups}
}

func (h *prettyHandler) WithGroup(name string) slog.Handler {
	groups := make([]string, 0, len(h.groups)+1)
	groups = append(groups, h.groups...)
	groups = append(groups, name)
	return &prettyHandler{w: h.w, level: h.level, color: h.color, mu: h.mu, attrs: h.attrs, groups: groups}
}

func (h *prettyHandler) style(code, s string) string {
	if !h.color {
		return s
	}
	return code + s + ansiReset
}

func (h *prettyHandler) levelString(l slog.Level) string {
	name := l.String()
	switch {
	case l >= slog.LevelError:
		return h.style(ansiBold+ansiRed, name)
	case l >= slog.LevelWarn:
		return h.style(ansiBold+ansiYello, name)
	case l >= slog.LevelDebug && l < slog.LevelInfo:
		return h.style(ansiCyan, name)
	default:
		return h.style(ansiBold+ansiGreen, name)
	}
}

func (h *prettyHandler) formatAttr(a slog.Attr) string {
	key := strings.Join(append(append([]string{}, h.groups...), a.Key), ".")
	return h.style(ansiDim, key+"=") + formatValue(a.Value)
}

func formatValue(v slog.Value) string {
	var s string
	switch v.Kind() {
	case slog.KindString:
		s = v.String()
	case slog.KindInt64:
		return strconv.FormatInt(v.Int64(), 10)
	case slog.KindUint64:
		return strconv.FormatUint(v.Uint64(), 10)
	case slog.KindFloat64:
		return strconv.FormatFloat(v.Float64(), 'f', -1, 64)
	case slog.KindBool:
		return strconv.FormatBool(v.Bool())
	case slog.KindDuration:
		return v.Duration().String()
	case slog.KindTime:
		return v.Time().Format(time.RFC3339)
	case slog.KindGroup:
		parts := make([]string, 0, len(v.Group()))
		for _, a := range v.Group() {
			parts = append(parts, a.Key+"="+formatValue(a.Value))
		}
		s = "(" + strings.Join(parts, " ") + ")"
	case slog.KindAny:
		fallthrough
	default:
		if err, ok := v.Any().(error); ok {
			s = err.Error()
		} else {
			s = fmt.Sprintf("%v", v.Any())
		}
	}
	if strings.ContainsAny(s, " \t\"'= ") {
		return strconv.Quote(s)
	}
	return s
}

// colorize reports whether w deserves ANSI colors: a real terminal,
// with NO_COLOR unset and a non-dumb TERM.
func colorize(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	st, err := f.Stat()
	if err != nil {
		return false
	}
	return st.Mode()&os.ModeCharDevice != 0
}

// frameFileLine resolves a program counter to a trimmed file:line.
func frameFileLine(pc uintptr) string {
	frames := runtime.CallersFrames([]uintptr{pc})
	f, _ := frames.Next()
	if f.File == "" {
		return ""
	}
	return shortFile(f.File) + ":" + strconv.Itoa(f.Line)
}
