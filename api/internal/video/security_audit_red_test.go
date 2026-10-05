package video

import (
	"context"
	"database/sql/driver"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

type securityRoundTripper struct {
	calls int
}

func (t *securityRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	t.calls++
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("fake-media")),
	}, nil
}

func TestSecurityAuditMergeDownloaderRejectsPrivateAndLinkLocalURLsBeforeNetwork(t *testing.T) {
	for _, rawURL := range []string{
		"https://127.0.0.1/private.mp4",
		"https://169.254.169.254/latest/meta-data/",
		"https://10.1.2.3/internal.mp4",
	} {
		t.Run(rawURL, func(t *testing.T) {
			transport := &securityRoundTripper{}
			downloader := &HTTPMediaDownloader{Client: &http.Client{Transport: transport}}
			tmp, err := os.CreateTemp("", "security-merge-*")
			if err != nil {
				t.Fatal(err)
			}
			path := tmp.Name()
			_ = tmp.Close()
			defer os.Remove(path)

			err = downloader.Download(context.Background(), rawURL, path)
			if err == nil {
				t.Fatalf("private/link-local URL %q reached downloader", rawURL)
			}
			if transport.calls != 0 {
				t.Fatalf("network calls=%d, want 0 for rejected URL %q", transport.calls, rawURL)
			}
		})
	}
}

func TestSecurityAuditArtifactStoreRejectsPrivateAndLinkLocalURLsBeforeNetwork(t *testing.T) {
	for _, rawURL := range []string{
		"https://127.0.0.1/provider.mp4",
		"https://169.254.169.254/latest/meta-data/",
		"https://192.168.1.10/provider.mp4",
	} {
		t.Run(rawURL, func(t *testing.T) {
			transport := &securityRoundTripper{}
			store, err := NewArtifactStore(ArtifactStoreConfig{
				Bucket:        "video-bucket",
				PublicBaseURL: "https://cdn.example/video-bucket",
				HTTPClient:    &http.Client{Transport: transport},
				Uploader:      &recordingObjectUploader{},
			})
			if err != nil {
				t.Fatal(err)
			}
			_, err = store.Persist(context.Background(), rawURL, "security/out.mp4")
			if err == nil {
				t.Fatalf("private/link-local URL %q reached artifact store", rawURL)
			}
			if transport.calls != 0 {
				t.Fatalf("network calls=%d, want 0 for rejected URL %q", transport.calls, rawURL)
			}
		})
	}
}

func TestSecurityAuditAssignedLocalExecutorTaskCannotBeCompletedByAnotherExecutor(t *testing.T) {
	ctx := context.Background()
	store := newMemoryLocalExecutorStore()
	service := NewLocalExecutorService(store, time.Now)
	register := func(name string) LocalExecutorRegistrationResult {
		result, err := service.Register(ctx, LocalExecutorRegistrationInput{
			Name: name, ProviderKey: ProviderDoubaoLocalExecutor, Model: ModelDoubaoSeedance,
		})
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	owner := register("executor-a")
	attacker := register("executor-b")
	store.tasks["let_assigned"] = LocalExecutorTask{
		ID: "let_assigned", ProviderKey: ProviderDoubaoLocalExecutor, Model: ModelDoubaoSeedance,
		Status: TaskRunning, ExecutorID: owner.Executor.ID,
	}

	err := service.CompleteTask(ctx, attacker.Token, "let_assigned", LocalExecutorCompleteInput{ArtifactURL: "https://cdn.example/forged.mp4"})
	if !errors.Is(err, ErrLocalExecutorUnauthorized) {
		t.Fatalf("err=%v, want ErrLocalExecutorUnauthorized for cross-executor completion", err)
	}
	got, err := store.GetLocalExecutorTask(ctx, "let_assigned")
	if err != nil {
		t.Fatal(err)
	}
	if got.ExecutorID != owner.Executor.ID || got.Status != TaskRunning || got.ArtifactURL != "" {
		t.Fatalf("assigned task was modified by wrong executor: %+v", got)
	}
}

func TestSecurityAuditDuplicateIdenticalMySQLCompletionIsIdempotent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)

	mock.ExpectExec("UPDATE video_local_executor_tasks").
		WithArgs(TaskSucceeded, "lex_owner", "https://cdn.example/out.mp4", now, "let_1").
		WillReturnResult(sqlmock.NewResult(0, 0))
	rows := sqlmock.NewRows([]string{"id", "source_task_id", "provider_key", "model", "prompt", "request_id", "status", "executor_id", "artifact_url", "error_code", "error_message", "created_at", "updated_at"}).
		AddRow("let_1", "101", ProviderDoubaoLocalExecutor, ModelDoubaoSeedance, "prompt", "req", TaskSucceeded, "lex_owner", "https://cdn.example/out.mp4", "", "", now, now)
	mock.ExpectQuery("SELECT id, source_task_id").WithArgs("let_1").WillReturnRows(rows)

	err = store.CompleteLocalExecutorTask(context.Background(), "let_1", "lex_owner", "https://cdn.example/out.mp4", now)
	if err != nil {
		t.Fatalf("duplicate identical completion should be idempotent, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

var _ driver.Value = TaskSucceeded
