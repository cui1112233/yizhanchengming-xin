package httpapi

import (
	"net"
	"strings"
	"sync"
	"time"
)

type LoginRateLimitOptions struct {
	MaxFailures int
	Window      time.Duration
	Now         func() time.Time
}

type loginFailureBucket struct {
	failures int
	start    time.Time
}

// LoginRateLimiter is intentionally local to the Go API process: it is a basic
// brute-force boundary, not a long-term source of truth. A distributed limiter
// can later replace it without changing auth/session persistence semantics.
type LoginRateLimiter struct {
	mu          sync.Mutex
	buckets     map[string]loginFailureBucket
	maxFailures int
	window      time.Duration
	now         func() time.Time
}

func NewLoginRateLimiter(options LoginRateLimitOptions) *LoginRateLimiter {
	maxFailures := options.MaxFailures
	if maxFailures <= 0 {
		maxFailures = 5
	}
	window := options.Window
	if window <= 0 {
		window = 5 * time.Minute
	}
	now := options.Now
	if now == nil {
		now = time.Now
	}
	return &LoginRateLimiter{
		buckets:     make(map[string]loginFailureBucket),
		maxFailures: maxFailures,
		window:      window,
		now:         now,
	}
}

func (l *LoginRateLimiter) Allow(key string) bool {
	if l == nil || key == "" {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	bucket, ok := l.buckets[key]
	if !ok {
		return true
	}
	now := l.now()
	if now.Sub(bucket.start) >= l.window {
		delete(l.buckets, key)
		return true
	}
	return bucket.failures < l.maxFailures
}

func (l *LoginRateLimiter) RecordFailure(key string) {
	if l == nil || key == "" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	bucket, ok := l.buckets[key]
	if !ok || now.Sub(bucket.start) >= l.window {
		l.buckets[key] = loginFailureBucket{failures: 1, start: now}
		return
	}
	bucket.failures++
	l.buckets[key] = bucket
}

func (l *LoginRateLimiter) Reset(key string) {
	if l == nil || key == "" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.buckets, key)
}

func limiterKey(remoteAddr, username string) string {
	host := strings.TrimSpace(remoteAddr)
	if parsed, _, err := net.SplitHostPort(host); err == nil {
		host = parsed
	}
	return strings.ToLower(host + "|" + strings.TrimSpace(username))
}
