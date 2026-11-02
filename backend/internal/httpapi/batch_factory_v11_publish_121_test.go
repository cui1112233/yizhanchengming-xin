package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/videomanagementconnector"
)

type recording121Connector struct {
	uploadNames []string
	submissions []videomanagementconnector.SubmissionInput
	readbacks   int
}

func (c *recording121Connector) Login(context.Context, videomanagementconnector.LoginInput) (videomanagementconnector.Session, error) {
	return videomanagementconnector.Session{}, nil
}

func (c *recording121Connector) CheckSession(context.Context, videomanagementconnector.Session) error {
	return nil
}

func (c *recording121Connector) RefreshCatalogs(context.Context, videomanagementconnector.Session) (videomanagementconnector.Catalogs, error) {
	return videomanagementconnector.Catalogs{}, nil
}

func (c *recording121Connector) UploadVideo(_ context.Context, _ videomanagementconnector.Session, filename string, _ []byte) (string, error) {
	c.uploadNames = append(c.uploadNames, filename)
	raw, _ := json.Marshal(map[string]any{"name": filename, "url": "https://video.test/" + filename})
	return string(raw), nil
}

func (c *recording121Connector) Submit(_ context.Context, _ videomanagementconnector.Session, input videomanagementconnector.SubmissionInput) (videomanagementconnector.Submission, error) {
	c.submissions = append(c.submissions, input)
	return videomanagementconnector.Submission{ID: "submission-1", Status: "uploaded", ReadbackRaw: map[string]any{"accepted": true}}, nil
}

func (c *recording121Connector) GetSubmission(_ context.Context, _ videomanagementconnector.Session, _ string) (videomanagementconnector.Submission, error) {
	c.readbacks++
	return videomanagementconnector.Submission{ID: "submission-1", Status: "uploaded", ReadbackRaw: map[string]any{"confirmed": true}}, nil
}

func TestPublishOrganizationUsesNestedBatchPublishSettings(t *testing.T) {
	settings := map[string]json.RawMessage{
		"publishSettings": json.RawMessage(`{"organization":1}`),
	}
	if got := publishOrganization(map[string]any{}, settings); got != "1" {
		t.Fatalf("publishOrganization() = %q, want 1", got)
	}
}

func TestPublisherRequiresGoVideoManagementConnector(t *testing.T) {
	publisher := &BatchFactory121Publisher{}
	_, err := publisher.Submit(context.Background(), "owner", "batch", "book", nil)
	if err == nil || !strings.Contains(err.Error(), "视频管理系统上传服务未配置") {
		t.Fatalf("Submit() error = %v, want explicit Go connector configuration error", err)
	}
}

func TestSelectPublishMergeJobsChoosesNewestCompleteSetByOpening(t *testing.T) {
	book := batchfactoryv11.Book{
		ID:     "book-1",
		Videos: []batchfactoryv11.Video{{ID: "video-1"}, {ID: "video-2"}},
		SettingsState: batchfactoryv11.SettingsState{Patch: batchfactoryv11.SettingsPatch{
			"openingEnabled": json.RawMessage(`true`),
			"openingCount":   json.RawMessage(`4`),
		}},
	}
	batch := batchfactoryv11.Batch{ID: "batch-1", Books: []batchfactoryv11.Book{book}}
	base := time.Date(2026, 10, 17, 10, 0, 0, 0, time.UTC)
	jobs := []batchfactoryv11.MergeJob{
		{ID: "opening-2-old", BookID: book.ID, OpeningVariantIndex: 2, Status: batchfactoryv11.MergeSucceeded, OutputURL: "merge://opening-2-old", UpdatedAt: base},
		{ID: "opening-3", BookID: book.ID, OpeningVariantIndex: 3, Status: batchfactoryv11.MergeSucceeded, OutputURL: "merge://opening-3", UpdatedAt: base.Add(3 * time.Minute)},
		{ID: "opening-0", BookID: book.ID, OpeningVariantIndex: 0, Status: batchfactoryv11.MergeSucceeded, OutputURL: "merge://opening-0", UpdatedAt: base.Add(time.Minute)},
		{ID: "opening-2-new", BookID: book.ID, OpeningVariantIndex: 2, Status: batchfactoryv11.MergeSucceeded, OutputURL: "merge://opening-2-new", UpdatedAt: base.Add(4 * time.Minute)},
		{ID: "opening-1", BookID: book.ID, OpeningVariantIndex: 1, Status: batchfactoryv11.MergeSucceeded, OutputURL: "merge://opening-1", UpdatedAt: base.Add(2 * time.Minute)},
		{ID: "other-book", BookID: "book-2", OpeningVariantIndex: 0, Status: batchfactoryv11.MergeSucceeded, OutputURL: "merge://other", UpdatedAt: base.Add(5 * time.Minute)},
	}

	selected, err := selectPublishMergeJobs(batch, book, jobs)
	if err != nil {
		t.Fatalf("selectPublishMergeJobs() error = %v", err)
	}
	gotIndexes := make([]int, 0, len(selected))
	gotIDs := make([]string, 0, len(selected))
	for _, job := range selected {
		gotIndexes = append(gotIndexes, job.OpeningVariantIndex)
		gotIDs = append(gotIDs, job.ID)
	}
	if want := []int{0, 1, 2, 3}; !reflect.DeepEqual(gotIndexes, want) {
		t.Fatalf("opening indexes = %v, want %v", gotIndexes, want)
	}
	if want := []string{"opening-0", "opening-1", "opening-2-new", "opening-3"}; !reflect.DeepEqual(gotIDs, want) {
		t.Fatalf("merge IDs = %v, want %v", gotIDs, want)
	}
}

