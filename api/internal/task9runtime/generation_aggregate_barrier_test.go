package task9runtime

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"sync"
	"testing"
	"time"
)

func TestGenerationAggregateBarrierCoversPreparedFallback(t *testing.T) {
	for _, contextStatement := range []bool{false, true} {
		t.Run(map[bool]string{false: "Query", true: "QueryContext"}[contextStatement], func(t *testing.T) {
			barrier := &aggregateBarrierConnector{inner: &aggregateFallbackConnector{contextStatement: contextStatement}, read: make(chan struct{}), release: make(chan struct{})}
			db := sql.OpenDB(barrier)
			defer db.Close()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			defer close(barrier.release)
			type result struct {
				value string
				err   error
			}
			done := make(chan result, 1)
			go func() {
				var value string
				err := db.QueryRowContext(ctx, "SELECT br.status FROM book_runs br WHERE br.run_id=?", int64(11)).Scan(&value)
				done <- result{value, err}
			}()
			select {
			case <-barrier.read:
			case result := <-done:
				t.Fatalf("prepared query bypassed barrier: %+v", result)
			case <-ctx.Done():
				t.Fatal("prepared query did not reach barrier")
			}
			// Cancellation must unblock prepared rows and preserve the context error.
			cancel()
			select {
			case result := <-done:
				if !errors.Is(result.err, context.Canceled) {
					t.Fatalf("cancellation=%v", result.err)
				}
			case <-time.After(time.Second):
				t.Fatal("prepared query ignored cancellation")
			}
		})
	}
}

func TestGenerationAggregateBarrierPreparedReleasePreservesRows(t *testing.T) {
	for _, contextStatement := range []bool{false, true} {
		t.Run(map[bool]string{false: "Query", true: "QueryContext"}[contextStatement], func(t *testing.T) {
			barrier := &aggregateBarrierConnector{inner: &aggregateFallbackConnector{contextStatement: contextStatement}, read: make(chan struct{}), release: make(chan struct{})}
			db := sql.OpenDB(barrier)
			defer db.Close()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			var releaseOnce sync.Once
			release := func() { releaseOnce.Do(func() { close(barrier.release) }) }
			defer release()
			type result struct {
				value string
				err   error
			}
			done := make(chan result, 1)
			go func() {
				var value string
				err := db.QueryRowContext(ctx, "SELECT br.status FROM book_runs br WHERE br.run_id=?", int64(11)).Scan(&value)
				done <- result{value, err}
			}()
			select {
			case <-barrier.read:
			case result := <-done:
				t.Fatalf("prepared query bypassed barrier: %+v", result)
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
			release()
			select {
			case result := <-done:
				if result.err != nil || result.value != "failed" {
					t.Fatalf("result=%+v", result)
				}
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
		})
	}
}

func TestGenerationAggregateBarrierPreparedErrorsBypassReadBarrier(t *testing.T) {
	for _, contextStatement := range []bool{false, true} {
		t.Run(map[bool]string{false: "Query", true: "QueryContext"}[contextStatement], func(t *testing.T) {
			barrier := &aggregateBarrierConnector{inner: &aggregateFallbackConnector{contextStatement: contextStatement}, read: make(chan struct{}), release: make(chan struct{})}
			db := sql.OpenDB(barrier)
			defer db.Close()
			defer close(barrier.release)
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()
			var status string
			err := db.QueryRowContext(ctx, "SELECT br.status FROM book_runs br WHERE br.run_id=?", int64(12)).Scan(&status)
			if !errors.Is(err, errAggregateFallbackArguments) {
				t.Fatalf("underlying error was lost: %v", err)
			}
			select {
			case <-barrier.read:
				t.Fatal("failed query signalled successful-read barrier")
			default:
			}
		})
	}
}

var errAggregateFallbackArguments = errors.New("query/arguments not forwarded")

// This narrow driver forces database/sql's ErrSkip -> Prepare path, including
// legacy Stmt.Query and StmtQueryContext, without requiring a MySQL server.
type aggregateFallbackConnector struct{ contextStatement bool }

func (c *aggregateFallbackConnector) Driver() driver.Driver { return aggregateFallbackDriver{} }
func (c *aggregateFallbackConnector) Connect(context.Context) (driver.Conn, error) {
	return &aggregateFallbackConn{contextStatement: c.contextStatement}, nil
}

type aggregateFallbackDriver struct{}

func (aggregateFallbackDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("use connector")
}

type aggregateFallbackConn struct{ contextStatement bool }

func (c *aggregateFallbackConn) Prepare(query string) (driver.Stmt, error) {
	stmt := &aggregateFallbackStmt{query: query}
	if c.contextStatement {
		return &aggregateFallbackContextStmt{aggregateFallbackStmt: stmt}, nil
	}
	return stmt, nil
}
func (*aggregateFallbackConn) Close() error { return nil }
func (*aggregateFallbackConn) Begin() (driver.Tx, error) {
	return nil, errors.New("unused transaction")
}
func (*aggregateFallbackConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return nil, driver.ErrSkip
}

type aggregateFallbackStmt struct{ query string }

func (*aggregateFallbackStmt) Close() error  { return nil }
func (*aggregateFallbackStmt) NumInput() int { return 1 }
func (*aggregateFallbackStmt) Exec([]driver.Value) (driver.Result, error) {
	return nil, errors.New("unused execution")
}
func (s *aggregateFallbackStmt) Query(args []driver.Value) (driver.Rows, error) {
	if s.query != "SELECT br.status FROM book_runs br WHERE br.run_id=?" || len(args) != 1 || args[0] != int64(11) {
		return nil, errAggregateFallbackArguments
	}
	return &aggregateFallbackRows{}, nil
}

type aggregateFallbackContextStmt struct{ *aggregateFallbackStmt }

func (s *aggregateFallbackContextStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(args) != 1 || args[0].Ordinal != 1 || args[0].Value != int64(11) {
		return nil, errAggregateFallbackArguments
	}
	return s.Query([]driver.Value{args[0].Value})
}

type aggregateFallbackRows struct{ read bool }

func (*aggregateFallbackRows) Columns() []string { return []string{"status"} }
func (*aggregateFallbackRows) Close() error      { return nil }
func (r *aggregateFallbackRows) Next(values []driver.Value) error {
	if r.read {
		return io.EOF
	}
	r.read = true
	values[0] = "failed"
	return nil
}
