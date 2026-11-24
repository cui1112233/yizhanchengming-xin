package yzcmusage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type identityKey struct{}
type identity struct{ Owner, Feature string }

func WithIdentity(ctx context.Context, owner, feature string) context.Context {
	previous, _ := ctx.Value(identityKey{}).(identity)
	if owner != "" && previous.Owner == "" {
		previous.Owner = owner
	}
	if feature != "" {
		previous.Feature = feature
	}
	return context.WithValue(ctx, identityKey{}, previous)
}
func Fingerprint(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

type Tracker struct {
	ReceiptDir        string
	Store             Store
	Client            Client
	ResolveCredential func(context.Context, string, string) (string, error)
	wake              chan struct{}
	once              sync.Once
	credentialMu      sync.Mutex
	credentials       map[string]cachedCredential
}

type cachedCredential struct {
	key   string
	until time.Time
}

func (t *Tracker) remember(key string) {
	t.credentialMu.Lock()
	defer t.credentialMu.Unlock()
	if t.credentials == nil {
		t.credentials = map[string]cachedCredential{}
	}
	hash := Fingerprint(key)
	if len(t.credentials) >= 256 {
		for name, value := range t.credentials {
			if time.Now().After(value.until) {
				delete(t.credentials, name)
			}
		}
		if len(t.credentials) >= 256 {
			for name := range t.credentials {
				delete(t.credentials, name)
				break
			}
		}
	}
	t.credentials[hash] = cachedCredential{key: key, until: time.Now().Add(15 * time.Minute)}
}
func (t *Tracker) credential(hash string) string {
	t.credentialMu.Lock()
	defer t.credentialMu.Unlock()
	if value, ok := t.credentials[hash]; ok && time.Now().Before(value.until) {
		return value.key
	}
	delete(t.credentials, hash)
	return ""
}

func (t *Tracker) Run(ctx context.Context) {
	t.once.Do(func() {
		t.wake = make(chan struct{}, 1)
		go func() {
			t.flushReceipts(ctx)
			timer := time.NewTicker(20 * time.Second)
			defer timer.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-timer.C:
				case <-t.wake:
				}
				t.flushReceipts(ctx)
				t.reconcile(ctx)
			}
		}()
	})
}
func (t *Tracker) reconcile(parent context.Context) {
	ctx, cancel := context.WithTimeout(parent, 15*time.Second)
	defer cancel()
	pending, err := t.Store.Pending(ctx)
	if err != nil {
		log.Print("YZCM usage: pending lookup failed")
		return
	}
	if len(pending) == 0 {
		return
	}
	rates, err := t.Client.Rates(ctx)
	if err != nil {
		return
	}
	groups := map[string][]Pending{}
	for _, p := range pending {
		groups[p.CredentialHash] = append(groups[p.CredentialHash], p)
	}
	for _, g := range groups {
		key := t.credential(g[0].CredentialHash)
		if key == "" && t.ResolveCredential != nil {
			owners := map[string]bool{}
			for _, pending := range g {
				if owners[pending.Owner] {
					continue
				}
				owners[pending.Owner] = true
				if resolved, resolveErr := t.ResolveCredential(ctx, pending.Owner, pending.CredentialHash); resolveErr == nil && resolved != "" {
					key = resolved
					t.remember(key)
					break
				}
			}
		}
		if key == "" {
			continue
		}
		for start := 0; start < len(g); start += 50 {
			end := start + 50
			if end > len(g) {
				end = len(g)
			}
			requests := []Request{}
			for _, p := range g[start:end] {
				requests = append(requests, p.Request)
			}
			logs, err := t.Client.Logs(ctx, key, requests)
			if err != nil {
				continue
			}
			owners := map[string]bool{}
			for _, request := range requests {
				owners[request.Owner] = true
			}
			for owner := range owners {
				entries, err := Match(owner, requests, logs, rates)
				if err != nil {
					continue
				}
				for _, entry := range entries {
					if err := t.Store.Settle(ctx, entry); err != nil {
						log.Print("YZCM usage: settlement persistence failed")
					}
				}
			}
		}
	}
}

type trackedTransport struct {
	base http.RoundTripper
	t    *Tracker
}

func (t *Tracker) Wrap(base http.RoundTripper) http.RoundTripper {
	return &trackedTransport{base: base, t: t}
}
func (t *trackedTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	i, _ := r.Context().Value(identityKey{}).(identity)
	path := strings.TrimRight(r.URL.Path, "/")
	paid := r.Method == http.MethodPost && (path == "/v1/chat/completions" || path == "/v1/responses" || path == "/v1/embeddings" || path == "/v1/rerank" || path == "/v1/completions" || path == "/v1/images/generations" || path == "/v1/images/edits" || path == "/v1/audio/speech" || path == "/v1/audio/transcriptions")
	if r.URL.Hostname() != "newapi.yzcm.site" || !paid || i.Owner == "" {
		return t.base.RoundTrip(r)
	}
	response, err := t.base.RoundTrip(r)
	if err != nil {
		return response, err
	}
	id := strings.TrimSpace(response.Header.Get("X-Oneapi-Request-Id"))
	key := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	if id == "" || key == "" || len(id) > 64 {
		return response, nil
	}
	feature := i.Feature
	if feature == "" {
		feature = "model"
	}
	if strings.Contains(path, "/images/") {
		feature = "image"
	}
	if strings.Contains(path, "/audio/") {
		feature = "audio"
	}
	t.t.remember(key)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.t.recordReceipt(ctx, Request{ID: id, Owner: i.Owner, Feature: feature}, Fingerprint(key), response.StatusCode, time.Now().UTC())
	cancel()
	// A bounded worker, rather than a goroutine per response, performs lookups.
	return response, nil
}
