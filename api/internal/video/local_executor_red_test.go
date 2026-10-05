package video

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestLocalExecutorRegistrationHeartbeatIdentityAndCapabilities(t *testing.T) {
	now := time.Date(2026, 10, 5, 10, 40, 0, 0, time.UTC)
	store := newMemoryLocalExecutorStore()
	service := NewLocalExecutorService(store, func() time.Time { return now })

	registered, err := service.Register(context.Background(), LocalExecutorRegistrationInput{
		Name:         "doubao-mac-01",
		ProviderKey:  ProviderDoubaoLocalExecutor,
		Model:        "doubao-seedance",
		Capabilities: []string{"text_to_video", "reference_images"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if registered.Token == "" || registered.Executor.ID == "" {
		t.Fatalf("registration did not return one-time token/identity: %+v", registered)
	}
	if registered.Executor.ProviderKey != ProviderDoubaoLocalExecutor || registered.Executor.Model != "doubao-seedance" {
		t.Fatalf("executor provider/model = %+v", registered.Executor)
	}
	if !registered.Executor.Online || registered.Executor.LastSeenAt.IsZero() {
		t.Fatalf("new executor must be online with lastSeen: %+v", registered.Executor)
	}

	identity, err := service.Identity(context.Background(), registered.Token)
	if err != nil {
		t.Fatal(err)
	}
	if identity.ID != registered.Executor.ID || identity.TokenConfigured == false {
		t.Fatalf("identity = %+v", identity)
	}

	now = now.Add(20 * time.Second)
	if err := service.Heartbeat(context.Background(), registered.Token, LocalExecutorHeartbeatInput{
		Capabilities: []string{"text_to_video", "reference_images", "last_frame"},
	}); err != nil {
		t.Fatal(err)
	}
	identity, err = service.Identity(context.Background(), registered.Token)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(identity.Capabilities); got != 3 {
		t.Fatalf("capabilities len = %d want 3", got)
	}

	now = now.Add(LocalExecutorOnlineThreshold + time.Second)
	identity, err = service.Identity(context.Background(), registered.Token)
	if err != nil {
		t.Fatal(err)
	}
	if identity.Online {
		t.Fatal("stale heartbeat must report executor offline")
	}
}

func TestLocalExecutorRegistrationRejectsWrongProviderModel(t *testing.T) {
	service := NewLocalExecutorService(newMemoryLocalExecutorStore(), time.Now)
	_, err := service.Register(context.Background(), LocalExecutorRegistrationInput{
		Name: "wrong",
		ProviderKey: ProviderPersonalAPI,
		Model: "doubao-seedance",
	})
	if err == nil {
		t.Fatal("expected provider/model rejection")
	}
}

func TestDoubaoLocalProviderCreatesDurableTaskAndPollsCompleteOrFail(t *testing.T) {
	ctx := context.Background()
	store := newMemoryLocalExecutorStore()
	provider := NewDoubaoLocalExecutorProvider(store)

	submitted, err := provider.Submit(ctx, SubmitRequest{
		Model:        "doubao-seedance",
		Prompt:       "durable prompt",
		RequestID:    "req-local-1",
		SourceTaskID: "1024",
	})
	if err != nil {
		t.Fatal(err)
	}
	if submitted.ProviderJobID == "" || submitted.Status != TaskQueued {
		t.Fatalf("submit = %+v", submitted)
	}
	job, err := store.GetLocalExecutorTask(ctx, submitted.ProviderJobID)
	if err != nil {
		t.Fatal(err)
	}
	if job.SourceTaskID != "1024" || job.ProviderKey != ProviderDoubaoLocalExecutor || job.Model != "doubao-seedance" {
		t.Fatalf("durable job = %+v", job)
	}

	executorService := NewLocalExecutorService(store, time.Now)
	reg, err := executorService.Register(ctx, LocalExecutorRegistrationInput{
		Name: "doubao-win",
		ProviderKey: ProviderDoubaoLocalExecutor,
		Model: "doubao-seedance",
		Capabilities: []string{"text_to_video"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := executorService.CompleteTask(ctx, reg.Token, submitted.ProviderJobID, LocalExecutorCompleteInput{ArtifactURL: "https://cdn.example/local.mp4"}); err != nil {
		t.Fatal(err)
	}
	polled, err := provider.Poll(ctx, submitted.ProviderJobID)
	if err != nil {
		t.Fatal(err)
	}
	if polled.Status != TaskSucceeded || polled.ArtifactURL != "https://cdn.example/local.mp4" {
		t.Fatalf("completed poll = %+v", polled)
	}

	second, err := provider.Submit(ctx, SubmitRequest{Model: "doubao-seedance", Prompt: "fail me", SourceTaskID: "1025"})
	if err != nil {
		t.Fatal(err)
	}
	if err := executorService.FailTask(ctx, reg.Token, second.ProviderJobID, LocalExecutorFailInput{Code: "doubao_submit_failed", Message: "provider rejected"}); err != nil {
		t.Fatal(err)
	}
	polled, err = provider.Poll(ctx, second.ProviderJobID)
	if err != nil {
		t.Fatal(err)
	}
	if polled.Status != TaskFailed || polled.ErrorCode == "" || polled.ErrorMessage == "" {
		t.Fatalf("failed poll = %+v", polled)
	}
}

func TestLocalExecutorLeaseBoundaryHasNoRuntimeImplementation(t *testing.T) {
	var _ LocalExecutorLeaseCoordinator = (*recordingLeaseCoordinator)(nil)
}

type recordingLeaseCoordinator struct{}

func (*recordingLeaseCoordinator) Claim(context.Context, LocalExecutorIdentity) (LocalExecutorLease, error) {
	return LocalExecutorLease{}, errors.New("test")
}
func (*recordingLeaseCoordinator) Renew(context.Context, LocalExecutorLease) (LocalExecutorLease, error) {
	return LocalExecutorLease{}, errors.New("test")
}
func (*recordingLeaseCoordinator) Release(context.Context, LocalExecutorLease) error { return errors.New("test") }
func (*recordingLeaseCoordinator) RequeueExpired(context.Context, time.Time, int) (int, error) { return 0, errors.New("test") }
