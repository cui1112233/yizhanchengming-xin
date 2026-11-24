// Package yzcmusage reconciles application requests against YZCM's billing log.
// Provider usage.cost_rmb is deliberately ignored: it is the provider cost,
// which can differ from the amount deducted by YZCM.
package yzcmusage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const Origin = "http://newapi.yzcm.site"

type Client struct{ HTTP *http.Client }
type BillingLog struct {
	RequestID    string `json:"request_id"`
	Model        string `json:"model_name"`
	CreatedAt    int64  `json:"created_at"`
	Type         int    `json:"type"`
	Quota        int64  `json:"quota"`
	InputTokens  int64  `json:"prompt_tokens"`
	OutputTokens int64  `json:"completion_tokens"`
}
type Rates struct {
	QuotaPerUnit    float64 `json:"quota_per_unit"`
	USDExchangeRate float64 `json:"usd_exchange_rate"`
}
type Request struct {
	ID      string
	Owner   string
	Feature string
}
type Entry struct {
	ID           string    `json:"id"`
	Owner        string    `json:"-"`
	At           time.Time `json:"at"`
	Feature      string    `json:"feature"`
	Model        string    `json:"model"`
	InputTokens  int64     `json:"inputTokens"`
	OutputTokens int64     `json:"outputTokens"`
	TotalTokens  int64     `json:"totalTokens"`
	Quota        int64     `json:"quota"`
	Cost         float64   `json:"actualCost"`
	Currency     string    `json:"costCurrency"`
	CostKnown    bool      `json:"costKnown"`
	UsageKnown   bool      `json:"usageKnown"`
	Source       string    `json:"source"`
}

func (c Client) read(ctx context.Context, path, key string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, Origin+path, nil)
	if err != nil {
		return err
	}
	if key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	client := c.HTTP
	if client == nil {
		client = &http.Client{Timeout: 12 * time.Second}
	}
	copyClient := *client
	copyClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	response, err := copyClient.Do(req)
	if err != nil {
		return errors.New("YZCM API connection failed")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("YZCM API returned HTTP %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 4<<20+1))
	if err != nil || len(body) > 4<<20 {
		return errors.New("invalid YZCM API response size")
	}
	if json.Unmarshal(body, target) != nil {
		return errors.New("invalid YZCM API response")
	}
	return nil
}

func (c Client) Rates(ctx context.Context) (Rates, error) {
	var response struct {
		Success bool  `json:"success"`
		Data    Rates `json:"data"`
	}
	if err := c.read(ctx, "/api/status", "", &response); err != nil {
		return Rates{}, err
	}
	r := response.Data
	if !response.Success || r.QuotaPerUnit <= 0 || r.USDExchangeRate <= 0 || math.IsInf(r.QuotaPerUnit, 0) || math.IsInf(r.USDExchangeRate, 0) || math.IsNaN(r.QuotaPerUnit) || math.IsNaN(r.USDExchangeRate) {
		return Rates{}, errors.New("invalid YZCM quota conversion")
	}
	return r, nil
}

// Logs requests only IDs already recorded by the application. Keys remain in
// server memory and are never put in URLs or returned in a response.
func (c Client) Logs(ctx context.Context, key string, requests []Request) ([]BillingLog, error) {
	if strings.TrimSpace(key) == "" || len(requests) == 0 || len(requests) > 50 {
		return nil, errors.New("invalid billing lookup")
	}
	ids := make([]string, 0, len(requests))
	seen := map[string]bool{}
	for _, r := range requests {
		if r.ID == "" || len(r.ID) > 64 || strings.ContainsAny(r.ID, ",\r\n") {
			return nil, errors.New("invalid request ID")
		}
		if !seen[r.ID] {
			ids = append(ids, r.ID)
			seen[r.ID] = true
		}
	}
	var response struct {
		Success bool         `json:"success"`
		Data    []BillingLog `json:"data"`
	}
	path := "/api/log/token?request_ids=" + url.QueryEscape(strings.Join(ids, ","))
	if err := c.read(ctx, path, key, &response); err != nil {
		return nil, err
	}
	if !response.Success {
		return nil, errors.New("YZCM billing lookup failed")
	}
	// Older servers may ignore the filter. Never expose their unrelated rows.
	result := []BillingLog{}
	for _, log := range response.Data {
		if seen[log.RequestID] {
			result = append(result, log)
		}
	}
	return result, nil
}

// Match enforces owner isolation even when many members share one API key.
// No timestamp/model-name guessing and no substitution of upstream cost.
func Match(owner string, requests []Request, logs []BillingLog, rates Rates) ([]Entry, error) {
	if rates.QuotaPerUnit <= 0 || rates.USDExchangeRate <= 0 || math.IsNaN(rates.QuotaPerUnit) || math.IsNaN(rates.USDExchangeRate) || math.IsInf(rates.QuotaPerUnit, 0) || math.IsInf(rates.USDExchangeRate, 0) {
		return nil, errors.New("invalid billing conversion")
	}
	owned := map[string]Request{}
	conflict := map[string]bool{}
	for _, r := range requests {
		if old, exists := owned[r.ID]; exists && old.Owner != r.Owner {
			conflict[r.ID] = true
		}
		owned[r.ID] = r
	}
	result := []Entry{}
	seen := map[string]bool{}
	for _, log := range logs {
		r, ok := owned[log.RequestID]
		if !ok || r.Owner != owner || owner == "" || r.ID == "" || conflict[r.ID] || seen[r.ID] || log.Type != 2 || log.Quota < 0 || log.InputTokens < 0 || log.OutputTokens < 0 {
			continue
		}
		cost := float64(log.Quota) / rates.QuotaPerUnit * rates.USDExchangeRate
		if math.IsNaN(cost) || math.IsInf(cost, 0) {
			return nil, errors.New("invalid billed amount")
		}
		seen[r.ID] = true
		result = append(result, Entry{ID: r.ID, Owner: owner, At: time.Unix(log.CreatedAt, 0).UTC(), Feature: r.Feature, Model: log.Model, InputTokens: log.InputTokens, OutputTokens: log.OutputTokens, TotalTokens: log.InputTokens + log.OutputTokens, Quota: log.Quota, Cost: cost, Currency: "CNY", CostKnown: true, UsageKnown: true, Source: "yzcm_billing_log"})
	}
	return result, nil
}
