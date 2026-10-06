package shuihuo

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"testing"
)

type memoryStore struct {
	nextProjectID int64
	nextSegmentID int64
	projects      map[int64]Project
	segments      map[int64][]Segment
}

func newMemoryStore() *memoryStore {
	return &memoryStore{nextProjectID: 1, nextSegmentID: 1, projects: map[int64]Project{}, segments: map[int64][]Segment{}}
}

func (m *memoryStore) CreateProject(_ context.Context, p Project) (Project, error) {
	p.ID = m.nextProjectID
	m.nextProjectID++
	m.projects[p.ID] = p
	return p, nil
}
func (m *memoryStore) ListProjects(_ context.Context, actor Actor) ([]Project, error) {
	out := []Project{}
	for _, p := range m.projects {
		if CanAccess(actor, p) {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out, nil
}
func (m *memoryStore) GetProject(_ context.Context, id int64) (Project, error) {
	p, ok := m.projects[id]
	if !ok {
		return Project{}, ErrNotFound
	}
	return p, nil
}
func (m *memoryStore) SaveSourceAndReset(_ context.Context, id int64, source string) (Project, error) {
	p, ok := m.projects[id]
	if !ok {
		return Project{}, ErrNotFound
	}
	p.SourceText = source
	p.SegmentationStatus = SegmentationDraft
	m.projects[id] = p
	delete(m.segments, id)
	return p, nil
}
func (m *memoryStore) DeleteProject(_ context.Context, id int64) error {
	if _, ok := m.projects[id]; !ok {
		return ErrNotFound
	}
	delete(m.projects, id)
	delete(m.segments, id)
	return nil
}
func (m *memoryStore) ReplaceSegments(_ context.Context, projectID int64, candidates []Candidate) ([]Segment, error) {
	p, ok := m.projects[projectID]
	if !ok {
		return nil, ErrNotFound
	}
	rows := make([]Segment, 0, len(candidates))
	for i, c := range candidates {
		row := Segment{ID: m.nextSegmentID, ProjectID: projectID, Position: i + 1, SourceText: c.Text, SubtitleText: c.Text, Speaker: normalizeSpeaker(c.Speaker)}
		m.nextSegmentID++
		rows = append(rows, row)
	}
	m.segments[projectID] = rows
	p.SegmentationStatus = SegmentationConfirmed
	m.projects[projectID] = p
	return append([]Segment(nil), rows...), nil
}
func (m *memoryStore) ListSegments(_ context.Context, projectID int64) ([]Segment, error) {
	if _, ok := m.projects[projectID]; !ok {
		return nil, ErrNotFound
	}
	return append([]Segment(nil), m.segments[projectID]...), nil
}
func (m *memoryStore) GetSegment(_ context.Context, segmentID int64) (Segment, error) {
	for _, rows := range m.segments {
		for _, row := range rows {
			if row.ID == segmentID {
				return row, nil
			}
		}
	}
	return Segment{}, ErrNotFound
}
func (m *memoryStore) CreateSegment(_ context.Context, projectID int64, input SegmentInput) (Segment, error) {
	if _, ok := m.projects[projectID]; !ok {
		return Segment{}, ErrNotFound
	}
	row := Segment{ID: m.nextSegmentID, ProjectID: projectID, Position: len(m.segments[projectID]) + 1, SourceText: input.SourceText, SubtitleText: input.SubtitleText, Speaker: normalizeSpeaker(input.Speaker)}
	if row.SubtitleText == "" {
		row.SubtitleText = row.SourceText
	}
	m.nextSegmentID++
	m.segments[projectID] = append(m.segments[projectID], row)
	return row, nil
}
func (m *memoryStore) UpdateSegment(_ context.Context, segmentID int64, input SegmentInput) (Segment, error) {
	for projectID, rows := range m.segments {
		for i, row := range rows {
			if row.ID == segmentID {
				row.SourceText = input.SourceText
				row.SubtitleText = input.SubtitleText
				if row.SubtitleText == "" {
					row.SubtitleText = row.SourceText
				}
				row.Speaker = normalizeSpeaker(input.Speaker)
				rows[i] = row
				m.segments[projectID] = rows
				return row, nil
			}
		}
	}
	return Segment{}, ErrNotFound
}
func (m *memoryStore) DeleteSegment(_ context.Context, segmentID int64) error {
	for projectID, rows := range m.segments {
		for i, row := range rows {
			if row.ID == segmentID {
				rows = append(rows[:i], rows[i+1:]...)
				for j := range rows {
					rows[j].Position = j + 1
				}
				m.segments[projectID] = rows
				return nil
			}
		}
	}
	return ErrNotFound
}
func (m *memoryStore) ReorderSegments(_ context.Context, projectID int64, ids []int64) ([]Segment, error) {
	rows := m.segments[projectID]
	if len(rows) != len(ids) {
		return nil, ErrInvalid
	}
	byID := map[int64]Segment{}
	for _, row := range rows {
		byID[row.ID] = row
	}
	next := make([]Segment, 0, len(ids))
	seen := map[int64]bool{}
	for i, id := range ids {
		row, ok := byID[id]
		if !ok || seen[id] {
			return nil, ErrInvalid
		}
		seen[id] = true
		row.Position = i + 1
		next = append(next, row)
	}
	m.segments[projectID] = next
	return append([]Segment(nil), next...), nil
}

type fakeSmartSegmenter struct {
	candidates []Candidate
	err        error
}

func (f fakeSmartSegmenter) Segment(context.Context, string) ([]Candidate, error) {
	return f.candidates, f.err
}

func TestServiceProjectLifecycleAndOwnership(t *testing.T) {
	ctx := context.Background()
	store := newMemoryStore()
	svc := NewService(store, nil)
	owner := Actor{UserID: 1, TeamID: 10}
	teammate := Actor{UserID: 2, TeamID: 10}
	outsider := Actor{UserID: 3, TeamID: 20}
	created, err := svc.CreateProject(ctx, owner, CreateProjectInput{Name: "作品A", ProductionMode: "commentary"})
	if err != nil {
		t.Fatal(err)
	}
	if created.Project.OwnerUserID != 1 || created.Project.TeamID != 10 || created.Project.Name != "作品A" {
		t.Fatalf("unexpected project: %#v", created.Project)
	}
	if _, err := svc.GetProject(ctx, teammate, created.Project.ID); err != nil {
		t.Fatalf("team access should work: %v", err)
	}
	if _, err := svc.GetProject(ctx, outsider, created.Project.ID); !errors.Is(err, ErrForbidden) {
		t.Fatalf("outsider error=%v, want forbidden", err)
	}
	saved, err := svc.ReplaceSource(ctx, owner, created.Project.ID, "第一段\n\n第二段")
	if err != nil {
		t.Fatal(err)
	}
	if saved.Project.SourceText != "第一段\n\n第二段" || saved.Project.SegmentationStatus != SegmentationDraft {
		t.Fatalf("source not persisted/reset: %#v", saved.Project)
	}
	if err := svc.DeleteProject(ctx, owner, created.Project.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GetProject(ctx, owner, created.Project.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("after delete error=%v", err)
	}
}

func TestSegmentationModesAndConfirm(t *testing.T) {
	ctx := context.Background()
	store := newMemoryStore()
	smart := fakeSmartSegmenter{candidates: []Candidate{{Text: "AI一"}, {Text: "AI二", Speaker: "角色甲"}}}
	svc := NewService(store, smart)
	actor := Actor{UserID: 1}
	created, _ := svc.CreateProject(ctx, actor, CreateProjectInput{Name: "作品"})
	id := created.Project.ID
	_, _ = svc.ReplaceSource(ctx, actor, id, "第一段\n\n第二段\n第三行")
	paragraphs, err := svc.ParagraphSegmentation(ctx, actor, id, SegmentationInput{})
	if err != nil {
		t.Fatal(err)
	}
	wantParagraphs := []Candidate{{Text: "第一段", Speaker: "旁白"}, {Text: "第二段\n第三行", Speaker: "旁白"}}
	if !reflect.DeepEqual(paragraphs, wantParagraphs) {
		t.Fatalf("paragraphs=%#v", paragraphs)
	}
	fixed, err := svc.FixedSegmentation(ctx, actor, id, FixedSegmentationInput{LinesPerSegment: 2})
	if err != nil {
		t.Fatal(err)
	}
	wantFixed := []Candidate{{Text: "第一段\n第二段", Speaker: "旁白"}, {Text: "第三行", Speaker: "旁白"}}
	if !reflect.DeepEqual(fixed, wantFixed) {
		t.Fatalf("fixed=%#v", fixed)
	}
	imported, err := svc.ImportSegmentation(ctx, actor, id, SegmentationInput{Text: "1\t导入甲\n2\t导入乙"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(imported, []Candidate{{Text: "导入甲", Speaker: "旁白"}, {Text: "导入乙", Speaker: "旁白"}}) {
		t.Fatalf("imported=%#v", imported)
	}
	ai, err := svc.SmartSegmentation(ctx, actor, id, SegmentationInput{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ai, []Candidate{{Text: "AI一", Speaker: "旁白"}, {Text: "AI二", Speaker: "角色甲"}}) {
		t.Fatalf("smart=%#v", ai)
	}
	confirmed, err := svc.ConfirmSegmentation(ctx, actor, id, ai)
	if err != nil {
		t.Fatal(err)
	}
	if confirmed.Project.SegmentationStatus != SegmentationConfirmed || len(confirmed.Segments) != 2 || confirmed.Segments[1].Speaker != "角色甲" {
		t.Fatalf("confirmed=%#v", confirmed)
	}
}

func TestSegmentCRUDAndReorder(t *testing.T) {
	ctx := context.Background()
	store := newMemoryStore()
	svc := NewService(store, nil)
	actor := Actor{UserID: 1}
	created, _ := svc.CreateProject(ctx, actor, CreateProjectInput{Name: "作品"})
	id := created.Project.ID
	confirmed, err := svc.ConfirmSegmentation(ctx, actor, id, []Candidate{{Text: "一"}, {Text: "二"}, {Text: "三"}})
	if err != nil {
		t.Fatal(err)
	}
	fourth, err := svc.CreateSegment(ctx, actor, id, SegmentInput{SourceText: "四", Speaker: "旁白"})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := svc.UpdateSegment(ctx, actor, fourth.ID, SegmentInput{SourceText: "四改", SubtitleText: "字幕四", Speaker: "角色乙"})
	if err != nil {
		t.Fatal(err)
	}
	if updated.SourceText != "四改" || updated.SubtitleText != "字幕四" || updated.Speaker != "角色乙" {
		t.Fatalf("updated=%#v", updated)
	}
	order := []int64{fourth.ID, confirmed.Segments[0].ID, confirmed.Segments[1].ID, confirmed.Segments[2].ID}
	reordered, err := svc.ReorderSegments(ctx, actor, id, order)
	if err != nil {
		t.Fatal(err)
	}
	for i, row := range reordered.Segments {
		if row.ID != order[i] || row.Position != i+1 {
			t.Fatalf("bad reorder: %#v", reordered.Segments)
		}
	}
	if err := svc.DeleteSegment(ctx, actor, confirmed.Segments[1].ID); err != nil {
		t.Fatal(err)
	}
	read, err := svc.GetProject(ctx, actor, id)
	if err != nil {
		t.Fatal(err)
	}
	if len(read.Segments) != 3 {
		t.Fatalf("segments after delete=%d", len(read.Segments))
	}
	for i, row := range read.Segments {
		if row.Position != i+1 {
			t.Fatalf("positions not compact: %#v", read.Segments)
		}
	}
}

func TestSmartSegmentationUnavailable(t *testing.T) {
	ctx := context.Background()
	store := newMemoryStore()
	svc := NewService(store, nil)
	actor := Actor{UserID: 1}
	created, _ := svc.CreateProject(ctx, actor, CreateProjectInput{Name: "作品"})
	_, _ = svc.ReplaceSource(ctx, actor, created.Project.ID, "正文")
	if _, err := svc.SmartSegmentation(ctx, actor, created.Project.ID, SegmentationInput{}); !errors.Is(err, ErrSmartUnavailable) {
		t.Fatalf("error=%v", err)
	}
}
