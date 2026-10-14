// qiantie-executor is the Go replacement for the Node/Electron control loop.
// The browser profile is deliberately local to this process; the server only
// receives leases, progress, and the final MP4 artifact.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"qiantie/backend/internal/doubaoexecutor"
	"qiantie/backend/internal/localexecutorclient"
)

type state struct {
	BaseURL    string `json:"baseUrl"`
	ExecutorID string `json:"executorId"`
	Token      string `json:"token"`
}

type retryConfig struct {
	Attempts int
	Delay    time.Duration
}

type leaseRenewer interface {
	Renew(context.Context, string, string, localexecutorclient.Lease) error
}

type executorAdapter interface {
	Prepare(context.Context, string, doubaoexecutor.Job) error
	Submit(context.Context, string) (doubaoexecutor.Submission, error)
	Wait(context.Context, string, doubaoexecutor.Submission) (doubaoexecutor.Completion, error)
	Download(context.Context, doubaoexecutor.Completion, string) error
}

type accountPool struct {
	mu       sync.Mutex
	accounts map[string]bool
}

func newAccountPool(ids []string) *accountPool {
	pool := &accountPool{accounts: make(map[string]bool)}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" {
			pool.accounts[id] = false
		}
	}
	return pool
}

func (p *accountPool) Acquire() (string, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for id, busy := range p.accounts {
		if !busy {
			p.accounts[id] = true
			return id, true
		}
	}
	return "", false
}

func (p *accountPool) Release(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.accounts[id]; ok {
		p.accounts[id] = false
	}
}

func (p *accountPool) Stats() localexecutorclient.AccountStats {
	p.mu.Lock()
	defer p.mu.Unlock()
	stats := localexecutorclient.AccountStats{Total: len(p.accounts)}
	for _, busy := range p.accounts {
		if busy {
			stats.Busy++
		} else {
			stats.Available++
		}
	}
	return stats
}

func retryOperation[T any](ctx context.Context, cfg retryConfig, operation func(context.Context) (T, error)) (T, error) {
	var zero T
	if cfg.Attempts < 1 {
		cfg.Attempts = 1
	}
	var last error
	for attempt := 0; attempt < cfg.Attempts; attempt++ {
		value, err := operation(ctx)
		if err == nil {
			return value, nil
		}
		last = err
		if attempt+1 == cfg.Attempts {
			break
		}
		if cfg.Delay <= 0 {
			continue
		}
		timer := time.NewTimer(cfg.Delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return zero, ctx.Err()
		case <-timer.C:
		}
	}
	return zero, last
}

