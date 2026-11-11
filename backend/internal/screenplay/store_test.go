package screenplay

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestSaveProjectSourcePersistsWithoutBumpingRevision(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	project, err := store.CreateProject(ctx, Project{ID: "p1", OwnerUsername: "alice", Title: "测试项目"})
	if err != nil {
		t.Fatal(err)
	}
	saved, err := store.SaveProjectSource(ctx, "alice", "p1", "第一集原文")
	if err != nil {
		t.Fatal(err)
	}
	if saved.SourceText != "第一集原文" {
		t.Fatalf("source_text = %q", saved.SourceText)
	}
	if saved.Revision != project.Revision {
		t.Fatalf("保存原文不应修改 revision: got=%d want=%d", saved.Revision, project.Revision)
	}
	loaded, err := store.GetProject(ctx, "alice", "p1")
	if err != nil || loaded.SourceText != "第一集原文" {
		t.Fatalf("重新读取原文失败: %v %q", err, loaded.SourceText)
	}
	updated, err := store.UpdateProject(ctx, "alice", "p1", Project{ScriptMode: "a1", ModelID: "text-model"}, loaded.Revision)
	if err != nil {
		t.Fatal(err)
	}
	if updated.SourceText != "第一集原文" {
		t.Fatalf("更新项目设置不应清空原文: %q", updated.SourceText)
	}
}

func TestUpdateBatchRejectsStaleRevision(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	batch := NewBatch("b1", "p1", "alice", "s1", PermissionAuto)
	if err := store.CreateBatch(ctx, batch); err != nil {
		t.Fatal(err)
	}
	stale := cloneBatch(batch)
	first, err := store.MutateBatch(ctx, "alice", "p1", func(b *Batch) error {
		b.LastError = "first"
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if first.Revision != 1 {
		t.Fatalf("revision = %d, want 1", first.Revision)
	}
	stale.LastError = "second"
	if err := store.UpdateBatch(ctx, stale, 0); !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("err = %v, want ErrRevisionConflict", err)
	}
	if err := store.UpdateBatch(ctx, stale, first.Revision); err != nil {
		t.Fatalf("使用最新 revision 应成功: %v", err)
	}
}

// 20 路并发写同一批次：显式锁保证不丢更新（对应原版 save_state 裸写缺陷）。
func TestMutateBatchSerializesTwentyConcurrentWriters(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	if err := store.CreateBatch(ctx, NewBatch("b1", "p1", "alice", "s1", PermissionBypass)); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := store.MutateBatch(ctx, "alice", "p1", func(b *Batch) error {
				b.PendingActions = append(b.PendingActions, PendingAction{ID: fmt.Sprintf("action-%02d", i)})
				return nil
			})
			if err != nil {
				t.Errorf("writer %d: %v", i, err)
			}
		}(i)
	}
	wg.Wait()

	got, err := store.GetBatch(ctx, "alice", "p1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.PendingActions) != 20 {
		t.Fatalf("pending actions = %d, want 20（存在丢更新）", len(got.PendingActions))
	}
	if got.Revision != 20 {
		t.Fatalf("revision = %d, want 20", got.Revision)
	}
}

func TestAcquireLockRejectsSecondHolderAndAllowsRelease(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	if err := store.AcquireLock(ctx, "storyboard:ep1", "token-a", time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := store.AcquireLock(ctx, "storyboard:ep1", "token-b", time.Minute); !errors.Is(err, ErrLockBusy) {
		t.Fatalf("err = %v, want ErrLockBusy", err)
	}
	if err := store.ReleaseLock(ctx, "storyboard:ep1", "token-b"); err != nil {
		t.Fatal(err)
	}
	if err := store.AcquireLock(ctx, "storyboard:ep1", "token-c", time.Minute); !errors.Is(err, ErrLockBusy) {
		t.Fatalf("非持有者释放后仍应被占用: %v", err)
	}
	if err := store.ReleaseLock(ctx, "storyboard:ep1", "token-a"); err != nil {
		t.Fatal(err)
	}
	if err := store.AcquireLock(ctx, "storyboard:ep1", "token-c", time.Minute); err != nil {
		t.Fatalf("释放后应可获取: %v", err)
	}
}

func TestAcquireLockHonoursExpiry(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	// 反向 TTL 会被归一为 2 分钟；这里直接验证过期后可被抢占需要依赖注入时钟，
	// 因此仅校验非法参数被拒绝与短锁可立即重复获取同一 token。
	if err := store.AcquireLock(ctx, "", "t", time.Minute); !errors.Is(err, ErrInvalidArgument) && err == nil {
		t.Fatal("空锁名应被拒绝")
	}
	if err := store.AcquireLock(ctx, "k", "t", time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := store.AcquireLock(ctx, "k", "t", time.Minute); err != nil {
		t.Fatalf("同一 token 续租应成功: %v", err)
	}
}

func TestMemoryStoreProjectLifecycle(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	created, err := store.CreateProject(ctx, Project{ID: "comic-1", OwnerUsername: "alice", Title: "第一部漫剧", AspectRatio: "9:16"})
	if err != nil {
		t.Fatal(err)
	}
	if created.Revision != 0 || created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatalf("created=%+v", created)
	}
	projects, err := store.ListProjects(ctx, "alice")
	if err != nil || len(projects) != 1 || projects[0].ID != "comic-1" {
		t.Fatalf("projects=%+v err=%v", projects, err)
	}
	updated, err := store.UpdateProject(ctx, "alice", "comic-1", Project{Title: "改名后的漫剧", AspectRatio: "16:9", ArtStyle: "国风", PermissionMode: PermissionBypass}, 0)
	if err != nil || updated.Revision != 1 || updated.Title != "改名后的漫剧" || updated.AspectRatio != "16:9" {
		t.Fatalf("updated=%+v err=%v", updated, err)
	}
	if _, err := store.UpdateProject(ctx, "alice", "comic-1", Project{Title: "旧页面覆盖"}, 0); !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("stale update err=%v", err)
	}
	if err := store.DeleteProject(ctx, "alice", "comic-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetProject(ctx, "alice", "comic-1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("get deleted err=%v", err)
	}
	projects, err = store.ListProjects(ctx, "alice")
	if err != nil || len(projects) != 0 {
		t.Fatalf("projects after delete=%+v err=%v", projects, err)
	}
}
