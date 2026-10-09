package generation

import (
	"context"
	"errors"
	"testing"
)

func TestSplitStoryboardPreservesCardBoundariesAndFallback(t *testing.T) {
	cards := splitStoryboard("### 分镜一（总时长：10s）\n甲进入房间\n---\n### 分镜二（总时长：10s）\n乙回头")
	if len(cards) != 2 || cards[0].Position != 1 || cards[1].Title != "分镜二（总时长：10s）" || cards[0].Content != "甲进入房间" {
		t.Fatalf("cards = %#v", cards)
	}
	fallback := splitStoryboard("没有标题的导演内容")
	if len(fallback) != 1 || fallback[0].Title != "分镜一" || fallback[0].Content != "没有标题的导演内容" {
		t.Fatalf("fallback = %#v", fallback)
	}
}

func TestRecompileStoryboardRejectsRuntimeBeforeStoryboardWrite(t *testing.T) {
	base := newMemoryStore()
	base.seedPrompts()
	base.bookRuns = append(base.bookRuns, BookRun{ID: 90, RunID: 80, BatchProjectID: 3, BookID: 11, Status: StatusCompleted})
	base.stageRuns = append(base.stageRuns, StageRun{ID: 91, BookRunID: 90, BookID: 11, Stage: StageDirector, Status: StatusCompleted, OutputText: "director"})
	store := &runtimeStoryboardStore{memoryStore: base}
	service := NewService(store, &fakeProvider{errors: map[int]error{}}, nil)
	if _, err := service.RecompileStoryboard(context.Background(), 3, 11, "legacy"); !errors.Is(err, ErrConflict) {
		t.Fatalf("err=%v", err)
	}
	if store.creates != 0 {
		t.Fatalf("runtime recompile wrote storyboard %d times", store.creates)
	}
}

type runtimeStoryboardStore struct {
	*memoryStore
	creates int
}

func (*runtimeStoryboardStore) GetStoryboard(context.Context, int64, int64) (StoryboardDocument, error) {
	return StoryboardDocument{}, ErrNotFound
}
func (s *runtimeStoryboardStore) CreateStoryboard(_ context.Context, doc StoryboardDocument) (StoryboardDocument, error) {
	s.creates++
	return doc, nil
}
func (*runtimeStoryboardStore) SaveStoryboardCard(context.Context, int64, int64, StoryboardCard, int) (StoryboardDocument, error) {
	return StoryboardDocument{}, ErrUnavailable
}
func (*runtimeStoryboardStore) DeleteStoryboardCard(context.Context, int64, int64, int64, int) (StoryboardDocument, error) {
	return StoryboardDocument{}, ErrUnavailable
}
func (*runtimeStoryboardStore) ReorderStoryboard(context.Context, int64, int64, []int64, int) (StoryboardDocument, error) {
	return StoryboardDocument{}, ErrUnavailable
}
