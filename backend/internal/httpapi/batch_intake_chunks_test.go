package httpapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/mergeworker"
	"qiantie/backend/internal/novelfetchworkshop"
)

// An isolated Redis executes the real quota scripts, including races and expiry.
func intakeTestServer(t *testing.T) *mergeworker.RedisStore {
	t.Helper()
	url := os.Getenv("BATCH_INTAKE_TEST_REDIS_URL")
	if url == "" {
		bin, err := exec.LookPath("redis-server")
		if err != nil && runtime.GOOS == "windows" {
			bin = filepath.Join("C:\\Redis", "redis-server.exe")
			_, err = os.Stat(bin)
		}
		if err != nil {
			t.Skip("install redis-server or set BATCH_INTAKE_TEST_REDIS_URL to an isolated Redis")
		}
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		port := ln.Addr().(*net.TCPAddr).Port
		ln.Close()
		cmd := exec.Command(bin, "--bind", "127.0.0.1", "--port", strconv.Itoa(port), "--dir", t.TempDir(), "--save", "", "--appendonly", "no")
		if err = cmd.Start(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
		url = fmt.Sprintf("redis://127.0.0.1:%d", port)
	}
	s, err := mergeworker.NewRedisStore(url, "test")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		if _, err = s.Do(context.Background(), "PING"); err == nil {
			return s
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal(err)
	return nil
}

func TestLargeBatchStagedSourcesCreateOneOwnedBatch(t *testing.T) {
	now := time.Unix(100, 0)
	redis := intakeTestServer(t)
	store := &intakeFailOnceStore{MemoryStore: batchfactoryv11.NewMemoryStore()}
	api := NewRouter(RouterOptions{Store: store, Slice: 1, BridgeSecret: "secret", Now: func() time.Time { return now }, NovelFetchQueue: novelfetchworkshop.NewRedisQueueWithDo("test", novelfetchworkshop.RedisDoFunc(redis.Do))})
	oversized := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/intakes/source-chunks", intakeChunkInput{MaxFieldBytes: intakeFieldLimit + 1, BookID: "877777", Platform: "3"})
	if oversized.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized field: HTTP %d", oversized.Code)
	}
	source := strings.Repeat("甲", 3000000)
	content, _ := json.Marshal(map[string]string{"sourceText": "截取", "originalRaw": source})
	const size = 512 << 10
	group := func(platform, ref string) map[string]any {
		return map[string]any{"platformId": platform, "inputText": "877777 书", "sourceRefsByBookId": map[string]string{"877777": ref}}
	}
	create := func(owner string, groups ...map[string]any) int {
		return signedJSONRequest(t, api, now, owner, http.MethodPost, "/api/batch-factory/v12/intakes/manual", map[string]any{"title": "批次", "groups": groups}).Code
	}
	groups := []map[string]any{}
	for _, platform := range []string{"3", "15"} {
		ref := ""
		for i := 0; i*size < len(content); i++ {
			end := min((i+1)*size, len(content))
			rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/intakes/source-chunks", map[string]any{"platform": platform, "bookId": "877777", "bytes": len(content), "index": i, "ref": ref, "data": base64.StdEncoding.EncodeToString(content[i*size : end])})
			if rec.Code != 200 {
				t.Fatalf("chunk %d: %d %s", i, rec.Code, rec.Body.String())
			}
			newRef := decodeBody[map[string]string](t, rec)["ref"]
			if newRef == "" || (ref != "" && newRef != ref) {
				t.Fatalf("ref=%q", newRef)
			}
			ref = newRef
			if i == 0 && create("alice", group(platform, ref)) < 400 {
				t.Fatal("missing chunks accepted")
			}
		}
		if create("bob", group(platform, ref)) < 400 || create("alice", group("wrong", ref)) < 400 {
			t.Fatal("foreign ref accepted")
		}
		groups = append(groups, group(platform, ref))
	}
	store.fail = true
	if code := create("alice", groups...); code < 400 {
		t.Fatalf("failed create discarded retry data: status %d", code)
	}
	for _, id := range store.intakeIDs {
		if record, err := store.GetIntake(context.Background(), "alice", id); err == nil && record.ConsumedAt == nil {
			t.Fatal("failed transaction left an unconsumed intake")
		}
	}
	if code := create("alice", groups...); code != http.StatusCreated {
		t.Fatalf("create code=%d", code)
	}
	batches, _ := store.ListBatches(context.Background(), "alice")
	if len(batches) != 1 || len(batches[0].Books) != 2 {
		t.Fatalf("batch count %d", len(batches))
	}
	for _, book := range batches[0].Books {
		if book.TxtText != source || book.SourceText != "截取" || book.SourceMetadata["rawSourceStatus"] != "verified" {
			t.Fatal("lost source")
		}
	}
	if create("alice", groups...) < 400 {
		t.Fatal("consumed ref accepted again")
	}
	if len(store.intakeIDs) != 0 {
		t.Fatalf("staged creation persisted %d redundant intakes", len(store.intakeIDs))
	}
	for _, group := range groups {
		key, err := intakeChunkKey("alice", group["platformId"].(string), "877777", group["sourceRefsByBookId"].(map[string]string)["877777"])
		if err != nil {
			t.Fatal(err)
		}
		exists, err := redis.Do(context.Background(), "HEXISTS", intakeOwnerKey("alice"), key)
		if err != nil || exists != int64(0) {
			t.Fatalf("committed allocation not released: %v %v", exists, err)
		}
	}
}

type intakeFailOnceStore struct {
	*batchfactoryv11.MemoryStore
	fail      bool
	intakeIDs []string
}

func (s *intakeFailOnceStore) CreateIntake(ctx context.Context, owner string, in batchfactoryv11.NovelFetchIntakeInput) (batchfactoryv11.Intake, error) {
	v, err := s.MemoryStore.CreateIntake(ctx, owner, in)
	if err == nil {
		s.intakeIDs = append(s.intakeIDs, v.ID)
	}
	return v, err
}

func (s *intakeFailOnceStore) CreateBatch(ctx context.Context, owner string, in batchfactoryv11.CreateBatchInput) (batchfactoryv11.Batch, error) {
	if s.fail {
		s.fail = false
		return batchfactoryv11.Batch{}, fmt.Errorf("simulated failed transaction")
	}
	return s.MemoryStore.CreateBatch(ctx, owner, in)
}

func (s *intakeFailOnceStore) CreateBatchFromIntake(ctx context.Context, owner, id string, in batchfactoryv11.CreateBatchInput) (batchfactoryv11.Batch, error) {
	if s.fail {
		s.fail = false
		return batchfactoryv11.Batch{}, fmt.Errorf("simulated failed transaction")
	}
	return s.MemoryStore.CreateBatchFromIntake(ctx, owner, id, in)
}

func TestStagedNovelTransferPreservesExplicitRawAndExpires(t *testing.T) {
	ctx := context.Background()
	redis := intakeTestServer(t)
	stager := intakeChunks{do: redis.Do}
	store := batchfactoryv11.NewMemoryStore()
	put := func(body, bookID string) string {
		ref, err := stager.put(ctx, "alice", intakeChunkInput{Platform: "3", BookID: bookID, Bytes: len(body), Data: base64.StdEncoding.EncodeToString([]byte(body))})
		if err != nil {
			t.Fatal(err)
		}
		return ref
	}
	ref := put(`{"sourceText":"已确认 AI 版","originalRaw":" 全文\n"}`, "123")
	in := batchfactoryv11.NovelFetchIntakeInput{Books: []batchfactoryv11.CreateBookInput{{BookID: "123", Platform: "3", Title: "甲", SourceRef: ref, SourceMetadata: map[string]any{"sourceMode": "novel_fetch", "sourceContentVersion": "ai1"}}}}
	result, err := stager.createNovelIntake(ctx, "alice", store, in)
	if err != nil {
		t.Fatal(err)
	}
	batch, err := store.CreateBatchFromIntake(ctx, "alice", result.ID, batchfactoryv11.CreateBatchInput{Title: "迁移"})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	if book.SourceText != "已确认 AI 版" || book.TxtText != " 全文\n" || book.SourceMetadata["rawSourceOrigin"] != "novel_fetch_original_raw" || book.SourceMetadata["sourceContentVersion"] != "ai1" {
		t.Fatalf("bad transfer %+v", book)
	}
	ref = put(`{"sourceText":"旧生产稿"}`, "456")
	key, _ := intakeChunkKey("alice", "3", "456", ref)
	redis.Do(ctx, "DEL", key)
	read := intakeSourceRead{chunks: stager}
	if _, err := read.book(ctx, "alice", "3", "456", ref); err == nil || !strings.Contains(err.Error(), "过期") {
		t.Fatalf("expired stage: %v", err)
	}
}

func TestExternalIntakeCannotAttestViaMetadata(t *testing.T) {
	s := batchfactoryv11.NewMemoryStore()
	ctx := context.Background()
	for _, mode := range []string{"novel_fetch", "custom_copy", ""} {
		in := batchfactoryv11.NovelFetchIntakeInput{Books: []batchfactoryv11.CreateBookInput{{BookID: "123", Title: "旧书", TxtText: "旧稿", SourceText: "旧稿", SourceMetadata: map[string]any{"sourceMode": mode, "rawSourceStatus": "verified", "rawSourceOrigin": "bookstore_raw"}}}}
		i, err := (intakeChunks{}).createNovelIntake(ctx, "alice", s, in)
		if err != nil {
			t.Fatal(err)
		}
		b, err := s.CreateBatchFromIntake(ctx, "alice", i.ID, batchfactoryv11.CreateBatchInput{Title: "批次"})
		if err != nil {
			t.Fatal(err)
		}
		if b.Books[0].SourceMetadata["rawSourceStatus"] != "unverified" || b.Books[0].SourceMetadata["rawSourceOrigin"] != nil {
			t.Fatalf("forged provenance accepted for %q", mode)
		}
	}
}

func TestExternalIntakeBatchOverridesCannotAttest(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	now := time.Now()
	api := NewRouter(RouterOptions{Store: store, Slice: 1, BridgeSecret: "secret", Now: func() time.Time { return now }})
	intake, err := store.CreateIntake(ctx, "alice", batchfactoryv11.NovelFetchIntakeInput{Books: []batchfactoryv11.CreateBookInput{{BookID: "123", Title: "甲"}}})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/api/batch-factory/v11/batches", "/api/batch-factory/v11/intakes/" + intake.ID + "/batches"} {
		in := batchfactoryv11.CreateBatchInput{Title: "批次", Books: []batchfactoryv11.CreateBookInput{{BookID: "123", Title: "甲", TxtText: "旧稿", SourceMetadata: map[string]any{"rawSourceStatus": "verified", "rawSourceOrigin": "forged"}}}}
		rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, in)
		if rec.Code != http.StatusCreated {
			t.Fatalf("%s %d %s", path, rec.Code, rec.Body.String())
		}
		got := decodeBody[map[string]batchfactoryv11.Batch](t, rec)["batch"].Books[0]
		if got.SourceMetadata["rawSourceStatus"] != "unverified" || got.SourceMetadata["rawSourceOrigin"] != nil {
			t.Fatalf("accepted caller status at %s", path)
		}
	}
}

