package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
)

func TestAgentImageGeneratorUsesOnlineLocalDoubaoExecutor(t *testing.T) {
	now := time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC)
	executorStore := localexecutor.NewMemoryStore()
	executors := localexecutor.NewService(executorStore, func() time.Time { return now })
	pairing, err := executors.CreatePairing(context.Background(), "alice", localexecutor.PlatformDoubao)
	if err != nil {
		t.Fatal(err)
	}
	paired, err := executors.Pair(context.Background(), localexecutor.PairInput{Code: pairing.Code, Platform: localexecutor.PlatformDoubao, DeviceName: "test", OS: "windows", Version: "1.2.0"})
	if err != nil {
		t.Fatal(err)
	}
	if err := executors.Heartbeat(context.Background(), paired.Token, localexecutor.HeartbeatInput{DeviceName: "test", OS: "windows", Version: "1.2.0", Accounts: localexecutor.AccountStats{Total: 1, Available: 1}}); err != nil {
		t.Fatal(err)
	}
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	assets := agentworkspace.NewMemoryAssetStore()
	reference := agentworkspace.Asset{ID: "reference-1", ProjectID: "agent-1", Kind: "image", Name: "reference.png", ContentType: "image/png", ObjectKey: "agent-assets/alice/agent-1/reference-1.png", ByteSize: 4, CreatedAt: now}
	if err := assets.CreateAsset(context.Background(), "alice", reference); err != nil {
		t.Fatal(err)
	}
	pngBytes, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII=")
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		claim, claimErr := executors.ClaimJobWait(context.Background(), paired.Token, 2*time.Second)
		if claimErr != nil {
			done <- claimErr
			return
		}
		var payload map[string]any
		if decodeErr := json.Unmarshal(claim.Job.Payload, &payload); decodeErr != nil {
			done <- decodeErr
			return
		}
		lease := localexecutor.LeaseCredential{Token: claim.LeaseToken, Generation: claim.LeaseGeneration}
		urls, _ := payload["referenceImages"].([]any)
		if len(urls) != 1 || urls[0] != agentExecutorReferenceAssetURL("agent-1", "reference-1") {
			_ = executors.FailJob(context.Background(), paired.Token, claim.Job.ID, lease, localexecutor.FailureInput{Code: "assertion", Message: "referenceImages missing"})
			done <- fmt.Errorf("referenceImages=%v", payload["referenceImages"])
			return
		}
		for _, state := range []localexecutor.JobState{localexecutor.JobPreparing, localexecutor.JobSubmitting} {
			if progressErr := executors.RecordProgress(context.Background(), paired.Token, claim.Job.ID, lease, state); progressErr != nil {
				done <- progressErr
				return
			}
		}
		if acceptErr := executors.RecordAcceptance(context.Background(), paired.Token, claim.Job.ID, lease, localexecutor.AcceptanceInput{AccountID: "account-1", SubmissionID: "submission-1"}); acceptErr != nil {
			done <- acceptErr
			return
		}
		for _, state := range []localexecutor.JobState{localexecutor.JobGenerating, localexecutor.JobDownloading, localexecutor.JobUploading} {
			if progressErr := executors.RecordProgress(context.Background(), paired.Token, claim.Job.ID, lease, state); progressErr != nil {
				done <- progressErr
				return
			}
		}
		saved, saveErr := files.SaveImage("lea_agent_test", "image/png", bytes.NewReader(pngBytes))
		if saveErr != nil {
			done <- saveErr
			return
		}
		digest := sha256.Sum256(pngBytes)
		artifact, artifactErr := executors.RecordArtifact(context.Background(), paired.Token, claim.Job.ID, lease, localexecutor.ArtifactInput{ID: saved.ID, MediaType: saved.MediaType, ByteSize: saved.ByteSize, SHA256: hex.EncodeToString(digest[:]), StorageRef: saved.StorageRef})
		if artifactErr != nil {
			done <- artifactErr
			return
		}
		done <- executors.CompleteJob(context.Background(), paired.Token, claim.Job.ID, lease, localexecutor.ResultInput{ArtifactID: artifact.ID})
	}()

	generator := newAgentImageGenerator(nil, nil, nil, executors, files, assets, nil)
	results, status, err := generator(context.Background(), "alice", AgentImageGenerationInput{ModelID: agentLocalDoubaoImageModelID, Prompt: "一只蓝色小鸟", Size: "1024x1536", AspectRatio: "9:16", Quality: "standard", Resolution: "1K", Count: 1, ProjectID: "agent-1", ReferenceAssetIDs: []string{"reference-1"}})
	if err != nil || status != 200 || len(results) != 1 || !strings.HasPrefix(results[0], "data:image/png;base64,") {
		t.Fatalf("status=%d results=%q err=%v", status, results, err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
