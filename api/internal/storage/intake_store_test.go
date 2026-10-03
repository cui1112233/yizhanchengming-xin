package storage

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/batchfactory"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/novel"
)

type intakeStoreTxFake struct {
	queries []string
	args    [][]any
	commit  bool
}

func (f *intakeStoreTxFake) ExecContext(_ context.Context, q string, args ...any) (sql.Result, error) {
	f.queries = append(f.queries, q)
	f.args = append(f.args, append([]any(nil), args...))
	return intakeStoreResult(1), nil
}
func (f *intakeStoreTxFake) Commit() error { f.commit = true; return nil }
func (f *intakeStoreTxFake) Rollback() error { return nil }

type intakeStoreResult int64
func (r intakeStoreResult) LastInsertId() (int64, error) { return int64(r), nil }
func (r intakeStoreResult) RowsAffected() (int64, error) { return int64(r), nil }

func TestSQLIntakeStoreCreatesIntakeAndGroupedBooksAtomically(t *testing.T) {
	tx := &intakeStoreTxFake{}
	store := newSQLIntakeStoreWithBegin(func(context.Context) (intakeStoreTx, error) { return tx, nil })
	err := store.CreateIntake(context.Background(), batchfactory.IntakeRecord{
		ID: "intake-1", Owner: "user-1", Title: "批量",
		Books: []batchfactory.IntakeBookRecord{
			{PlatformID: "zhihu", PlatformName: "知乎", MaxTxt: 2000, BookID: "z1", ManualGender: novel.GenderFemale, Style: "现代女主"},
			{PlatformID: "dianzhong", PlatformName: "点众", MaxTxt: 4000, BookID: "d1"},
		},
	})
	if err != nil { t.Fatal(err) }
	if !tx.commit { t.Fatal("expected transaction commit") }
	if len(tx.queries) != 3 { t.Fatalf("queries=%d", len(tx.queries)) }
	if !strings.Contains(tx.queries[0], "INSERT INTO intakes") { t.Fatalf("first query=%s", tx.queries[0]) }
	if !strings.Contains(tx.queries[1], "max_txt") || !strings.Contains(tx.queries[1], "manual_gender") || !strings.Contains(tx.queries[1], "resolved_gender") { t.Fatalf("book query=%s", tx.queries[1]) }
	if got := tx.args[1][3]; got != 2000 { t.Fatalf("max_txt=%v", got) }
	if got := tx.args[1][5]; got != string(novel.GenderFemale) { t.Fatalf("manual_gender=%v", got) }
	if got := tx.args[1][6]; got != string(novel.GenderFemale) { t.Fatalf("resolved_gender=%v", got) }
	if got := tx.args[1][7]; got != string(novel.GenderSourceManual) { t.Fatalf("gender_source=%v", got) }
}
