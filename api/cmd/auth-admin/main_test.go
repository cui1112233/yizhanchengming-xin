package main

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

func TestValidateLocalFixtureTargetRejectsNonStagingOrNonLoopbackDSN(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		dsn         string
		wantErr     bool
	}{
		{name: "local staging", environment: "local", dsn: "fixture:secret@tcp(127.0.0.1:3306)/ycm_staging?parseTime=true", wantErr: false},
		{name: "wrong environment", environment: "production", dsn: "fixture:secret@tcp(127.0.0.1:3306)/ycm_staging", wantErr: true},
		{name: "production database", environment: "local", dsn: "fixture:secret@tcp(127.0.0.1:3306)/production", wantErr: true},
		{name: "remote host", environment: "local", dsn: "fixture:secret@tcp(10.0.0.8:3306)/ycm_staging", wantErr: true},
		{name: "localhost alias is not exact loopback contract", environment: "local", dsn: "fixture:secret@tcp(localhost:3306)/ycm_staging", wantErr: true},
		{name: "unix socket", environment: "local", dsn: "fixture:secret@unix(/tmp/mysql.sock)/ycm_staging", wantErr: true},
		{name: "malformed", environment: "local", dsn: "not-a-dsn", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLocalFixtureTarget(tt.environment, tt.dsn)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateLocalAcceptanceMemberHashesPasswordAndCreatesRestrictedUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	password := "local-member-password"
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO auth_users (username, display_name, password_hash, role, active) VALUES (?, ?, ?, 'member', TRUE)`)).
		WithArgs("ycm-staging-member-review", "Local acceptance member", passwordHashMatcher{password: password}).
		WillReturnResult(sqlmock.NewResult(9, 1))

	if err := createLocalAcceptanceMember(context.Background(), db, "ycm-staging-member-review", "Local acceptance member", password); err != nil {
		t.Fatalf("create member: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestLocalAcceptanceMemberRequiresReservedUsernameAndStrongPassword(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	for _, tc := range []struct {
		username string
		password string
	}{
		{username: "ordinary-member", password: "local-member-password"},
		{username: "ycm-staging-member-review", password: "short"},
	} {
		if err := createLocalAcceptanceMember(context.Background(), db, tc.username, "Local member", tc.password); err == nil {
			t.Fatalf("username=%q password length=%d expected error", tc.username, len(tc.password))
		}
	}
}

func TestDeleteLocalAcceptanceMemberOnlyDeletesReservedMember(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM auth_users WHERE username = ? AND role = 'member' AND team_id IS NULL`)).
		WithArgs("ycm-staging-member-review").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := deleteLocalAcceptanceMember(context.Background(), db, "ycm-staging-member-review"); err != nil {
		t.Fatalf("delete member: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteLocalAcceptanceMemberRejectsWrongRoleOrMissingUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM auth_users WHERE username = ? AND role = 'member' AND team_id IS NULL`)).
		WithArgs("ycm-staging-member-review").
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := deleteLocalAcceptanceMember(context.Background(), db, "ycm-staging-member-review"); err == nil {
		t.Fatal("expected guarded delete to reject missing or non-member user")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

type passwordHashMatcher struct {
	password string
}

func (m passwordHashMatcher) Match(value driver.Value) bool {
	hash, ok := value.(string)
	return ok && authn.VerifyPassword(hash, m.password) == nil && authn.VerifyPassword(hash, "wrong-password") != nil
}
