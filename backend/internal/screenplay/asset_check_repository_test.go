package screenplay

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"sync"
	"testing"
)

func newCheckFixture() AssetCheckJob {
	return AssetCheckJob{ID: "check-1", Owner: "alice", ProjectID: "p1", Actor: "alice", RequestKey: "request-1", InputDigest: "digest", Degree: DegreeStandard, Status: "queued"}
}
func candidate(name, prompt string) AssetCheckCandidate {
	return AssetCheckCandidate{Asset: Asset{Type: AssetCharacter, Name: name, Prompt: prompt, Identity: "身份", Description: "描述", Voice: "声音", Episodes: "1", Tag: "主角", Extra: map[string]string{"custom": "retained"}}, Original: json.RawMessage(`{"original_key":{"nested":true}}`)}
}
func TestAssetCheckMemoryRequestIsolationAndSingleActive(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	s.CreateProject(ctx, Project{ID: "p1", OwnerUsername: "alice"})
	s.CreateProject(ctx, Project{ID: "p1", OwnerUsername: "bob"})
	j, err := s.CreateAssetCheckJob(ctx, newCheckFixture())
	if err != nil {
		t.Fatal(err)
	}
	repeat, err := s.CreateAssetCheckJob(ctx, newCheckFixture())
	if err != nil || repeat.ID != j.ID {
		t.Fatal(repeat, err)
	}
	bad := newCheckFixture()
	bad.InputDigest = "changed"
	if _, err = s.CreateAssetCheckJob(ctx, bad); !errors.Is(err, ErrRevisionConflict) {
		t.Fatal(err)
	}
	bad = newCheckFixture()
	bad.RequestKey = "new"
	bad.ID = "check-2"
	if _, err = s.CreateAssetCheckJob(ctx, bad); !errors.Is(err, ErrInvalidState) {
		t.Fatal(err)
	}
	if _, err = s.GetAssetCheckJob(ctx, "bob", "p1", j.ID); !errors.Is(err, ErrNotFound) {
		t.Fatal("owner leak", err)
	}
	if _, err = s.MutateAssetCheckJob(ctx, "bob", "p1", j.ID, func(*AssetCheckJob) error { return nil }); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	jobs, _ := s.ListAssetCheckJobs(ctx, "bob", "p1", 10)
	if len(jobs) != 0 {
		t.Fatal(jobs)
	}
	_, err = s.MutateAssetCheckJob(ctx, "alice", "p1", j.ID, func(j *AssetCheckJob) error {
		j.Status = "cancelled"
		j.Candidates = []AssetCheckCandidate{candidate("A", "first")}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateAssetCheckJob(ctx, bad); err != nil {
		t.Fatal("completed/cancelled draft must allow a new check", err)
	}
}
func TestAssetCheckMemoryAtomicExactNamesAndIdempotentApply(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	s.CreateProject(ctx, Project{ID: "p1", OwnerUsername: "alice"})
	existing, _ := s.CreateAsset(ctx, Asset{Type: AssetCharacter, Name: " A ", Prompt: "do not replace", ProjectID: "p1", OwnerUsername: "alice"})
	j, _ := s.CreateAssetCheckJob(ctx, newCheckFixture())
	j, err := s.MutateAssetCheckJob(ctx, "alice", "p1", j.ID, func(j *AssetCheckJob) error {
		j.Status = "completed"
		j.Candidates = []AssetCheckCandidate{candidate("A", "wrong"), candidate("a", "first"), candidate(" a ", "second"), candidate("B", "third")}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ApplyAssetCheckJob(ctx, "alice", "p1", j.ID, j.Revision-1); !errors.Is(err, ErrRevisionConflict) {
		t.Fatal(err)
	}
	applied, err := s.ApplyAssetCheckJob(ctx, "alice", "p1", j.ID, j.Revision)
	if err != nil || applied.AddedCount != 2 || applied.SkippedCount != 2 {
		t.Fatal(applied, err)
	}
	rows, _ := s.ListAssets(ctx, "alice", "p1", AssetCharacter)
	if len(rows) != 3 || rows[0].ID != existing.ID || rows[0].Prompt != "do not replace" || rows[1].Name != "a" || rows[1].Prompt != "first" || rows[1].Extra["custom"] != "retained" || rows[1].Extra["original_asset_check"] != string(j.Candidates[1].Original) {
		t.Fatal(rows)
	}
	again, err := s.ApplyAssetCheckJob(ctx, "alice", "p1", j.ID, 0)
	if err != nil || again.Revision != applied.Revision || again.AddedCount != 2 {
		t.Fatal(again, err)
	}
	rows, _ = s.ListAssets(ctx, "alice", "p1", AssetCharacter)
	if len(rows) != 3 {
		t.Fatal(rows)
	}
	j.Candidates[0].Asset.Name = "external mutation"
	stored, _ := s.GetAssetCheckJob(ctx, "alice", "p1", j.ID)
	if stored.Candidates[0].Asset.Name != "A" {
		t.Fatal(stored)
	}
}
func TestAssetCheckMemoryConcurrentRequestsHaveOneWinner(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	s.CreateProject(ctx, Project{ID: "p1", OwnerUsername: "alice"})
	var wg sync.WaitGroup
	wins := make(chan bool, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			j := newCheckFixture()
			j.ID = fmt.Sprint(i)
			j.RequestKey = fmt.Sprint(i)
			_, err := s.CreateAssetCheckJob(ctx, j)
			wins <- err == nil
		}(i)
	}
	wg.Wait()
	close(wins)
	count := 0
	for win := range wins {
		if win {
			count++
		}
	}
	if count != 1 {
		t.Fatal(count)
	}
}
func checkJSONRows(j AssetCheckJob) *sqlmock.Rows {
	raw, _ := json.Marshal(j)
	return sqlmock.NewRows([]string{"state_json"}).AddRow(string(raw))
}
func expectCheckProject(m sqlmock.Sqlmock) {
	m.ExpectQuery("SELECT id FROM screenplay_projects").WithArgs("p1", "alice").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("p1"))
}

type checkStateMatcher struct {
	applied  bool
	status   string
	revision int64
}

func (m checkStateMatcher) Match(v driver.Value) bool {
	raw, ok := v.(string)
	if !ok {
		return false
	}
	var j AssetCheckJob
	return json.Unmarshal([]byte(raw), &j) == nil && j.Applied == m.applied && j.Status == m.status && j.Revision == m.revision
}
func TestAssetCheckMySQLApplyExactNamesAndRollback(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprint(fail), func(t *testing.T) {
			db, m, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			store := NewMySQLStore(db)
			j := newCheckFixture()
			j.Status = "completed"
			j.Revision = 4
			j.Candidates = []AssetCheckCandidate{candidate("A", "skip"), candidate("a", "first"), candidate(" a ", "skip2"), candidate("B", "third")}
			m.ExpectBegin()
			expectCheckProject(m)
			m.ExpectQuery("SELECT state_json FROM screenplay_asset_check_jobs.*FOR UPDATE").WithArgs("check-1", "alice", "p1").WillReturnRows(checkJSONRows(j))
			m.ExpectQuery("SELECT asset_type,name,sort_order FROM screenplay_assets").WithArgs("alice", "p1").WillReturnRows(sqlmock.NewRows([]string{"asset_type", "name", "sort_order"}).AddRow("character", "A", 7))
			for i, name := range []string{"a", "B"} {
				expect := m.ExpectExec("INSERT INTO screenplay_assets").WithArgs(sqlmock.AnyArg(), "p1", "alice", "character", "standard", name, "身份", "描述", map[string]string{"a": "first", "B": "third"}[name], "声音", "1", "主角", assetCheckOriginalMatcher{}, i+8, 0, sqlmock.AnyArg(), sqlmock.AnyArg())
				if fail && i == 1 {
					expect.WillReturnError(fmt.Errorf("insert interrupted"))
				} else {
					expect.WillReturnResult(sqlmock.NewResult(1, 1))
				}
			}
			if fail {
				m.ExpectRollback()
			} else {
				m.ExpectExec("UPDATE screenplay_asset_check_jobs").WithArgs("completed", true, 4, checkStateMatcher{true, "completed", 5}, sqlmock.AnyArg(), "check-1", "alice", "p1").WillReturnResult(sqlmock.NewResult(0, 1))
				m.ExpectCommit()
			}
			got, err := store.ApplyAssetCheckJob(context.Background(), "alice", "p1", "check-1", 4)
			if fail {
				if err == nil {
					t.Fatal("expected rollback")
				}
			} else if err != nil || got.AddedCount != 2 || got.SkippedCount != 2 {
				t.Fatal(got, err)
			}
			if err = m.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestAssetCheckMySQLCASAndAlreadyApplied(t *testing.T) {
	for _, applied := range []bool{false, true} {
		t.Run(fmt.Sprint(applied), func(t *testing.T) {
			db, m, _ := sqlmock.New()
			defer db.Close()
			j := newCheckFixture()
			j.Status = "completed"
			j.Revision = 5
			j.Applied = applied
			j.AddedCount = 2
			j.Candidates = []AssetCheckCandidate{candidate("a", "a")}
			m.ExpectBegin()
			expectCheckProject(m)
			m.ExpectQuery("SELECT state_json.*FOR UPDATE").WithArgs("check-1", "alice", "p1").WillReturnRows(checkJSONRows(j))
			if applied {
				m.ExpectCommit()
			} else {
				m.ExpectRollback()
			}
			got, err := NewMySQLStore(db).ApplyAssetCheckJob(context.Background(), "alice", "p1", "check-1", 0)
			if applied {
				if err != nil || got.AddedCount != 2 || got.Revision != 5 {
					t.Fatal(got, err)
				}
			} else if !errors.Is(err, ErrRevisionConflict) {
				t.Fatal(err)
			}
			if err = m.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestAssetCheckMySQLRequestReuseDigestAndOwner(t *testing.T) {
	db, m, _ := sqlmock.New()
	defer db.Close()
	s := NewMySQLStore(db)
	j := newCheckFixture()
	j.Revision = 1
	m.ExpectBegin()
	expectCheckProject(m)
	m.ExpectQuery("SELECT state_json.*request_key.*FOR UPDATE").WithArgs("alice", "p1", "request-1").WillReturnRows(checkJSONRows(j))
	m.ExpectCommit()
	if _, err := s.CreateAssetCheckJob(context.Background(), j); err != nil {
		t.Fatal(err)
	}
	m.ExpectBegin()
	expectCheckProject(m)
	m.ExpectQuery("SELECT state_json.*request_key.*FOR UPDATE").WithArgs("alice", "p1", "request-1").WillReturnRows(checkJSONRows(j))
	m.ExpectRollback()
	changed := j
	changed.InputDigest = "new"
	if _, err := s.CreateAssetCheckJob(context.Background(), changed); !errors.Is(err, ErrRevisionConflict) {
		t.Fatal(err)
	}
	m.ExpectQuery("SELECT state_json FROM screenplay_asset_check_jobs WHERE id").WithArgs("check-1", "bob", "p1").WillReturnError(sql.ErrNoRows)
	if _, err := s.GetAssetCheckJob(context.Background(), "bob", "p1", "check-1"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	if err := m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
func TestAssetCheckMySQLMutatePersistsStatusAndRevision(t *testing.T) {
	db, m, _ := sqlmock.New()
	defer db.Close()
	j := newCheckFixture()
	j.Revision = 3
	m.ExpectBegin()
	m.ExpectQuery("SELECT state_json.*FOR UPDATE").WithArgs("check-1", "alice", "p1").WillReturnRows(checkJSONRows(j))
	m.ExpectExec("UPDATE screenplay_asset_check_jobs").WithArgs("cancelled", false, 0, checkStateMatcher{false, "cancelled", 4}, sqlmock.AnyArg(), "check-1", "alice", "p1").WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectCommit()
	got, err := NewMySQLStore(db).MutateAssetCheckJob(context.Background(), "alice", "p1", "check-1", func(j *AssetCheckJob) error { j.Status = "cancelled"; j.Owner = "foreign"; return nil })
	if err != nil || got.Owner != "alice" || got.Revision != 4 {
		t.Fatal(got, err)
	}
	if err = m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

type assetCheckOriginalMatcher struct{}

func (assetCheckOriginalMatcher) Match(v driver.Value) bool {
	raw, ok := v.(string)
	if !ok {
		return false
	}
	var extra map[string]string
	if json.Unmarshal([]byte(raw), &extra) != nil {
		return false
	}
	return extra["custom"] == "retained" && extra["original_asset_check"] == `{"original_key":{"nested":true}}`
}
func TestAssetCheckMySQLCreateAndBlockActiveDraft(t *testing.T) {
	for _, blocked := range []bool{false, true} {
		t.Run(fmt.Sprint(blocked), func(t *testing.T) {
			db, m, _ := sqlmock.New()
			defer db.Close()
			j := newCheckFixture()
			m.ExpectBegin()
			expectCheckProject(m)
			m.ExpectQuery("SELECT state_json.*request_key.*FOR UPDATE").WithArgs("alice", "p1", "request-1").WillReturnError(sql.ErrNoRows)
			active := m.ExpectQuery("SELECT id FROM screenplay_asset_check_jobs.*status IN.*FOR UPDATE").WithArgs("alice", "p1")
			if blocked {
				active.WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("existing-draft"))
				m.ExpectRollback()
			} else {
				active.WillReturnError(sql.ErrNoRows)
				m.ExpectExec("INSERT INTO screenplay_asset_check_jobs").WithArgs("check-1", "alice", "p1", "request-1", "digest", "queued", false, 0, checkStateMatcher{false, "queued", 1}, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectCommit()
			}
			_, err := NewMySQLStore(db).CreateAssetCheckJob(context.Background(), j)
			if blocked {
				if !errors.Is(err, ErrInvalidState) {
					t.Fatal(err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if err = m.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestAssetCheckMySQLRunnableUsesIndexedStatusAndMutationErrorRollsBack(t *testing.T) {
	db, m, _ := sqlmock.New()
	defer db.Close()
	s := NewMySQLStore(db)
	j := newCheckFixture()
	m.ExpectQuery(`SELECT state_json FROM screenplay_asset_check_jobs WHERE status IN \('queued','running'\)`).WillReturnRows(checkJSONRows(j))
	jobs, err := s.ListRunnableAssetCheckJobs(context.Background())
	if err != nil || len(jobs) != 1 {
		t.Fatal(jobs, err)
	}
	m.ExpectBegin()
	m.ExpectQuery("SELECT state_json.*FOR UPDATE").WithArgs("check-1", "alice", "p1").WillReturnRows(checkJSONRows(j))
	m.ExpectRollback()
	_, err = s.MutateAssetCheckJob(context.Background(), "alice", "p1", "check-1", func(*AssetCheckJob) error { return ErrLockBusy })
	if !errors.Is(err, ErrLockBusy) {
		t.Fatal(err)
	}
	if err = m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAssetCheckApplyRejectsDiscardedAndRollsBackAppliedWriteFailure(t *testing.T) {
	ctx := context.Background()
	memory := NewMemoryStore()
	memory.CreateProject(ctx, Project{ID: "p1", OwnerUsername: "alice"})
	j, _ := memory.CreateAssetCheckJob(ctx, newCheckFixture())
	j, _ = memory.MutateAssetCheckJob(ctx, "alice", "p1", j.ID, func(j *AssetCheckJob) error {
		j.Status = "completed"
		j.Candidates = []AssetCheckCandidate{candidate("a", "first")}
		j.Discarded = true
		return nil
	})
	if _, err := memory.ApplyAssetCheckJob(ctx, "alice", "p1", j.ID, j.Revision); !errors.Is(err, ErrInvalidState) {
		t.Fatal(err)
	}
	db, m, _ := sqlmock.New()
	defer db.Close()
	j.Discarded = false
	m.ExpectBegin()
	expectCheckProject(m)
	m.ExpectQuery("SELECT state_json.*FOR UPDATE").WithArgs("check-1", "alice", "p1").WillReturnRows(checkJSONRows(j))
	m.ExpectQuery("SELECT asset_type,name,sort_order FROM screenplay_assets").WithArgs("alice", "p1").WillReturnRows(sqlmock.NewRows([]string{"asset_type", "name", "sort_order"}))
	m.ExpectExec("INSERT INTO screenplay_assets").WithArgs(sqlmock.AnyArg(), "p1", "alice", "character", "standard", "a", "身份", "描述", "first", "声音", "1", "主角", assetCheckOriginalMatcher{}, 1, 0, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	m.ExpectExec("UPDATE screenplay_asset_check_jobs").WillReturnError(fmt.Errorf("job journal unavailable"))
	m.ExpectRollback()
	if _, err := NewMySQLStore(db).ApplyAssetCheckJob(ctx, "alice", "p1", j.ID, j.Revision); err == nil {
		t.Fatal("journal failure must roll back asset insertion")
	}
	if err := m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAssetCheckRequestKeyLookupIsScopedAndCloned(t *testing.T) {
	ctx := context.Background()
	memory := NewMemoryStore()
	memory.CreateProject(ctx, Project{ID: "p1", OwnerUsername: "alice"})
	j := newCheckFixture()
	j.Existing = []Asset{{Name: "frozen", Extra: map[string]string{"original": "keep"}}}
	stored, err := memory.CreateAssetCheckJob(ctx, j)
	if err != nil {
		t.Fatal(err)
	}
	found, err := memory.FindAssetCheckJobByRequestKey(ctx, "alice", "p1", j.RequestKey)
	if err != nil || found.ID != stored.ID {
		t.Fatal(found, err)
	}
	found.Existing[0].Extra["original"] = "changed"
	frozen, _ := memory.FindAssetCheckJobByRequestKey(ctx, "alice", "p1", j.RequestKey)
	if frozen.Existing[0].Extra["original"] != "keep" {
		t.Fatal("lookup exposed mutable ledger")
	}
	for _, scope := range [][3]string{{"bob", "p1", j.RequestKey}, {"alice", "p2", j.RequestKey}, {"alice", "p1", "other-key"}} {
		if _, err = memory.FindAssetCheckJobByRequestKey(ctx, scope[0], scope[1], scope[2]); !errors.Is(err, ErrNotFound) {
			t.Fatal(scope, err)
		}
	}
	db, m, _ := sqlmock.New()
	defer db.Close()
	mysql := NewMySQLStore(db)
	m.ExpectQuery("SELECT state_json FROM screenplay_asset_check_jobs WHERE owner_username=\\? AND project_id=\\? AND request_key=\\?").WithArgs("alice", "p1", j.RequestKey).WillReturnRows(checkJSONRows(stored))
	found, err = mysql.FindAssetCheckJobByRequestKey(ctx, "alice", "p1", j.RequestKey)
	if err != nil || found.ID != stored.ID {
		t.Fatal(found, err)
	}
	m.ExpectQuery("SELECT state_json FROM screenplay_asset_check_jobs WHERE owner_username=\\? AND project_id=\\? AND request_key=\\?").WithArgs("bob", "p1", j.RequestKey).WillReturnError(sql.ErrNoRows)
	if _, err = mysql.FindAssetCheckJobByRequestKey(ctx, "bob", "p1", j.RequestKey); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	if err = m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
