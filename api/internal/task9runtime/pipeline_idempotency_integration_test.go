package task9runtime_test

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
	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
	_ "github.com/go-sql-driver/mysql"
)

// TestPERFRunIdempotency001BusinessCreatePath is the performance-audit P0
// reproduction, executed inside the mandatory Task9 real-MySQL CI job.
func TestPERFRunIdempotency001BusinessCreatePath(t *testing.T) {
	dsn:=os.Getenv("TASK9_MYSQL_DSN"); if dsn==""{t.Skip("TASK9_MYSQL_DSN not configured")}
	db,err:=sql.Open("mysql",dsn);if err!=nil{t.Fatal(err)};defer db.Close();if err:=db.Ping();err!=nil{t.Fatal(err)}
	ctx:=context.Background();name:=fmt.Sprintf("perf-idempotency-%d",time.Now().UnixNano())
	res,err:=db.ExecContext(ctx,`INSERT INTO intakes (name,status) VALUES (?, 'completed')`,name);if err!=nil{t.Fatal(err)}
	intakeID,_:=res.LastInsertId();t.Cleanup(func(){_,_=db.ExecContext(context.Background(),`DELETE FROM intakes WHERE id=?`,intakeID)})
	fixedNow:=time.Date(2026,10,5,12,0,0,0,time.UTC);service:=pipeline.NewService(intake.NewMySQLStore(db),func()time.Time{return fixedNow})
	const concurrency=20;start:=make(chan struct{});errCh:=make(chan error,concurrency);var wg sync.WaitGroup;wg.Add(concurrency);before:=runtime.NumGoroutine()
	for i:=0;i<concurrency;i++{go func(){defer wg.Done();<-start;_,err:=service.Create(ctx,pipeline.CreateRequest{IntakeID:intakeID,Name:name});errCh<-err}()};close(start);wg.Wait();close(errCh)
	failed:=0;for err:=range errCh{if err!=nil{failed++;t.Logf("create error: %v",err)}}
	var projects,runs int;if err:=db.QueryRowContext(ctx,`SELECT COUNT(*) FROM batch_projects WHERE intake_id=?`,intakeID).Scan(&projects);err!=nil{t.Fatal(err)}
	if err:=db.QueryRowContext(ctx,`SELECT COUNT(*) FROM runs r JOIN batch_projects bp ON bp.id=r.batch_project_id WHERE bp.intake_id=?`,intakeID).Scan(&runs);err!=nil{t.Fatal(err)}
	t.Logf("PERF-RUN-IDEMPOTENCY-001 concurrency=%d failed=%d projects=%d logical_runs=%d goroutines_before=%d goroutines_after=%d",concurrency,failed,projects,runs,before,runtime.NumGoroutine())
	if failed!=0{t.Fatalf("concurrent create failures=%d want 0",failed)};if projects!=1{t.Fatalf("BatchProjects=%d want 1",projects)};if runs!=1{t.Fatalf("logical Runs=%d want 1",runs)}
}
