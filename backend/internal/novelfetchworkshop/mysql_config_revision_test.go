package novelfetchworkshop

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLConfigRevisionLocksChecksAndWritesUnicode(t *testing.T) {
	for _, stale := range []bool{false, true} {
		t.Run(map[bool]string{false: "fresh", true: "stale"}[stale], func(t *testing.T) {
			_, mock, s := newBodySQLMock(t)
			mock.ExpectBegin()
			mock.ExpectExec("INSERT IGNORE INTO novel_fetch_workshop_configs").WithArgs("alice").WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectQuery(regexp.QuoteMeta("SELECT settings_json FROM novel_fetch_workshop_configs WHERE owner_username=? FOR UPDATE")).WithArgs("alice").WillReturnRows(sqlmock.NewRows([]string{"settings_json"}).AddRow(`{"configRevision":"4","label":"旧配置"}`))
			revision := "4"
			if stale {
				revision = "3"
				mock.ExpectRollback()
			} else {
				mock.ExpectExec(regexp.QuoteMeta("UPDATE novel_fetch_workshop_configs SET settings_json=CAST(? AS CHAR CHARACTER SET utf8mb4), updated_at=CURRENT_TIMESTAMP(6) WHERE owner_username=?")).WithArgs(`{"configRevision":"5","label":"中文规则"}`, "alice").WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			}
			err := s.PutConfig(context.Background(), "alice", map[string]any{"configRevision": revision, "label": "中文规则"})
			if stale && !errors.Is(err, ErrConfigConflict) || !stale && err != nil {
				t.Fatal(err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
