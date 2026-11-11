package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"qiantie/backend/internal/safehttp"
)

// HTTPProvider is deliberately provider-neutral: the deployment supplies the
// provider's approved gateway endpoint. It never logs the credential payload.
type HTTPProvider struct {
	Provider Provider
	Endpoint string
	APIKey   string
	Client   *http.Client
}

func (p *HTTPProvider) Validate() error {
	if _, err := safeEndpoint(p.Endpoint); err != nil {
		return fmt.Errorf("external provider endpoint: %w", err)
	}
	if strings.TrimSpace(p.APIKey) == "" {
		return fmt.Errorf("external provider API key is required")
	}
	if p.Provider != Provider121 && p.Provider != ProviderYadi {
		return ErrInvalid
	}
	return nil
}

func (p *HTTPProvider) Submit(ctx context.Context, credential CredentialInput, intent SubmissionIntent) (ProviderReference, error) {
	if err := p.Validate(); err != nil {
		return ProviderReference{}, err
	}
	endpoint, _ := safeEndpoint(p.Endpoint)
	payload := map[string]any{
		"provider": string(p.Provider), "intentId": intent.ID, "batchId": intent.BatchID, "bookId": intent.BookID,
		"payloadDigest": intent.PayloadDigest, "payload": json.RawMessage(intent.Payload),
		"credential": credential,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return ProviderReference{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return ProviderReference{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)
	client := p.Client
	if client == nil {
		client = safehttp.NewClientWithTimeout(60 * time.Second)
	}
	resp, err := client.Do(req)
	if err != nil {
		return ProviderReference{}, fmt.Errorf("external provider request failed")
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return ProviderReference{}, fmt.Errorf("read external provider response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ProviderReference{}, fmt.Errorf("external provider returned HTTP %d", resp.StatusCode)
	}
	var value map[string]any
	if err := json.Unmarshal(responseBody, &value); err != nil {
		return ProviderReference{}, fmt.Errorf("invalid external provider response")
	}
	reference := firstString(value, "reference", "providerReference", "provider_reference", "id", "taskId", "task_id")
	if reference == "" {
		return ProviderReference{}, fmt.Errorf("external provider response did not contain a reference")
	}
	status := firstString(value, "status", "state")
	if status == "" {
		status = "submitted"
	}
	return ProviderReference{Provider: p.Provider, Reference: boundedMessage(reference), Status: boundedMessage(status)}, nil
}

func safeEndpoint(raw string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed == nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, fmt.Errorf("endpoint must be an https URL without credentials, query, or fragment")
	}
	return parsed, nil
}

func firstString(value map[string]any, keys ...string) string {
	for _, key := range keys {
		if text, ok := value[key].(string); ok && strings.TrimSpace(text) != "" {
			return strings.TrimSpace(text)
		}
		if number, ok := value[key].(float64); ok {
			return fmt.Sprintf("%.0f", number)
		}
	}
	return ""
}
