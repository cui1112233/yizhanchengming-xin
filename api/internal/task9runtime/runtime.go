package task9runtime

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/taskruntime"
)

type BookState string
type RunState string

const (
	BookPending BookState = "pending"
	BookQueued BookState = "queued"
	BookRunning BookState = "running"
	BookSucceeded BookState = "succeeded"
	BookFailed BookState = "failed"
	BookRetryableFailed BookState = "retryable_failed"

	RunPending RunState = "pending"
	RunRunning RunState = "running"
	RunSucceeded RunState = "succeeded"
	RunPartialFailed RunState = "partial_failed"
	RunFailed RunState = "failed"
)

var ErrStaleExecution = errors.New("task9 stale execution")

type WorkItem struct { BookRunID int64; BookID int64; Attempt int }
type Execution struct { BookRunID int64; Attempt int; FencingToken uint64; Owner string }
type Failure struct { Code string; Message string; Retryable bool }
type BookAttempt struct { BookRunID int64; BookID int64; Attempt int; State BookState; Retryable bool }
type RecoveryCandidate struct { BookRunID int64; BookID int64; Attempt int; FencingToken uint64; RunningSince time.Time; LeaseDeadline time.Time; MaxAttempts int; Retryable bool }
type RunClaim struct { RunID int64 }

type Store interface {
	Claim(context.Context, WorkItem, string, time.Time) (Execution, bool, error)
	Renew(context.Context, Execution, time.Time) (bool, error)
	Complete(context.Context, Execution) (bool, error)
	Fail(context.Context, Execution, Failure) (bool, error)
}
type Executor interface { Execute(context.Context, Execution) error }
type RuntimeCoordinator interface { Enqueue(context.Context, WorkItem) error }

type QueueCoordinator struct { queue taskruntime.Queue }
func NewQueueCoordinator(q taskruntime.Queue)*QueueCoordinator{return &QueueCoordinator{queue:q}}
func taskKey(item WorkItem) string { return fmt.Sprintf("bookrun:%d:book:%d",item.BookRunID,item.BookID) }
func decodeWorkItem(msg taskruntime.Message)(WorkItem,error){var br,b int64;if _,err:=fmt.Sscanf(msg.TaskKey,"bookrun:%d:book:%d",&br,&b);err!=nil{return WorkItem{},err};if br<=0{return WorkItem{},errors.New("invalid book run id")};return WorkItem{BookRunID:br,BookID:b,Attempt:msg.Attempt},nil}
func (c *QueueCoordinator) Enqueue(ctx context.Context,item WorkItem)error{return c.queue.Enqueue(ctx,taskruntime.Message{TaskKey:taskKey(item),Attempt:item.Attempt})}

type Worker struct { store Store; leases taskruntime.LeaseStore; executor Executor; owner string; leaseTTL time.Duration; now func() time.Time }
func NewWorker(store Store, leases taskruntime.LeaseStore, executor Executor, owner string, leaseTTL time.Duration, now func() time.Time) *Worker { if now==nil{now=time.Now};return &Worker{store:store,leases:leases,executor:executor,owner:owner,leaseTTL:leaseTTL,now:now} }

func (w *Worker) Process(ctx context.Context, item WorkItem) error {
	execution, ok, err := w.store.Claim(ctx, item, w.owner, w.now().Add(w.leaseTTL)); if err != nil || !ok { return err }
	if err := w.executor.Execute(ctx, execution); err != nil { code,message:=SafeError(err);_,failErr:=w.store.Fail(ctx,execution,Failure{Code:code,Message:message,Retryable:retryableError(err)});return failErr }
	ok,err=w.store.Complete(ctx,execution);if err==nil&&!ok{return ErrStaleExecution};return err
}

type retryability interface{ Retryable() bool }
func retryableError(err error)bool{var r retryability;if errors.As(err,&r){return r.Retryable()};return true}

