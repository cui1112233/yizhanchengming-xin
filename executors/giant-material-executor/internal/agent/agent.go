package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"qiantie/giant-material-executor/internal/worker"
)

var (
	ErrNoClaimableJob = errors.New("no claimable giant material job")
	ErrUnauthorized   = errors.New("giant material executor is unauthorized")
)

// progressReportInterval prevents a frame-level OCR stream from overwhelming
// the public control plane. The first update and terminal 100% update are
// always persisted so the UI remains responsive and accurate.
const progressReportInterval = time.Second

type Job struct {
	ID                string     `json:"id"`
	MaterialID        string     `json:"materialId"`
	PlatformBookID    string     `json:"platformBookId"`
	Title             string     `json:"title"`
	VideoURL          string     `json:"videoUrl"`
	VideoExpiresAt    *time.Time `json:"videoExpiresAt,omitempty"`
	DurationSeconds   float64    `json:"durationSeconds"`
	ModelVersion      string     `json:"modelVersion"`
	ContentRangeLines string     `json:"contentRangeLines,omitempty"`
}

type ClaimResult struct {
	Job             Job       `json:"job"`
	LeaseToken      string    `json:"leaseToken"`
	LeaseGeneration int64     `json:"leaseGeneration"`
	LeaseExpiresAt  time.Time `json:"leaseExpiresAt"`
}

type PairInput struct {
	Code       string `json:"code"`
	DeviceName string `json:"deviceName"`
	OS         string `json:"os"`
	Version    string `json:"version"`
	Platform   string `json:"platform"`
}

type PairResult struct {
	ExecutorID               string `json:"executorId"`
	Token                    string `json:"token"`
	HeartbeatIntervalSeconds int    `json:"heartbeatIntervalSeconds"`
}

type LeaseCredential struct {
	Token      string `json:"leaseToken"`
	Generation int64  `json:"leaseGeneration"`
}

type JobProgress struct {
	Completed int
	Total     int
	Percent   int
}

type Result struct {
	Text      string
	WordCount int
}

type PublicClient interface {
	Heartbeat(context.Context, string, HeartbeatRequest) error
	Claim(context.Context, string) (ClaimResult, error)
	Renew(context.Context, string, string, LeaseCredential) error
	Progress(context.Context, string, string, LeaseCredential, State, JobProgress) error
	Complete(context.Context, string, string, LeaseCredential, Result) error
	Fail(context.Context, string, string, LeaseCredential, string, string) error
}

type HeartbeatRequest struct {
	DeviceName string
	OS         string
	Version    string
}

type Config struct {
	Client         PublicClient
	Supervisor     *worker.Supervisor
	PrepareModel   func(context.Context, string) error
	OnBindingState func(BindingState)
	Token          string
	DeviceName     string
	OS             string
	Version        string
	PollInterval   time.Duration
	HeartbeatEvery time.Duration
}

type Agent struct {
	config Config
	state  *StateMachine
}

func New(config Config) (*Agent, error) {
	if config.Client == nil || config.Supervisor == nil || strings.TrimSpace(config.Token) == "" {
		return nil, errors.New("agent client, worker, and token are required")
	}
	if config.PollInterval <= 0 {
		config.PollInterval = 2 * time.Second
	}
	if config.HeartbeatEvery <= 0 {
		config.HeartbeatEvery = 15 * time.Second
	}
	return &Agent{config: config, state: NewStateMachine()}, nil
}

func (a *Agent) Snapshot() Snapshot { return a.state.Snapshot() }

