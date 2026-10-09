package generation

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type perfGenerationProvider struct {
	failBookID int64
	calls      atomic.Int64
	active     atomic.Int64
	maxActive  atomic.Int64
}

func (p *perfGenerationProvider) Complete(ctx context.Context, req TextRequest) (string, error) {
	active := p.active.Add(1)
	defer p.active.Add(-1)
	p.calls.Add(1)
	for {
		current := p.maxActive.Load()
		if active <= current || p.maxActive.CompareAndSwap(current, active) {
			break
		}
	}
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(500 * time.Microsecond):
	}
	if p.failBookID > 0 && req.BookID == p.failBookID && req.Stage == StageScript {
		return "", fmt.Errorf("controlled provider failure")
	}
	return fmt.Sprintf("controlled-%s-book-%d", req.Stage, req.BookID), nil
}

func TestPerformanceGenerationControlled10_50_100(t *testing.T) {
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

	for _, total := range []int{10, 50, 100} {
		total := total
		t.Run(fmt.Sprintf("books_%d", total), func(t *testing.T) {
			intakeID, projectID, bookIDs := seedGenerationPerfProject(t, db, total)
			defer db.Exec(`DELETE FROM intakes WHERE id = ?`, intakeID)
			provider := &perfGenerationProvider{}
			service := NewService(NewMySQLStore(db), provider, nil)

			runtime.GC()
			var before runtime.MemStats
			runtime.ReadMemStats(&before)
			rssBefore := perfGenerationRSS()
			gorBefore := runtime.NumGoroutine()
			dbBefore := db.Stats()

			latencies := make([]time.Duration, 0, len(bookIDs))
			failed := 0
			started := time.Now()
			for _, bookID := range bookIDs {
				one := time.Now()
				_, runErr := service.RunBook(context.Background(), RunBookRequest{
					BatchProjectID: projectID,
					BookID:         bookID,
					DirectorMode:   DirectorNormal,
					RequestID:      fmt.Sprintf("perf-%d", bookID),
				})
				latencies = append(latencies, time.Since(one))
				if runErr != nil {
					failed++
				}
			}
			duration := time.Since(started)
			sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

			var loaded runtime.MemStats
			runtime.ReadMemStats(&loaded)
			rssPeak := perfGenerationRSS()
			runtime.GC()
			time.Sleep(20 * time.Millisecond)
			var post runtime.MemStats
			runtime.ReadMemStats(&post)
			dbAfter := db.Stats()
			throughput := float64(total) / duration.Seconds()

			t.Logf("PERF_GENERATION books=%d total=%d success=%d failed=%d duration=%s throughput=%.2f p50=%s p95=%s p99=%s provider_calls=%d provider_max_active=%d heap_before=%d heap_peak=%d heap_post_gc=%d rss_before=%d rss_peak=%d rss_post_gc=%d goroutines_before=%d goroutines_after=%d db_open=%d db_in_use=%d db_idle=%d db_wait_count_delta=%d db_wait_duration_delta=%s",
				total, total, total-failed, failed, duration, throughput,
				perfGenerationPercentile(latencies, .50), perfGenerationPercentile(latencies, .95), perfGenerationPercentile(latencies, .99),
				provider.calls.Load(), provider.maxActive.Load(), before.HeapAlloc, loaded.HeapAlloc, post.HeapAlloc,
				rssBefore, rssPeak, perfGenerationRSS(), gorBefore, runtime.NumGoroutine(),
				dbAfter.OpenConnections, dbAfter.InUse, dbAfter.Idle, dbAfter.WaitCount-dbBefore.WaitCount, dbAfter.WaitDuration-dbBefore.WaitDuration)

			if failed != 0 {
				t.Fatalf("generation failed=%d", failed)
			}
			if provider.maxActive.Load() != 1 {
				t.Fatalf("current synchronous model unexpectedly executed provider concurrently: max_active=%d", provider.maxActive.Load())
			}
		})
	}
}

