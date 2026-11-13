package screenplay

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	mysql "github.com/go-sql-driver/mysql"
	"os"
	"qiantie/backend/internal/prompt"
	"qiantie/backend/internal/storage"
	"strings"
	"testing"
	"time"
)

// Opt-in only: this creates and drops a uniquely named disposable database.
func TestCreativeMySQLPersistence(t *testing.T) {
	dsn := os.Getenv("QIANTIE_CREATIVE_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("isolated MySQL test connection is not configured")
	}
	cfg, e := mysql.ParseDSN(dsn)
	if e != nil {
		t.Fatal("invalid MySQL test connection")
	}
	cfg.DBName = ""
	cfg.ParseTime = true
	admin, e := sql.Open("mysql", cfg.FormatDSN())
	if e != nil {
		t.Fatal("open MySQL test server")
	}
	defer admin.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	name := fmt.Sprintf("qiantie_creative_test_%d", time.Now().UnixNano())
	if _, e = admin.ExecContext(ctx, "CREATE DATABASE `"+name+"` CHARACTER SET utf8mb4"); e != nil {
		t.Fatal("create isolated test database:", e)
	}
	defer func() {
		if strings.HasPrefix(name, "qiantie_creative_test_") {
			_, _ = admin.ExecContext(context.Background(), "DROP DATABASE `"+name+"`")
		}
	}()
	cfg.DBName = name
	db, e := sql.Open("mysql", cfg.FormatDSN())
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	var serverVersion, connectionCharset string
	if e = db.QueryRowContext(ctx, "SELECT VERSION(), @@character_set_connection").Scan(&serverVersion, &connectionCharset); e != nil {
		t.Fatal(e)
	}
	t.Logf("MySQL server=%s connection_charset=%s", serverVersion, connectionCharset)
	if e = storage.RunMigrations(ctx, db, storage.ScreenplayMigrations()); e != nil {
		t.Fatal("migrate:", e)
	}
	if e = storage.RunMigrations(ctx, db, storage.ScreenplayMigrations()); e != nil {
		t.Fatal("repeat migration:", e)
	}
	s := NewMySQLStore(db)
	_, e = s.CreateProject(ctx, Project{ID: "p", OwnerUsername: "alice", Title: "真实MySQL验收", ScriptMode: "b"})
	if e != nil {
		t.Fatal(e)
	}
	v, e := s.SaveUserStyle(ctx, UserStyle{ID: "style", OwnerUsername: "alice", Name: "私人线稿", Prompt: "细线描绘", Images: []string{"/api/reference-assets/file/character/a/main"}}, nil)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.GetUserStyle(ctx, "bob", v.ID); !errors.Is(e, ErrNotFound) {
		t.Fatalf("owner isolation=%v", e)
	}
	pmt, e := prompt.ResolveCreativePrompt(prompt.CreativeScript144, "script")
	if e != nil {
		t.Fatal(e)
	}
	c := CreativeConfig{ScriptPromptID: pmt.ID, Style: &CreativeStyleSnapshot{ID: v.ID, Name: v.Name, Prompt: v.Prompt, Images: v.Images}, PromptSnapshots: map[string]CreativePromptSnapshot{"script": pmt}, ImageCoreInstruction: "核心要求"}
	p, e := s.SaveCreativeConfig(ctx, "alice", "p", c, 0)
	if e != nil || p.Revision != 1 {
		t.Fatalf("save rev=%d err=%v", p.Revision, e)
	}
	// A fresh store must read both selected text and server-only template body from MySQL.
	fresh := NewMySQLStore(db)
	p, e = fresh.GetProject(ctx, "alice", "p")
	if e != nil {
		t.Fatal(e)
	}
	if p.CreativeConfig == nil || p.CreativeConfig.PromptSnapshots["script"].Body != pmt.Body || p.StyleDesc != v.Prompt {
		t.Fatal("persistent snapshot mismatch")
	}
	v.Prompt = "新版本，不应影响项目"
	zero := int64(0)
	if _, e = s.SaveUserStyle(ctx, v, &zero); e != nil {
		t.Fatal(e)
	}
	p, _ = fresh.GetProject(ctx, "alice", "p")
	if p.StyleDesc != "细线描绘" {
		t.Fatal("frozen selection changed with library")
	}
	if _, e = s.SaveCreativeConfig(ctx, "alice", "p", c, 0); !errors.Is(e, ErrRevisionConflict) {
		t.Fatalf("stale conflict=%v", e)
	}
	c.Style = nil
	if p, e = s.SaveCreativeConfig(ctx, "alice", "p", c, 1); e != nil || p.StyleDesc != "" || p.ArtStyle != "" || p.SelectedStyle != "" {
		t.Fatalf("clear=%+v %v", p, e)
	}
	if _, e = s.SaveCreativeConfig(ctx, "bob", "p", c, 2); !errors.Is(e, ErrNotFound) {
		t.Fatalf("foreign project=%v", e)
	}
	if e = s.DeleteUserStyle(ctx, "alice", v.ID, 0); !errors.Is(e, ErrRevisionConflict) {
		t.Fatalf("stale style=%v", e)
	}
	if e = s.DeleteUserStyle(ctx, "alice", v.ID, 1); e != nil {
		t.Fatal(e)
	}
	raw := "  {\"scenes\":[{\"location\":\"书房\"}]}\r\n "
	editable := "# 第1集：下一站\n\n## 1-1：日 内 书房\n△ 林晚推开木窗。"
	if _, e = s.SaveScreenplay(ctx, Document{ID: "raw-doc", ProjectID: "p", OwnerUsername: "alice", Source: "b", Content: editable, RawResult: raw}, nil, nil, nil); e != nil {
		t.Fatal("save original and editable:", e)
	}
	doc, e := fresh.ActiveDocument(ctx, "alice", "p")
	if e != nil || doc.RawResult != raw || doc.Content != editable {
		t.Fatal("real MySQL original/editable roundtrip failed", e)
	}
	t.Log("real MySQL: migration repeat, owner isolation, frozen body/style persistence, revision conflict, clear and delete passed")
}
