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
	Store             Store
	Client            Client
	ResolveCredential func(context.Context, string, string) (string, error)
	wake              chan struct{}
	once              sync.Once
}

func (t *Tracker) Run(ctx context.Context) {
	t.once.Do(func() {
		t.wake = make(chan struct{}, 1)
		go func() {
			timer := time.NewTicker(20 * time.Second)
			defer timer.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-timer.C:
				case <-t.wake:
				}
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
		groups[p.Owner+":"+p.CredentialHash] = append(groups[p.Owner+":"+p.CredentialHash], p)
	}
	for _, g := range groups {
		if t.ResolveCredential == nil {
			return
		}
		key, err := t.ResolveCredential(ctx, g[0].Owner, g[0].CredentialHash)
		if err != nil || key == "" {
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
			entries, err := Match(g[0].Owner, requests, logs, rates)
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
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	err = t.t.Store.Record(ctx, Request{ID: id, Owner: i.Owner, Feature: feature}, Fingerprint(key), response.StatusCode)
	cancel()
	if err != nil {
		log.Print("YZCM usage: request persistence failed")
	}
	// A bounded worker, rather than a goroutine per response, performs lookups.
	return response, nil
}
