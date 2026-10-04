package media

import (
	"database/sql"
	"testing"
)

func TestNewSQLStoreKeepsDatabaseForReadsAndWrites(t *testing.T) {
	db := &sql.DB{}
	store := NewSQLStore(db)
	if store.db != db {
		t.Fatal("NewSQLStore must retain db for Get queries")
	}
	if store.exec != db {
		t.Fatal("NewSQLStore must use the same db for Create writes")
	}
}
