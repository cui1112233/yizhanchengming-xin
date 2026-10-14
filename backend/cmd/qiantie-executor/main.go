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
	go heartbeatLoop(ctx, client, cred.Token)
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
		claim, err := client.Claim(ctx, cred.Token)
		if errors.Is(err, localexecutorclient.ErrNoJob) {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(2 * time.Second):
			}
			continue
		}
		if err != nil {
			return err
		}
		if err := runJob(ctx, client, adapter, cred.Token, accountID, claim); err != nil {
			log.Printf("job %s failed: %v", claim.Job.ID, err)
		}
	}
}

func runJob(ctx context.Context, client *localexecutorclient.Client, adapter *doubaoexecutor.Adapter, token, accountID string, claim *localexecutorclient.Claim) error {
	lease := localexecutorclient.Lease{Token: claim.LeaseToken, Generation: claim.LeaseGeneration}
	job := doubaoexecutor.Job{ID: claim.Job.ID, Payload: claim.Job.Payload}
	if err := client.Progress(ctx, token, job.ID, lease, "preparing"); err != nil {
		return err
	}
	if err := adapter.Prepare(ctx, accountID, job); err != nil {
		_ = client.Release(ctx, token, job.ID, lease, err.Error())
		return err
	}
	if err := client.Progress(ctx, token, job.ID, lease, "submitting"); err != nil {
		return err
	}
	submission, err := adapter.Submit(ctx, accountID)
	if err != nil {
		_ = client.Release(ctx, token, job.ID, lease, err.Error())
		return err
	}
	if err := client.Acceptance(ctx, token, job.ID, lease, accountID, submission.ID); err != nil {
		return err
	}
	if err := client.Progress(ctx, token, job.ID, lease, "generating"); err != nil {
		return err
	}
	completion, err := adapter.Wait(ctx, accountID, submission)
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
	if err := adapter.Download(ctx, completion, tmpPath); err != nil {
		_ = client.Fail(ctx, token, job.ID, lease, "DOWNLOAD_FAILED", err.Error())
		return err
	}
	if err := client.Progress(ctx, token, job.ID, lease, "uploading"); err != nil {
		return err
	}
	uploaded, err := client.UploadArtifact(ctx, token, job.ID, lease, tmpPath)
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

func heartbeatLoop(ctx context.Context, client *localexecutorclient.Client, token string) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		stats := localexecutorclient.AccountStats{}
		if os.Getenv("QIANTIE_DOUBAO_ACCOUNT_READY") == "1" {
			stats = localexecutorclient.AccountStats{Total: 1, Available: 1}
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

func (s state) String() string { return fmt.Sprintf("%s/%s", s.BaseURL, s.ExecutorID) }