func startLeaseRenewer(ctx context.Context, client leaseRenewer, token, jobID string, lease localexecutorclient.Lease, interval time.Duration) func() {
	if interval <= 0 {
		interval = 15 * time.Second
	}
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-stop:
				return
			case <-ticker.C:
				if err := client.Renew(ctx, token, jobID, lease); err != nil {
					log.Printf("lease renewal failed for job %s: %v", jobID, err)
				}
			}
		}
	}()
	return func() {
		close(stop)
		<-done
	}
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("QIANTIE_EXECUTOR_BASE_URL")), "/")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:4001"
	}
	configDir, _ := os.UserConfigDir()
	if strings.TrimSpace(configDir) == "" {
		configDir = "."
	}
	statePath := env("QIANTIE_EXECUTOR_STATE_FILE", filepath.Join(configDir, "qiantie", "executor.json"))
	cred := loadState(statePath)
	if cred.BaseURL == "" {
		cred.BaseURL = baseURL
	}
	client, err := localexecutorclient.New(cred.BaseURL, &http.Client{Timeout: 5 * time.Minute})
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if cred.Token == "" {
		code := strings.TrimSpace(os.Getenv("QIANTIE_EXECUTOR_PAIR_CODE"))
		if code == "" {
			return errors.New("QIANTIE_EXECUTOR_PAIR_CODE is required for first pairing")
		}
		paired, err := client.Pair(ctx, localexecutorclient.PairInput{Code: code, DeviceName: env("QIANTIE_EXECUTOR_DEVICE_NAME", "Go 豆包执行器"), Platform: "doubao", OS: "windows", Version: env("QIANTIE_EXECUTOR_VERSION", "go-0.1.0")})
		if err != nil {
			return err
		}
		cred.ExecutorID, cred.Token = paired.ExecutorID, paired.Token
		if err := saveState(statePath, cred); err != nil {
			return err
		}
		log.Printf("paired executor %s", cred.ExecutorID)
	}
	adapter, err := doubaoexecutor.New(doubaoexecutor.Config{ChromePath: os.Getenv("QIANTIE_EXECUTOR_CHROME_PATH"), UserDataDir: env("QIANTIE_EXECUTOR_CHROME_DATA", filepath.Join(filepath.Dir(statePath), "chrome")), Headless: os.Getenv("QIANTIE_EXECUTOR_HEADLESS") == "1"})
	if err != nil {
		return err
	}
	accountID := env("QIANTIE_DOUBAO_ACCOUNT_ID", "default")
	accountReady := os.Getenv("QIANTIE_DOUBAO_ACCOUNT_READY") == "1"
	accountIDs := splitEnv("QIANTIE_DOUBAO_ACCOUNT_IDS")
	if len(accountIDs) == 0 {
		accountIDs = []string{accountID}
	}
	pool := newAccountPool(accountIDs)
	if accountReady {
		log.Printf("Go executor accounts ready: %s", strings.Join(accountIDs, ", "))
	}
	go heartbeatLoop(ctx, client, cred.Token, pool)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		if !accountReady {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(5 * time.Second):
			}
			continue
		}
		selectedAccount, ok := pool.Acquire()
		if !ok {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(2 * time.Second):
			}
			continue
		}
		claim, err := client.Claim(ctx, cred.Token)
		if errors.Is(err, localexecutorclient.ErrNoJob) {
			pool.Release(selectedAccount)
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(2 * time.Second):
			}
			continue
		}
		if err != nil {
			pool.Release(selectedAccount)
			return err
		}
		if err := runJob(ctx, client, adapter, cred.Token, selectedAccount, claim); err != nil {
			log.Printf("job %s failed: %v", claim.Job.ID, err)
		}
		pool.Release(selectedAccount)
	}
}

