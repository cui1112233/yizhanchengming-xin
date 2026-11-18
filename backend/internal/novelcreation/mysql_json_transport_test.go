package novelcreation

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/go-sql-driver/mysql"
	"os"
	"testing"
)

// Read-only probe, including on an application database: SELECT only, no schema
// or data changes. This reproduces the production interpolateParams setting.
func TestMySQLJSONTransportReadOnly(t *testing.T) {
	raw := os.Getenv("QIANTIE_TEST_NOVEL_JSON_DSN")
	if raw == "" {
		t.Skip("read-only live DSN not supplied")
	}
	for _, interpolate := range []bool{false, true} {
		c, e := mysql.ParseDSN(raw)
		if e != nil {
			t.Fatal("invalid probe DSN")
		}
		c.InterpolateParams = interpolate
		db, e := sql.Open("mysql", c.FormatDSN())
		if e != nil {
			t.Fatal("cannot open probe")
		}
		defer db.Close()
		d := sqlDatabase{db}
		tx, e := d.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})
		if e != nil {
			t.Fatal(e)
		}
		defer tx.Rollback()
		for _, executor := range []sqlExecutor{d, tx} {
			for _, arg := range []any{[]byte(`{"text":"雨夜"}`), json.RawMessage(`{"text":"雨夜"}`)} {
				var kind string
				e = executor.QueryRowContext(context.Background(), "SELECT JSON_TYPE(CAST(? AS JSON))", arg).Scan(&kind)
				if e != nil {
					t.Errorf("interpolate=%v JSON transport: %v", interpolate, e)
					continue
				}
				if kind != "OBJECT" {
					t.Errorf("interpolate=%v kind=%s", interpolate, kind)
				}
				rows, err := executor.QueryContext(context.Background(), "SELECT JSON_TYPE(CAST(? AS JSON))", arg)
				if err != nil {
					t.Fatal(err)
				}
				if !rows.Next() {
					t.Fatal("probe has no row")
				}
				if err = rows.Scan(&kind); err != nil {
					t.Fatal(err)
				}
				rows.Close()
				if kind != "OBJECT" {
					t.Fatal("row query changed JSON type")
				}
			}
		}
	}
}

func TestNovelSQLArgsPreserveCallerAndBinary(t *testing.T) {
	raw := json.RawMessage(`{"a":1}`)
	binary := []byte{0, 255}
	args := []any{raw, []byte(`[]`), binary, nil, "text"}
	converted := novelSQLArgs(args)
	if converted[0] != `{"a":1}` || converted[1] != "[]" {
		t.Fatal("JSON not bound as text")
	}
	if _, ok := args[0].(json.RawMessage); !ok {
		t.Fatal("frozen input mutated")
	}
	if got, ok := converted[2].([]byte); !ok || len(got) != 2 || got[1] != 255 {
		t.Fatal("non-JSON binary changed")
	}
	if converted[3] != nil || converted[4] != "text" {
		t.Fatal("unrelated arguments changed")
	}
}
