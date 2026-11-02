package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
)

func TestLocalExecutorArtifactUploadAcceptsVerifiedPNG(t *testing.T) {
	now := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	memory := localexecutor.NewMemoryStore()
	service := localexecutor.NewService(memory, func() time.Time { return now })
	pairing, err := service.CreatePairing(context.Background(), "alice", localexecutor.PlatformDoubao)
	if err != nil {
		t.Fatal(err)
	}
	paired, err := service.Pair(context.Background(), localexecutor.PairInput{Code: pairing.Code, Platform: localexecutor.PlatformDoubao, DeviceName: "test", OS: "windows", Version: "1.1.0"})
	if err != nil {
		t.Fatal(err)
	}
	job, err := service.CreateJob(context.Background(), "alice", localexecutor.CreateJobInput{SourceTaskID: "settings-test:image", Platform: localexecutor.PlatformDoubao, Payload: map[string]any{"type": "image", "prompt": "bird"}})
	if err != nil {
		t.Fatal(err)
	}
	claim, err := service.ClaimJob(context.Background(), paired.Token)
	if err != nil {
		t.Fatal(err)
	}
	lease := localexecutor.LeaseCredential{Token: claim.LeaseToken, Generation: claim.LeaseGeneration}
	for _, state := range []localexecutor.JobState{localexecutor.JobPreparing, localexecutor.JobSubmitting} {
		if err := service.RecordProgress(context.Background(), paired.Token, job.ID, lease, state); err != nil {
			t.Fatal(err)
		}
	}
	if err := service.RecordAcceptance(context.Background(), paired.Token, job.ID, lease, localexecutor.AcceptanceInput{AccountID: "account", SubmissionID: "relay-task"}); err != nil {
		t.Fatal(err)
	}
	for _, state := range []localexecutor.JobState{localexecutor.JobGenerating, localexecutor.JobDownloading} {
		if err := service.RecordProgress(context.Background(), paired.Token, job.ID, lease, state); err != nil {
			t.Fatal(err)
		}
	}

	mux := http.NewServeMux()
	artifactDir := t.TempDir()
	files := localartifact.NewStore(artifactDir, 1<<20)
	RegisterLocalExecutorArtifactRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, service, files)
	png := append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte{0}, 16)...)
	req := httptest.NewRequest(http.MethodPost, "/api/local-executor/v1/jobs/"+job.ID+"/artifact", bytes.NewReader(png))
	req.Header.Set("Authorization", "Bearer "+paired.Token)
	req.Header.Set("Content-Type", "image/png")
	req.Header.Set("X-Lease-Token", claim.LeaseToken)
	req.Header.Set("X-Lease-Generation", fmt.Sprint(claim.LeaseGeneration))
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, req)
	if response.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var artifact localexecutor.ArtifactView
	if err := json.Unmarshal(response.Body.Bytes(), &artifact); err != nil {
		t.Fatal(err)
	}
	stored, globErr := filepath.Glob(filepath.Join(artifactDir, "*.png"))
	if globErr != nil || len(stored) != 1 {
		t.Fatalf("stored=%v err=%v entries=%v", stored, globErr, mustReadDir(t, artifactDir))
	}
	if artifact.MediaType != "image/png" || !strings.HasSuffix(stored[0], ".png") {
		t.Fatalf("artifact=%+v", artifact)
	}
}

func mustReadDir(t *testing.T, dir string) []os.DirEntry {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	return entries
}

func TestArtifactResponseMetadataUsesStoredMediaType(t *testing.T) {
	for _, test := range []struct{ mediaType, filename string }{
		{"video/mp4", "artifact.mp4"}, {"image/png", "artifact.png"},
		{"image/jpeg", "artifact.jpg"}, {"image/webp", "artifact.webp"},
	} {
		gotType, gotName, err := artifactResponseMetadata("artifact", test.mediaType)
		if err != nil {
			t.Fatalf("%s: %v", test.mediaType, err)
		}
		if gotType != test.mediaType || gotName != test.filename {
			t.Fatalf("got %q %q", gotType, gotName)
		}
	}
}
