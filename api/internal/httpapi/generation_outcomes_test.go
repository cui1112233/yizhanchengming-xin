package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/observability"
)

const generationDiagnostic = "provider-canary-short\npassword=pass-canary\nCookie: session=cookie-canary\nAuthorization: Bearer bearer-canary\nprovider_api_key=key-canary\nuser:dsn-canary@tcp(localhost:3306)/db"
const generationSafeMessage = "生成阶段执行失败，请稍后重试"

type outcomeGenerationFake struct {
	fakeGenerationService
	err     error
	readErr error
	batch   generation.BatchGenerationResult
}

func (f *outcomeGenerationFake) RunBook(ctx context.Context, r generation.RunBookRequest) (generation.BookGenerationResult, error) {
	f.runBookReq = r
	return f.book, f.err
}
func (f *outcomeGenerationFake) RetryStage(ctx context.Context, r generation.RetryStageRequest) (generation.BookGenerationResult, error) {
	f.retryReq = r
	return f.book, f.err
}
func (f *outcomeGenerationFake) RunBatch(ctx context.Context, r generation.RunBatchRequest) (generation.BatchGenerationResult, error) {
	f.runBatchReq = r
	return f.batch, f.err
}
func (f *outcomeGenerationFake) ProjectSummary(context.Context, int64) (generation.ProjectSummary, error) {
	return f.project, f.readErr
}
func (f *outcomeGenerationFake) BookSummary(context.Context, int64, int64) (generation.BookGenerationResult, error) {
	return f.book, f.readErr
}
func (f *outcomeGenerationFake) StageResult(context.Context, int64, int64, generation.Stage) (generation.StageRun, error) {
	return f.stage, f.readErr
}

type outcomeStoryboardFake struct {
	ScriptStoryboardService
	book generation.BookGenerationResult
	err  error
}

func (f outcomeStoryboardFake) RecompileStoryboard(context.Context, int64, int64, string) (generation.BookGenerationResult, error) {
	return f.book, f.err
}

func legacyOutcomeFixture(diagnostic string) *outcomeGenerationFake {
	snapshot, _ := json.Marshal(map[string]any{"source": "source-business-text", "script": "script-business-text", "prompt": `{"error":"literal story text"}`, "matchAudio": true, "audioDurationSec": json.Number("28.250"), "large": json.Number("900719925474099312345"), "nested": []any{map[string]any{"error": map[string]any{"detail": diagnostic}, "code": diagnostic}, map[string]any{"errorMessage": []string{diagnostic}, "errorCode": diagnostic}}})
	validation, _ := json.Marshal(map[string]any{"valid": false, "repaired": true, "durationMs": 28250, "error": diagnostic, "code": diagnostic})
	now := time.Date(2026, 10, 9, 8, 0, 0, 0, time.UTC)
	run := generation.BookRun{ID: 17, BatchProjectID: 3, BookID: 11, Status: generation.StatusFailed, RequestID: "execution-old", ErrorMessage: diagnostic, StartedAt: &now}
	stage := generation.StageRun{ID: 29, BookRunID: 17, BookID: 11, Stage: generation.StageDirector, Status: generation.StatusFailed, Attempt: 4, RequestID: "execution-old", PromptKey: generation.PromptDirector, PromptVersion: 7, OutputText: "output-business-text", InputSnapshot: string(snapshot), ValidationResult: string(validation), ErrorMessage: diagnostic, StartedAt: &now}
	return &outcomeGenerationFake{fakeGenerationService: fakeGenerationService{
		book:    generation.BookGenerationResult{Run: run, Stages: []generation.StageRun{stage}, Latest: map[generation.Stage]generation.StageRun{generation.StageDirector: stage}, Error: diagnostic},
		stage:   stage,
		project: generation.ProjectSummary{BatchProjectID: 3, Failed: 1, Pending: 1, Books: []generation.BookGenerationSummary{{BookID: 11, Run: &run, Stages: map[generation.Stage]generation.StageRun{generation.StageDirector: stage}}, {BookID: 12, Stages: map[generation.Stage]generation.StageRun{}}}},
	}, batch: generation.BatchGenerationResult{BatchProjectID: 3, Failed: 1, Completed: 1, Books: []generation.BatchBookResult{{BookID: 11, Run: &run, Error: diagnostic}, {BookID: 12, Run: &generation.BookRun{ID: 18, BookID: 12, Status: generation.StatusCompleted}}}}}
}