func TestPerformanceGenerationBatchFailureIsolation(t *testing.T) {
	dsn := os.Getenv("PERF_MYSQL_DSN")
	if dsn == "" {
		t.Skip("PERF_MYSQL_DSN not configured")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	intakeID, projectID, bookIDs := seedGenerationPerfProject(t, db, 10)
	defer db.Exec(`DELETE FROM intakes WHERE id = ?`, intakeID)

	provider := &perfGenerationProvider{failBookID: bookIDs[4]}
	service := NewService(NewMySQLStore(db), provider, nil)
	started := time.Now()
	result, runErr := service.RunBatch(context.Background(), RunBatchRequest{
		BatchProjectID: projectID,
		DirectorMode:   DirectorNormal,
		RequestID:      "perf-failure-isolation",
	})
	duration := time.Since(started)

	t.Logf("PERF_GENERATION_ISOLATION total=%d completed=%d failed=%d provider_calls=%d duration=%s failed_book=%d last_book=%d",
		len(result.Books), result.Completed, result.Failed, provider.calls.Load(), duration, bookIDs[4], bookIDs[len(bookIDs)-1])
	if runErr == nil {
		t.Fatal("expected aggregate error for controlled failed book")
	}
	if len(result.Books) != 10 || result.Completed != 9 || result.Failed != 1 {
		t.Fatalf("isolation result total=%d completed=%d failed=%d, want 10/9/1", len(result.Books), result.Completed, result.Failed)
	}
	last, err := service.store.LatestBookRun(context.Background(), projectID, bookIDs[len(bookIDs)-1])
	if err != nil {
		t.Fatal(err)
	}
	if last.Status != StatusCompleted {
		t.Fatalf("book after controlled failure status=%s, want completed", last.Status)
	}
}

func seedGenerationPerfProject(t *testing.T, db *sql.DB, count int) (int64, int64, []int64) {
	t.Helper()
	ctx := context.Background()
	name := fmt.Sprintf("perf-generation-%d-%d", count, time.Now().UnixNano())
	res, err := db.ExecContext(ctx, `INSERT INTO intakes (name, status) VALUES (?, 'completed')`, name)
	if err != nil {
		t.Fatal(err)
	}
	intakeID, _ := res.LastInsertId()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	bookIDs := make([]int64, 0, count)
	body := strings.Repeat("正文容量测试。", 256)
	for i := 0; i < count; i++ {
		result, err := tx.ExecContext(ctx, `INSERT INTO books (intake_id, source, platform_id, external_book_id, title, original_text, gender, gender_source, style, status) VALUES (?, 'perf-generation', '1', ?, ?, ?, 'male', 'perf', 'baseline', 'fetched')`, intakeID, fmt.Sprintf("g-%d", i), fmt.Sprintf("Generation %d", i), body)
		if err != nil {
			_ = tx.Rollback()
			t.Fatal(err)
		}
		id, _ := result.LastInsertId()
		bookIDs = append(bookIDs, id)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	project, err := db.ExecContext(ctx, `INSERT INTO batch_projects (intake_id, name) VALUES (?, ?)`, intakeID, name)
	if err != nil {
		t.Fatal(err)
	}
	projectID, _ := project.LastInsertId()
	return intakeID, projectID, bookIDs
}

func perfGenerationPercentile(values []time.Duration, p float64) time.Duration {
	if len(values) == 0 {
		return 0
	}
	index := int(float64(len(values))*p+0.999999) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(values) {
		index = len(values) - 1
	}
	return values[index]
}

func perfGenerationRSS() uint64 {
	file, err := os.Open("/proc/self/status")
	if err != nil {
		return 0
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(strings.TrimSpace(scanner.Text()))
		if len(fields) >= 2 && fields[0] == "VmRSS:" {
			kb, _ := strconv.ParseUint(fields[1], 10, 64)
			return kb * 1024
		}
	}
	return 0
}
