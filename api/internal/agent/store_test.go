package agent

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

type storeTxFake struct {
	queries []string
	args    [][]any
	commit  bool
}

func (f *storeTxFake) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	f.queries = append(f.queries, query)
	f.args = append(f.args, append([]any(nil), args...))
	return storeResult(1), nil
}
func (f *storeTxFake) Commit() error   { f.commit = true; return nil }
func (f *storeTxFake) Rollback() error { return nil }

type storeResult int64
func (r storeResult) LastInsertId() (int64, error) { return int64(r), nil }
func (r storeResult) RowsAffected() (int64, error) { return int64(r), nil }

func TestNormalizeMediaAssetIDsAllowsOnlyOpaqueIDs(t *testing.T) {
	got, err := normalizeMediaAssetIDs([]string{"asset_1", "asset-2", "asset_1"})
	if err != nil { t.Fatal(err) }
	if len(got) != 2 || got[0] != "asset_1" || got[1] != "asset-2" {
		t.Fatalf("unexpected media ids: %#v", got)
	}
	for _, bad := range []string{"/tmp/a.png", "https://example.com/a.png", "../a", "asset id"} {
		if _, err := normalizeMediaAssetIDs([]string{bad}); err == nil {
			t.Fatalf("expected %q to be rejected", bad)
		}
	}
}

func TestSQLStoreAppendMessageWritesMessageAndMediaAtomically(t *testing.T) {
	tx := &storeTxFake{}
	store := newSQLStoreWithBegin(nil, func(context.Context) (agentTx, error) { return tx, nil })
	message, err := store.AppendMessage(context.Background(), AppendMessageInput{
		ThreadID: "thread-1",
		Owner: "owner-1",
		Role: RoleUser,
		Content: "这张图重新生成",
		MediaAssetIDs: []string{"asset_1", "asset_2"},
	})
	if err != nil { t.Fatal(err) }
	if message.ID == "" { t.Fatal("expected message id") }
	if !tx.commit { t.Fatal("expected transaction commit") }
	if len(tx.queries) != 4 { t.Fatalf("queries=%d", len(tx.queries)) }
	if !strings.Contains(tx.queries[0], "INSERT INTO agent_messages") { t.Fatalf("message query=%s", tx.queries[0]) }
	if !strings.Contains(tx.queries[1], "INSERT INTO agent_message_media") || !strings.Contains(tx.queries[2], "INSERT INTO agent_message_media") {
		t.Fatalf("media queries=%#v", tx.queries[1:3])
	}
	if !strings.Contains(tx.queries[3], "UPDATE agent_threads") { t.Fatalf("thread query=%s", tx.queries[3]) }
	if got := tx.args[1][1]; got != "asset_1" { t.Fatalf("first media=%v", got) }
	if got := tx.args[2][1]; got != "asset_2" { t.Fatalf("second media=%v", got) }
}

func TestValidateMessageInputRejectsEmptyAndOversizedContent(t *testing.T) {
	if err := validateMessageInput(AppendMessageInput{ThreadID: "t", Owner: "o", Role: RoleUser, Content: "  "}); err == nil {
		t.Fatal("expected empty content error")
	}
	if err := validateMessageInput(AppendMessageInput{ThreadID: "t", Owner: "o", Role: RoleUser, Content: strings.Repeat("a", MaxMessageRunes+1)}); err == nil {
		t.Fatal("expected oversized content error")
	}
}
