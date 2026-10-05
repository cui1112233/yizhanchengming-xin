package task9runtime

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrRunNotFound = errors.New("task9 run not found")
	ErrBookRunNotRetryable = errors.New("task9 book run not retryable")
)

type RunRecord struct {
	ID int64
	BatchProjectID int64
	IdempotencyKey string
	RunAt time.Time
	Status RunState
	MaxAttempts int
}

type MySQLStore struct { db *sql.DB }
func NewMySQLStore(db *sql.DB)*MySQLStore{return &MySQLStore{db:db}}

func (s *MySQLStore) CreateRun(ctx context.Context, projectID int64, key string, runAt time.Time, maxAttempts int)(RunRecord,bool,error){
	if key=="" {return RunRecord{},false,errors.New("idempotency key required")}
	if maxAttempts<=0 {maxAttempts=3}
	res,err:=s.db.ExecContext(ctx,`INSERT INTO runs (batch_project_id,idempotency_key,run_at,status,max_attempts) VALUES (?,?,?,'pending',?) ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id)`,projectID,key,runAt,maxAttempts)
	if err!=nil{return RunRecord{},false,err}
	id,err:=res.LastInsertId();if err!=nil{return RunRecord{},false,err}
	affected,_:=res.RowsAffected();created:=affected==1
	var r RunRecord;var status string
	if err:=s.db.QueryRowContext(ctx,`SELECT id,batch_project_id,idempotency_key,run_at,status,max_attempts FROM runs WHERE id=?`,id).Scan(&r.ID,&r.BatchProjectID,&r.IdempotencyKey,&r.RunAt,&status,&r.MaxAttempts);err!=nil{return RunRecord{},false,err}
	r.Status=RunState(status);return r,created,nil
}

func (s *MySQLStore) ClaimDueRun(ctx context.Context, now time.Time)(RunClaim,bool,error){
	tx,err:=s.db.BeginTx(ctx,&sql.TxOptions{Isolation:sql.LevelReadCommitted});if err!=nil{return RunClaim{},false,err};defer tx.Rollback()
	var id int64
	err=tx.QueryRowContext(ctx,`SELECT id FROM runs WHERE status='pending' AND run_at<=? ORDER BY run_at,id LIMIT 1 FOR UPDATE SKIP LOCKED`,now).Scan(&id)
	if errors.Is(err,sql.ErrNoRows){return RunClaim{},false,nil};if err!=nil{return RunClaim{},false,err}
	res,err:=tx.ExecContext(ctx,`UPDATE runs SET status='running',started_at=COALESCE(started_at,?) WHERE id=? AND status='pending'`,now,id);if err!=nil{return RunClaim{},false,err}
	n,_:=res.RowsAffected();if n!=1{return RunClaim{},false,nil}
	if err:=tx.Commit();err!=nil{return RunClaim{},false,err};return RunClaim{RunID:id},true,nil
}

