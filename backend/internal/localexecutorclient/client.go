// Package localexecutorclient is the Go client for the local-executor control
// plane.  It deliberately mirrors the versioned HTTP contract served by the
// main Go API so the desktop executor does not need a Node API client.
package localexecutorclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

var ErrNoJob = errors.New("no claimable job")

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

type PairInput struct {
	Code       string `json:"code"`
	DeviceName string `json:"deviceName"`
	Platform   string `json:"platform"`
	OS         string `json:"os"`
	Version    string `json:"version"`
}

type PairResult struct {
	ExecutorID               string `json:"executorId"`
	Token                    string `json:"token"`
	HeartbeatIntervalSeconds int    `json:"heartbeatIntervalSeconds"`
}

type Lease struct {
	Token      string `json:"leaseToken"`
	Generation int64  `json:"leaseGeneration"`
}

type HeartbeatInput struct {
	DeviceName string       `json:"deviceName"`
	OS         string       `json:"os"`
	Version    string       `json:"version"`
	Accounts   AccountStats `json:"accounts"`
}

type AccountStats struct {
	Total             int `json:"total"`
	Available         int `json:"available"`
	Busy              int `json:"busy"`
	QuotaExhausted    int `json:"quotaExhausted"`
	LoginError        int `json:"loginError"`
	HumanVerification int `json:"humanVerification"`
}

type Job struct {
	ID           string         `json:"id"`
	SourceTaskID string         `json:"sourceTaskId,omitempty"`
	State        string         `json:"state,omitempty"`
	Payload      map[string]any `json:"payload,omitempty"`
}

type Claim struct {
	Job             Job    `json:"job"`
	LeaseToken      string `json:"leaseToken"`
	LeaseGeneration int64  `json:"leaseGeneration"`
}

func New(baseURL string, httpClient *http.Client) (*Client, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, errors.New("base URL is required")
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{BaseURL: baseURL, HTTPClient: httpClient}, nil
}

func (c *Client) Pair(ctx context.Context, input PairInput) (PairResult, error) {
	var out PairResult
	err := c.json(ctx, http.MethodPost, "/api/local-executor/v1/pair", "", input, &out)
	if err == nil && (out.ExecutorID == "" || out.Token == "") {
		err = errors.New("pair response is incomplete")
	}
	return out, err
}

func (c *Client) Heartbeat(ctx context.Context, token string, input HeartbeatInput) error {
	return c.json(ctx, http.MethodPost, "/api/local-executor/v1/heartbeat", token, input, nil)
}

func (c *Client) Claim(ctx context.Context, token string) (*Claim, error) {
	req, err := c.newRequest(ctx, http.MethodPost, "/api/local-executor/v1/jobs/claim", token, nil, "")
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		return nil, ErrNoJob
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, readAPIError(resp)
	}
	var out Claim
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.Job.ID == "" || out.LeaseToken == "" || out.LeaseGeneration < 1 {
		return nil, errors.New("claim response is incomplete")
	}
	return &out, nil
}

func (c *Client) Renew(ctx context.Context, token, jobID string, lease Lease) error {
	return c.leaseJSON(ctx, token, jobID, "renew", lease, nil)
}

func (c *Client) Progress(ctx context.Context, token, jobID string, lease Lease, state string) error {
	return c.leaseJSON(ctx, token, jobID, "progress", lease, map[string]any{"state": state})
}

func (c *Client) Acceptance(ctx context.Context, token, jobID string, lease Lease, accountID, submissionID string) error {
	return c.leaseJSON(ctx, token, jobID, "acceptance", lease, map[string]any{"accountId": accountID, "submissionId": submissionID})
}

func (c *Client) Release(ctx context.Context, token, jobID string, lease Lease, reason string) error {
	return c.leaseJSON(ctx, token, jobID, "release", lease, map[string]any{"reason": reason})
}

func (c *Client) Fail(ctx context.Context, token, jobID string, lease Lease, code, message string) error {
	return c.leaseJSON(ctx, token, jobID, "fail", lease, map[string]any{"code": code, "message": message})
}

func (c *Client) Result(ctx context.Context, token, jobID string, lease Lease, artifactID string) error {
	return c.leaseJSON(ctx, token, jobID, "result", lease, map[string]any{"artifactId": artifactID})
}

func (c *Client) UploadArtifact(ctx context.Context, token, jobID string, lease Lease, filePath string) (map[string]any, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	req, err := c.newRequest(ctx, http.MethodPost, "/api/local-executor/v1/jobs/"+jobID+"/artifact", token, file, "video/mp4")
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Lease-Token", lease.Token)
	req.Header.Set("X-Lease-Generation", fmt.Sprint(lease.Generation))
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, readAPIError(resp)
	}
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) leaseJSON(ctx context.Context, token, jobID, action string, lease Lease, extra map[string]any) error {
	body := map[string]any{"leaseToken": lease.Token, "leaseGeneration": lease.Generation}
	for key, value := range extra {
		body[key] = value
	}
	return c.json(ctx, http.MethodPost, "/api/local-executor/v1/jobs/"+jobID+"/"+action, token, body, nil)
}

func (c *Client) json(ctx context.Context, method, path, token string, value any, target any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	req, err := c.newRequest(ctx, method, path, token, bytes.NewReader(encoded), "application/json")
	if err != nil {
		return err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return readAPIError(resp)
	}
	if target == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func (c *Client) newRequest(ctx context.Context, method, path, token string, body io.Reader, contentType string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req, nil
}

func readAPIError(resp *http.Response) error {
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	var payload struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(data, &payload)
	if payload.Error != "" {
		return fmt.Errorf("local executor API: %s (HTTP %d)", payload.Error, resp.StatusCode)
	}
	return fmt.Errorf("local executor API: HTTP %d", resp.StatusCode)
}
