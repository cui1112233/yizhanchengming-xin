package task9runtime

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestRuntimeFailurePersistsSafeCodeAndRedactedMessage(t *testing.T) {
	f:=newIntegrationFixture(t,1);ctx:=context.Background()
	run,_,err:=f.store.CreateRun(ctx,f.projectID,"safe-error",time.Now().Add(-time.Second),1);if err!=nil{t.Fatal(err)}
	claim,ok,err:=f.store.ClaimDueRun(ctx,time.Now());if err!=nil||!ok{t.Fatalf("claim ok=%v err=%v",ok,err)}
	items,err:=f.store.EnsureQueuedBooks(ctx,claim);if err!=nil||len(items)!=1{t.Fatalf("items=%v err=%v",items,err)}
	exec,ok,err:=f.store.Claim(ctx,items[0],"worker-safe",time.Now().Add(time.Second));if err!=nil||!ok{t.Fatalf("book claim ok=%v err=%v",ok,err)}
	secret:="Authorization: Bearer SUPER-SECRET-TOKEN"
	ok,err=f.store.Fail(ctx,exec,Failure{Code:"provider_error",Message:secret,Retryable:false});if err!=nil||!ok{t.Fatalf("fail ok=%v err=%v",ok,err)}
	var status,code,message string
	if err:=f.db.QueryRow(`SELECT status,error_code,error_message FROM book_runs WHERE id=?`,items[0].BookRunID).Scan(&status,&code,&message);err!=nil{t.Fatal(err)}
	if status!="failed"{t.Fatalf("status=%s",status)}
	if code!="internal_error"{t.Fatalf("error_code=%s want internal_error",code)}
	if message!="internal provider error"{t.Fatalf("error_message=%q",message)}
	if strings.Contains(message,"SUPER-SECRET")||strings.Contains(strings.ToLower(message),"bearer"){t.Fatalf("secret leaked: %q",message)}
	state,err:=f.store.AggregateRunStatus(ctx,run.ID);if err!=nil{t.Fatal(err)};if state!=RunFailed{t.Fatalf("run state=%s",state)}
}
