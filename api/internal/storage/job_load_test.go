package storage

import (
	"context"
	"database/sql"
	"io"
	"testing"
	"time"
)

type fakeRows struct {
	values [][]any
	index int
}
func (f *fakeRows) Next() bool { return f.index < len(f.values) }
func (f *fakeRows) Scan(dest ...any) error {
	row := f.values[f.index]
	f.index++
	for i := range dest {
		switch ptr := dest[i].(type) {
		case *string:
			*ptr = row[i].(string)
		case *sql.NullString:
			if row[i] == nil { *ptr = sql.NullString{} } else { *ptr = sql.NullString{String: row[i].(string), Valid:true} }
		case *sql.NullTime:
			if row[i] == nil { *ptr = sql.NullTime{} } else { *ptr = sql.NullTime{Time: row[i].(time.Time), Valid:true} }
		case *int:
			*ptr = row[i].(int)
		default:
			return io.ErrUnexpectedEOF
		}
	}
	return nil
}
func (f *fakeRows) Close() error { return nil }
func (f *fakeRows) Err() error { return nil }

type fakeQueryer struct { calls int; runAt time.Time }
func (f *fakeQueryer) QueryContext(_ context.Context, query string, args ...any) (jobRows, error) {
	f.calls++
	if f.calls == 1 {
		return &fakeRows{values:[][]any{{"job-1","intake-1",nil,"queued",f.runAt}}}, nil
	}
	return &fakeRows{values:[][]any{{"fetch_book",1},{"resolve_metadata",2},{"create_batch",3}}}, nil
}

func TestSQLJobStoreLoadsIntakeJobAndOrderedStages(t *testing.T) {
	runAt := time.Date(2026,10,4,9,0,0,0,time.UTC)
	queryer := &fakeQueryer{runAt:runAt}
	store := newSQLJobStoreWithQuery(queryer.QueryContext)
	job, err := store.LoadJob(context.Background(), "job-1")
	if err != nil { t.Fatal(err) }
	if job.ID != "job-1" || job.IntakeID != "intake-1" || job.BatchID != "" || !job.RunAt.Equal(runAt) { t.Fatalf("job=%#v", job) }
	if len(job.Stages) != 3 || job.Stages[0] != "fetch_book" || job.Stages[2] != "create_batch" { t.Fatalf("stages=%#v", job.Stages) }
}
