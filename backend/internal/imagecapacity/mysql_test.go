package imagecapacity

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLResolveUsesAutomaticConcurrency(t *testing.T) {
	db, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	q := Request{Endpoint: "https://provider/v1", Credential: "never-store-this", Model: "unverified", Size: "other", References: true}
	p, e := NewMySQLRepository(db).Resolve(context.Background(), q)
	if e != nil || p.AccountLimit != AutoAccountLimit || p.ModelLimit != AutoModelLimit || p.InitialLimit != AutoInitialLimit || p.Verified {
		t.Fatalf("auto policy %+v %v", p, e)
	}
	if !strings.HasPrefix(p.GroupKey, "auto:") {
		t.Fatalf("group key %q should be auto-derived", p.GroupKey)
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
	m.ExpectExec("INSERT INTO image_capacity_accounts").WithArgs(p.AccountKey, p.GroupKey, 20).WillReturnResult(sqlmock.NewResult(0, 1))
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
func TestMySQLResolveIgnoresLegacyManualLimits(t *testing.T) {
	db, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	// 历史手工配置是 4 路；自动并发不应再被这个旧值压回去。
	m.ExpectQuery("SELECT model_limit").WillReturnRows(sqlmock.NewRows([]string{"model", "initial"}).AddRow(4, 4))
	p, e := NewMySQLRepository(db).Resolve(context.Background(), Request{Endpoint: "https://provider/v1", Credential: "same"})
	if e != nil {
		t.Fatal(e)
	}
	if p.ModelLimit != AutoModelLimit || p.InitialLimit != AutoInitialLimit {
		t.Fatalf("legacy limits leaked into auto policy: %+v", p)
	}
	if e = m.ExpectationsWereMet(); e == nil {
		t.Fatal("manual policy table should not be queried any more")
	}
}

func TestMySQLResolveIsCountIndependent(t *testing.T) {
	db, _, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	repo := NewMySQLRepository(db)
	for _, count := range []int{1, 4} {
		p, e := repo.Resolve(context.Background(), Request{Endpoint: "https://provider/v1", Credential: "same", Model: "verified", Size: "1024x1536", Count: count})
		if e != nil {
			t.Fatal(e)
		}
		if p.ModelLimit != AutoModelLimit || p.InitialLimit != AutoInitialLimit || p.Count != count {
			t.Fatalf("count%d resolved wrong policy %+v", count, p)
		}
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
