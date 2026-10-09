package pipeline

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	_ "github.com/go-sql-driver/mysql"
)

func TestConcurrentCreateSameIntakeIsIdempotent(t *testing.T) {
	dsn := os.Getenv("PERF_MYSQL_DSN")
	if dsn == "" {
		t.Skip("PERF_MYSQL_DSN not configured")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	name := fmt.Sprintf("perf-idempotency-%d", time.Now().UnixNano())
	result, err := db.ExecContext(ctx, `INSERT INTO intakes (name, status) VALUES (?, 'completed')`, name)
	if err != nil {
		t.Fatal(err)
	}
	intakeID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	defer db.ExecContext(ctx, `DELETE FROM intakes WHERE id = ?`, intakeID)

	fixedNow := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	service := NewService(intake.NewMySQLStore(db), func() time.Time { return fixedNow })

	const concurrency = 20
	start := make(chan struct{})
	errCh := make(chan error, concurrency)
	var wg sync.WaitGroup
	wg.Add(concurrency)

	goroutinesBefore := runtime.NumGoroutine()
	started := time.Now()
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			<-start
			_, err := service.Create(ctx, CreateRequest{IntakeID: intakeID, Name: name})
			errCh <- err
		}()
	}
	close(start)
	wg.Wait()
	elapsed := time.Since(started)
	close(errCh)

	succeeded := 0
	failed := 0
	for err := range errCh {
		if err != nil {
			failed++
			t.Logf("create error: %v", err)
			continue
		}
		succeeded++
	}

	var projectCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM batch_projects WHERE intake_id = ?`, intakeID).Scan(&projectCount); err != nil {
		t.Fatal(err)
	}
	var runCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM runs r JOIN batch_projects bp ON bp.id = r.batch_project_id WHERE bp.intake_id = ?`, intakeID).Scan(&runCount); err != nil {
		t.Fatal(err)
	}

	stats := db.Stats()
	goroutinesAfter := runtime.NumGoroutine()
	t.Logf("PERF_RED concurrency=%d succeeded=%d failed=%d projects=%d logical_runs=%d duration=%s goroutines_before=%d goroutines_after=%d db_open=%d db_in_use=%d db_idle=%d db_wait_count=%d db_wait_duration=%s",
		concurrency,
		succeeded,
		failed,
		projectCount,
		runCount,
		elapsed,
		goroutinesBefore,
		goroutinesAfter,
		stats.OpenConnections,
		stats.InUse,
		stats.Idle,
		stats.WaitCount,
		stats.WaitDuration,
	)

	if failed != 0 {
		t.Fatalf("concurrent create failures = %d, want 0", failed)
	}
	if projectCount != 1 {
		t.Fatalf("batch projects = %d, want 1", projectCount)
	}
	if runCount != 1 {
		t.Fatalf("logical runs = %d, want 1", runCount)
	}
}