func TestSelectPublishMergeJobsRejectsIncompleteOpeningSet(t *testing.T) {
	book := batchfactoryv11.Book{
		ID:     "book-1",
		Videos: []batchfactoryv11.Video{{ID: "video-1"}, {ID: "video-2"}},
		SettingsState: batchfactoryv11.SettingsState{Patch: batchfactoryv11.SettingsPatch{
			"openingEnabled": json.RawMessage(`true`),
			"openingCount":   json.RawMessage(`4`),
		}},
	}
	jobs := []batchfactoryv11.MergeJob{
		{ID: "opening-0", BookID: book.ID, OpeningVariantIndex: 0, Status: batchfactoryv11.MergeSucceeded, OutputURL: "merge://opening-0"},
		{ID: "opening-1", BookID: book.ID, OpeningVariantIndex: 1, Status: batchfactoryv11.MergeSucceeded, OutputURL: "merge://opening-1"},
		{ID: "opening-3", BookID: book.ID, OpeningVariantIndex: 3, Status: batchfactoryv11.MergeSucceeded, OutputURL: "merge://opening-3"},
	}

	_, err := selectPublishMergeJobs(batchfactoryv11.Batch{Books: []batchfactoryv11.Book{book}}, book, jobs)
	if err == nil || !strings.Contains(err.Error(), "换开头 2") {
		t.Fatalf("selectPublishMergeJobs() error = %v, want missing opening 2", err)
	}
}

func TestSelectPublishMergeJobsUsesOnlyOriginalForSingleStoryboard(t *testing.T) {
	book := batchfactoryv11.Book{
		ID:     "book-1",
		Videos: []batchfactoryv11.Video{{ID: "video-1"}},
		SettingsState: batchfactoryv11.SettingsState{Patch: batchfactoryv11.SettingsPatch{
			"openingEnabled": json.RawMessage(`true`),
			"openingCount":   json.RawMessage(`4`),
		}},
	}
	jobs := []batchfactoryv11.MergeJob{{ID: "opening-0", BookID: book.ID, OpeningVariantIndex: 0, Status: batchfactoryv11.MergeSucceeded, OutputURL: "merge://opening-0"}}

	selected, err := selectPublishMergeJobs(batchfactoryv11.Batch{Books: []batchfactoryv11.Book{book}}, book, jobs)
	if err != nil {
		t.Fatalf("selectPublishMergeJobs() error = %v", err)
	}
	if len(selected) != 1 || selected[0].OpeningVariantIndex != 0 {
		t.Fatalf("selected = %+v, want only original opening", selected)
	}
}

func TestBatchFactory121PublisherUploadsAllOpeningsInOneSubmission(t *testing.T) {
	publisher, connector, store, batch, book := seedMultiOpening121Publisher(t, []int{0, 1, 2, 3})

	result, err := publisher.Submit(context.Background(), "alice", batch.ID, book.ID, map[string]any{"organization": "1"})
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	wantNames := []string{
		book.BookID + "-opening-0.mp4",
		book.BookID + "-opening-1.mp4",
		book.BookID + "-opening-2.mp4",
		book.BookID + "-opening-3.mp4",
	}
	if !reflect.DeepEqual(connector.uploadNames, wantNames) {
		t.Fatalf("uploaded names = %v, want %v", connector.uploadNames, wantNames)
	}
	if len(connector.submissions) != 1 {
		t.Fatalf("submission count = %d, want 1", len(connector.submissions))
	}
	if connector.readbacks != 1 {
		t.Fatalf("readback count = %d, want 1", connector.readbacks)
	}
	input := connector.submissions[0]
	if input.Filename != book.BookID+".txt" || string(input.Content) != "正文内容" {
		t.Fatalf("submitted text = %q/%q, want %q/%q", input.Filename, input.Content, book.BookID+".txt", "正文内容")
	}
	var heads []map[string]any
	if err := json.Unmarshal([]byte(input.FormFields["jieya_ai_head_video"]), &heads); err != nil {
		t.Fatalf("jieya_ai_head_video is not JSON: %v", err)
	}
	if len(heads) != 4 {
		t.Fatalf("head count = %d, want 4", len(heads))
	}
	if got := result["videoCount"]; got != 4 {
		t.Fatalf("result videoCount = %#v, want 4", got)
	}

	updated, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	history, ok := updated.Books[0].SourceMetadata["websiteSubmitHistory"].([]any)
	if !ok || len(history) != 1 {
		t.Fatalf("websiteSubmitHistory = %#v, want one entry", updated.Books[0].SourceMetadata["websiteSubmitHistory"])
	}
	entry, ok := history[0].(map[string]any)
	if !ok || entry["videoCount"] != 4 {
		t.Fatalf("history entry = %#v, want videoCount 4", history[0])
	}
}

