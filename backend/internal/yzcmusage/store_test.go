package yzcmusage

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestTrackerPersistsOnlyAuthenticatedAttributionAndPreservesBody(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectExec("INSERT IGNORE INTO yzcm_usage_requests").WithArgs("req-1", "alice", "image", Fingerprint("private-key"), sqlmock.AnyArg(), 200).WillReturnResult(sqlmock.NewResult(0, 1))
	tracker := &Tracker{Store: Store{DB: db}}
	transport := tracker.Wrap(transportFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{"X-Oneapi-Request-Id": []string{"req-1"}}, Body: io.NopCloser(strings.NewReader("unchanged result"))}, nil
	}))
	req, _ := http.NewRequestWithContext(WithIdentity(context.Background(), "alice", "script"), "POST", Origin+"/v1/images/generations", strings.NewReader(`{"model":"image"}`))
	req.Header.Set("Authorization", "Bearer private-key")
	response, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)
	if string(body) != "unchanged result" {
		t.Fatal("generation response changed")
	}
	// An unscoped background call must not be assigned to a key's account owner.
	req = req.WithContext(context.Background())
	if _, err = transport.RoundTrip(req); err != nil {
		t.Fatal(err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPendingReceiptIsNotReportedAsFree(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for i := 0; i < 2; i++ {
		mock.ExpectQuery("SELECT COUNT").WithArgs("alice", sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"calls", "priced", "success", "input", "output", "cost"}).AddRow(1, 0, 1, 0, 0, nil))
	}
	mock.ExpectQuery("SELECT feature,COUNT").WithArgs("alice", sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"feature", "calls", "tokens"}).AddRow("image", 1, 0))
	mock.ExpectQuery("SELECT request_id,created_at").WithArgs("alice").WillReturnRows(sqlmock.NewRows([]string{"id", "created", "feature", "model", "input", "output", "cost", "currency", "status"}))
	result, err := (Store{DB: db}).ReadSummary(context.Background(), "alice")
	if err != nil {
		t.Fatal(err)
	}
	month := result["month"].(map[string]any)
	if month["actualCost"] != nil || month["pendingCalls"] != int64(1) || month["pricedCalls"] != int64(0) {
		t.Fatal(month)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
