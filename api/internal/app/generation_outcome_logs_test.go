package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/httpapi"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/observability"
)

type outcomeLogService struct {
	httpapi.GenerationService
	cause error
}

func (s outcomeLogService) RunBook(context.Context, generation.RunBookRequest) (generation.BookGenerationResult, error) {
	return generation.BookGenerationResult{}, s.cause
}
func (s outcomeLogService) RetryStage(context.Context, generation.RetryStageRequest) (generation.BookGenerationResult, error) {
	return generation.BookGenerationResult{}, s.cause
}
func (s outcomeLogService) RunBatch(context.Context, generation.RunBatchRequest) (generation.BatchGenerationResult, error) {
	return generation.BatchGenerationResult{}, s.cause
}

func TestGenerationOutcomeLogsObservedCauses(t *testing.T) {
	const diagnostic = "book 11: provider-canary-short\npassword=pass-canary\nCookie: session=cookie-canary\nAuthorization: Bearer bearer-canary\nprovider_api_key=key-canary\nuser:dsn-canary@tcp(localhost:3306)/db"
	for _, operation := range []string{"run_book", "retry_stage", "run_batch"} {
		t.Run(operation, func(t *testing.T) {
			var logs bytes.Buffer
			cause := errors.New(diagnostic)
			s := observedGenerationService{next: outcomeLogService{cause: cause}, logger: observability.NewJSONLogger(&logs)}
			ctx := observability.WithRequestID(context.Background(), "http-current")
			var err error
			switch operation {
			case "run_book":
				_, err = s.RunBook(ctx, generation.RunBookRequest{BatchProjectID: 3, BookID: 11, RequestID: "execution-old"})
			case "retry_stage":
				_, err = s.RetryStage(ctx, generation.RetryStageRequest{BatchProjectID: 3, BookID: 11, Stage: generation.StageDirector, RequestID: "execution-old"})
			case "run_batch":
				_, err = s.RunBatch(ctx, generation.RunBatchRequest{BatchProjectID: 3, RequestID: "execution-old"})
			}
			if err != cause {
				t.Fatal("original cause lost")
			}
			found := false
			for _, line := range strings.Split(strings.TrimSpace(logs.String()), "\n") {
				var entry map[string]any
				if json.Unmarshal([]byte(line), &entry) != nil {
					t.Fatal("invalid JSON log")
				}
				if entry["request_id"] != "http-current" || entry["generation_request_id"] != "execution-old" || entry["batch_project_id"] != float64(3) {
					t.Fatal(entry)
				}
				if entry["operation"] == operation {
					found = true
					safe, _ := entry["safe_error"].(string)
					if !strings.Contains(safe, "provider-canary-short") || entry["subsystem"] != "generation" {
						t.Fatal(entry)
					}
					if operation != "run_batch" && entry["book_id"] != float64(11) {
						t.Fatal(entry)
					}
				}
			}
			if !found {
				t.Fatal("missing failure log")
			}
			for _, secret := range []string{"pass-canary", "cookie-canary", "bearer-canary", "key-canary", "dsn-canary"} {
				if strings.Contains(logs.String(), secret) {
					t.Fatalf("secret leaked: %s", secret)
				}
			}
		})
	}
}
