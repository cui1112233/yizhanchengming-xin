package publishing

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLCreateAccountAndEncryptedCredentialAreAtomic(t *testing.T) {
	db,mock,err:=sqlmock.New(); if err!=nil{t.Fatal(err)}; defer db.Close()
	store:=NewMySQLStore(db); now:=time.Date(2026,10,5,12,0,0,0,time.UTC)
	credential:=EncryptedCredential{Ref:CredentialRef{ID:"cred_abc",Platform:"douyin",Name:"alice",CreatedAt:now,UpdatedAt:now},OwnerUserID:7,TeamID:3,KeyID:"key-id",Nonce:[]byte("123456789012"),Ciphertext:[]byte("encrypted-only")}
	account:=Account{OwnerUserID:7,TeamID:3,Platform:"douyin",DisplayName:"Alice 抖音",CredentialRefID:"cred_abc",Active:true,CreatedAt:now,UpdatedAt:now}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO publishing_credentials (ref, owner_user_id, team_id, platform, name, key_id, nonce, ciphertext, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)" )).
		WithArgs("cred_abc",int64(7),int64(3),"douyin","alice","key-id",[]byte("123456789012"),[]byte("encrypted-only"),now,now).
		WillReturnResult(sqlmock.NewResult(0,1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO publishing_accounts (owner_user_id, team_id, platform, display_name, credential_ref, active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)" )).
		WithArgs(int64(7),int64(3),"douyin","Alice 抖音","cred_abc",true,now,now).
		WillReturnResult(sqlmock.NewResult(41,1))
	mock.ExpectCommit()

	created,err:=store.CreateAccountWithCredential(context.Background(),account,credential); if err!=nil{t.Fatalf("create account: %v",err)}
	if created.ID!=41{t.Fatalf("id=%d want 41",created.ID)}
	if err:=mock.ExpectationsWereMet();err!=nil{t.Fatal(err)}
}

func TestMySQLIntentAndAuditAreAtomic(t *testing.T) {
	db,mock,err:=sqlmock.New(); if err!=nil{t.Fatal(err)}; defer db.Close()
	store:=NewMySQLStore(db); now:=time.Date(2026,10,5,12,5,0,0,time.UTC)
	intent:=Intent{BatchProjectID:21,BookID:22,PublishingAccountID:11,RequestedByUserID:7,Platform:"douyin",Status:IntentStatusPending,RequestedAt:now,UpdatedAt:now}
	audit:=Audit{BatchProjectID:21,AccountID:11,ActorUserID:7,Platform:"douyin",Action:"intent.created",Result:AuditResultAccepted,CreatedAt:now}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO publish_intents (batch_project_id, book_id, publishing_account_id, requested_by_user_id, platform, status, requested_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)" )).
		WithArgs(int64(21),int64(22),int64(11),int64(7),"douyin",IntentStatusPending,now,now).
		WillReturnResult(sqlmock.NewResult(73,1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO publish_audits (intent_id, batch_project_id, publishing_account_id, actor_user_id, platform, action, result, error_summary, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)" )).
		WithArgs(int64(73),int64(21),int64(11),int64(7),"douyin","intent.created",AuditResultAccepted,"",now).
		WillReturnResult(sqlmock.NewResult(91,1))
	mock.ExpectCommit()

	created,err:=store.CreateIntentWithAudit(context.Background(),intent,audit); if err!=nil{t.Fatalf("create intent: %v",err)}
	if created.ID!=73{t.Fatalf("id=%d want 73",created.ID)}
	if err:=mock.ExpectationsWereMet();err!=nil{t.Fatal(err)}
}

func TestMySQLIntentRollsBackWhenAuditInsertFails(t *testing.T) {
	db,mock,err:=sqlmock.New(); if err!=nil{t.Fatal(err)}; defer db.Close()
	store:=NewMySQLStore(db); now:=time.Date(2026,10,5,12,10,0,0,time.UTC)
	intent:=Intent{BatchProjectID:21,PublishingAccountID:11,RequestedByUserID:7,Platform:"douyin",Status:IntentStatusPending,RequestedAt:now,UpdatedAt:now}
	audit:=Audit{BatchProjectID:21,AccountID:11,ActorUserID:7,Platform:"douyin",Action:"intent.created",Result:AuditResultAccepted,CreatedAt:now}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO publish_intents").WillReturnResult(sqlmock.NewResult(73,1))
	mock.ExpectExec("INSERT INTO publish_audits").WillReturnError(context.DeadlineExceeded)
	mock.ExpectRollback()
	if _,err:=store.CreateIntentWithAudit(context.Background(),intent,audit);err==nil{t.Fatal("expected audit failure to roll back intent transaction")}
	if err:=mock.ExpectationsWereMet();err!=nil{t.Fatal(err)}
}
