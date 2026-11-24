package yzcmusage

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFailedReceiptSurvivesRestartWithoutKeyOrContent(t *testing.T) {
	db, m, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	dir := filepath.Join(t.TempDir(), "receipts")
	at := time.Date(2026, 11, 24, 2, 0, 0, 0, time.UTC)
	r := Request{ID: "request1", Owner: "alice", Feature: "image"}
	hash := Fingerprint("secret-key")
	tracker := &Tracker{Store: Store{DB: db}, ReceiptDir: dir}
	m.ExpectExec("INSERT IGNORE INTO yzcm_usage_requests").WithArgs(r.ID, r.Owner, r.Feature, hash, at, 200).WillReturnError(errors.New("db temporarily unavailable"))
	tracker.recordReceipt(context.Background(), r, hash, 200, at)
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("durable receipt absent %v %v", entries, err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "secret-key") || strings.Contains(string(raw), "prompt") || strings.Contains(string(raw), "Authorization") {
		t.Fatal("sensitive content persisted")
	}
	restart := &Tracker{Store: Store{DB: db}, ReceiptDir: dir}
	m.ExpectExec("INSERT IGNORE INTO yzcm_usage_requests").WithArgs(r.ID, r.Owner, r.Feature, hash, at, 200).WillReturnResult(sqlmock.NewResult(0, 1))
	restart.flushReceipts(context.Background())
	entries, _ = os.ReadDir(dir)
	if len(entries) != 0 {
		t.Fatal("successful durable receipt not removed")
	}
	if err = m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
func TestReceiptDuplicateCannotMoveOwnerAndFailedFlushRetainsFile(t *testing.T) {
	db, m, _ := sqlmock.New()
	defer db.Close()
	dir := filepath.Join(t.TempDir(), "receipts")
	tracker := &Tracker{Store: Store{DB: db}, ReceiptDir: dir}
	at := time.Now().UTC()
	hash := Fingerprint("secret")
	if err := tracker.saveReceipt(receipt{Request: Request{"same", "alice", "image"}, CredentialHash: hash, Status: 200, At: at}); err != nil {
		t.Fatal(err)
	}
	if err := tracker.saveReceipt(receipt{Request: Request{"same", "bob", "image"}, CredentialHash: hash, Status: 200, At: at}); err != nil {
		t.Fatal(err)
	}
	m.ExpectExec("INSERT IGNORE INTO yzcm_usage_requests").WithArgs("same", "alice", "image", hash, at, 200).WillReturnError(errors.New("still down"))
	tracker.flushReceipts(context.Background())
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatal("uncommitted receipt deleted")
	}
	m.ExpectExec("INSERT IGNORE INTO yzcm_usage_requests").WithArgs("same", "alice", "image", hash, at, 200).WillReturnResult(sqlmock.NewResult(0, 0))
	tracker.flushReceipts(context.Background())
	entries, _ = os.ReadDir(dir)
	if len(entries) != 0 {
		t.Fatal("idempotent alreadyrecorded receipt retained")
	}
	if err := m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
func TestReceiptRejectsUntrustedNamesAndDirectories(t *testing.T) {
	tracker := &Tracker{ReceiptDir: filepath.Join(t.TempDir(), "r")}
	if err := tracker.saveReceipt(receipt{Request: Request{"../escape", "alice", "image"}, CredentialHash: "invalid", At: time.Now()}); err == nil {
		t.Fatal("invalid receipt accepted")
	}
	if _, err := os.Stat(tracker.ReceiptDir); !os.IsNotExist(err) {
		t.Fatal("invalid receipt created directory")
	}
}
func TestTrackerDatabaseFailureKeepsGenerationResponse(t *testing.T) {
	db, m, _ := sqlmock.New()
	defer db.Close()
	dir := filepath.Join(t.TempDir(), "r")
	tracker := &Tracker{Store: Store{DB: db}, ReceiptDir: dir}
	m.ExpectExec("INSERT IGNORE INTO yzcm_usage_requests").WillReturnError(errors.New("db down"))
	transport := tracker.Wrap(transportFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{"X-Oneapi-Request-Id": []string{"actual-call"}}, Body: io.NopCloser(strings.NewReader("unchanged private generation"))}, nil
	}))
	r, _ := http.NewRequestWithContext(WithIdentity(context.Background(), "alice", "script"), http.MethodPost, Origin+"/v1/chat/completions", strings.NewReader("private prompt"))
	r.Header.Set("Authorization", "Bearer private-key")
	response, err := transport.RoundTrip(r)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(response.Body)
	if string(raw) != "unchanged private generation" {
		t.Fatal("response changed")
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("no receipt %v", err)
	}
	receiptRaw, _ := os.ReadFile(filepath.Join(dir, entries[0].Name()))
	if strings.Contains(string(receiptRaw), "private") {
		t.Fatal("secret content persisted")
	}
	if err = m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
func TestTrackerStartupFlushesOldReceipt(t *testing.T) {
	db, m, _ := sqlmock.New()
	defer db.Close()
	tracker := &Tracker{Store: Store{DB: db}, ReceiptDir: filepath.Join(t.TempDir(), "r")}
	at := time.Now().UTC()
	hash := Fingerprint("key")
	if err := tracker.saveReceipt(receipt{Request: Request{"old", "alice", "image"}, CredentialHash: hash, Status: 200, At: at}); err != nil {
		t.Fatal(err)
	}
	m.ExpectExec("INSERT IGNORE INTO yzcm_usage_requests").WithArgs("old", "alice", "image", hash, at, 200).WillReturnResult(sqlmock.NewResult(0, 1))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tracker.Run(ctx)
	deadline := time.Now().Add(2 * time.Second)
	for {
		entries, _ := os.ReadDir(tracker.ReceiptDir)
		if len(entries) == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("startup did not flush old receipt")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err := m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
