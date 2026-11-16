package httpapi

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"qiantie/backend/internal/screenplay"
	"qiantie/backend/internal/subjectupload"
	"strings"
	"testing"
	"time"
)

func TestSubjectUploadWholeLibraryRoutesRegister(t *testing.T) {
	library, _, _, _, now := assetLibraryServiceFixture(t)
	library.SubjectUploads = &ScreenplaySubjectUploadService{Repository: screenplay.NewMemorySubjectUploadRepository(), Library: library}
	defer func() {
		if p := recover(); p != nil {
			t.Fatalf("whole library route registration must not panic: %v", p)
		}
	}()
	mux := http.NewServeMux()
	RegisterScreenplayAssetLibraryRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, library)
	w := signedJSONRequest(t, mux, now, "alice", "GET", "/api/comic-screenplays/asset-library/subject-upload-capability", nil)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
}

type subjectTestRunner struct {
	calls   int
	pending bool
}

func TestSubjectUploadDefaultProjectMatchesSettingsAndFilenameIdentity(t *testing.T) {
	library, _, _, _, _ := assetLibraryServiceFixture(t)
	configs := NewUserConfigStore(t.TempDir())
	if e := configs.write("alice", map[string]any{"apiIntegrations": map[string]any{"values": map[string]any{"realpersonChannel": "volcengine", "volcanoAccessKeyId": "own-test-ak", "volcanoSecretAccessKey": "own-test-sk", "imageHostUrl": "https://public.example/upload"}}}); e != nil {
		t.Fatal(e)
	}
	runner := &subjectTestRunner{}
	service := &ScreenplaySubjectUploadService{Repository: screenplay.NewMemorySubjectUploadRepository(), Library: library, Configs: configs, Runner: runner, RedisDo: func(context.Context, ...string) (any, error) { return "OK", nil }}
	c, e := service.configuration(context.Background(), "alice")
	if e != nil || c.ProjectName != "default" {
		t.Fatalf("settings promise default project: %q %v", c.ProjectName, e)
	}
	post := func(filename string) int {
		var b bytes.Buffer
		m := multipart.NewWriter(&b)
		_ = m.WriteField("requestKey", "same-filename-key")
		_ = m.WriteField("name", "角色")
		f, _ := m.CreateFormFile("file", filename)
		_, _ = f.Write(testImageToolPNG(t))
		m.Close()
		r := httptest.NewRequest("POST", "/", &b)
		r.Header.Set("Content-Type", m.FormDataContentType())
		w := httptest.NewRecorder()
		service.serveUpload(w, r, "alice")
		return w.Code
	}
	if n := post("first.png"); n != 200 {
		t.Fatal(n)
	}
	if n := post("second.png"); n != 409 || runner.calls != 1 {
		t.Fatal("changed provider filename accepted", n, runner.calls)
	}
}

func (r *subjectTestRunner) Run(ctx context.Context, c subjectupload.Config, in subjectupload.Input, p subjectupload.Progress, save subjectupload.Checkpoint) (subjectupload.Result, error) {
	r.calls++
	if r.pending {
		p.PendingAction = "CreateAsset"
		if e := save(ctx, p); e != nil {
			return subjectupload.Result{}, e
		}
		return subjectupload.Result{}, fmt.Errorf("unknown outcome")
	}
	return subjectupload.Result{AssetID: "fake-asset", AssetURI: "asset://fake-asset", Provider: c.Provider, Status: "Active"}, nil
}
func TestSubjectUploadServiceIdempotentAndUnknownCreate(t *testing.T) {
	library, _, _, _, _ := assetLibraryServiceFixture(t)
	configs := NewUserConfigStore(t.TempDir())
	if e := configs.write("alice", map[string]any{"apiIntegrations": map[string]any{"values": map[string]any{"realpersonChannel": "volcengine", "volcanoProjectName": "p", "volcanoAccessKeyId": "test-ak", "volcanoSecretAccessKey": "test-sk", "imageHostUrl": "https://public.example/upload"}}}); e != nil {
		t.Fatal(e)
	}
	runner := &subjectTestRunner{}
	s := &ScreenplaySubjectUploadService{Repository: screenplay.NewMemorySubjectUploadRepository(), Library: library, Configs: configs, Runner: runner, RedisDo: func(context.Context, ...string) (any, error) { return "OK", nil }}
	post := func(key, name string) int {
		var b bytes.Buffer
		m := multipart.NewWriter(&b)
		_ = m.WriteField("requestKey", key)
		_ = m.WriteField("name", name)
		f, _ := m.CreateFormFile("file", "test.png")
		_, _ = f.Write(testImageToolPNG(t))
		m.Close()
		req := httptest.NewRequest("POST", "/", &b)
		req.Header.Set("Content-Type", m.FormDataContentType())
		w := httptest.NewRecorder()
		s.serveUpload(w, req, "alice")
		if strings.Contains(w.Body.String(), "test-sk") {
			t.Fatal("credential leak")
		}
		return w.Code
	}
	if n := post("request00", ""); n != 400 || runner.calls != 0 {
		t.Fatal(n, runner.calls)
	}
	if tomb, e := s.Repository.GetSubjectUpload(context.Background(), "alice", "request00"); e != nil || tomb.Status != "rejected" || tomb.Digest != "rejected-v1" {
		t.Fatal(tomb, e)
	}
	if n := post("request00", "修正"); n != 200 || runner.calls != 0 {
		t.Fatal(n, runner.calls)
	}
	if n := post("request01", "角色"); n != 200 {
		t.Fatal(n)
	}
	if n := post("request01", "角色"); n != 200 || runner.calls != 1 {
		t.Fatal(n, runner.calls)
	}
	if n := post("request01", "变更"); n != 409 {
		t.Fatal(n)
	}
	runner.pending = true
	if n := post("request02", "角色"); n != 200 {
		t.Fatal(n)
	}
	if n := post("request02", "角色"); n != 200 || runner.calls != 2 {
		t.Fatal(n, runner.calls)
	}
	v, e := s.Repository.GetSubjectUpload(context.Background(), "alice", "request02")
	if e != nil || v.Status != "needs_review" {
		t.Fatal(v, e)
	}
	runner.pending = false
	cfg := configs.read("alice")
	values := apiIntegrationValues(apiIntegrationStoredMap(cfg))
	delete(values, "imageHostUrl")
	if e = configs.write("alice", cfg); e != nil {
		t.Fatal(e)
	}
	if n := post("request03", "角色"); n != 200 || runner.calls != 2 {
		t.Fatal(n, runner.calls)
	}
	v, e = s.Repository.GetSubjectUpload(context.Background(), "alice", "request03")
	if e != nil || v.Status != "paused" {
		t.Fatal(v, e)
	}
	values["imageHostUrl"] = "https://public.example/upload"
	if e = configs.write("alice", cfg); e != nil {
		t.Fatal(e)
	}
	if n := post("request03", "角色"); n != 200 || runner.calls != 3 {
		t.Fatal(n, runner.calls)
	}
	v, e = s.Repository.GetSubjectUpload(context.Background(), "alice", "request03")
	if e != nil || v.Status != "succeeded" {
		t.Fatal(v, e)
	}
	if n := post("request03", "角色"); n != 200 || runner.calls != 3 {
		t.Fatal(n, runner.calls)
	}
}