func TestIntakeMediumTextByteLimit(t *testing.T) {
	ctx := context.Background()
	store := intakeTestServer(t)
	stager := intakeChunks{do: store.Do}
	owner, _ := intakeToken()
	exact := strings.Repeat("甲", intakeFieldLimit/3)
	if len(exact) != intakeFieldLimit || validateIntakeFields(exact, exact, exact) != nil {
		t.Fatal("valid MEDIUMTEXT boundary rejected")
	}
	over := exact + "甲"
	for _, fields := range [][3]string{{over, "", ""}, {"", over, ""}, {"", "", over}} {
		if !errors.Is(validateIntakeFields(fields[0], fields[1], fields[2]), errIntakeFieldBytes) {
			t.Fatal("accepted oversized field")
		}
	}
	input := intakeChunkInput{BookID: "123", Platform: "3", MaxFieldBytes: len(over)}
	if _, err := stager.put(ctx, owner, input); !errors.Is(err, errIntakeFieldBytes) {
		t.Fatalf("declared oversized field accepted: %v", err)
	}
	// Old clients omit field-size hints. Verify real UTF-8 bytes on the last
	// chunk, return a 413 and immediately release the invalid reservation.
	data, _ := json.Marshal(stagedBookBody{OriginalRaw: over})
	input.MaxFieldBytes = 0
	input.Bytes = len(data)
	for i := 0; i < len(data); i += intakeChunkSize {
		input.Index = i / intakeChunkSize
		input.Data = base64.StdEncoding.EncodeToString(data[i:min(i+intakeChunkSize, len(data))])
		ref, err := stager.put(ctx, owner, input)
		if i+intakeChunkSize < len(data) {
			if err != nil {
				t.Fatal(err)
			}
			input.Ref = ref
			continue
		}
		if !errors.Is(err, errIntakeFieldBytes) {
			t.Fatalf("actual field size was not validated: %v", err)
		}
		rec := httptest.NewRecorder()
		writeIntakeSourceError(rec, err)
		if rec.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("status=%d", rec.Code)
		}
	}
	left, err := store.Do(ctx, "HLEN", intakeOwnerKey(owner))
	if err != nil || left != int64(0) {
		t.Fatalf("invalid upload leaked quota %v %v", left, err)
	}
}

