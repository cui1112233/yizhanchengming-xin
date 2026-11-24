package yzcmusage

import (
	"context"
	"io"
	"math"
	"net/http"
	"strings"
	"testing"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestMatchUsesDeductedQuotaAndExactOwner(t *testing.T) {
	requests := []Request{{"a", "alice", "image"}, {"b", "bob", "image"}}
	logs := []BillingLog{{RequestID: "a", Quota: 3356, InputTokens: 12, Type: 2}, {RequestID: "b", Quota: 9000, Type: 2}, {RequestID: "unknown", Quota: 9000, Type: 2}}
	entries, err := Match("alice", requests, logs, Rates{500000, 7.3})
	if err != nil || len(entries) != 1 {
		t.Fatal(entries, err)
	}
	if math.Abs(entries[0].Cost-.0489976) > 1e-12 || entries[0].TotalTokens != 12 || entries[0].Source != "yzcm_billing_log" {
		t.Fatal(entries)
	}
	// Sharing the key cannot make Bob's charge part of Alice's usage.
	requests = append(requests, Request{"a", "bob", "image"})
	entries, err = Match("alice", requests, logs, Rates{500000, 7.3})
	if err != nil || len(entries) != 0 {
		t.Fatal(entries, err)
	}
}
func TestZeroChargeIsKnownAndInvalidRatesFail(t *testing.T) {
	entries, err := Match("alice", []Request{{"a", "alice", "text"}}, []BillingLog{{RequestID: "a", Type: 2}}, Rates{500000, 7.3})
	if err != nil || len(entries) != 1 || !entries[0].CostKnown || entries[0].Cost != 0 {
		t.Fatal(entries, err)
	}
	for _, r := range []Rates{{0, 7.3}, {500000, math.Inf(1)}, {math.NaN(), 7.3}} {
		if _, err = Match("alice", nil, nil, r); err == nil {
			t.Fatal("invalid conversion accepted")
		}
	}
}
func TestLogLookupNeverLeaksCredentialOrUnmatchedRows(t *testing.T) {
	c := Client{HTTP: &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		if strings.Contains(r.URL.String(), "private-key") || r.Header.Get("Authorization") != "Bearer private-key" {
			t.Fatal("unsafe credential transport")
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"success":true,"data":[{"request_id":"a","quota":3356,"type":2},{"request_id":"other","quota":9000,"type":2}]}`))}, nil
	})}}
	logs, err := c.Logs(context.Background(), "private-key", []Request{{"a", "alice", "image"}})
	if err != nil || len(logs) != 1 || logs[0].RequestID != "a" {
		t.Fatal(logs, err)
	}
}

func TestProviderCredentialOwnerCannotOverrideAuthenticatedMember(t *testing.T) {
	ctx := WithIdentity(context.Background(), "alice", "script")
	ctx = WithIdentity(ctx, "manager", "")
	identity := ctx.Value(identityKey{}).(identity)
	if identity.Owner != "alice" || identity.Feature != "script" {
		t.Fatal(identity)
	}
}
