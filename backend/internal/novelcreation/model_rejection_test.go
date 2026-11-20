package novelcreation

import (
	"context"
	"qiantie/backend/internal/batchfactoryv11"
	"testing"
)

func TestJournalExplicitModelUnavailable503IsRejectedWithoutResending(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		known      bool
	}{
		{"explicit", `{"error":{"code":"model_not_found","message":"当前模型暂不可用"}}`, true},
		{"generic503", `{"error":{"code":"internal_error","message":"upstream processing failed"}}`, false},
		{"unstructured", `gateway model_not_found`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newServiceFixture(t, nil)
			run := queuedWorkerRun(t, f)
			lease, _, err := f.service.claim(context.Background(), "alice", run.ID)
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.WithValue(context.Background(), journalKey{}, &callJournal{f.store, lease})
			p := &workerProvider{err: &batchfactoryv11.DirectorHTTPError{StatusCode: 503, Model: "configured-model", Body: tc.body}}
			request := batchfactoryv11.TextCompletionRequest{UserPrompt: "test"}
			_, err = journalComplete(ctx, "draft", p, request)
			state, _, _ := f.store.GetExecution(ctx, "alice", run.ID)
			if tc.known {
				if !IsCode(err, ErrModelUnavailable) || state.Calls["draft"].State != "rejected" {
					t.Fatalf("explicit upstream rejection misclassified: %#v %v", state.Calls["draft"], err)
				}
			} else if !IsCode(err, ErrResultUnknown) || state.Calls["draft"].State != "may_have_sent" {
				t.Fatal("uncertain 503 safety lost", err)
			}
			_, _ = journalComplete(ctx, "draft", p, request)
			if p.calls != 1 {
				t.Fatal("duplicate paid request", p.calls)
			}
		})
	}
}
