package auth

import (
	"sync"
	"time"
)

// Auth rate limits: 5 attempts per email and 20 per IP per sliding window.
const (
	authLimitWindow   = 15 * time.Minute
	authLimitPerEmail = 5
	authLimitPerIP    = 20
)

// Limiter gates auth attempts per email and per IP. Implemented in-memory;
// swap for a Redis backend later without changing callers.
type Limiter interface {
	Allow(email, ip string) (allowed bool, retryAfter time.Duration)
}

// MemoryLimiter is an in-memory sliding-window Limiter.
type MemoryLimiter struct {
	mu     sync.Mutex
	emails map[string][]time.Time
	ips    map[string][]time.Time
	now    func() time.Time
}

// NewMemoryLimiter builds a Limiter with the auth limits above.
func NewMemoryLimiter() *MemoryLimiter {
	return &MemoryLimiter{
		emails: make(map[string][]time.Time),
		ips:    make(map[string][]time.Time),
		now:    time.Now,
	}
}

// Allow records an attempt for email+ip when both quotas allow it,
// otherwise reports how long the caller must wait.
func (l *MemoryLimiter) Allow(email, ip string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	emailHits := pruneHits(l.emails[email], now)
	ipHits := pruneHits(l.ips[ip], now)
	if len(emailHits) >= authLimitPerEmail {
		return false, emailHits[0].Add(authLimitWindow).Sub(now)
	}
	if len(ipHits) >= authLimitPerIP {
		return false, ipHits[0].Add(authLimitWindow).Sub(now)
	}
	l.emails[email] = append(emailHits, now)
	l.ips[ip] = append(ipHits, now)
	return true, 0
}

// pruneHits drops timestamps outside the current window (oldest first).
func pruneHits(hits []time.Time, now time.Time) []time.Time {
	cutoff := now.Add(-authLimitWindow)
	kept := make([]time.Time, 0, len(hits))
	for _, h := range hits {
		if h.After(cutoff) {
			kept = append(kept, h)
		}
	}
	return kept
}
