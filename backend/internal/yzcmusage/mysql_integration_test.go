package yzcmusage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"os"
	"testing"
	"time"
)

// This opt-in test is restricted to the existing isolated local review database.
// It cannot run against the production DSN or another database by accident.
func TestMySQLSharedKeyLedgerIsolation(t *testing.T) {
	dsn := os.Getenv("QIANTIE_USAGE_TEST_DSN")
	if dsn == "" {
		t.Skip("local review database not selected")
	}
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Net != "tcp" || cfg.Addr != "127.0.0.1:3307" || cfg.DBName != "qiantie_v88" {
		t.Fatal("only the isolated local review database is permitted")
	}
	cfg.InterpolateParams = true // Match the application runtime connection settings.
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for _, ddl := range SchemaStatements() {
		if _, err = db.ExecContext(ctx, ddl); err != nil {
			t.Fatal(err)
		}
	}
	prefix := fmt.Sprintf("qa_usage_%d", time.Now().UnixNano())
	a, b, pending := prefix+"a", prefix+"b", prefix+"p"
	ownerA, ownerB := fmt.Sprintf("qa_a_%d", time.Now().UnixNano()), fmt.Sprintf("qa_b_%d", time.Now().UnixNano())
	defer db.ExecContext(context.Background(), "DELETE FROM yzcm_usage_requests WHERE request_id IN (?,?,?)", a, b, pending)
	store := Store{DB: db}
	for _, r := range []Request{{a, ownerA, "image"}, {b, ownerB, "script"}, {pending, ownerA, "image"}} {
		if err = store.Record(ctx, r, Fingerprint("shared-test-key"), 200); err != nil {
			t.Fatal(err)
		}
	}
	// A duplicate delivery cannot move an existing request to a different member.
	if err = store.Record(ctx, Request{a, ownerB, "image"}, Fingerprint("shared-test-key"), 200); err != nil {
		t.Fatal(err)
	}
	for _, e := range []Entry{{ID: a, Owner: ownerA, Model: "QA-FIXTURE", Quota: 3356, InputTokens: 12, OutputTokens: 8, Cost: .049, Currency: "CNY"}, {ID: b, Owner: ownerB, Model: "QA-FIXTURE", Quota: 1000, InputTokens: 40, OutputTokens: 20, Cost: .5, Currency: "CNY"}} {
		if err = store.Settle(ctx, e); err != nil {
			t.Fatal(err)
		}
	}
	result, err := store.ReadSummary(ctx, ownerA)
	if err != nil {
		t.Fatal(err)
	}
	month := result["month"].(map[string]any)
	if month["calls"] != int64(2) || month["pricedCalls"] != int64(1) || month["pendingCalls"] != int64(1) || month["actualCost"] != float64(.049) || month["totalTokens"] != int64(20) {
		t.Fatal(month)
	}
	for _, row := range result["recent"].([]map[string]any) {
		if row["id"] == b {
			t.Fatal("another member's charge leaked")
		}
	}
	total, err := store.TeamMonthTokens(ctx, []string{ownerA, ownerB})
	if err != nil || total != 80 {
		t.Fatal(total, err)
	}
}
