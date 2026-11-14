package storage

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/go-sql-driver/mysql"
)

func TestV11SliceResumeValidatesBeforeCompletingPartialDDL(t *testing.T) {
	for _, conflict := range []bool{false, true} {
		t.Run(map[bool]string{false: "partial", true: "conflicting_key"}[conflict], func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			mock.ExpectBegin()
			tx, _ := db.Begin()
			first := ""
			for _, statement := range V11SliceOneStatements() {
				parts := v11SliceForeignKey.FindStringSubmatch(statement)
				if parts == nil {
					continue
				}
				rows := sqlmock.NewRows([]string{"table", "column", "target", "target_column", "delete_rule", "update_rule"})
				if first == "" {
					first = statement
					table := parts[1]
					if conflict {
						table = "batch_factory_v12_books"
					}
					rows.AddRow(table, parts[3], parts[4], parts[5], "RESTRICT", "RESTRICT")
				}
				mock.ExpectQuery(sliceForeignKeyQuery).WithArgs(parts[2]).WillReturnRows(rows)
				if conflict {
					break
				}
			}
			if !conflict {
				for _, statement := range V11SliceOneStatements() {
					if statement != first {
						mock.ExpectExec(statement).WillReturnResult(sqlmock.NewResult(0, 0))
					}
				}
			}
			mock.ExpectRollback()
			adopted, err := adoptV11SliceOneLedger(context.Background(), tx)
			if conflict {
				if err == nil || adopted || !strings.Contains(err.Error(), "does not match") {
					t.Fatalf("adopted=%v err=%v", adopted, err)
				}
			} else if err != nil || !adopted {
				t.Fatalf("adopted=%v err=%v", adopted, err)
			}
			_ = tx.Rollback()
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestV12CanonicalResumePlansAllRenamesBeforeWriting(t *testing.T) {
	for _, mode := range []string{"partial", "completed", "conflict", "missing"} {
		t.Run(mode, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			mock.ExpectBegin()
			tx, _ := db.Begin()
			query := `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=?`
			pending := []string{}
			for i, statement := range V12CanonicalTableRenameStatements() {
				parts := strings.Fields(statement)
				source, target := 0, 1
				if mode == "partial" && i%2 == 1 {
					source, target = 1, 0
					pending = append(pending, parts[2]+" TO "+parts[4])
				}
				if mode == "conflict" {
					source, target = 1, 1
				}
				if mode == "missing" {
					source, target = 0, 0
				}
				mock.ExpectQuery(query).WithArgs(parts[2]).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(source))
				mock.ExpectQuery(query).WithArgs(parts[4]).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(target))
				if mode == "conflict" || mode == "missing" {
					break
				}
			}
			if len(pending) > 0 {
				mock.ExpectExec("RENAME TABLE " + strings.Join(pending, ", ")).WillReturnResult(sqlmock.NewResult(0, 0))
			}
			mock.ExpectRollback()
			adopted, err := adoptV12CanonicalTableRename(context.Background(), tx)
			wantSuccess := mode == "partial" || mode == "completed"
			if adopted != wantSuccess || (err == nil) != wantSuccess {
				t.Fatalf("adopted=%v err=%v", adopted, err)
			}
			_ = tx.Rollback()
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// The caller must supply an empty disposable database. This test keeps all
// rows, loses a migration checkpoint deliberately, and repeats the startup.
func TestMySQLMigrationResumePreservesRowsAfterInterruptedDDL(t *testing.T) {
	dsn := os.Getenv("QIANTIE_TEST_MIGRATION_RESUME_DSN")
	if dsn == "" {
		t.Skip("requires an empty disposable MySQL database")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	t.Log("creating foundation and interrupted slice migration")
	if err := RunMigrations(ctx, db, V11FoundationMigrations()); err != nil {
		t.Fatal(err)
	}
	for _, statement := range V11SliceOneStatements() {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			t.Fatal(err)
		}
		if v11SliceForeignKey.MatchString(statement) {
			break
		}
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO batch_factory_v11_batches(id,owner_username,created_at,updated_at) VALUES ('preserved-batch','resume-owner',NOW(6),NOW(6))`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO batch_factory_v11_books(id,batch_id,owner_username,ordinal,created_at,updated_at) VALUES ('preserved-book','preserved-batch','resume-owner',1,NOW(6),NOW(6))`); err != nil {
		t.Fatal(err)
	}
	slice := V11Migrations()[1]
	t.Log("recovering interrupted foreign keys")
	for i := 0; i < 2; i++ {
		if err := RunMigrations(ctx, db, []Migration{slice}); err != nil {
			t.Fatalf("slice restart %d: %v", i, err)
		}
	}
	for _, statement := range V12CanonicalTableRenameStatements() {
		parts := strings.Fields(statement)
		if _, err := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS "+parts[2]+" (id INT PRIMARY KEY)"); err != nil {
			t.Fatal(err)
		}
	}
	for _, statement := range V12CanonicalTableRenameStatements()[:2] {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}
	canonical := V12Migrations()[4]
	t.Log("recovering interrupted canonical table rename")
	for i := 0; i < 2; i++ {
		if err := RunMigrations(ctx, db, []Migration{canonical}); err != nil {
			t.Fatalf("canonical restart %d: %v", i, err)
		}
	}
	var owner string
	if err := db.QueryRowContext(ctx, `SELECT owner_username FROM batch_factory_v12_books WHERE id='preserved-book'`).Scan(&owner); err != nil || owner != "resume-owner" {
		t.Fatalf("preserved owner=%q err=%v", owner, err)
	}
	var legacyTables int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name LIKE 'batch_factory_v11_%'`).Scan(&legacyTables); err != nil || legacyTables != 0 {
		t.Fatalf("legacy tables=%d err=%v", legacyTables, err)
	}
}
