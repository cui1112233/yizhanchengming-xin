package perfaudit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/httpapi"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	_ "github.com/go-sql-driver/mysql"
)

func TestPerformanceRepeatedRSSAndGoroutineRetention(t *testing.T) {
	dsn := os.Getenv("PERF_MYSQL_DSN")
	if dsn == "" {
		t.Skip("PERF_MYSQL_DSN not configured")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := intake.NewMySQLStore(db)
	handler := httpapi.NewHandler(httpapi.Dependencies{BatchProjects: store, BatchProjectDetails: store})
	server := httptest.NewServer(handler)
	defer server.Close()

	intakeID, projectID := seedProject(t, db, 100)
	defer db.Exec(`DELETE FROM intakes WHERE id = ?`, intakeID)
	url := fmt.Sprintf("%s/api/v1/batch-projects/%d", server.URL, projectID)

	// Warm reusable HTTP/client goroutines before the retention series.
	warm := measureConcurrentHTTP(server.Client(), url, 20)
	if warm.Failed != 0 {
		t.Fatalf("warmup failures=%d", warm.Failed)
	}
	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	baseGoroutines := runtime.NumGoroutine()
	baseRSS := readRSSBytes()

	var previousRSS uint64
	strictRSSIncreases := 0
	for round := 1; round <= 10; round++ {
		result := measureConcurrentHTTP(server.Client(), url, 100)
		if result.Failed != 0 {
			t.Fatalf("round=%d failures=%d", round, result.Failed)
		}
		runtime.GC()
		time.Sleep(50 * time.Millisecond)
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		rss := readRSSBytes()
		goroutines := runtime.NumGoroutine()
		stats := db.Stats()
		if previousRSS > 0 && rss > previousRSS {
			strictRSSIncreases++
		}
		previousRSS = rss
		t.Logf("PERF_RETENTION round=%d requests=100 p95=%s heap_alloc=%d heap_inuse=%d sys=%d rss=%d goroutines=%d db_open=%d db_in_use=%d db_idle=%d db_wait_count=%d db_wait_duration=%s",
			round, result.P95, mem.HeapAlloc, mem.HeapInuse, mem.Sys, rss, goroutines,
			stats.OpenConnections, stats.InUse, stats.Idle, stats.WaitCount, stats.WaitDuration)
	}
	finalGoroutines := runtime.NumGoroutine()
	finalRSS := readRSSBytes()
	t.Logf("PERF_RETENTION_SUMMARY rounds=10 base_rss=%d final_rss=%d rss_delta=%d strict_rss_increase_steps=%d base_goroutines=%d final_goroutines=%d goroutine_delta=%d",
		baseRSS, finalRSS, int64(finalRSS)-int64(baseRSS), strictRSSIncreases, baseGoroutines, finalGoroutines, finalGoroutines-baseGoroutines)
}

func TestPerformanceProjectDetailDecomposition(t *testing.T) {
	dsn := os.Getenv("PERF_MYSQL_DSN")
	if dsn == "" {
		t.Skip("PERF_MYSQL_DSN not configured")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := intake.NewMySQLStore(db)
	intakeID, projectID := seedProject(t, db, 100)
	defer db.Exec(`DELETE FROM intakes WHERE id = ?`, intakeID)

	handler := httpapi.NewHandler(httpapi.Dependencies{BatchProjects: store, BatchProjectDetails: store})
	server := httptest.NewServer(handler)
	defer server.Close()
	path := fmt.Sprintf("/api/v1/batch-projects/%d", projectID)
	url := server.URL + path

	var objectBytes int
	books, err := store.ListBooks(context.Background(), intakeID)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(books)
	if err != nil {
		t.Fatal(err)
	}
	objectBytes = len(encoded)

	directStore := measureConcurrentFunc(100, func() error {
		_, err := store.ListBooks(context.Background(), intakeID)
		return err
	})
	storeAndJSON := measureConcurrentFunc(100, func() error {
		values, err := store.ListBooks(context.Background(), intakeID)
		if err != nil {
			return err
		}
		_, err = json.Marshal(values)
		return err
	})
	inProcessHandler := measureConcurrentFunc(100, func() error {
		req := httptest.NewRequest("GET", path, nil)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		if recorder.Code != 200 {
			return fmt.Errorf("status=%d", recorder.Code)
		}
		return nil
	})
	network := measureConcurrentHTTP(server.Client(), url, 100)

	t.Logf("PERF_DETAIL_DECOMP concurrency=100 object_json_bytes=%d auth_session=not_in_baseline direct_store_p50=%s direct_store_p95=%s direct_store_p99=%s store_json_p50=%s store_json_p95=%s store_json_p99=%s handler_inprocess_p50=%s handler_inprocess_p95=%s handler_inprocess_p99=%s network_p50=%s network_p95=%s network_p99=%s direct_failed=%d json_failed=%d handler_failed=%d network_failed=%d",
		objectBytes,
		directStore.P50, directStore.P95, directStore.P99,
		storeAndJSON.P50, storeAndJSON.P95, storeAndJSON.P99,
		inProcessHandler.P50, inProcessHandler.P95, inProcessHandler.P99,
		network.P50, network.P95, network.P99,
		directStore.Failed, storeAndJSON.Failed, inProcessHandler.Failed, network.Failed)
	if directStore.Failed+storeAndJSON.Failed+inProcessHandler.Failed+network.Failed != 0 {
		t.Fatal("project detail decomposition had request failures")
	}
}

func measureConcurrentFunc(total int, fn func() error) latencySummary {
	latencies := make([]time.Duration, total)
	failed := 0
	var mu sync.Mutex
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
			err := fn()
			latencies[i] = time.Since(one)
			if err != nil {
				mu.Lock()
				failed++
				mu.Unlock()
			}
		}()
	}
	close(start)
	wg.Wait()
	return summarize(latencies, time.Since(started), failed)
}
