package storage

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestBridgeUsersReuseSuccessfulResolution(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	users := &BridgeUsers{DB: db}
	mock.ExpectExec("INSERT INTO batch_factory_v12_bridge_users").WithArgs("reader", true).WillReturnResult(sqlmock.NewResult(0, 1))
	for i := 0; i < 3; i++ {
		if err := users.ResolveBridgeUser(context.Background(), "reader", true); err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestBridgeUsersRoleChangesAndFailedWritesDoNotReuseStaleCache(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	users := &BridgeUsers{DB: db}
	mock.ExpectExec("INSERT INTO batch_factory_v12_bridge_users").WithArgs("reader", true).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := users.ResolveBridgeUser(context.Background(), "reader", true); err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec("INSERT INTO batch_factory_v12_bridge_users").WithArgs("reader", false).WillReturnError(errors.New("database unavailable"))
	if err := users.ResolveBridgeUser(context.Background(), "reader", false); err == nil {
		t.Fatal("changed role must be saved before success")
	}
	mock.ExpectExec("INSERT INTO batch_factory_v12_bridge_users").WithArgs("reader", false).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := users.ResolveBridgeUser(context.Background(), "reader", false); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestBridgeUsersExpiredResolutionIsRefreshed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	users := &BridgeUsers{DB: db, cacheTTL: -time.Second}
	for i := 0; i < 2; i++ {
		mock.ExpectExec("INSERT INTO batch_factory_v12_bridge_users").WithArgs("reader", true).WillReturnResult(sqlmock.NewResult(0, 1))
		if err := users.ResolveBridgeUser(context.Background(), "reader", true); err != nil {
			t.Fatal(err)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestBridgeUsersConcurrentFirstRequestsShareOneWrite(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	users := &BridgeUsers{DB: db}
	mock.ExpectExec("INSERT INTO batch_factory_v12_bridge_users").WithArgs("reader", true).WillReturnResult(sqlmock.NewResult(0, 1))
	var wg sync.WaitGroup
	errors := make(chan error, 12)
	for i := 0; i < cap(errors); i++ {
		wg.Add(1)
		go func() { defer wg.Done(); errors <- users.ResolveBridgeUser(context.Background(), "reader", true) }()
	}
	wg.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