func (s *MySQLStore) EnsureQueuedBooks(ctx context.Context, claim RunClaim)([]WorkItem,error){
	tx,err:=s.db.BeginTx(ctx,&sql.TxOptions{Isolation:sql.LevelReadCommitted});if err!=nil{return nil,err};defer tx.Rollback()
	var projectID int64;var maxAttempts int;var status string
	if err:=tx.QueryRowContext(ctx,`SELECT batch_project_id,max_attempts,status FROM runs WHERE id=? FOR UPDATE`,claim.RunID).Scan(&projectID,&maxAttempts,&status);err!=nil{return nil,err}
	if status!="running" {return nil,fmt.Errorf("run %d not running",claim.RunID)}
	rows,err:=tx.QueryContext(ctx,`SELECT b.id FROM books b JOIN batch_projects p ON p.intake_id=b.intake_id WHERE p.id=? ORDER BY b.id`,projectID);if err!=nil{return nil,err}
	var bookIDs []int64;for rows.Next(){var id int64;if err:=rows.Scan(&id);err!=nil{rows.Close();return nil,err};bookIDs=append(bookIDs,id)};if err:=rows.Close();err!=nil{return nil,err}
	for _,bookID:=range bookIDs {if _,err:=tx.ExecContext(ctx,`INSERT INTO book_runs (run_id,batch_project_id,book_id,attempt,max_attempts,retryable,status) VALUES (?,?,?,?,?,1,'queued') ON DUPLICATE KEY UPDATE id=id`,claim.RunID,projectID,bookID,1,maxAttempts);err!=nil{return nil,err}}
	rows,err=tx.QueryContext(ctx,`SELECT id,book_id,attempt FROM book_runs WHERE run_id=? AND attempt=1 ORDER BY book_id`,claim.RunID);if err!=nil{return nil,err}
	items:=make([]WorkItem,0,len(bookIDs));for rows.Next(){var w WorkItem;if err:=rows.Scan(&w.BookRunID,&w.BookID,&w.Attempt);err!=nil{rows.Close();return nil,err};items=append(items,w)};if err:=rows.Close();err!=nil{return nil,err}
	if err:=tx.Commit();err!=nil{return nil,err};return items,nil
}

func (s *MySQLStore) Claim(ctx context.Context,item WorkItem,owner string,deadline time.Time)(Execution,bool,error){
	now:=time.Now();res,err:=s.db.ExecContext(ctx,`UPDATE book_runs SET status='running',execution_token=execution_token+1,execution_owner=?,running_since=?,lease_deadline=?,heartbeat_at=? WHERE id=? AND attempt=? AND status='queued'`,owner,now,deadline,now,item.BookRunID,item.Attempt);if err!=nil{return Execution{},false,err}
	n,_:=res.RowsAffected();if n!=1{return Execution{},false,nil}
	var token uint64;var attempt int
	if err:=s.db.QueryRowContext(ctx,`SELECT attempt,execution_token FROM book_runs WHERE id=?`,item.BookRunID).Scan(&attempt,&token);err!=nil{return Execution{},false,err}
	return Execution{BookRunID:item.BookRunID,Attempt:attempt,FencingToken:token,Owner:owner},true,nil
}

func (s *MySQLStore) Renew(ctx context.Context,e Execution,deadline time.Time)(bool,error){
	res,err:=s.db.ExecContext(ctx,`UPDATE book_runs SET lease_deadline=?,heartbeat_at=? WHERE id=? AND attempt=? AND execution_token=? AND execution_owner=? AND status='running'`,deadline,time.Now(),e.BookRunID,e.Attempt,e.FencingToken,e.Owner);if err!=nil{return false,err};n,_:=res.RowsAffected();return n==1,nil
}

func (s *MySQLStore) Complete(ctx context.Context,e Execution)(bool,error){
	res,err:=s.db.ExecContext(ctx,`UPDATE book_runs SET status='succeeded',finished_at=?,lease_deadline=NULL,heartbeat_at=? WHERE id=? AND attempt=? AND execution_token=? AND execution_owner=? AND status='running'`,time.Now(),time.Now(),e.BookRunID,e.Attempt,e.FencingToken,e.Owner);if err!=nil{return false,err};n,_:=res.RowsAffected();return n==1,nil
}

func safeFailure(code,message string)(string,string){
	if message=="" {if code==""{code="execution_error"};return code,""}
	safeCode,safeMessage:=SafeError(errors.New(message));if safeCode=="internal_error"{return safeCode,safeMessage};if code==""{code=safeCode};return code,safeMessage
}
func (s *MySQLStore) Fail(ctx context.Context,e Execution,f Failure)(bool,error){
	code,msg:=safeFailure(f.Code,f.Message);retryable:=0;if f.Retryable{retryable=1}
	res,err:=s.db.ExecContext(ctx,`UPDATE book_runs SET status='failed',retryable=?,error_code=?,error_message=?,finished_at=?,lease_deadline=NULL,heartbeat_at=? WHERE id=? AND attempt=? AND execution_token=? AND execution_owner=? AND status='running'`,retryable,code,msg,time.Now(),time.Now(),e.BookRunID,e.Attempt,e.FencingToken,e.Owner);if err!=nil{return false,err};n,_:=res.RowsAffected();return n==1,nil
}