func (a *Agent) Run(ctx context.Context) error {
	a.setBindingState(BindingConnecting)
	if err := a.config.Supervisor.Start(ctx); err != nil {
		a.setBindingState(BindingOffline)
		return err
	}
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = a.config.Supervisor.Stop(stopCtx)
	}()
	if a.config.PrepareModel != nil {
		if err := a.state.Transition(StateDownloadingModel); err != nil {
			return err
		}
		if err := a.config.PrepareModel(ctx, "windows-paddleocr-v1"); err != nil {
			_ = a.state.Transition(StateFailed)
			a.setBindingState(BindingOffline)
			return err
		}
	}
	if err := a.heartbeat(ctx); err != nil {
		a.setBindingState(bindingStateForError(err))
		return err
	}
	a.setBindingState(BindingOnline)
	if err := a.state.Transition(StateReady); err != nil {
		return err
	}
	poll := time.NewTicker(a.config.PollInterval)
	defer poll.Stop()
	heartbeat := time.NewTicker(a.config.HeartbeatEvery)
	defer heartbeat.Stop()
	for {
		select {
		case <-ctx.Done():
			_ = a.state.Transition(StateStopping)
			return ctx.Err()
		case <-heartbeat.C:
			if err := a.heartbeat(ctx); err != nil {
				a.setBindingState(bindingStateForError(err))
				return err
			}
		case <-poll.C:
			claim, err := a.config.Client.Claim(ctx, a.config.Token)
			if errors.Is(err, ErrNoClaimableJob) {
				continue
			}
			if err != nil {
				if errors.Is(err, ErrUnauthorized) {
					a.setBindingState(BindingNeedsPairing)
				} else {
					a.setBindingState(BindingOffline)
				}
				return err
			}
			if err := a.process(ctx, claim); err != nil {
				return err
			}
		}
	}
}

func (a *Agent) setBindingState(state BindingState) {
	a.state.SetBindingState(state)
	if a.config.OnBindingState != nil {
		a.config.OnBindingState(state)
	}
}

func bindingStateForError(err error) BindingState {
	if errors.Is(err, ErrUnauthorized) {
		return BindingNeedsPairing
	}
	return BindingOffline
}

func (a *Agent) heartbeat(ctx context.Context) error {
	return a.config.Client.Heartbeat(ctx, a.config.Token, HeartbeatRequest{DeviceName: a.config.DeviceName, OS: a.config.OS, Version: a.config.Version})
}

func (a *Agent) process(ctx context.Context, claim ClaimResult) error {
	a.state.SetJob(claim.Job.ID, claim.Job.ModelVersion)
	if err := a.state.Transition(StateRunning); err != nil {
		return err
	}
	request := worker.ExtractRequest{JobID: claim.Job.ID, VideoURL: claim.Job.VideoURL, DurationSeconds: claim.Job.DurationSeconds}
	if err := a.config.Supervisor.Submit(ctx, request); err != nil {
		_ = a.state.Transition(StateFailed)
		if failErr := a.config.Client.Fail(ctx, a.config.Token, claim.Job.ID, LeaseCredential{Token: claim.LeaseToken, Generation: claim.LeaseGeneration}, "WORKER_SUBMIT_FAILED", err.Error()); failErr != nil {
			return failErr
		}
		return nil
	}
	events := a.config.Supervisor.Events()
	renew := time.NewTicker(15 * time.Second)
	defer renew.Stop()
	completed := false
	var lastProgressReport time.Time
	for {
		select {
		case <-ctx.Done():
			_ = a.config.Supervisor.Cancel(context.Background(), claim.Job.ID)
			return ctx.Err()
		case <-renew.C:
			if err := a.config.Client.Renew(ctx, a.config.Token, claim.Job.ID, LeaseCredential{Token: claim.LeaseToken, Generation: claim.LeaseGeneration}); err != nil {
				return err
			}
		case event, ok := <-events:
			if !ok {
				return errors.New("OCR worker exited unexpectedly")
			}
			switch event.Type {
			case worker.EventProgress:
				progress := JobProgress{Completed: event.Completed, Total: event.Total, Percent: event.Percent}
				a.state.SetProgress(Progress{Completed: progress.Completed, Total: progress.Total, Percent: progress.Percent})
				now := time.Now()
				if !lastProgressReport.IsZero() && progress.Percent < 100 && now.Sub(lastProgressReport) < progressReportInterval {
					continue
				}
				lastProgressReport = now
				if err := a.config.Client.Progress(ctx, a.config.Token, claim.Job.ID, LeaseCredential{Token: claim.LeaseToken, Generation: claim.LeaseGeneration}, StateRunning, progress); err != nil {
					return err
				}
			case worker.EventComplete:
				if err := a.state.Transition(StateCleaning); err != nil {
					return err
				}
				progress := a.state.Snapshot().Progress
				jobProgress := JobProgress{Completed: progress.Completed, Total: progress.Total, Percent: progress.Percent}
				if err := a.config.Client.Progress(ctx, a.config.Token, claim.Job.ID, LeaseCredential{Token: claim.LeaseToken, Generation: claim.LeaseGeneration}, StateCleaning, jobProgress); err != nil {
					return err
				}
				if err := a.state.Transition(StateUploading); err != nil {
					return err
				}
				if err := a.config.Client.Progress(ctx, a.config.Token, claim.Job.ID, LeaseCredential{Token: claim.LeaseToken, Generation: claim.LeaseGeneration}, StateUploading, jobProgress); err != nil {
					return err
				}
				if err := a.config.Client.Complete(ctx, a.config.Token, claim.Job.ID, LeaseCredential{Token: claim.LeaseToken, Generation: claim.LeaseGeneration}, Result{Text: event.Text, WordCount: event.Characters}); err != nil {
					return err
				}
				completed = true
			case worker.EventFailed:
				_ = a.state.Transition(StateFailed)
				if err := a.config.Client.Fail(ctx, a.config.Token, claim.Job.ID, LeaseCredential{Token: claim.LeaseToken, Generation: claim.LeaseGeneration}, event.Code, event.Message); err != nil {
					return err
				}
				return nil
			case worker.EventIdle:
				if completed {
					a.state.ResetToIdle()
					return nil
				}
			}
		}
	}
}

