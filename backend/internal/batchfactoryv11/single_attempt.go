package batchfactoryv11

import (
	"context"
	"net/http"
)

type singleAttemptKey struct{}

// WithSingleAttempt disables implicit resends for callers with a durable
// per-attempt journal. Other modules retain their existing retry policy.
func WithSingleAttempt(ctx context.Context) context.Context {
	return context.WithValue(ctx, singleAttemptKey{}, true)
}
func SingleAttempt(ctx context.Context) bool { v, _ := ctx.Value(singleAttemptKey{}).(bool); return v }
func SingleAttemptClient(ctx context.Context, client *http.Client, req *http.Request) *http.Client {
	if !SingleAttempt(ctx) {
		return client
	}
	copy := *client
	copy.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	// A POST with no replay body cannot be transparently replayed by net/http.
	req.GetBody = nil
	req.Header.Del("Idempotency-Key")
	req.Header.Del("X-Idempotency-Key")
	return &copy
}
