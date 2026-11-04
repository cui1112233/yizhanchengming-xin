package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchYFAIBalanceParsesDollarStrings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/skills/balance" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer stub-key" {
			t.Fatalf("missing bearer auth")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"balance":"$12.345678","available_balance":"$9.876543","frozen":"$2.469135","currency":"USD"}`))
	}))
	defer server.Close()
	got := fetchYFAIBalance(context.Background(), server.URL, "stub-key")
	if got.Status != "available" || got.Available != 9.876543 || got.Total != 12.345678 {
		t.Fatalf("balance=%+v", got)
	}
	if got.Percent < 79.9 || got.Percent > 80.1 {
		t.Fatalf("percent=%v", got.Percent)
	}
}

func TestFetchYFAIBalanceReportsDepletedAtZero(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"balance":"$0.000000","available_balance":"$0.000000"}`))
	}))
	defer server.Close()
	if got := fetchYFAIBalance(context.Background(), server.URL, "stub-key"); got.Status != "depleted" {
		t.Fatalf("balance=%+v", got)
	}
}

func TestFetchYFAIBalanceMissingFieldsIsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"currency":"USD"}`))
	}))
	defer server.Close()
	if got := fetchYFAIBalance(context.Background(), server.URL, "stub-key"); got.Status != "error" {
		t.Fatalf("balance=%+v", got)
	}
}

func TestFetchYFAIBalanceUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid key"}`))
	}))
	defer server.Close()
	if got := fetchYFAIBalance(context.Background(), server.URL, "stub-key"); got.Status != "unauthorized" {
		t.Fatalf("balance=%+v", got)
	}
}

func TestFetchYFAIBalanceTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(120 * time.Millisecond):
		}
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	if got := fetchYFAIBalance(ctx, server.URL, "stub-key"); got.Status != "timeout" {
		t.Fatalf("balance=%+v", got)
	}
}

func TestFetchYFAIBalanceWithoutCredentialIsNotConfigured(t *testing.T) {
	if got := fetchYFAIBalance(context.Background(), "https://yf.token6688.com", ""); got.Status != "not_configured" {
		t.Fatalf("balance=%+v", got)
	}
}

func TestYFAIUSDAmountParsesGatewayForms(t *testing.T) {
	cases := []struct {
		name  string
		value any
		want  float64
		ok    bool
	}{
		{name: "dollar string", value: "$5.407224", want: 5.407224, ok: true},
		{name: "thousands separator", value: "$1,234.500000", want: 1234.5, ok: true},
		{name: "plain number", value: 3.5, want: 3.5, ok: true},
		{name: "empty", value: "", ok: false},
		{name: "nil", value: nil, ok: false},
		{name: "garbage", value: "n/a", ok: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := yfaiUSDAmount(tc.value)
			if ok != tc.ok || (ok && got != tc.want) {
				t.Fatalf("got=%v ok=%v", got, ok)
			}
		})
	}
}