type HTTPClient struct {
	BaseURL string
	Client  *http.Client
}

func NewHTTPClient(baseURL string, client *http.Client) *HTTPClient {
	if client == nil {
		client = http.DefaultClient
	}
	return &HTTPClient{BaseURL: strings.TrimRight(baseURL, "/"), Client: client}
}

func (c *HTTPClient) Heartbeat(ctx context.Context, token string, input HeartbeatRequest) error {
	var response struct{}
	return c.do(ctx, http.MethodPost, "/api/giant-material-executor/v1/heartbeat", token, map[string]string{"deviceName": input.DeviceName, "os": input.OS, "version": input.Version}, &response)
}

func (c *HTTPClient) Pair(ctx context.Context, input PairInput) (PairResult, error) {
	var result PairResult
	err := c.do(ctx, http.MethodPost, "/api/giant-material-executor/v1/pair", "", input, &result)
	return result, err
}

func (c *HTTPClient) Claim(ctx context.Context, token string) (ClaimResult, error) {
	var result ClaimResult
	err := c.do(ctx, http.MethodPost, "/api/giant-material-executor/v1/jobs/claim", token, nil, &result)
	return result, err
}

func (c *HTTPClient) Renew(ctx context.Context, token, jobID string, lease LeaseCredential) error {
	var result struct{}
	return c.do(ctx, http.MethodPost, "/api/giant-material-executor/v1/jobs/"+jobID+"/renew", token, lease, &result)
}

func (c *HTTPClient) Progress(ctx context.Context, token, jobID string, lease LeaseCredential, state State, progress JobProgress) error {
	var result struct{}
	body := map[string]any{"leaseToken": lease.Token, "leaseGeneration": lease.Generation, "state": string(state), "completed": progress.Completed, "total": progress.Total, "percent": progress.Percent}
	return c.do(ctx, http.MethodPost, "/api/giant-material-executor/v1/jobs/"+jobID+"/progress", token, body, &result)
}

func (c *HTTPClient) Complete(ctx context.Context, token, jobID string, lease LeaseCredential, result Result) error {
	var response struct{}
	body := map[string]any{"leaseToken": lease.Token, "leaseGeneration": lease.Generation, "text": result.Text, "wordCount": result.WordCount}
	return c.do(ctx, http.MethodPost, "/api/giant-material-executor/v1/jobs/"+jobID+"/result", token, body, &response)
}

func (c *HTTPClient) Fail(ctx context.Context, token, jobID string, lease LeaseCredential, code, message string) error {
	var response struct{}
	body := map[string]any{"leaseToken": lease.Token, "leaseGeneration": lease.Generation, "code": code, "message": message}
	return c.do(ctx, http.MethodPost, "/api/giant-material-executor/v1/jobs/"+jobID+"/fail", token, body, &response)
}

func (c *HTTPClient) do(ctx context.Context, method, path, token string, body any, output any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = strings.NewReader(string(data))
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reader)
	if err != nil {
		return err
	}
	if strings.TrimSpace(token) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(token))
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNoContent {
		return ErrNoClaimableJob
	}
	if response.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("%w: public executor API HTTP %d", ErrUnauthorized, response.StatusCode)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("public executor API HTTP %d", response.StatusCode)
	}
	if output == nil {
		return nil
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 3<<20)).Decode(output); err != nil {
		return err
	}
	return nil
}
