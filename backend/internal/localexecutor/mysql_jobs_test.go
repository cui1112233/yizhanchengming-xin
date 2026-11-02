package localexecutor

import (
	"strings"
	"testing"
)

func TestClaimJobCandidateQuerySortsOnlyTheJobID(t *testing.T) {
	query := strings.ToLower(strings.Join(strings.Fields(claimJobCandidateQuery), " "))
	selectClause, _, found := strings.Cut(query, " from ")
	if !found {
		t.Fatalf("claim candidate query has no FROM clause: %q", query)
	}
	if selectClause != "select id" {
		t.Fatalf("claim candidate query must not sort payload_json: %q", selectClause)
	}
	if !strings.Contains(query, "for update skip locked") {
		t.Fatalf("claim candidate query must keep row locking: %q", query)
	}
}