func TestIntakeRefCountAndLostFirstReply(t *testing.T) {
	ctx := context.Background()
	redis := intakeTestServer(t)
	owner, _ := intakeToken()
	open := func(account, id string) (string, error) {
		s := intakeChunks{do: redis.Do, secret: "shared"}
		return s.put(ctx, account, intakeChunkInput{Platform: "3", BookID: "123", UploadID: id, Bytes: 2, Data: "e30="})
	}
	id, _ := intakeToken()
	ref, err := open(owner, id)
	if err != nil {
		t.Fatal(err)
	}
	refAgain, err := open(owner, id)
	if err != nil || refAgain != ref {
		t.Fatalf("lost acknowledgement created second reservation %s %v", refAgain, err)
	}
	for i := 1; i < intakeOwnerRefs; i++ {
		id, _ := intakeToken()
		if _, err := open(owner, id); err != nil {
			t.Fatalf("ref %d: %v", i, err)
		}
	}
	extra, _ := intakeToken()
	if _, err := open(owner, extra); err == nil || !strings.Contains(err.Error(), "50") {
		t.Fatalf("owner ref cap missed: %v", err)
	}
	other, _ := intakeToken()
	if _, err := open(other, extra); err != nil {
		t.Fatalf("budget shared across users: %v", err)
	}
}

