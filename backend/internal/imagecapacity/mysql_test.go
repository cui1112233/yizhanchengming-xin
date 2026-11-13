package imagecapacity

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
)

func TestMySQLUnknownProfileUsesSharedAccountCapAndDefaultFour(t *testing.T) {
	db, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	q := Request{Endpoint: "https://provider/v1", Credential: "never-store-this", Model: "unverified", Size: "other", References: true}
	m.ExpectQuery("SELECT group_key, account_limit").WithArgs(AccountKey(q)).WillReturnRows(sqlmock.NewRows([]string{"group_key", "account_limit"}).AddRow("shared", 20))
	m.ExpectQuery("SELECT MIN").WithArgs("shared").WillReturnRows(sqlmock.NewRows([]string{"min"}).AddRow(16))
	m.ExpectQuery("SELECT MIN\\(COALESCE").WithArgs(ProfileKey(q), "shared").WillReturnRows(sqlmock.NewRows([]string{"model", "initial", "verified"}).AddRow(4, 4, false))
	m.ExpectQuery("SELECT COALESCE\\(MAX").WithArgs("shared", modelIdentity(q.Model)).WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(4))
	p, e := NewMySQLRepository(db).Resolve(context.Background(), q)
	if e != nil || p.AccountLimit != 16 || p.ModelLimit != 4 || p.InitialLimit != 4 || p.Verified {
		t.Fatalf("unknown policy %+v %v", p, e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
func TestMySQLSaveUsesTransactionAndExactProfile(t *testing.T) {
	db, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	q := Request{Endpoint: "https://provider/v1", Credential: "secret", Model: "verified", Size: "1024x1536"}
	p := defaultPolicy(q)
	p.AccountLimit = 20
	p.ModelLimit = 20
	p.InitialLimit = 12
	p.Verified = true
	m.ExpectBegin()
	m.ExpectExec("INSERT INTO image_capacity_accounts").WithArgs(p.AccountKey, "", 20).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectExec("INSERT INTO image_capacity_policies").WithArgs(p.AccountKey, p.ProfileKey, p.Model, p.Size, p.Quality, p.Resolution, false, 1, 20, 12, true).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectCommit()
	if e = NewMySQLRepository(db).Save(context.Background(), p); e != nil {
		t.Fatal(e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
	p.Size = "different"
	if e = NewMySQLRepository(db).Save(context.Background(), p); e == nil {
		t.Fatal("mismatched request profile accepted")
	}
}
func TestMySQLFailureDoesNotFallbackToOpen(t *testing.T) {
	db, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	m.ExpectQuery("SELECT group_key").WillReturnError(sql.ErrConnDone)
	if _, e = NewMySQLRepository(db).Resolve(context.Background(), Request{}); e == nil {
		t.Fatal("database error hidden")
	}
}

func TestMySQLExactCountProfileNeverInheritsSingleImageVerification(t *testing.T) {
	db, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	repo := NewMySQLRepository(db)
	q := Request{Endpoint: "https://provider/v1", Credential: "same", Model: "verified", Size: "1024x1536", Count: 1}
	for _, count := range []int{1, 4} {
		q.Count = count
		m.ExpectQuery("SELECT group_key, account_limit").WithArgs(AccountKey(q)).WillReturnRows(sqlmock.NewRows([]string{"group", "cap"}).AddRow("", 20))
		query := m.ExpectQuery("SELECT model_limit").WithArgs(AccountKey(q), ProfileKey(q))
		if count == 1 {
			query.WillReturnRows(sqlmock.NewRows([]string{"model", "initial", "verified"}).AddRow(20, 12, true))
		} else {
			query.WillReturnError(sql.ErrNoRows)
		}
		m.ExpectQuery("SELECT COALESCE\\(MAX").WithArgs(AccountKey(q), modelIdentity(q.Model)).WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(20))
		p, e := repo.Resolve(context.Background(), q)
		if e != nil {
			t.Fatal(e)
		}
		if count == 1 && (!p.Verified || p.ModelLimit != 20) || count == 4 && (p.Verified || p.ModelLimit != 4 || p.Count != 4) {
			t.Fatalf("count%d inherited wrong policy %+v", count, p)
		}
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}

func TestEnsureSchemaMigratesLegacyCountColumn(t *testing.T) {
	db, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	for range SchemaStatements() {
		m.ExpectExec("CREATE TABLE IF NOT EXISTS").WillReturnResult(sqlmock.NewResult(0, 0))
	}
	m.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	m.ExpectExec("ALTER TABLE image_capacity_policies ADD COLUMN item_count").WillReturnResult(sqlmock.NewResult(0, 0))
	if e = NewMySQLRepository(db).EnsureSchema(context.Background()); e != nil {
		t.Fatal(e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
