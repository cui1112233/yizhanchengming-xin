package shuihuo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// MediaProvider is deliberately small: providers receive a prompt and return a
// short-lived media URL.  The worker, rather than the provider, owns durable
// TOS persistence and task state.
type MediaProvider interface {
	Available(MediaKind) bool
	Generate(context.Context, MediaTask, string) (GeneratedMedia, error)
}

type GeneratedMedia struct{ URL, ContentType string }

type HTTPProviderConfig struct {
	BaseURL, APIKey, ImageModel, TTSModel string
	Client                                *http.Client
}
type HTTPProvider struct {
	baseURL, key, imageModel, ttsModel string
	client                             *http.Client
}
type ProviderSet struct{ Image, Audio MediaProvider }

func (p ProviderSet) Available(kind MediaKind) bool {
	if kind == MediaImage {
		return p.Image != nil && p.Image.Available(kind)
	}
	return kind == MediaAudio && p.Audio != nil && p.Audio.Available(kind)
}
func (p ProviderSet) Generate(ctx context.Context, task MediaTask, prompt string) (GeneratedMedia, error) {
	if task.Kind == MediaImage && p.Image != nil {
		return p.Image.Generate(ctx, task, prompt)
	}
	if task.Kind == MediaAudio && p.Audio != nil {
		return p.Audio.Generate(ctx, task, prompt)
	}
	return GeneratedMedia{}, ErrProviderUnavailable
}

func NewHTTPProvider(c HTTPProviderConfig) *HTTPProvider {
	return &HTTPProvider{baseURL: strings.TrimRight(strings.TrimSpace(c.BaseURL), "/"), key: strings.TrimSpace(c.APIKey), imageModel: strings.TrimSpace(c.ImageModel), ttsModel: strings.TrimSpace(c.TTSModel), client: c.Client}
}
func (p *HTTPProvider) Available(kind MediaKind) bool {
	if p == nil || p.baseURL == "" || p.key == "" {
		return false
	}
	return (kind == MediaImage && p.imageModel != "") || (kind == MediaAudio && p.ttsModel != "")
}
func (p *HTTPProvider) Generate(ctx context.Context, task MediaTask, prompt string) (GeneratedMedia, error) {
	if !p.Available(task.Kind) {
		return GeneratedMedia{}, ErrProviderUnavailable
	}
	model := p.imageModel
	if task.Kind == MediaAudio {
		model = p.ttsModel
	}
	// The browser may record a requested model for display, but the executable
	// model is server configuration so callers cannot select an unapproved or
	// billable provider model.
	body, _ := json.Marshal(map[string]string{"kind": string(task.Kind), "model": model, "prompt": prompt, "requestId": task.RequestID})
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/generate", bytes.NewReader(body))
	if err != nil {
		return GeneratedMedia{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+p.key)
	client := p.client
	if client == nil {
		client = &http.Client{Timeout: 90 * time.Second}
	}
	response, err := client.Do(request)
	if err != nil {
		return GeneratedMedia{}, fmt.Errorf("provider request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return GeneratedMedia{}, fmt.Errorf("provider returned status %d", response.StatusCode)
	}
	var out struct {
		URL         string `json:"url"`
		ContentURL  string `json:"contentUrl"`
		ContentType string `json:"contentType"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&out); err != nil {
		return GeneratedMedia{}, fmt.Errorf("provider response: %w", err)
	}
	if out.URL == "" {
		out.URL = out.ContentURL
	}
	if strings.TrimSpace(out.URL) == "" {
		return GeneratedMedia{}, errors.New("provider response has no media URL")
	}
	return GeneratedMedia{URL: out.URL, ContentType: strings.ToLower(strings.TrimSpace(strings.Split(out.ContentType, ";")[0]))}, nil
}

var ErrProviderUnavailable = errors.New("shuihuo: provider unavailable")