func TestSubjectUploadLeaseRequiresRedisAndComparesToken(t *testing.T) {
	if _, e := subjectUploadLease(context.Background(), nil, "alice", "request01"); e == nil {
		t.Fatal("missing Redis accepted")
	}
	var key, token string
	calls := 0
	release, e := subjectUploadLease(context.Background(), func(_ context.Context, args ...string) (any, error) {
		calls++
		if args[0] == "SET" {
			if strings.Join(args[3:], " ") != "NX EX 300" {
				t.Fatal(args)
			}
			key = args[1]
			token = args[2]
			return "OK", nil
		}
		if args[0] != "EVAL" || args[3] != key || args[4] != token || !strings.Contains(args[1], "ARGV[1]") {
			t.Fatal(args)
		}
		return int64(1), nil
	}, "alice", "request01")
	if e != nil {
		t.Fatal(e)
	}
	release()
	if calls != 2 {
		t.Fatal(calls)
	}
	if _, e = subjectUploadLease(context.Background(), func(context.Context, ...string) (any, error) { return nil, nil }, "alice", "request01"); e == nil {
		t.Fatal("busy lock accepted")
	}
}
func TestSubjectUploadPublicNeverExposesPrivateProgress(t *testing.T) {
	v := subjectUploadPublic(screenplay.SubjectUploadOperation{Owner: "private-owner", RequestKey: "request01", Digest: "private-digest", Progress: subjectupload.Progress{ImageURL: "https://secret.invalid/?credential=secret", PendingAction: "secret"}, Result: &subjectupload.Result{AssetID: "asset1", Provider: "volcengine"}})
	out := fmt.Sprint(v)
	for _, secret := range []string{"private-owner", "private-digest", "credential=secret", "pending_action"} {
		if strings.Contains(out, secret) {
			t.Fatal(out)
		}
	}
}

func TestSubjectUploadRejectedPublicAndStaleReconciliation(t *testing.T) {
	public := subjectUploadPublic(screenplay.SubjectUploadOperation{RequestKey: "rejected01", Status: "rejected"})["operation"].(map[string]any)
	if public["status"] != "failed" || public["safeToStartNew"] != true {
		t.Fatal(public)
	}
	repo := screenplay.NewMemorySubjectUploadRepository()
	row, _ := repo.SaveSubjectUpload(context.Background(), screenplay.SubjectUploadOperation{Owner: "alice", RequestKey: "stale001", Digest: "d", Status: "running"}, 0)
	s := &ScreenplaySubjectUploadService{Repository: repo, RedisDo: func(context.Context, ...string) (any, error) { return nil, nil }}
	if got := s.reconcileSubjectUpload(context.Background(), row); got.Status != "running" {
		t.Fatal("active changed", got)
	}
	s.RedisDo = func(context.Context, ...string) (any, error) { return "OK", nil }
	row = s.reconcileSubjectUpload(context.Background(), row)
	if row.Status != "paused" {
		t.Fatal(row)
	}
	row.Status = "running"
	row.Progress.PendingAction = "CreateAsset"
	row, _ = repo.SaveSubjectUpload(context.Background(), row, row.Revision)
	if got := s.reconcileSubjectUpload(context.Background(), row); got.Status != "needs_review" {
		t.Fatal(got)
	}
}
