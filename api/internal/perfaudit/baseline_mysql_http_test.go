package perfaudit

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/httpapi"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	_ "github.com/go-sql-driver/mysql"
)

type latencySummary struct {
	Total      int
	Success    int
	Failed     int
	Duration   time.Duration
	P50        time.Duration
	P95        time.Duration
	P99        time.Duration
	Throughput float64
}

func TestMySQLAndHTTPBaseline(t *testing.T) {
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

	store := intake.NewMySQLStore(db)
	handler := httpapi.NewHandler(httpapi.Dependencies{
		BatchProjects:       store,
		BatchProjectDetails: store,
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	for _, bookCount := range []int{10, 50, 100} {
		bookCount := bookCount
		t.Run(fmt.Sprintf("books_%d", bookCount), func(t *testing.T) {
			ctx := context.Background()
			intakeID, projectID := seedProject(t, db, bookCount)
			defer db.ExecContext(ctx, `DELETE FROM intakes WHERE id = ?`, intakeID)

			runtime.GC()
			var memBefore runtime.MemStats
			runtime.ReadMemStats(&memBefore)
			rssBefore := readRSSBytes()
			goroutinesBefore := runtime.NumGoroutine()
			dbBefore := db.Stats()

			runStarted := time.Now()
			if _, err := store.CreateRun(ctx, intake.Run{
				BatchProjectID: projectID,
				RunAt:          time.Now().UTC(),
				Status:         intake.RunStatusPending,
			}); err != nil {
				t.Fatalf("create run: %v", err)
			}
			runCreateLatency := time.Since(runStarted)

			mysqlList := measureSequential(25, func() error {
				_, err := store.ListBatchProjects(ctx)
				return err
			})
			mysqlDetail := measureSequential(25, func() error {
				_, err := store.ListBooks(ctx, intakeID)
				return err
			})

			listURL := server.URL + "/api/v1/batch-projects"
			detailURL := fmt.Sprintf("%s/api/v1/batch-projects/%d", server.URL, projectID)
			httpList := measureConcurrentHTTP(server.Client(), listURL, bookCount)
			httpDetail := measureConcurrentHTTP(server.Client(), detailURL, bookCount)

			var memAfterLoad runtime.MemStats
			runtime.ReadMemStats(&memAfterLoad)
			rssAfterLoad := readRSSBytes()
			runtime.GC()
			var memPostGC runtime.MemStats
			runtime.ReadMemStats(&memPostGC)
			rssPostGC := readRSSBytes()
			goroutinesAfter := runtime.NumGoroutine()
			dbAfter := db.Stats()

			t.Logf("PERF_BASELINE books=%d run_create=%s mysql_list_p50=%s mysql_list_p95=%s mysql_list_p99=%s mysql_detail_p50=%s mysql_detail_p95=%s mysql_detail_p99=%s http_list_total=%d http_list_success=%d http_list_failed=%d http_list_p50=%s http_list_p95=%s http_list_p99=%s http_list_rps=%.2f http_detail_total=%d http_detail_success=%d http_detail_failed=%d http_detail_p50=%s http_detail_p95=%s http_detail_p99=%s http_detail_rps=%.2f heap_before=%d heap_after_load=%d heap_post_gc=%d rss_before=%d rss_after_load=%d rss_post_gc=%d goroutines_before=%d goroutines_after=%d db_open_before=%d db_open_after=%d db_wait_count_delta=%d db_wait_duration_delta=%s",
				bookCount,
				runCreateLatency,
				mysqlList.P50,
				mysqlList.P95,
				mysqlList.P99,
				mysqlDetail.P50,
				mysqlDetail.P95,
				mysqlDetail.P99,
				httpList.Total,
				httpList.Success,
				httpList.Failed,
				httpList.P50,
				httpList.P95,
				httpList.P99,
				httpList.Throughput,
				httpDetail.Total,
				httpDetail.Success,
				httpDetail.Failed,
				httpDetail.P50,
				httpDetail.P95,
				httpDetail.P99,
				httpDetail.Throughput,
				memBefore.HeapAlloc,
				memAfterLoad.HeapAlloc,
				memPostGC.HeapAlloc,
				rssBefore,
				rssAfterLoad,
				rssPostGC,
				goroutinesBefore,
				goroutinesAfter,
				dbBefore.OpenConnections,
				dbAfter.OpenConnections,
				dbAfter.WaitCount-dbBefore.WaitCount,
				dbAfter.WaitDuration-dbBefore.WaitDuration,
			)

			if mysqlList.Failed != 0 || mysqlDetail.Failed != 0 {
				t.Fatalf("mysql baseline errors: list=%d detail=%d", mysqlList.Failed, mysqlDetail.Failed)
			}
			if httpList.Failed != 0 || httpDetail.Failed != 0 {
				t.Fatalf("http baseline errors: list=%d detail=%d", httpList.Failed, httpDetail.Failed)
			}
		})
	}
}

func seedProject(t *testing.T, db *sql.DB, bookCount int) (int64, int64) {
	t.Helper()
	ctx := context.Background()
	name := fmt.Sprintf("perf-baseline-%d-%d", bookCount, time.Now().UnixNano())
	result, err := db.ExecContext(ctx, `INSERT INTO intakes (name, status) VALUES (?, 'completed')`, name)
	if err != nil {
		t.Fatal(err)
	}
	intakeID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	body := strings.Repeat("正文", 2048)
	for i := 0; i < bookCount; i++ {
		if _, err := tx.ExecContext(ctx, `INSERT INTO books (intake_id, source, platform_id, external_book_id, title, original_text, gender, gender_source, style, status) VALUES (?, 'perf', '1', ?, ?, ?, 'male', 'perf', 'baseline', 'fetched')`, intakeID, fmt.Sprintf("book-%d", i), fmt.Sprintf("Book %d", i), body); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	committed = true

	projectResult, err := db.ExecContext(ctx, `INSERT INTO batch_projects (intake_id, name) VALUES (?, ?)`, intakeID, name)
	if err != nil {
		t.Fatal(err)
	}
	projectID, err := projectResult.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return intakeID, projectID
}

func measureSequential(total int, fn func() error) latencySummary {
	latencies := make([]time.Duration, 0, total)
	failed := 0
	started := time.Now()
	for i := 0; i < total; i++ {
		one := time.Now()
		if err := fn(); err != nil {
			failed++
		}
		latencies = append(latencies, time.Since(one))
	}
	return summarize(latencies, time.Since(started), failed)
}

func measureConcurrentHTTP(client *http.Client, url string, total int) latencySummary {
	latencies := make([]time.Duration, total)
	failed := 0
	var failedMu sync.Mutex
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(total)
	started := time.Now()
	for i := 0; i < total; i++ {
		i := i
		go func() {
			defer wg.Done()
			<-start
			one := time.Now()
			resp, err := client.Get(url)
			latencies[i] = time.Since(one)
			bad := err != nil
			if resp != nil {
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
				bad = bad || resp.StatusCode != http.StatusOK
			}
			if bad {
				failedMu.Lock()
				failed++
				failedMu.Unlock()
			}
		}()
	}
	close(start)
	wg.Wait()
	return summarize(latencies, time.Since(started), failed)
}

func summarize(latencies []time.Duration, duration time.Duration, failed int) latencySummary {
	ordered := append([]time.Duration(nil), latencies...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	total := len(ordered)
	throughput := 0.0
	if duration > 0 {
		throughput = float64(total) / duration.Seconds()
	}
	return latencySummary{
		Total:      total,
		Success:    total - failed,
		Failed:     failed,
		Duration:   duration,
		P50:        percentile(ordered, 0.50),
		P95:        percentile(ordered, 0.95),
		P99:        percentile(ordered, 0.99),
		Throughput: throughput,
	}
}

func percentile(values []time.Duration, p float64) time.Duration {
	if len(values) == 0 {
		return 0
	}
	index := int(float64(len(values))*p + 0.999999) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(values) {
		index = len(values) - 1
	}
	return values[index]
}

func readRSSBytes() uint64 {
	file, err := os.Open("/proc/self/status")
	if err != nil {
		return 0
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "VmRSS:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return 0
		}
		kb, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return 0
		}
		return kb * 1024
	}
	return 0
}