func (s *MySQLStore) RetryBookRun(ctx context.Context, failedBookRunID int64)(WorkItem,bool,error){
	tx,err:=s.db.BeginTx(ctx,&sql.TxOptions{Isolation:sql.LevelReadCommitted});if err!=nil{return WorkItem{},false,err};defer tx.Rollback()
	var runID sql.NullInt64;var projectID,bookID int64;var attempt,maxAttempts int;var retryable bool;var status string
	if err:=tx.QueryRowContext(ctx,`SELECT run_id,batch_project_id,book_id,attempt,max_attempts,retryable,status FROM book_runs WHERE id=? FOR UPDATE`,failedBookRunID).Scan(&runID,&projectID,&bookID,&attempt,&maxAttempts,&retryable,&status);err!=nil{return WorkItem{},false,err}
	if !runID.Valid||status!="failed"||!retryable||attempt>=maxAttempts{return WorkItem{},false,ErrBookRunNotRetryable}
	next:=attempt+1
	res,err:=tx.ExecContext(ctx,`INSERT INTO book_runs (run_id,batch_project_id,book_id,attempt,max_attempts,retryable,status) VALUES (?,?,?,?,?,1,'queued') ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id)`,runID.Int64,projectID,bookID,next,maxAttempts);if err!=nil{return WorkItem{},false,err}
	id,err:=res.LastInsertId();if err!=nil{return WorkItem{},false,err};affected,_:=res.RowsAffected();created:=affected==1
	if err:=tx.Commit();err!=nil{return WorkItem{},false,err};return WorkItem{BookRunID:id,BookID:bookID,Attempt:next},created,nil
}

func (s *MySQLStore) AggregateRunStatus(ctx context.Context,runID int64)(RunState,error){
	tx,err:=s.db.BeginTx(ctx,&sql.TxOptions{Isolation:sql.LevelRepeatableRead});if err!=nil{return RunRunning,err};defer tx.Rollback()
	rows,err:=tx.QueryContext(ctx,`SELECT br.status,br.retryable,br.attempt,br.max_attempts FROM book_runs br JOIN (SELECT book_id,MAX(attempt) max_attempt FROM book_runs WHERE run_id=? GROUP BY book_id) latest ON latest.book_id=br.book_id AND latest.max_attempt=br.attempt WHERE br.run_id=?`,runID,runID);if err!=nil{return RunRunning,err}
	total,succeeded,failed:=0,0,0;active:=false
	for rows.Next(){var status string;var retryable bool;var attempt,maxAttempts int;if err:=rows.Scan(&status,&retryable,&attempt,&maxAttempts);err!=nil{rows.Close();return RunRunning,err};total++;switch status{case "succeeded":succeeded++;case "failed":if retryable&&attempt<maxAttempts{active=true}else{failed++};default:active=true}}
	if err:=rows.Close();err!=nil{return RunRunning,err}
	state:=RunRunning;if total==0{state=RunRunning}else if active{state=RunRunning}else if succeeded==total{state=RunSucceeded}else if failed==total{state=RunFailed}else if succeeded>0&&failed>0{state=RunPartialFailed}
	if state==RunSucceeded||state==RunPartialFailed||state==RunFailed{_,err=tx.ExecContext(ctx,`UPDATE runs SET status=?,finished_at=? WHERE id=?`,string(state),time.Now(),runID)}else{_,err=tx.ExecContext(ctx,`UPDATE runs SET status='running',finished_at=NULL WHERE id=?`,runID)};if err!=nil{return state,err}
	if err:=tx.Commit();err!=nil{return state,err};return state,nil
}