func TestIntakeOwnerOutstandingBudgetAcrossWorkers(t *testing.T) {
	ctx := context.Background()
	redis := intakeTestServer(t)
	owner, _ := intakeToken()
	input := intakeChunkInput{Platform: "3", BookID: "123", Index: 0, Bytes: 24 << 20, Data: base64.StdEncoding.EncodeToString(make([]byte, 512<<10))}
	start := make(chan struct{})
	results := make(chan string, 3)
	for i := 0; i < 3; i++ {
		go func() {
			<-start
			ref, err := (intakeChunks{do: redis.Do}).put(ctx, owner, input)
			if err != nil {
				ref = ""
			}
			results <- ref
		}()
	}
	close(start)
	refs := []string{}
	for i := 0; i < 3; i++ {
		if ref := <-results; ref != "" {
			refs = append(refs, ref)
		}
	}
	if len(refs) != 2 {
		t.Fatalf("48 MiB budget accepted %d of 3 x 24 MiB concurrent reservations", len(refs))
	}
	input.Ref = refs[0]
	if _, err := (intakeChunks{do: redis.Do}).put(ctx, owner, input); err != nil {
		t.Fatalf("retry charged twice: %v", err)
	}
	key, _ := intakeChunkKey(owner, "3", "123", input.Ref)
	redis.Do(ctx, "PEXPIRE", key, "1")
	time.Sleep(10 * time.Millisecond)
	input.Ref = ""
	if _, err := (intakeChunks{do: redis.Do}).put(ctx, owner, input); err != nil {
		t.Fatalf("expired partial reservation not recovered: %v", err)
	}
}