func TestBatchFactory121PublisherUploadsEditedContentAndKeepsRemainingSource(t *testing.T) {
	publisher, connector, store, batch, book := seedMultiOpening121Publisher(t, []int{0, 1, 2, 3})

	updatedBook, err := store.CaptureBookSource(context.Background(), "alice", batch.ID, book.ID, batchfactoryv11.CaptureBookSourceInput{
		SourceText:       "原文第一行\n原文第二行\n原文第三行",
		ReplaceSource:    true,
		ExpectedRevision: book.Revision,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.UpdateBookMetadata(context.Background(), "alice", batch.ID, book.ID, batchfactoryv11.UpdateBookMetadataInput{
		Metadata:         map[string]any{"gender": "女", "style": "现代甜文", "contentRangeLines": 2},
		ExpectedRevision: updatedBook.Revision,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.SaveDraft(context.Background(), "alice", batchfactoryv11.Draft{
		Key:     "working-front:" + book.ID,
		Kind:    "working-front-content",
		Scope:   batch.ID,
		Content: "编辑第一行\n编辑第二行",
	}); err != nil {
		t.Fatal(err)
	}

	_, err = publisher.Submit(context.Background(), "alice", batch.ID, book.ID, map[string]any{"organization": "1"})
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if len(connector.submissions) != 1 {
		t.Fatalf("submission count = %d, want 1", len(connector.submissions))
	}
	if got, want := string(connector.submissions[0].Content), "编辑第一行\n编辑第二行\n原文第三行"; got != want {
		t.Fatalf("submitted content = %q, want %q", got, want)
	}
}

func TestBatchFactory121PublisherRejectsIncompleteOpeningSetBeforeUpload(t *testing.T) {
	publisher, connector, _, batch, book := seedMultiOpening121Publisher(t, []int{0, 1, 3})

	_, err := publisher.Submit(context.Background(), "alice", batch.ID, book.ID, map[string]any{"organization": "1"})
	if err == nil || !strings.Contains(err.Error(), "换开头 2") {
		t.Fatalf("Submit() error = %v, want missing opening 2", err)
	}
	if len(connector.uploadNames) != 0 || len(connector.submissions) != 0 {
		t.Fatalf("uploads/submissions = %d/%d, want 0/0", len(connector.uploadNames), len(connector.submissions))
	}
}

func seedMultiOpening121Publisher(t *testing.T, openingIndexes []int) (*BatchFactory121Publisher, *recording121Connector, *batchfactoryv11.MemoryStore, batchfactoryv11.Batch, batchfactoryv11.Book) {
	t.Helper()
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{
		Title: "batch",
		InitialBatchSettings: batchfactoryv11.SettingsPatch{
			"openingEnabled": json.RawMessage(`true`),
			"openingCount":   json.RawMessage(`4`),
		},
		Books: []batchfactoryv11.CreateBookInput{{
			BookID:     "123456",
			Title:      "测试小说",
			Platform:   "15",
			SourceText: "正文内容",
			SourceMetadata: map[string]any{
				"gender": "女",
				"style":  "现代甜文",
			},
			Videos: []batchfactoryv11.CreateVideoInput{{Label: "分镜一"}, {Label: "分镜二"}},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	mediaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		_, _ = w.Write([]byte("mp4:" + request.URL.Path))
	}))
	t.Cleanup(mediaServer.Close)
	base := time.Date(2026, 10, 17, 10, 0, 0, 0, time.UTC)
	for _, index := range openingIndexes {
		_, err := store.CreateMergeJob(ctx, batchfactoryv11.MergeJob{
			Owner:               "alice",
			BatchID:             batch.ID,
			BookID:              book.ID,
			OpeningVariantIndex: index,
			RequestID:           "merge-opening-" + string(rune('0'+index)),
			Status:              batchfactoryv11.MergeSucceeded,
			OutputURL:           mediaServer.URL + "/opening-" + string(rune('0'+index)) + ".mp4",
			CreatedAt:           base.Add(time.Duration(index) * time.Minute),
			UpdatedAt:           base.Add(time.Duration(index) * time.Minute),
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	usersDir := t.TempDir()
	userDir := filepath.Join(usersDir, "alice")
	if err := os.MkdirAll(userDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(userDir, "upload-target.json"), []byte(`{"cookie":"session=ok","targetUsername":"operator"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	connector := &recording121Connector{}
	return &BatchFactory121Publisher{Store: store, UsersDir: usersDir, Connector: connector}, connector, store, batch, book
}
