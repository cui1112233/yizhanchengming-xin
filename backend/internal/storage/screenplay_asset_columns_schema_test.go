package storage

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"regexp"
	"testing"
)

func TestScreenplayAssetColumnsUpgradePreservesHealthyAndPartialSchemas(t *testing.T) {
	for _, test := range []struct {
		name         string
		existing     []string
		incompatible bool
	}{
		{"old database", nil, false}, {"healthy database", []string{"episodes", "tag", "extra"}, false},
		{"interrupted upgrade", []string{"episodes"}, false}, {"conflicting definition", []string{"episodes"}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			mock.ExpectBegin()
			rows := sqlmock.NewRows([]string{"column_name", "data_type", "character_maximum_length", "is_nullable"})
			found := map[string]bool{}
			for _, name := range test.existing {
				found[name] = true
				switch name {
				case "episodes":
					length := 191
					if test.incompatible {
						length = 10
					}
					rows.AddRow(name, "varchar", length, "NO")
				case "tag":
					rows.AddRow(name, "varchar", 64, "NO")
				case "extra":
					rows.AddRow(name, "json", nil, "YES")
				}
			}
			mock.ExpectQuery(regexp.QuoteMeta(screenplayAssetColumnsQuery)).WillReturnRows(rows)
			if !test.incompatible {
				for _, c := range []struct{ name, ddl string }{
					{"episodes", "ALTER TABLE screenplay_assets ADD COLUMN episodes VARCHAR(191) NOT NULL DEFAULT ''"},
					{"tag", "ALTER TABLE screenplay_assets ADD COLUMN tag VARCHAR(64) NOT NULL DEFAULT ''"},
					{"extra", "ALTER TABLE screenplay_assets ADD COLUMN extra JSON NULL"},
				} {
					if !found[c.name] {
						mock.ExpectExec(regexp.QuoteMeta(c.ddl)).WillReturnResult(sqlmock.NewResult(0, 0))
					}
				}
			}
			mock.ExpectRollback()
			tx, err := db.Begin()
			if err != nil {
				t.Fatal(err)
			}
			err = reconcileScreenplayAssetColumns(context.Background(), tx)
			if (err != nil) != test.incompatible {
				t.Fatalf("unexpected reconciliation error: %v", err)
			}
			tx.Rollback()
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