func (s *MySQLStore) ListQueuedBookRuns(ctx context.Context,limit int)([]WorkItem,error){
	if limit<=0{return nil,nil};rows,err:=s.db.QueryContext(ctx,`SELECT id,book_id,attempt FROM book_runs WHERE run_id IS NOT NULL AND status='queued' ORDER BY id LIMIT ?`,limit);if err!=nil{return nil,err};defer rows.Close();var out []WorkItem;for rows.Next(){var w WorkItem;if err:=rows.Scan(&w.BookRunID,&w.BookID,&w.Attempt);err!=nil{return nil,err};out=append(out,w)};return out,rows.Err()
}

func (s *MySQLStore) RecoverStaleBookRuns(ctx context.Context,now time.Time,limit int)([]WorkItem,error){
	if limit<=0{return nil,nil};rows,err:=s.db.QueryContext(ctx,`SELECT id FROM book_runs WHERE run_id IS NOT NULL AND status='running' AND lease_deadline IS NOT NULL AND lease_deadline<=? ORDER BY lease_deadline,id LIMIT ?`,now,limit);if err!=nil{return nil,err};var ids []int64;for rows.Next(){var id int64;if err:=rows.Scan(&id);err!=nil{rows.Close();return nil,err};ids=append(ids,id)};if err:=rows.Close();err!=nil{return nil,err}
	out:=make([]WorkItem,0,len(ids));for _,id:=range ids{item,created,err:=s.recoverOne(ctx,id,now);if err!=nil{return out,err};if created{out=append(out,item)}};return out,nil
}
func (s *MySQLStore) recoverOne(ctx context.Context,id int64,now time.Time)(WorkItem,bool,error){
	tx,err:=s.db.BeginTx(ctx,&sql.TxOptions{Isolation:sql.LevelReadCommitted});if err!=nil{return WorkItem{},false,err};defer tx.Rollback()
	var runID sql.NullInt64;var projectID,bookID int64;var attempt,maxAttempts int;var retryable bool;var status string;var deadline sql.NullTime
	if err:=tx.QueryRowContext(ctx,`SELECT run_id,batch_project_id,book_id,attempt,max_attempts,retryable,status,lease_deadline FROM book_runs WHERE id=? FOR UPDATE`,id).Scan(&runID,&projectID,&bookID,&attempt,&maxAttempts,&retryable,&status,&deadline);err!=nil{return WorkItem{},false,err}
	if !runID.Valid||status!="running"||!deadline.Valid||deadline.Time.After(now){return WorkItem{},false,nil}
	if !retryable||attempt>=maxAttempts{_,err:=tx.ExecContext(ctx,`UPDATE book_runs SET status='failed',retryable=0,error_code='worker_lease_expired',error_message='worker lease expired',finished_at=?,lease_deadline=NULL WHERE id=? AND status='running'`,now,id);if err!=nil{return WorkItem{},false,err};if err:=tx.Commit();err!=nil{return WorkItem{},false,err};return WorkItem{},false,nil}
	if _,err:=tx.ExecContext(ctx,`UPDATE book_runs SET status='failed',error_code='worker_lease_expired',error_message='worker lease expired',finished_at=?,lease_deadline=NULL WHERE id=? AND status='running'`,now,id);err!=nil{return WorkItem{},false,err}
	next:=attempt+1;res,err:=tx.ExecContext(ctx,`INSERT INTO book_runs (run_id,batch_project_id,book_id,attempt,max_attempts,retryable,status) VALUES (?,?,?,?,?,1,'queued') ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id)`,runID.Int64,projectID,bookID,next,maxAttempts);if err!=nil{return WorkItem{},false,err};nextID,err:=res.LastInsertId();if err!=nil{return WorkItem{},false,err};if err:=tx.Commit();err!=nil{return WorkItem{},false,err};return WorkItem{BookRunID:nextID,BookID:bookID,Attempt:next},true,nil
}
