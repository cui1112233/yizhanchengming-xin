package worker

import (
    "context"
    "testing"

    one21 "github.com/cui1112233/yizhanchengming-xin/api/internal/integrations/121"
    "github.com/cui1112233/yizhanchengming-xin/api/internal/novel"
    "github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

type fakeIntakeRepo struct {
    books []IntakeBook
    fetched map[int64]FetchedBook
    resolved map[int64]ResolvedMetadata
}
func (f *fakeIntakeRepo) ListBooks(context.Context, string) ([]IntakeBook, error) { return append([]IntakeBook(nil), f.books...), nil }
func (f *fakeIntakeRepo) SaveFetched(_ context.Context, id int64, book FetchedBook) error {
    if f.fetched == nil { f.fetched = map[int64]FetchedBook{} }
    f.fetched[id] = book
    return nil
}
func (f *fakeIntakeRepo) SaveResolved(_ context.Context, id int64, meta ResolvedMetadata) error {
    if f.resolved == nil { f.resolved = map[int64]ResolvedMetadata{} }
    f.resolved[id] = meta
    return nil
}

type fake121Fetcher struct{ response one21.BookResponse }
func (f *fake121Fetcher) FetchBook(context.Context, string, string, int) (one21.BookResponse, error) { return f.response, nil }

func TestIntakeExecutorFetchesBookThrough121AndStoresMetadata(t *testing.T) {
    repo := &fakeIntakeRepo{books: []IntakeBook{{ID:1, BookID:"7673480334440139800", PlatformID:"2", MaxTxt:2000}}}
    fetcher := &fake121Fetcher{response: one21.BookResponse{Code:200, Data:"正文", BookInfo:one21.BookInfo{BookID:"7673480334440139800", BookName:"测试书", Category:"男生生活", Genre:8}}}
    exec := IntakeExecutor{Books:repo, Fetcher:fetcher}
    job := pipeline.Job{ID:"job-1", IntakeID:"intake-1"}

    if err := exec.Execute(context.Background(), job, pipeline.StageFetchBook); err != nil { t.Fatal(err) }
    got := repo.fetched[1]
    if got.SourceText != "正文" || got.Category != "男生生活" || got.Genre != 8 { t.Fatalf("fetched=%#v", got) }
}

func TestIntakeExecutorResolvesGenderWithoutAIWhenCategoryIsExplicit(t *testing.T) {
    repo := &fakeIntakeRepo{books: []IntakeBook{{ID:1, BookID:"b1", Category:"男生生活", Genre:8}}}
    exec := IntakeExecutor{Books:repo}
    job := pipeline.Job{ID:"job-1", IntakeID:"intake-1"}

    if err := exec.Execute(context.Background(), job, pipeline.StageResolveMetadata); err != nil { t.Fatal(err) }
    got := repo.resolved[1]
    if got.Gender.Gender != novel.GenderMale || got.Gender.Source != novel.GenderSource121Category { t.Fatalf("resolved=%#v", got) }
}

func TestIntakeExecutorMarksAIDecisionOnlyForMissingMetadata(t *testing.T) {
    repo := &fakeIntakeRepo{books: []IntakeBook{
        {ID:1, BookID:"b1", Category:"男生生活", Style:"都市"},
        {ID:2, BookID:"b2", Category:"", Style:""},
    }}
    exec := IntakeExecutor{Books:repo}
    job := pipeline.Job{ID:"job-1", IntakeID:"intake-1"}

    if err := exec.Execute(context.Background(), job, pipeline.StageAIClassify); err != nil { t.Fatal(err) }
    if repo.resolved[1].NeedsAI { t.Fatalf("deterministic book should skip AI: %#v", repo.resolved[1]) }
    if !repo.resolved[2].NeedsAI { t.Fatalf("missing metadata should require AI: %#v", repo.resolved[2]) }
}
