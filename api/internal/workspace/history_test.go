package workspace

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"strings"
	"testing"
	"time"
)

// This checks emitted SQL boundaries. Result fixtures do not execute MySQL predicates.
func historyMatcher(elevated bool) sqlmock.QueryMatcher {
	return sqlmock.QueryMatcherFunc(func(expected, actual string) error {
		for _, part := range []string{"WITH visible_intakes AS", "JOIN visible_intakes vi ON vi.id=bp.intake_id", "b.intake_id=vp.intake_id", "sr.book_id=br.book_id", "linked_vpj.batch_project_id=smt.batch_project_id", "linked_vpj.book_id=smt.book_id", "smt.media_kind='video'", "normalized_history AS", "FROM filtered_history"} {
			if !strings.Contains(actual, part) {
				return fmt.Errorf("missing history isolation %q", part)
			}
		}
		if !elevated {
			for _, part := range []string{"auth_intake_ownership", "auth_batch_project_ownership", "io.owner_user_id=? OR (? > 0 AND io.team_id IS NOT NULL AND io.team_id=?)", "po.owner_user_id=? OR (? > 0 AND po.team_id IS NOT NULL AND po.team_id=?)"} {
				if !strings.Contains(actual, part) {
					return fmt.Errorf("missing owner scope %q", part)
				}
			}
		} else if strings.Contains(actual, "ownership") {
			return fmt.Errorf("elevated query must expose all existing facts")
		}
		if !strings.Contains(actual, expected) {
			return fmt.Errorf("missing operation %s", expected)
		}
		return nil
	})
}
func historyRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"kind", "origin", "source_id", "intake_id", "project_id", "book_id", "book_run_id", "attempt", "revision", "title", "project_name", "status", "source_status", "updated_at", "archived_at"})
}

func TestHistoryProjectionNormalizesUnionStatusCollation(t *testing.T) {
	for _, fragment := range []string{
		"vi.status COLLATE utf8mb4_unicode_ci AS raw_status",
		") COLLATE utf8mb4_unicode_ci AS raw_status",
		"br.status COLLATE utf8mb4_unicode_ci",
		"sr.status COLLATE utf8mb4_unicode_ci",
		"'saved' COLLATE utf8mb4_unicode_ci",
		"'measured' COLLATE utf8mb4_unicode_ci",
		"vm.status COLLATE utf8mb4_unicode_ci",
		"vv.status COLLATE utf8mb4_unicode_ci",
		"vma.status COLLATE utf8mb4_unicode_ci",
	} {
		if !strings.Contains(historyProjection, fragment) {
			t.Fatalf("history UNION status is not normalized with %q", fragment)
		}
	}
}

func TestHistoryOwnedFactsQueryBoundary(t *testing.T) {
	for _, tc := range []struct {
		name                string
		uid, team           int64
		elevated, hasIntake bool
	}{{"TeamID zero never shared", 8, 0, false, false}, {"only Intake owner", 7, 0, false, true}, {"same nonzero team", 8, 3, false, true}, {"cross user", 8, 9, false, false}, {"no ownership", 9, 0, false, false}, {"admin all facts", 7, 0, true, true}, {"owner all facts", 7, 0, true, true}, {"dev no elevation", 9, 0, false, false}} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(historyMatcher(tc.elevated)))
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			args := []driver.Value{}
			if !tc.elevated {
				args = []driver.Value{tc.uid, tc.team, tc.team, tc.uid, tc.team, tc.team}
			}
			n := 0
			rows := historyRows()
			if tc.hasIntake {
				n = 1
				rows.AddRow("intake", "intake", "42", 42, 0, 0, 0, 0, 0, "独立 Intake", "", "unknown", "unknown", time.Now(), nil)
			}
			mock.ExpectQuery("SELECT COUNT(*)").WithArgs(args...).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(n))
			mock.ExpectQuery("ORDER BY updated_at DESC,kind ASC,origin ASC,source_id DESC").WithArgs(append(args, 20, 0)...).WillReturnRows(rows)
			got, err := NewMySQLStore(db).ListHistory(context.Background(), HistoryQuery{UserID: tc.uid, TeamID: tc.team, Elevated: tc.elevated, Page: 1, Limit: 20, Archived: "all"})
			if err != nil {
				t.Fatal(err)
			}
			if got.Total != n || len(got.Entries) != n || got.Entries == nil {
				t.Fatalf("%+v", got)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestHistorySafeDTOAndFixedNavigation(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(historyMatcher(false)))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now()
	args := []driver.Value{int64(7), int64(3), int64(3), int64(7), int64(3), int64(3), "%hero%", "%hero%", "%hero%", "%hero%", "%hero%", "script", "completed"}
	mock.ExpectQuery("SELECT COUNT(*)").WithArgs(args...).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectQuery("archived_at IS NOT NULL").WithArgs(append(args, 2, 2)...).WillReturnRows(historyRows().AddRow("batch", "project", "12", 42, 12, 0, 0, 0, 0, "项目", "项目", "completed", "succeeded", now, now).AddRow("script", "stage", "18", 42, 12, 3, 9, 2, 0, "小说", "项目", "failed", "failed", now, now))
	got, err := NewMySQLStore(db).ListHistory(context.Background(), HistoryQuery{UserID: 7, TeamID: 3, Page: 2, Limit: 2, Q: "hero", Kind: "script", Status: "completed", Archived: "archived"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Total != 3 || got.Entries[0].ID != "batch:project:12" || got.Entries[0].Href != "/batch-factory?projectId=12" || got.Entries[1].Href != "" || got.Entries[1].CanPreview || got.Entries[1].Attempt != 2 || got.Entries[1].ErrorCode != "HISTORY_SOURCE_FAILED" {
		t.Fatalf("%+v", got)
	}
	raw, _ := json.Marshal(got)
	var decoded map[string]any
	_ = json.Unmarshal(raw, &decoded)
	entry := decoded["entries"].([]any)[0].(map[string]any)
	if len(entry) != 19 {
		t.Fatalf("safe DTO fields=%d body=%s", len(entry), raw)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
func TestHistoryStatusWhitelist(t *testing.T) {
	for _, tc := range []struct{ in, want string }{{"succeeded", "completed"}, {"completed", "completed"}, {"pending_executor", "pending_executor"}, {"partial_failed", "partial_failed"}, {"retryable_failed", "retryable_failed"}, {"failed", "failed"}, {"", "unknown"}, {"provider says success secret=xyz", "unknown"}} {
		if got := normalizeHistoryStatus(tc.in); got != tc.want {
			t.Fatalf("%q => %q want %q", tc.in, got, tc.want)
		}
	}
}
