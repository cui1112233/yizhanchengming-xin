package worker

import (
    "context"
    "errors"
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

type fakeBatchCreator struct { intakeID string; jobID string; batchID string }
func (f *fakeBatchCreator) CreateFromIntake(_ context.Context, intakeID, jobID string) (string, error) {
    f.intakeID = intakeID; f.jobID = jobID
    if f.batchID == "" { f.batchID = "batch-1" }
    return f.batchID, nil
}

type fakeClassifier struct { calls int; result AIClassification; err error }
func (f *fakeClassifier) Classify(context.Context, IntakeBook) (AIClassification, error) {
    f.calls++
    return f.result, f.err
}

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

func TestIntakeExecutorSkipsClassifierWhenMetadataComplete(t *testing.T) {
    repo := &fakeIntakeRepo{books: []IntakeBook{{ID:1, BookID:"b1", Category:"男生生活", Style:"都市"}}}
    classifier := &fakeClassifier{err:errors.New("must not be called")}
    exec := IntakeExecutor{Books:repo, Classifier:classifier}
    job := pipeline.Job{ID:"job-1", IntakeID:"intake-1"}

    if err := exec.Execute(context.Background(), job, pipeline.StageAIClassify); err != nil { t.Fatal(err) }
    if classifier.calls != 0 { t.Fatalf("classifier calls=%d", classifier.calls) }
    if repo.resolved[1].NeedsAI { t.Fatalf("deterministic book should skip AI: %#v", repo.resolved[1]) }
}

func TestIntakeExecutorUsesAIOnlyForMissingMetadata(t *testing.T) {
    repo := &fakeIntakeRepo{books: []IntakeBook{{ID:2, BookID:"b2", SourceText:"正文", Category:"", Style:""}}}
    classifier := &fakeClassifier{result:AIClassification{Gender:novel.GenderFemale, Style:"现言"}}
    exec := IntakeExecutor{Books:repo, Classifier:classifier}
    job := pipeline.Job{ID:"job-1", IntakeID:"intake-1"}

    if err := exec.Execute(context.Background(), job, pipeline.StageAIClassify); err != nil { t.Fatal(err) }
    got := repo.resolved[2]
    if classifier.calls != 1 || got.Gender.Gender != novel.GenderFemale || got.Gender.Source != novel.GenderSourceAI || got.Style != "现言" || got.NeedsAI {
        t.Fatalf("classification=%#v calls=%d", got, classifier.calls)
    }
}

func TestIntakeExecutorDoesNotFakeAISuccessWithoutClassifier(t *testing.T) {
    repo := &fakeIntakeRepo{books: []IntakeBook{{ID:2, BookID:"b2", Category:"", Style:""}}}
    exec := IntakeExecutor{Books:repo}
    job := pipeline.Job{ID:"job-1", IntakeID:"intake-1"}
    if err := exec.Execute(context.Background(), job, pipeline.StageAIClassify); err == nil { t.Fatal("expected missing classifier error") }
}

func TestIntakeExecutorCreatesBatchFromPreparedIntake(t *testing.T) {
    repo := &fakeIntakeRepo{}
    creator := &fakeBatchCreator{batchID:"batch-9"}
    exec := IntakeExecutor{Books:repo, Batches:creator}
    job := pipeline.Job{ID:"job-1", IntakeID:"intake-1"}

    if err := exec.Execute(context.Background(), job, pipeline.StageCreateBatch); err != nil { t.Fatal(err) }
    if creator.intakeID != "intake-1" || creator.jobID != "job-1" { t.Fatalf("creator=%#v", creator) }
}