func (w *Worker) RunOnce(ctx context.Context,q taskruntime.Queue,poll time.Duration)error{
	delivery,err:=q.Claim(ctx,w.owner,poll);if err!=nil{return err}
	item,err:=decodeWorkItem(delivery.Message);if err!=nil{_ = q.Ack(ctx,delivery);return err}
	execution,ok,err:=w.store.Claim(ctx,item,w.owner,w.now().Add(w.leaseTTL));if err!=nil{_ = q.Nack(ctx,delivery,0);return err};if !ok{return q.Ack(ctx,delivery)}
	var lease taskruntime.Lease;leased:=false
	if w.leases!=nil { lease,ok,err=w.leases.Claim(ctx,delivery.Message.TaskKey,w.owner,execution.FencingToken,w.leaseTTL);if err!=nil{_ = q.Ack(ctx,delivery);return err};if !ok{_ = q.Ack(ctx,delivery);return nil};leased=true }
	execErr:=w.executeWithHeartbeat(ctx,execution,lease,leased)
	var durable bool
	if execErr!=nil {code,message:=SafeError(execErr);durable,err=w.store.Fail(ctx,execution,Failure{Code:code,Message:message,Retryable:retryableError(execErr)})} else {durable,err=w.store.Complete(ctx,execution)}
	if leased { if _,relErr:=w.leases.Release(ctx,lease);relErr!=nil&&!errors.Is(relErr,taskruntime.ErrLeaseNotOwner)&&err==nil{err=relErr} }
	if err!=nil{_ = q.Nack(ctx,delivery,0);return err}
	if !durable{_ = q.Ack(ctx,delivery);return ErrStaleExecution}
	return q.Ack(ctx,delivery)
}

func (w *Worker) executeWithHeartbeat(ctx context.Context,e Execution,lease taskruntime.Lease,leased bool)error{
	if !leased||w.leases==nil||w.leaseTTL<=0{return w.executor.Execute(ctx,e)}
	execCtx,cancel:=context.WithCancel(ctx);defer cancel();interval:=w.leaseTTL/3;if interval<10*time.Millisecond{interval=10*time.Millisecond}
	done:=make(chan struct{});hbErr:=make(chan error,1)
	go func(){defer close(done);ticker:=time.NewTicker(interval);defer ticker.Stop();for{select{case <-execCtx.Done():return;case <-ticker.C:
		ok,err:=w.leases.Renew(execCtx,lease,w.leaseTTL);if err!=nil||!ok{if err==nil{err=taskruntime.ErrLeaseNotOwner};select{case hbErr<-err:default:{}};cancel();return}
		ok,err=w.store.Renew(execCtx,e,w.now().Add(w.leaseTTL));if err!=nil||!ok{if err==nil{err=ErrStaleExecution};select{case hbErr<-err:default:{}};cancel();return}
	}}}()
	err:=w.executor.Execute(execCtx,e);cancel();<-done
	select{case h:=<-hbErr:if err==nil{return h};default:}
	return err
}

func (w *Worker) Run(ctx context.Context,q taskruntime.Queue,poll time.Duration)error{
	for { if err:=ctx.Err();err!=nil{return err};err:=w.RunOnce(ctx,q,poll);if err==nil||errors.Is(err,taskruntime.ErrQueueEmpty)||errors.Is(err,ErrStaleExecution){continue};if errors.Is(err,context.Canceled)||errors.Is(err,context.DeadlineExceeded){return err};return err }
}

