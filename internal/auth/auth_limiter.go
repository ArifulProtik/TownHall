package auth

import (
	"sync"
	"time"
)

// 5 attempts per email and 20 per IP per sliding window.
const (
	authLimitWindow   = 15 * time.Minute
	authLimitPerEmail = 5
	authLimitPerIP    = 20
)

// Limiter gates auth attempts. In-memory for now; swap for Redis later
// without changing callers.
type Limiter interface {
	Allow(email, ip string) (allowed bool, retryAfter time.Duration)
}

type MemoryLimiter struct {
	mu     sync.Mutex
	emails map[string][]time.Time
	ips    map[string][]time.Time
	now    func() time.Time
}

func NewMemoryLimiter() *MemoryLimiter {
	return &MemoryLimiter{
		emails: make(map[string][]time.Time),
		ips:    make(map[string][]time.Time),
		now:    time.Now,
	}
}

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
