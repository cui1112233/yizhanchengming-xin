package publishing

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLBookBelongsToBatchProjectUsesSharedIntakeRelationship(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	query := `SELECT EXISTS(SELECT 1 FROM batch_projects p JOIN books b ON b.intake_id = p.intake_id WHERE p.id = ? AND b.id = ?)`
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(int64(21), int64(22)).
		WillReturnRows(sqlmock.NewRows([]string{"allowed"}).AddRow(false))

	allowed, err := NewMySQLStore(db).BookBelongsToBatchProject(context.Background(), 21, 22)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("book from a different BatchProject intake was accepted")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
