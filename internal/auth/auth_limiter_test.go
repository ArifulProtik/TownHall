package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLimiter() (*MemoryLimiter, *time.Time) {
	now := time.Now()
	l := NewMemoryLimiter()
	l.now = func() time.Time { return now }
	advance := func(d time.Duration) { now = now.Add(d) }
	_ = advance
	return l, &now
}

func TestLimiter_PerEmailWindow(t *testing.T) {
	l, now := newTestLimiter()

	for range 5 {
		ok, _ := l.Allow("a@example.com", "1.1.1.1")
		assert.True(t, ok)
	}
	ok, retry := l.Allow("a@example.com", "1.1.1.1")
	assert.False(t, ok)
	assert.Greater(t, retry, time.Duration(0))
	assert.LessOrEqual(t, retry, 15*time.Minute)

	// Other email on the same IP still has quota.
	ok, _ = l.Allow("b@example.com", "1.1.1.1")
	assert.True(t, ok)

	// After the window slides past, quota returns.
	*now = now.Add(16 * time.Minute)
	ok, _ = l.Allow("a@example.com", "1.1.1.1")
	assert.True(t, ok)
}

func TestLimiter_PerIPWindow(t *testing.T) {
	l, _ := newTestLimiter()

	for i := range 20 {
		ok, _ := l.Allow("user"+string(rune('a'+i))+"@example.com", "9.9.9.9")
		require.True(t, ok)
	}
	ok, retry := l.Allow("last@example.com", "9.9.9.9")
	assert.False(t, ok)
	assert.Greater(t, retry, time.Duration(0))

	// Same email from a different IP still has email quota (used 1 of 5).
	ok, _ = l.Allow("last@example.com", "8.8.8.8")
	assert.True(t, ok)
}
