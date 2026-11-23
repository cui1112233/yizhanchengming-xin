package httpapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

// Redis is the only external dependency in this HTTP-to-database test.
type intakeTestRedis map[string]string

func (s intakeTestRedis) command(_ context.Context, args ...string) (any, error) {
	switch args[0] {
	case "SET":
		if _, ok := s[args[1]]; ok {
			return nil, nil
		}
		if !strings.Contains(strings.Join(args[3:], " "), "EX ") {
			return nil, fmt.Errorf("staged value needs expiry")
		}
		s[args[1]] = args[2]
		return "OK", nil
	case "GET":
		return s[args[1]], nil
	case "DEL":
		for _, key := range args[1:] {
			delete(s, key)
		}
		return int64(1), nil
	case "EVAL":
		n := 0
		fmt.Sscan(args[2], &n)
		if s[args[3]] == args[3+n] {
			for _, key := range args[3 : 3+n] {
				delete(s, key)
			}
		}
		return int64(1), nil
	}
	return nil, fmt.Errorf("unexpected command %q", args[0])
}

func TestLargeBatchStagedSourcesCreateOneOwnedBatch(t *testing.T) {
	now := time.Unix(100, 0)
	redis := intakeTestRedis{}
	store := &intakeFailOnceStore{MemoryStore: batchfactoryv11.NewMemoryStore()}
	api := NewRouter(RouterOptions{Store: store, Slice: 1, BridgeSecret: "secret", Now: func() time.Time { return now }, NovelFetchQueue: novelfetchworkshop.NewRedisQueueWithDo("test", redis.command)})
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
	if code := create("alice", groups...); code < 400 || len(redis) == 0 {
		t.Fatalf("failed create discarded retry data: status %d", code)
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
	if len(redis) != 0 {
		t.Fatalf("%d stage keys not cleaned", len(redis))
	}
	if create("alice", groups...) < 400 {
		t.Fatal("consumed ref accepted again")
	}
}

type intakeFailOnceStore struct {
	*batchfactoryv11.MemoryStore
	fail bool
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
	redis := intakeTestRedis{}
	stager := intakeChunks{do: redis.command}
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
	if book.SourceText != "已确认 AI 版" || book.TxtText != " 全文\n" || book.SourceMetadata["rawSourceOrigin"] != "novel_fetch_original_raw" || book.SourceMetadata["sourceContentVersion"] != "ai1" || len(redis) != 0 {
		t.Fatalf("bad transfer %+v", book)
	}
	ref = put(`{"sourceText":"旧生产稿"}`, "456")
	clear(redis)
	read := intakeSourceRead{chunks: stager}
	if _, err := read.book(ctx, "alice", "3", "456", ref); err == nil || !strings.Contains(err.Error(), "过期") {
		t.Fatalf("expired stage: %v", err)
	}
}
