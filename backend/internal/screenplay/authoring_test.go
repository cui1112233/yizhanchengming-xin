package screenplay

import (
	"context"
	"errors"
	"testing"
)

func TestAuthoringRevisionAndAccess(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	s.CreateProject(ctx, Project{ID: "p", OwnerUsername: "a"})
	r := NewMemoryAuthoringRepository()
	svc := AuthoringService{Store: s, Repository: r}
	st, e := svc.Load(ctx, "a", "p")
	if e != nil {
		t.Fatal(e)
	}
	st.Chapters = []Chapter{{ID: "c", Title: "章", SourceText: "原文"}}
	if _, e = svc.Save(ctx, "a", "p", st, 0); e != nil {
		t.Fatal(e)
	}
	if _, e = svc.Save(ctx, "a", "p", st, 0); !errors.Is(e, ErrRevisionConflict) {
		t.Fatal(e)
	}
	if _, e = svc.Resolve(ctx, "b", "p", "read"); !errors.Is(e, ErrAccessDenied) {
		t.Fatal(e)
	}
	st, _ = svc.Load(ctx, "a", "p")
	st.Members = []MemberGrant{{Username: "b", Read: true, Edit: true}}
	svc.Save(ctx, "a", "p", st, st.Revision)
	if _, e = svc.Resolve(ctx, "b", "p", "edit"); e != nil {
		t.Fatal(e)
	}
	if _, e = svc.Resolve(ctx, "b", "p", "generate"); !errors.Is(e, ErrAccessDenied) {
		t.Fatal(e)
	}
}
func TestAuthoringRejectUnsupportedProposal(t *testing.T) {
	if e := ValidateAuthoringActions([]AuthoringAction{{Kind: "generate_video"}}); !errors.Is(e, ErrInvalidArgument) {
		t.Fatal(e)
	}
}
func TestAuthoringPublishChaptersKeepsEpisodeNumbers(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	s.CreateProject(ctx, Project{ID: "p", OwnerUsername: "a"})
	doc, e := PublishChapterCollection(ctx, s, "a", "p", []Chapter{{ID: "1", Title: "一", ScriptText: "# 第1集：一\n## 1-1：日 内 家\n△ 一"}, {ID: "2", Title: "二", ScriptText: "# 第1集：二\n## 1-1：夜 外 山\n△ 二"}})
	if e != nil {
		t.Fatal(e)
	}
	shots, _ := s.ListShots(ctx, "a", "p", doc.ID)
	if doc.EpisodeCount != 2 || len(shots) != 2 || shots[1].EpisodeNo != 2 {
		t.Fatalf("%#v %#v", doc, shots)
	}
}
func TestAuthoringContentInvalidatesDoctorAndCompletion(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	s.CreateProject(ctx, Project{ID: "p", OwnerUsername: "a"})
	b := NewBatch("b", "p", "a", "", PermissionBypass)
	b.Status = BatchCompleted
	for _, st := range b.Steps {
		st.Status = StepDone
	}
	b.Steps["6_storyboard_doctor"].Episodes = map[int]*EpisodeState{1: {EpisodeNo: 1, Status: StepBlocked, Reason: "旧问题"}}
	b.Steps["2_screenplay_doctor"].Review = &Review{Pass: false, BlockNextStep: true}
	s.CreateBatch(ctx, b)
	if e := InvalidateAuthoringContent(ctx, s, "a", "p", "document"); e != nil {
		t.Fatal(e)
	}
	b, _ = s.GetBatch(ctx, "a", "p")
	if b.Status != BatchPaused || b.Step("2_screenplay_doctor").Review != nil || b.Step("6_storyboard_doctor").Episodes[1].Status != StepStale {
		t.Fatal(b)
	}
}