func AggregateRunStatus(states []BookState) RunState { if len(states)==0{return RunPending};succeeded,failed:=0,0;for _,state:=range states{switch state{case BookPending,BookQueued,BookRunning,BookRetryableFailed:return RunRunning;case BookSucceeded:succeeded++;case BookFailed:failed++}};if succeeded==len(states){return RunSucceeded};if failed==len(states){return RunFailed};if succeeded>0&&failed>0{return RunPartialFailed};return RunRunning }
func SafeError(err error)(string,string){if err==nil{return "",""};message:=err.Error();lower:=strings.ToLower(message);for _,marker:=range []string{"authorization","bearer ","password","cookie","access_token","refresh_token","api key","apikey","secret","dsn=","@tcp("}{if strings.Contains(lower,marker){return "internal_error","internal provider error"}};if len(message)>1024{message=message[:1024]};return "execution_error",message}
func PlanRetry(attempts []BookAttempt,maxAttempts int)[]WorkItem{latest:=map[int64]BookAttempt{};for _,a:=range attempts{if prev,ok:=latest[a.BookID];!ok||a.Attempt>prev.Attempt{latest[a.BookID]=a}};out:=make([]WorkItem,0);for _,a:=range latest{if a.State==BookFailed&&a.Retryable&&a.Attempt<maxAttempts{out=append(out,WorkItem{BookRunID:a.BookRunID,BookID:a.BookID,Attempt:a.Attempt+1})}};return out}
func ShouldRetry(a BookAttempt,maxAttempts int)bool{return a.State==BookFailed&&a.Retryable&&a.Attempt<maxAttempts}
func PlanRecovery(candidates []RecoveryCandidate,now time.Time)[]WorkItem{out:=make([]WorkItem,0);for _,c:=range candidates{if c.LeaseDeadline.IsZero()||!c.LeaseDeadline.Before(now)||c.Attempt>=c.MaxAttempts||!c.Retryable{continue};out=append(out,WorkItem{BookRunID:c.BookRunID,BookID:c.BookID,Attempt:c.Attempt+1})};return out}

type SchedulerStore interface { ClaimDueRun(context.Context,time.Time)(RunClaim,bool,error); EnsureQueuedBooks(context.Context,RunClaim)([]WorkItem,error) }
type Scheduler struct { store SchedulerStore; coordinator RuntimeCoordinator; now func() time.Time }
func NewScheduler(store SchedulerStore,coordinator RuntimeCoordinator,now func() time.Time)*Scheduler{if now==nil{now=time.Now};return &Scheduler{store:store,coordinator:coordinator,now:now}}
func (s *Scheduler) Tick(ctx context.Context,limit int)(int,error){if limit<=0{return 0,nil};claimed:=0;for i:=0;i<limit;i++{run,ok,err:=s.store.ClaimDueRun(ctx,s.now());if err!=nil{return claimed,err};if !ok{break};items,err:=s.store.EnsureQueuedBooks(ctx,run);if err!=nil{return claimed,err};if s.coordinator!=nil{for _,item:=range items{if err:=s.coordinator.Enqueue(ctx,item);err!=nil{return claimed,err}}};claimed++};return claimed,nil}
func (s *Scheduler) Run(ctx context.Context,interval time.Duration,limit int)error{if interval<=0{interval=time.Second};for{if err:=ctx.Err();err!=nil{return err};if _,err:=s.Tick(ctx,limit);err!=nil{return err};timer:=time.NewTimer(interval);select{case <-ctx.Done():if !timer.Stop(){<-timer.C};return ctx.Err();case <-timer.C:}}}

type RecoveryStore interface { ListQueuedBookRuns(context.Context,int)([]WorkItem,error); RecoverStaleBookRuns(context.Context,time.Time,int)([]WorkItem,error) }
type Recovery struct { store RecoveryStore; coordinator RuntimeCoordinator }
func NewRecovery(store RecoveryStore,coordinator RuntimeCoordinator)*Recovery{return &Recovery{store:store,coordinator:coordinator}}
func (r *Recovery) Rebuild(ctx context.Context,now time.Time,limit int)(int,error){if limit<=0{return 0,nil};queued,err:=r.store.ListQueuedBookRuns(ctx,limit);if err!=nil{return 0,err};remaining:=limit-len(queued);var recovered []WorkItem;if remaining>0{recovered,err=r.store.RecoverStaleBookRuns(ctx,now,remaining);if err!=nil{return 0,err}};seen:=map[int64]struct{}{};count:=0;for _,item:=range append(queued,recovered...){if _,ok:=seen[item.BookRunID];ok{continue};seen[item.BookRunID]=struct{}{};if r.coordinator!=nil{if err:=r.coordinator.Enqueue(ctx,item);err!=nil{return count,err}};count++};return count,nil}