func runJob(ctx context.Context, client *localexecutorclient.Client, adapter executorAdapter, token, accountID string, claim *localexecutorclient.Claim) error {
	lease := localexecutorclient.Lease{Token: claim.LeaseToken, Generation: claim.LeaseGeneration}
	job := doubaoexecutor.Job{ID: claim.Job.ID, Payload: claim.Job.Payload}
	stopRenew := startLeaseRenewer(ctx, client, token, job.ID, lease, time.Duration(durationEnv("QIANTIE_EXECUTOR_LEASE_RENEW_MS", 15000))*time.Millisecond)
	defer stopRenew()
	retries := retryConfig{Attempts: intEnv("QIANTIE_EXECUTOR_RETRY_ATTEMPTS", 3), Delay: time.Duration(durationEnv("QIANTIE_EXECUTOR_RETRY_DELAY_MS", 1000)) * time.Millisecond}
	if err := client.Progress(ctx, token, job.ID, lease, "preparing"); err != nil {
		return err
	}
	if _, err := retryOperation(ctx, retries, func(opCtx context.Context) (struct{}, error) {
		return struct{}{}, adapter.Prepare(opCtx, accountID, job)
	}); err != nil {
		_ = client.Release(ctx, token, job.ID, lease, err.Error())
		return err
	}
	if err := client.Progress(ctx, token, job.ID, lease, "submitting"); err != nil {
		return err
	}
	submission, err := retryOperation(ctx, retries, func(opCtx context.Context) (doubaoexecutor.Submission, error) {
		return adapter.Submit(opCtx, accountID)
	})
	if err != nil {
		_ = client.Release(ctx, token, job.ID, lease, err.Error())
		return err
	}
	if _, err := retryOperation(ctx, retries, func(opCtx context.Context) (struct{}, error) {
		return struct{}{}, client.Acceptance(opCtx, token, job.ID, lease, accountID, submission.ID)
	}); err != nil {
		return err
	}
	if err := client.Progress(ctx, token, job.ID, lease, "generating"); err != nil {
		return err
	}
	completion, err := retryOperation(ctx, retries, func(opCtx context.Context) (doubaoexecutor.Completion, error) {
		return adapter.Wait(opCtx, accountID, submission)
	})
	if err != nil {
		_ = client.Fail(ctx, token, job.ID, lease, "RESULT_TIMEOUT", err.Error())
		return err
	}
	if err := client.Progress(ctx, token, job.ID, lease, "downloading"); err != nil {
		return err
	}
	tmp, err := os.CreateTemp("", "qiantie-executor-*.mp4")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpPath)
	if _, err := retryOperation(ctx, retries, func(opCtx context.Context) (struct{}, error) {
		return struct{}{}, adapter.Download(opCtx, completion, tmpPath)
	}); err != nil {
		_ = client.Fail(ctx, token, job.ID, lease, "DOWNLOAD_FAILED", err.Error())
		return err
	}
	if err := client.Progress(ctx, token, job.ID, lease, "uploading"); err != nil {
		return err
	}
	uploaded, err := retryOperation(ctx, retries, func(opCtx context.Context) (map[string]any, error) {
		return client.UploadArtifact(opCtx, token, job.ID, lease, tmpPath)
	})
	if err != nil {
		_ = client.Fail(ctx, token, job.ID, lease, "UPLOAD_FAILED", err.Error())
		return err
	}
	artifactID, _ := uploaded["id"].(string)
	if artifactID == "" {
		artifactID, _ = uploaded["artifactId"].(string)
	}
	if artifactID == "" {
		return errors.New("artifact response is incomplete")
	}
	return client.Result(ctx, token, job.ID, lease, artifactID)
}

func heartbeatLoop(ctx context.Context, client *localexecutorclient.Client, token string, pool *accountPool) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		stats := localexecutorclient.AccountStats{}
		if os.Getenv("QIANTIE_DOUBAO_ACCOUNT_READY") == "1" && pool != nil {
			stats = pool.Stats()
		}
		if err := client.Heartbeat(ctx, token, localexecutorclient.HeartbeatInput{DeviceName: env("QIANTIE_EXECUTOR_DEVICE_NAME", "Go 豆包执行器"), OS: "windows", Version: env("QIANTIE_EXECUTOR_VERSION", "go-0.1.0"), Accounts: stats}); err != nil {
			log.Printf("heartbeat failed: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func loadState(path string) state {
	raw, err := os.ReadFile(path)
	if err != nil {
		return state{}
	}
	var out state
	_ = json.Unmarshal(raw, &out)
	return out
}
func saveState(path string, value state) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	raw, _ := json.MarshalIndent(value, "", "  ")
	return os.WriteFile(path, raw, 0o600)
}
func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func durationEnv(key string, fallback int) int {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		var parsed int
		if _, err := fmt.Sscanf(value, "%d", &parsed); err == nil && parsed > 0 {
			return parsed
		}
	}
	return fallback
}

func intEnv(key string, fallback int) int { return durationEnv(key, fallback) }

func splitEnv(key string) []string {
	parts := strings.FieldsFunc(os.Getenv(key), func(r rune) bool { return r == ',' || r == ';' || r == '\n' })
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			out = append(out, value)
		}
	}
	return out
}

func (s state) String() string { return fmt.Sprintf("%s/%s", s.BaseURL, s.ExecutorID) }
