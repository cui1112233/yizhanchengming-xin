package httpapi

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// AuthRateLimitConfig holds the thresholds for the sensitive authentication
// endpoints. Zero values are replaced by the defaults below.
type AuthRateLimitConfig struct {
	// Login is failure based: only rejected credentials / MFA codes count.
	LoginWindow      time.Duration
	LoginMaxFailures int
	LoginLockout     time.Duration
	// Register counts every application submission.
	RegisterWindow  time.Duration
	RegisterMax     int
	RegisterLockout time.Duration
	// PasswordReset counts every reset request / token attempt.
	PasswordResetWindow  time.Duration
	PasswordResetMax     int
	PasswordResetLockout time.Duration
}

// Default thresholds. The same values are documented in config.FromEnv and can
// be overridden with QIANTIE_AUTH_RATE_* environment variables.
const (
	defaultLoginRateWindow      = 10 * time.Minute
	defaultLoginRateMaxFailures = 10
	defaultLoginRateLockout     = 15 * time.Minute

	defaultRegisterRateWindow  = 10 * time.Minute
	defaultRegisterRateMax     = 5
	defaultRegisterRateLockout = 15 * time.Minute

	defaultPasswordResetRateWindow  = 60 * time.Second
	defaultPasswordResetRateMax     = 5
	defaultPasswordResetRateLockout = 15 * time.Minute
)

const (
	authRateScopeLogin         = "login"
	authRateScopeRegister      = "register"
	authRateScopePasswordReset = "password_reset"
)

type authRateLimitBucket struct {
	window  time.Duration
	max     int
	lockout time.Duration
}

type authRateLimitEntry struct {
	attempts     int
	windowStart  time.Time
	blockedUntil time.Time
}

// authRateLimiter is an in-process, per-instance limiter. It keeps the auth
// endpoints usable without Redis while still making brute force and mail
// bombing expensive. Entries are keyed by scope + dimension (IP or account).
type authRateLimiter struct {
	mu      sync.Mutex
	now     func() time.Time
	buckets map[string]authRateLimitBucket
	entries map[string]*authRateLimitEntry
}

func newAuthRateLimiter(config AuthRateLimitConfig, now func() time.Time) *authRateLimiter {
	if now == nil {
		now = time.Now
	}
	return &authRateLimiter{
		now: now,
		buckets: map[string]authRateLimitBucket{
			authRateScopeLogin: {
				window:  durationOr(config.LoginWindow, defaultLoginRateWindow),
				max:     intOr(config.LoginMaxFailures, defaultLoginRateMaxFailures),
				lockout: durationOr(config.LoginLockout, defaultLoginRateLockout),
			},
			authRateScopeRegister: {
				window:  durationOr(config.RegisterWindow, defaultRegisterRateWindow),
				max:     intOr(config.RegisterMax, defaultRegisterRateMax),
				lockout: durationOr(config.RegisterLockout, defaultRegisterRateLockout),
			},
			authRateScopePasswordReset: {
				window:  durationOr(config.PasswordResetWindow, defaultPasswordResetRateWindow),
				max:     intOr(config.PasswordResetMax, defaultPasswordResetRateMax),
				lockout: durationOr(config.PasswordResetLockout, defaultPasswordResetRateLockout),
			},
		},
		entries: map[string]*authRateLimitEntry{},
	}
}

func durationOr(value, fallback time.Duration) time.Duration {
	if value > 0 {
		return value
	}
	return fallback
}

func intOr(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func authRateEntryKey(scope, dimension string) string {
	return scope + "\x00" + dimension
}

// allow reports whether the key may proceed. A non-zero retryAfter means the key
// is locked; callers should return 429 with that delay.
func (l *authRateLimiter) allow(scope, dimension string) (time.Duration, bool) {
	if l == nil {
		return 0, true
	}
	bucket, ok := l.buckets[scope]
	if !ok || bucket.max <= 0 || bucket.window <= 0 {
		return 0, true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	key := authRateEntryKey(scope, dimension)
	entry := l.entries[key]
	if entry == nil {
		return 0, true
	}
	if entry.blockedUntil.After(now) {
		return entry.blockedUntil.Sub(now), false
	}
	if entry.blockedUntil.IsZero() && now.Sub(entry.windowStart) < bucket.window && entry.attempts >= bucket.max {
		// Defensive fallback: a full window without an explicit block.
		entry.blockedUntil = now.Add(bucket.lockout)
		return bucket.lockout, false
	}
	if now.Sub(entry.windowStart) >= bucket.window {
		delete(l.entries, key)
		return 0, true
	}
	return 0, true
}

// record counts one attempt. When the threshold is reached the key is locked
// immediately, so the next request receives 429.
func (l *authRateLimiter) record(scope, dimension string) {
	if l == nil {
		return
	}
	bucket, ok := l.buckets[scope]
	if !ok || bucket.max <= 0 || bucket.window <= 0 {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	key := authRateEntryKey(scope, dimension)
	entry := l.entries[key]
	if entry == nil || (!entry.blockedUntil.IsZero() && !entry.blockedUntil.After(now)) || now.Sub(entry.windowStart) >= bucket.window {
		entry = &authRateLimitEntry{windowStart: now}
		l.entries[key] = entry
	}
	entry.attempts++
	if entry.attempts >= bucket.max {
		entry.attempts = 0
		entry.windowStart = now
		entry.blockedUntil = now.Add(bucket.lockout)
	}
	l.pruneLocked(now)
}

// reset clears the counter after a successful authentication.
func (l *authRateLimiter) reset(scope, dimension string) {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, authRateEntryKey(scope, dimension))
}

// pruneLocked drops expired entries so the map cannot grow without bound.
func (l *authRateLimiter) pruneLocked(now time.Time) {
	if len(l.entries) < 1024 {
		return
	}
	for key, entry := range l.entries {
		bucket := l.buckets[strings.SplitN(key, "\x00", 2)[0]]
		expiredBlock := entry.blockedUntil.IsZero() || !entry.blockedUntil.After(now)
		expiredWindow := now.Sub(entry.windowStart) >= bucket.window
		if expiredBlock && expiredWindow {
			delete(l.entries, key)
		}
	}
}

// authRateLimitIP returns the direct peer address. Proxy headers are
// intentionally ignored: trusting X-Forwarded-For would let a client pick a new
// identity per request and bypass the limiter.
func authRateLimitIP(req *http.Request) string {
	if req == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(req.RemoteAddr))
	if err == nil {
		return host
	}
	return strings.TrimSpace(req.RemoteAddr)
}

func authRateLimitIPDimension(req *http.Request) string {
	return "ip:" + authRateLimitIP(req)
}

func authRateLimitUserDimension(username string) string {
	return "user:" + strings.ToLower(strings.TrimSpace(username))
}

// writeAuthRateLimited returns a Chinese 429 response with Retry-After.
func writeAuthRateLimited(w http.ResponseWriter, retryAfter time.Duration) {
	seconds := int(math.Ceil(retryAfter.Seconds()))
	if seconds < 1 {
		seconds = 1
	}
	w.Header().Set("Retry-After", strconv.Itoa(seconds))
	writeJSON(w, http.StatusTooManyRequests, map[string]any{
		"error": "操作过于频繁，请稍后再试",
		"code":  "RATE_LIMITED",
	})
}
