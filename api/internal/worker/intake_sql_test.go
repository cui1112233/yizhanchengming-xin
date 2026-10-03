package worker

import (
    "context"
    "database/sql"
    "io"
    "strings"
    "testing"

    "github.com/cui1112233/yizhanchengming-xin/api/internal/novel"
)

type intakeRows struct { rows [][]any; idx int }
func (r *intakeRows) Next() bool { return r.idx < len(r.rows) }
func (r *intakeRows) Scan(dest ...any) error {
    row := r.rows[r.idx]; r.idx++
    for i := range dest {
        switch p := dest[i].(type) {
        case *int64: *p = row[i].(int64)
        case *string: *p = row[i].(string)
        case *sql.NullInt64:
            if row[i] == nil { *p = sql.NullInt64{} } else { *p = sql.NullInt64{Int64:row[i].(int64), Valid:true} }
        default: return io.ErrUnexpectedEOF
        }
    }
    return nil
}
func (r *intakeRows) Close() error { return nil }
func (r *intakeRows) Err() error { return nil }

type intakeDBFake struct { queries []string; execs []string; args [][]any }
func (f *intakeDBFake) QueryContext(context.Context, string, ...any) (intakeBookRows, error) {
    return &intakeRows{rows:[][]any{{int64(7),"book-7","2","male","男生生活",int64(8),"都市","正文"}}}, nil
}
func (f *intakeDBFake) ExecContext(_ context.Context, q string, args ...any) (sql.Result, error) {
    f.execs = append(f.execs, q); f.args = append(f.args, append([]any(nil), args...)); return fakeSQLResult(1), nil
}

type fakeSQLResult int64
func (f fakeSQLResult) LastInsertId() (int64,error) { return int64(f),nil }
func (f fakeSQLResult) RowsAffected() (int64,error) { return int64(f),nil }

func TestSQLIntakeRepositoryLoadsAndPersistsBookData(t *testing.T) {
    db := &intakeDBFake{}
    repo := newSQLIntakeRepository(db.QueryContext, db.ExecContext)
    books, err := repo.ListBooks(context.Background(), "intake-1")
    if err != nil { t.Fatal(err) }
    if len(books)!=1 || books[0].ID!=7 || books[0].BookID!="book-7" || books[0].ManualGender!=novel.GenderMale || books[0].Genre!=8 || books[0].SourceText!="正文" { t.Fatalf("books=%#v", books) }

    if err := repo.SaveFetched(context.Background(), 7, FetchedBook{BookName:"书名", Author:"作者", SourceText:"正文", Category:"男生生活", Genre:8}); err != nil { t.Fatal(err) }
    if err := repo.SaveResolved(context.Background(), 7, ResolvedMetadata{Gender:novel.GenderResult{Gender:novel.GenderMale, Source:novel.GenderSource121Category}, Style:"都市", NeedsAI:false}); err != nil { t.Fatal(err) }
    if len(db.execs)!=2 { t.Fatalf("execs=%d", len(db.execs)) }
    if !strings.Contains(db.execs[0], "source_text") || !strings.Contains(db.execs[1], "resolved_gender") { t.Fatalf("queries=%v", db.execs) }
}