func requestOutcome(t *testing.T, f *outcomeGenerationFake, method, suffix string, logs *bytes.Buffer) *httptest.ResponseRecorder {
	t.Helper()
	h := NewHandler(Dependencies{Generation: f, ScriptStoryboards: outcomeStoryboardFake{book: f.book, err: f.err}, Logger: observability.NewJSONLogger(logs)})
	r := httptest.NewRequest(method, "/api/v1/batch-projects/3"+suffix, strings.NewReader(`{"requestId":"execution-new"}`))
	r.Header.Set(observability.RequestIDHeader, "http-current")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func assertSafeOutcomeBody(t *testing.T, raw string) {
	t.Helper()
	for _, canary := range []string{"provider-canary-short", "pass-canary", "cookie-canary", "bearer-canary", "key-canary", "dsn-canary"} {
		if strings.Contains(raw, canary) {
			t.Errorf("public diagnostic leaked %s", canary)
		}
	}
}

func TestGenerationOutcomeRoutesCanaryMatrix(t *testing.T) {
	for _, diagnostic := range append(strings.Split(generationDiagnostic, "\n"), generationDiagnostic) {
		for _, route := range []struct {
			method, suffix string
			status         int
		}{
			{"POST", "/books/11/generation", 500}, {"POST", "/books/11/generation/stages/DIRECTOR/retry", 500}, {"POST", "/generation", 207}, {"POST", "/books/11/storyboard/recompile", 500},
			{"GET", "/generation", 200}, {"GET", "/books/11/generation", 200}, {"GET", "/books/11/generation/stages/DIRECTOR", 200},
		} {
			t.Run(route.method+route.suffix+diagnostic, func(t *testing.T) {
				f := legacyOutcomeFixture(diagnostic)
				f.err = errors.New(diagnostic)
				originalBook, originalStage, originalProject, originalBatch := f.book, f.stage, f.project, f.batch
				// Snapshot the public structure independently; maps/slices in the originals may alias.
				originalJSON, _ := json.Marshal([]any{originalBook, originalStage, originalProject, originalBatch})
				var logs bytes.Buffer
				w := requestOutcome(t, f, route.method, route.suffix, &logs)
				if w.Code != route.status {
					t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
				}
				assertSafeOutcomeBody(t, w.Body.String())
				var body map[string]any
				if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if w.Header().Get(observability.RequestIDHeader) != "http-current" {
					t.Fatal("response correlation lost")
				}
				if route.method == "POST" {
					if body["code"] != "GENERATION_FAILED" || body["message"] != generationSafeMessage || body["error"] != generationSafeMessage {
						t.Errorf("missing safe failure envelope: %s", w.Body.String())
					}
					if route.status != 207 && body["request_id"] != "http-current" {
						t.Fatal("business requestId replaced current request_id")
					}
				}
				if route.suffix == "/generation" {
					if body["failed"] != float64(1) {
						t.Fatal("failure count changed")
					}
				} else {
					if !strings.Contains(w.Body.String(), "output-business-text") || !strings.Contains(w.Body.String(), `"attempt":4`) || !strings.Contains(w.Body.String(), "execution-old") {
						t.Fatal("execution facts changed")
					}
				}
				afterJSON, _ := json.Marshal([]any{f.book, f.stage, f.project, f.batch})
				if !bytes.Equal(originalJSON, afterJSON) {
					t.Fatal("HTTP projection mutated service-owned result")
				}
				if strings.Contains(route.suffix, "/retry") && f.retryReq.Stage != generation.StageDirector {
					t.Fatal("retry target changed")
				}
			})
		}
	}
}

func TestGenerationOutcomeRoutesBeforeRunAndStatus(t *testing.T) {
	for _, tc := range []struct {
		cause         error
		status        int
		code, message string
	}{
		{generation.ErrInvalid, 400, "GENERATION_INVALID", "生成参数无效，请检查后重试"}, {generation.ErrNotFound, 404, "GENERATION_NOT_FOUND", "生成记录不存在，请刷新后重试"},
		{generation.ErrConflict, 409, "GENERATION_CONFLICT", "当前生成状态不允许此操作，请刷新后重试"}, {generation.ErrAudioMeasurementRequired, 422, "AUDIO_MEASUREMENT_REQUIRED", "请先生成或检测音频"},
		{generation.ErrAudioProbeUnavailable, 503, "AUDIO_PROBE_UNAVAILABLE", "音频检测服务暂不可用，请稍后重试"}, {generation.ErrUnavailable, 503, "GENERATION_UNAVAILABLE", "生成服务暂不可用，请稍后重试"},
		{errors.New(generationDiagnostic), 500, "GENERATION_FAILED", generationSafeMessage},
	} {
		for _, suffix := range []string{"/books/11/generation", "/books/11/generation/stages/DIRECTOR/retry", "/books/11/storyboard/recompile"} {
			t.Run(tc.code+suffix, func(t *testing.T) {
				f := &outcomeGenerationFake{err: fmt.Errorf("%w: provider-canary-short", tc.cause)}
				var logs bytes.Buffer
				w := requestOutcome(t, f, "POST", suffix, &logs)
				var body map[string]any
				_ = json.Unmarshal(w.Body.Bytes(), &body)
				if w.Code != tc.status || body["code"] != tc.code || body["message"] != tc.message || body["error"] != tc.message || body["stages"] != nil {
					t.Fatalf("zero-valued failure = %d %s", w.Code, w.Body.String())
				}
				assertSafeOutcomeBody(t, w.Body.String())
			})
		}
	}
}

func TestGenerationLegacyProjectionSnapshotsAndShapes(t *testing.T) {
	for _, snapshot := range []string{`{"error":"provider-canary-short","code":"untrusted","large":900719925474099312345,"source":"provider-canary-short is story text","prompt":"{\"error\":\"literal\"}","nested":[{"errorMessage":["pass-canary"],"errorCode":"untrusted"}],"codeOnly":{"code":47}}`, `[{"error":{"detail":"pass-canary"},"valid":false}]`, `{"error":`, `"pass-canary"`, `null`, `1`, `{} {}`, ""} {
		t.Run(snapshot, func(t *testing.T) {
			f := legacyOutcomeFixture("")
			f.stage.Status = generation.StatusCompleted
			f.stage.ErrorMessage = "provider-canary-short"
			f.stage.InputSnapshot = snapshot
			f.stage.ValidationResult = snapshot
			original := f.stage
			var logs bytes.Buffer
			w := requestOutcome(t, f, "GET", "/books/11/generation/stages/DIRECTOR", &logs)
			var got generation.StageRun
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(f.stage, original) || got.Status != generation.StatusCompleted || got.OutputText != original.OutputText || got.ErrorMessage != generationSafeMessage {
				t.Fatalf("projection changed facts or left stale diagnostic: %#v", got)
			}
			if strings.HasPrefix(snapshot, `{"error":"`) {
				for _, expected := range []string{"900719925474099312345", "provider-canary-short is story text", `"code":47`, "GENERATION_FAILED"} {
					if !strings.Contains(got.InputSnapshot, expected) {
						t.Errorf("lost %q: %s", expected, got.InputSnapshot)
					}
				}
			} else if strings.HasPrefix(snapshot, "[") {
				if strings.Contains(got.InputSnapshot, "pass-canary") || !strings.Contains(got.InputSnapshot, `"valid":false`) {
					t.Fatal(got.InputSnapshot)
				}
			} else if got.InputSnapshot != "" || got.ValidationResult != "" {
				t.Fatal("malformed/scalar snapshots must be omitted")
			}
		})
	}
	for _, book := range []generation.BookGenerationResult{{}, {Stages: []generation.StageRun{}, Latest: map[generation.Stage]generation.StageRun{}}, {Run: generation.BookRun{Status: generation.StatusPending}, Stages: []generation.StageRun{{Status: generation.StatusSkipped}}}} {
		f := &outcomeGenerationFake{fakeGenerationService: fakeGenerationService{book: book}}
		var logs bytes.Buffer
		w := requestOutcome(t, f, "GET", "/books/11/generation", &logs)
		var got generation.BookGenerationResult
		_ = json.Unmarshal(w.Body.Bytes(), &got)
		// Empty latest is omitted by the existing JSON contract; inspect the original for nil/empty preservation separately.
		if !reflect.DeepEqual(got.Stages, book.Stages) || got.Run != book.Run || !reflect.DeepEqual(f.book, book) {
			t.Fatalf("empty/pending shape changed: %#v", got)
		}
	}
}

func TestGenerationOutcomeLogsRecompileAndReads(t *testing.T) {
	for _, suffix := range []string{"/books/11/storyboard/recompile", "/generation", "/books/11/generation", "/books/11/generation/stages/DIRECTOR"} {
		t.Run(suffix, func(t *testing.T) {
			f := legacyOutcomeFixture(generationDiagnostic)
			f.err = errors.New(generationDiagnostic)
			f.readErr = f.err
			method := "GET"
			operation := ""
			if strings.Contains(suffix, "recompile") {
				method = "POST"
				operation = "recompile_storyboard"
			}
			var logs bytes.Buffer
			w := requestOutcome(t, f, method, suffix, &logs)
			assertSafeOutcomeBody(t, w.Body.String())
			if w.Code != 500 {
				t.Fatal(w.Code)
			}
			found := false
			for _, line := range strings.Split(strings.TrimSpace(logs.String()), "\n") {
				var entry map[string]any
				if json.Unmarshal([]byte(line), &entry) != nil {
					t.Fatalf("invalid log %s", line)
				}
				if entry["safe_error"] != nil {
					found = true
					if entry["request_id"] != "http-current" || entry["subsystem"] != "generation" || (operation != "" && entry["operation"] != operation) {
						t.Fatal(entry)
					}
					if !strings.Contains(fmt.Sprint(entry["safe_error"]), "provider-canary-short") {
						t.Fatal("internal diagnostic lost")
					}
					if method == "POST" && (entry["book_id"] != float64(11) || entry["batch_project_id"] != float64(3) || entry["error_code"] != "GENERATION_FAILED") {
						t.Fatal(entry)
					}
				}
			}
			if !found {
				t.Fatal("missing original cause log")
			}
			for _, secret := range []string{"pass-canary", "cookie-canary", "bearer-canary", "key-canary", "dsn-canary"} {
				if strings.Contains(logs.String(), secret) {
					t.Errorf("secret leaked to log: %s", secret)
				}
			}
		})
	}
}

func TestGenerationOutcomeReadbackMySQLSelectOnly(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 10, 9, 8, 0, 0, 0, time.UTC)
	f := legacyOutcomeFixture(generationDiagnostic)
	mock.ExpectQuery("SELECT .* FROM book_runs WHERE batch_project_id=").WithArgs(int64(3), int64(11)).WillReturnRows(sqlmock.NewRows([]string{"id", "run_id", "batch_project_id", "book_id", "status", "request_id", "error_message", "started_at", "finished_at", "created_at", "updated_at"}).AddRow(17, nil, 3, 11, "failed", "execution-old", generationDiagnostic, now, now, now, now))
	mock.ExpectQuery("SELECT .* FROM stage_runs WHERE book_run_id=").WithArgs(int64(17)).WillReturnRows(sqlmock.NewRows([]string{"id", "book_run_id", "book_id", "stage", "status", "attempt", "request_id", "prompt_key", "prompt_version", "input_snapshot", "output_text", "error_message", "validation_result", "started_at", "finished_at", "created_at", "updated_at"}).AddRow(29, 17, 11, "DIRECTOR", "failed", 4, "execution-old", "director.default", 7, f.stage.InputSnapshot, f.stage.OutputText, generationDiagnostic, f.stage.ValidationResult, now, now, now, now))
	store := generation.NewMySQLStore(db)
	run, err := store.LatestBookRun(context.Background(), 3, 11)
	if err != nil {
		t.Fatal(err)
	}
	stages, err := store.ListStageRuns(context.Background(), 17)
	if err != nil {
		t.Fatal(err)
	}
	f.book = generation.BookGenerationResult{Run: run, Stages: stages, Latest: map[generation.Stage]generation.StageRun{generation.StageDirector: stages[0]}}
	f.stage = stages[0]
	f.project.Books[0].Run = &run
	f.project.Books[0].Stages = f.book.Latest
	for _, suffix := range []string{"/generation", "/books/11/generation", "/books/11/generation/stages/DIRECTOR"} {
		var logs bytes.Buffer
		w := requestOutcome(t, f, "GET", suffix, &logs)
		if w.Code != 200 {
			t.Fatal(w.Code)
		}
		assertSafeOutcomeBody(t, w.Body.String())
	}
	if run.ErrorMessage != generationDiagnostic || stages[0].ErrorMessage != generationDiagnostic || stages[0].InputSnapshot != f.stage.InputSnapshot {
		t.Fatal("historical SQL facts changed")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGenerationLegacyProjectionPureCopiesAndCatalogue(t *testing.T) {
	f := legacyOutcomeFixture(generationDiagnostic)
	f.book.Run.ErrorCode = "untrusted-code"
	f.book.ErrorCode = "untrusted-code"
	f.stage.ErrorCode = "untrusted-code"
	f.book.Stages[0] = f.stage
	f.book.Latest[generation.StageDirector] = f.stage
	before, _ := json.Marshal([]any{f.book, f.stage, f.batch, f.project})
	book, batch, project := projectBookGeneration(f.book), projectBatchGeneration(f.batch), projectGenerationSummary(f.project)
	if book.Run.ErrorCode != "GENERATION_FAILED" || book.ErrorCode != "GENERATION_FAILED" || book.Stages[0].ErrorCode != "GENERATION_FAILED" || book.Latest[generation.StageDirector].ErrorCode != "GENERATION_FAILED" || batch.Books[0].ErrorCode != "GENERATION_FAILED" || batch.Books[0].Run.ErrorCode != "GENERATION_FAILED" || project.Books[0].Stages[generation.StageDirector].ErrorCode != "GENERATION_FAILED" {
		t.Fatal("derived catalogue code missing")
	}
	// Mutations of every projected reference must stay local to that response.
	book.Stages[0].OutputText = "changed"
	book.Latest[generation.StageDirector] = generation.StageRun{}
	*book.Run.StartedAt = time.Time{}
	*book.Stages[0].StartedAt = time.Time{}
	batch.Books[0].Run.ErrorMessage = "changed"
	*batch.Books[0].Run.StartedAt = time.Time{}
	project.Books[0].Run.Status = generation.StatusCompleted
	project.Books[0].Stages[generation.StageDirector] = generation.StageRun{}
	after, _ := json.Marshal([]any{f.book, f.stage, f.batch, f.project})
	if !bytes.Equal(before, after) {
		t.Fatal("response copy aliases service-owned state")
	}
	for _, tc := range []struct{ message, code string }{
		{"", ""}, {generationSafeMessage, "GENERATION_FAILED"},
		{"生成参数无效，请检查后重试", "GENERATION_INVALID"}, {"生成记录不存在，请刷新后重试", "GENERATION_NOT_FOUND"},
		{"当前生成状态不允许此操作，请刷新后重试", "GENERATION_CONFLICT"}, {"请先生成或检测音频", "AUDIO_MEASUREMENT_REQUIRED"},
		{"音频检测服务暂不可用，请稍后重试", "AUDIO_PROBE_UNAVAILABLE"}, {"导演分镜时长校验失败，请重试导演阶段", "GENERATION_TIMELINE_INVALID"},
		{"生成服务暂不可用，请稍后重试", "GENERATION_UNAVAILABLE"},
	} {
		stage := projectStageRun(generation.StageRun{Status: generation.StatusCompleted, ErrorMessage: tc.message, ErrorCode: "untrusted"})
		if stage.ErrorMessage != tc.message || stage.ErrorCode != tc.code || stage.Status != generation.StatusCompleted {
			t.Fatalf("catalogue readback: %#v", stage)
		}
	}
	for _, values := range []generation.BookGenerationResult{{}, {Stages: []generation.StageRun{}, Latest: map[generation.Stage]generation.StageRun{}}} {
		if !reflect.DeepEqual(projectBookGeneration(values), values) {
			t.Fatal("nil/empty book shape changed")
		}
	}
	for _, values := range []generation.BatchGenerationResult{{}, {Books: []generation.BatchBookResult{}}, {Books: []generation.BatchBookResult{{BookID: 11}}}} {
		if !reflect.DeepEqual(projectBatchGeneration(values), values) {
			t.Fatal("nil/empty batch shape changed")
		}
	}
	for _, values := range []generation.ProjectSummary{{}, {Books: []generation.BookGenerationSummary{}}, {Books: []generation.BookGenerationSummary{{BookID: 11, Stages: map[generation.Stage]generation.StageRun{}}}}} {
		if !reflect.DeepEqual(projectGenerationSummary(values), values) {
			t.Fatal("nil/empty project shape changed")
		}
	}
}

func TestGenerationLegacyProjectionNullAndNonDiagnosticFacts(t *testing.T) {
	snapshot := `{"error":null,"errorCode":"provider-canary-short","source":"{\"error\":\"literal story\"}","extra":[],"nil":null,"large":900719925474099312345,"float":1.2300}`
	got := projectGenerationSnapshot(snapshot)
	for _, expected := range []string{`"error":null`, `"errorCode":""`, `"extra":[]`, `"nil":null`, `900719925474099312345`, `1.2300`, `literal story`} {
		if !strings.Contains(got, expected) {
			t.Errorf("lost %s: %s", expected, got)
		}
	}
}

func TestGenerationOutcomeRoutesSuccessfulLegacyResults(t *testing.T) {
	for _, suffix := range []string{"/generation", "/books/11/generation", "/books/11/generation/stages/DIRECTOR/retry", "/books/11/storyboard/recompile"} {
		t.Run(suffix, func(t *testing.T) {
			f := legacyOutcomeFixture(generationDiagnostic)
			var logs bytes.Buffer
			w := requestOutcome(t, f, "POST", suffix, &logs)
			if w.Code != 200 {
				t.Fatal(w.Code)
			}
			assertSafeOutcomeBody(t, w.Body.String())
			if strings.Contains(w.Body.String(), `"message":`) {
				t.Fatal("success acquired failure envelope")
			}
		})
	}
}
