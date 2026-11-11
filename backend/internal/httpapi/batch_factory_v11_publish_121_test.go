package httpapi

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
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
	uploadNames    []string
	uploadCookies  []string
	uploadFailures int
	loginInputs    []videomanagementconnector.LoginInput
	loginSession   videomanagementconnector.Session
	submissions    []videomanagementconnector.SubmissionInput
	readbacks      int
	readbackErr    error
	readbackRaw    map[string]any
}

func (c *recording121Connector) Login(_ context.Context, input videomanagementconnector.LoginInput) (videomanagementconnector.Session, error) {
	c.loginInputs = append(c.loginInputs, input)
	return c.loginSession, nil
}

func (c *recording121Connector) CheckSession(context.Context, videomanagementconnector.Session) error {
	return nil
}

func (c *recording121Connector) RefreshCatalogs(context.Context, videomanagementconnector.Session) (videomanagementconnector.Catalogs, error) {
	return videomanagementconnector.Catalogs{}, nil
}

func (c *recording121Connector) UploadVideo(_ context.Context, session videomanagementconnector.Session, filename string, _ []byte) (string, error) {
	c.uploadNames = append(c.uploadNames, filename)
	c.uploadCookies = append(c.uploadCookies, session.Cookie)
	if c.uploadFailures > 0 {
		c.uploadFailures--
		return "", videomanagementconnector.ErrSessionExpired
	}
	raw, _ := json.Marshal(map[string]any{"name": filename, "url": "https://video.test/" + filename})
	return string(raw), nil
}

func (c *recording121Connector) Submit(_ context.Context, _ videomanagementconnector.Session, input videomanagementconnector.SubmissionInput) (videomanagementconnector.Submission, error) {
	c.submissions = append(c.submissions, input)
	return videomanagementconnector.Submission{ID: "submission-1", Status: "uploaded", ReadbackRaw: map[string]any{"accepted": true}}, nil
}

func (c *recording121Connector) GetSubmission(_ context.Context, _ videomanagementconnector.Session, _ string) (videomanagementconnector.Submission, error) {
	c.readbacks++
	if c.readbackErr != nil {
		return videomanagementconnector.Submission{}, c.readbackErr
	}
	raw := c.readbackRaw
	if raw == nil {
		raw = map[string]any{"confirmed": true}
	}
	return videomanagementconnector.Submission{ID: "submission-1", Status: "uploaded", ReadbackRaw: raw}, nil
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

func TestBatchFactory121PublisherRelogsInOnceWhenUploadSessionExpires(t *testing.T) {
	publisher, connector, _, batch, book := seedMultiOpening121Publisher(t, []int{0, 1, 2, 3})
	const secret = "test-credential-secret"
	writeEncrypted121CredentialForTest(t, publisher.UsersDir, "alice", secret, map[string]string{
		"targetUsername": "operator",
		"password":       "secret-password",
		"baseUrl":        "https://two.121w.com/tttadmin",
	})
	publisher.CredentialSecret = secret
	connector.uploadFailures = 1
	connector.loginSession = videomanagementconnector.Session{Cookie: "PHPSESSID=fresh", TargetUsername: "operator"}

	if _, err := publisher.Submit(context.Background(), "alice", batch.ID, book.ID, map[string]any{"organization": "1"}); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if len(connector.loginInputs) != 1 {
		t.Fatalf("login calls = %d, want exactly one automatic relogin", len(connector.loginInputs))
	}
	if connector.loginInputs[0].Username != "operator" || connector.loginInputs[0].Password != "secret-password" {
		t.Fatalf("automatic relogin used the wrong saved credential")
	}
	if got := connector.uploadCookies; len(got) < 2 || got[0] != "session=ok" || got[1] != "PHPSESSID=fresh" {
		t.Fatalf("upload cookies = %v, want stale attempt followed by fresh session", got)
	}
	saved := readUploadTargetSession(publisher.UsersDir, "alice")
	if saved["cookie"] != "PHPSESSID=fresh" {
		t.Fatalf("saved upload session cookie = %#v, want refreshed session", saved["cookie"])
	}
}

func writeEncrypted121CredentialForTest(t *testing.T, usersDir, owner, secret string, value map[string]string) {
	t.Helper()
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		t.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}
	iv := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(iv); err != nil {
		t.Fatal(err)
	}
	plain, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	sealed := gcm.Seal(nil, iv, plain, nil)
	tagSize := gcm.Overhead()
	record := map[string]any{
		"version":    1,
		"algorithm":  "aes-256-gcm",
		"iv":         base64.StdEncoding.EncodeToString(iv),
		"tag":        base64.StdEncoding.EncodeToString(sealed[len(sealed)-tagSize:]),
		"ciphertext": base64.StdEncoding.EncodeToString(sealed[:len(sealed)-tagSize]),
	}
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(usersDir, owner, "novel-fetch-workshop")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "121-credentials.json"), raw, 0o600); err != nil {
		t.Fatal(err)
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

func TestPublish121IdempotentUploaded(t *testing.T) {
	publisher, connector, store, batch, book := seedMultiOpening121Publisher(t, []int{0, 1, 2, 3})
	if _, err := store.UpdateBookMetadata(context.Background(), "alice", batch.ID, book.ID, batchfactoryv11.UpdateBookMetadataInput{
		Metadata:         map[string]any{"websiteSubmitStatus": "uploaded"},
		ExpectedRevision: book.Revision,
	}); err != nil {
		t.Fatal(err)
	}

	result, err := publisher.Submit(context.Background(), "alice", batch.ID, book.ID, map[string]any{"organization": "1"})
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if result["status"] != "uploaded" || result["skipped"] != true {
		t.Fatalf("result = %#v, want uploaded short-circuit", result)
	}
	if len(connector.uploadNames) != 0 || len(connector.submissions) != 0 || connector.readbacks != 0 {
		t.Fatalf("uploads/submissions/readbacks = %d/%d/%d, want no remote calls", len(connector.uploadNames), len(connector.submissions), connector.readbacks)
	}
}

func TestPublish121ReadbackFailureReturnsUnconfirmed(t *testing.T) {
	publisher, connector, store, batch, book := seedMultiOpening121Publisher(t, []int{0, 1, 2, 3})
	connector.readbackErr = errors.New("network timeout")

	_, err := publisher.Submit(context.Background(), "alice", batch.ID, book.ID, map[string]any{"organization": "1"})
	if !errors.Is(err, errPublishUnconfirmed) {
		t.Fatalf("Submit() error = %v, want errPublishUnconfirmed", err)
	}
	if len(connector.uploadNames) != 4 || len(connector.submissions) != 1 {
		t.Fatalf("uploads/submissions = %d/%d, want 4/1 before readback", len(connector.uploadNames), len(connector.submissions))
	}
	updated, getErr := store.GetBatch(context.Background(), "alice", batch.ID)
	if getErr != nil {
		t.Fatal(getErr)
	}
	got := updated.Books[0].SourceMetadata["websiteSubmitStatus"]
	if got != "upload_pending_confirm" {
		t.Fatalf("websiteSubmitStatus = %#v, want upload_pending_confirm", got)
	}
	if n := publish121HeadCount(updated.Books[0].SourceMetadata["websiteSubmitHeads"]); n != 4 {
		t.Fatalf("websiteSubmitHeads count = %d, want 4 persisted heads", n)
	}
}

func publish121HeadCount(raw any) int {
	switch heads := raw.(type) {
	case []string:
		return len(heads)
	case []any:
		return len(heads)
	default:
		return 0
	}
}

func publish121WithPendingMetadata(t *testing.T, store *batchfactoryv11.MemoryStore, batchID string, book batchfactoryv11.Book) {
	t.Helper()
	metadata := map[string]any{}
	for key, value := range book.SourceMetadata {
		metadata[key] = value
	}
	metadata["websiteSubmitStatus"] = "upload_pending_confirm"
	metadata["websiteSubmitHeads"] = []any{`{"name":"h0"}`, `{"name":"h1"}`, `{"name":"h2"}`, `{"name":"h3"}`}
	if _, err := store.UpdateBookMetadata(context.Background(), "alice", batchID, book.ID, batchfactoryv11.UpdateBookMetadataInput{
		Metadata:         metadata,
		ExpectedRevision: book.Revision,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestPublish121PendingConfirmReadbackFirst(t *testing.T) {
	publisher, connector, store, batch, book := seedMultiOpening121Publisher(t, []int{0, 1, 2, 3})
	// 模拟上一轮：视频已上传、容器已提交、回读失败。
	publish121WithPendingMetadata(t, store, batch.ID, book)
	connector.readbackRaw = map[string]any{"data": []any{map[string]any{"bookid": book.BookID}}}

	result, err := publisher.Submit(context.Background(), "alice", batch.ID, book.ID, map[string]any{"organization": "1"})
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if result["status"] != "uploaded" {
		t.Fatalf("result = %#v, want uploaded after readback confirm", result)
	}
	if len(connector.uploadNames) != 0 || len(connector.submissions) != 0 {
		t.Fatalf("uploads/submissions = %d/%d, want readback-only path", len(connector.uploadNames), len(connector.submissions))
	}
	updated, _ := store.GetBatch(context.Background(), "alice", batch.ID)
	if got := updated.Books[0].SourceMetadata["websiteSubmitStatus"]; got != "uploaded" {
		t.Fatalf("websiteSubmitStatus = %#v, want uploaded", got)
	}
}

func TestPublish121PendingConfirmResubmitsWhenMissingRemotely(t *testing.T) {
	publisher, connector, store, batch, book := seedMultiOpening121Publisher(t, []int{0, 1, 2, 3})
	publish121WithPendingMetadata(t, store, batch.ID, book)
	// 第一次回读确认远端不存在 → 重新提交；提交后的回读返回正常。
	connector.readbackRaw = map[string]any{"data": []any{}}

	_, err := publisher.Submit(context.Background(), "alice", batch.ID, book.ID, map[string]any{"organization": "1"})
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if len(connector.uploadNames) != 0 {
		t.Fatalf("uploads = %d, want reuse persisted heads", len(connector.uploadNames))
	}
	if len(connector.submissions) != 1 {
		t.Fatalf("submissions = %d, want 1 resubmit", len(connector.submissions))
	}
}

func publishFieldsFixture(t *testing.T, publishSettings string) (batchfactoryv11.Batch, batchfactoryv11.Book) {
	t.Helper()
	patch := batchfactoryv11.SettingsPatch{}
	if publishSettings != "" {
		patch["publishSettings"] = json.RawMessage(publishSettings)
	}
	batch := batchfactoryv11.Batch{SettingsState: batchfactoryv11.SettingsState{Patch: patch}}
	book := batchfactoryv11.Book{
		ID:       "book-1",
		BookID:   "123456",
		Platform: "15",
		SourceMetadata: map[string]any{
			"gender": "女",
			"style":  "现代甜文",
		},
	}
	return batch, book
}

// Reproduces the production defect: the selected profile is 8 decompress +
// 0 scroll, but the form used hard-coded 4/4 because the profile was ignored.
func TestBuildPublishFieldsUsesSelectedWebsiteProfile(t *testing.T) {
	publish := `{
		"websiteProfileId": "121-420",
		"versionProfile": "8个解压",
		"organization": "1",
		"materialReuse": true,
		"horizontalFlip": true,
		"websiteProfiles": [{
			"id": "121-420", "name": "8个解压", "config_id": "420",
			"advanced": {"tl5":0,"ziti":1,"zitidx":36,"biaohong":"","jieyaNum":8,"keywords":"",
			             "gunpingNum":0,"jieyaPitch":0,"jieyaSpeed":1.7,"jieyaAiHead":0,
			             "gunpingSpeed":1,"biaohongReuse":"1","fontColorStyles":[1]}
		}]
	}`
	batch, book := publishFieldsFixture(t, publish)
	fields, err := buildPublishFields(batch, book, map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{
		"jieya_num":                     "8",
		"gunping_num":                   "0",
		"jieya_ai_head":                 "0",
		"jieya_speed":                   "1.7",
		"zitidx":                        "36",
		"font_color_styles":             "[1]",
		"biaohong_reuse":                "1",
		"jieya_ai_head_horizontal_flip": "1",
		"organization":                  "1",
	} {
		if got := fields[key]; got != want {
			t.Fatalf("%s = %q, want %q (fields=%v)", key, got, want, fields)
		}
	}
}

// Legacy patches only stored the profile display name. It must resolve to the
// same profile instead of silently falling back to 4/4 defaults.
func TestBuildPublishFieldsResolvesLegacyProfileName(t *testing.T) {
	publish := `{
		"versionProfile": "8个解压",
		"organization": 1,
		"websiteProfiles": [{
			"id": "121-420", "name": "8个解压", "config_id": "420",
			"advanced": {"jieyaNum":8,"gunpingNum":0,"biaohongReuse":"1"}
		}]
	}`
	batch, book := publishFieldsFixture(t, publish)
	fields, err := buildPublishFields(batch, book, map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if fields["jieya_num"] != "8" || fields["gunping_num"] != "0" {
		t.Fatalf("legacy name resolution = %v, want jieya_num=8 gunping_num=0", fields)
	}
}

// An unknown profile id must fail loudly; it must never produce 4/4 silently.
func TestBuildPublishFieldsRejectsUnknownWebsiteProfile(t *testing.T) {
	batch, book := publishFieldsFixture(t, `{"websiteProfileId":"121-999","organization":"1","websiteProfiles":[]}`)
	_, err := buildPublishFields(batch, book, map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "网站配置档 121-999") {
		t.Fatalf("error = %v, want explicit missing profile error", err)
	}
}

// Explicit publishSettings values must win over the profile preset.
func TestBuildPublishFieldsExplicitSettingsOverrideProfile(t *testing.T) {
	publish := `{
		"websiteProfileId": "121-420",
		"jieyaNum": 6,
		"gunpingNum": 2,
		"organization": "1",
		"websiteProfiles": [{
			"id": "121-420", "name": "8个解压", "config_id": "420",
			"advanced": {"jieyaNum":8,"gunpingNum":0}
		}]
	}`
	batch, book := publishFieldsFixture(t, publish)
	fields, err := buildPublishFields(batch, book, map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if fields["jieya_num"] != "6" || fields["gunping_num"] != "2" {
		t.Fatalf("explicit override = %v, want 6/2", fields)
	}
}

// The hard-coded 4/4 remains only as the documented fallback when no profile
// is selected at all.
func TestBuildPublishFieldsFallsBackToDefaultsWithoutProfile(t *testing.T) {
	batch, book := publishFieldsFixture(t, `{"organization":"1"}`)
	fields, err := buildPublishFields(batch, book, map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if fields["jieya_num"] != "4" || fields["gunping_num"] != "4" {
		t.Fatalf("default = %v, want 4/4", fields)
	}
}

func TestPublishOrganizationFallsBackToWebsiteProfile(t *testing.T) {
	settings := map[string]json.RawMessage{
		"publishSettings": json.RawMessage(`{
			"websiteProfileId": "121-420",
			"websiteProfiles": [{"id":"121-420","name":"8个解压","organization":"1","advanced":{}}]
		}`),
	}
	if got := publishOrganization(map[string]any{}, settings); got != "1" {
		t.Fatalf("publishOrganization() = %q, want profile-derived 1", got)
	}
}

func TestSelectedPublishProfileFallsBackToNameWhenStoredIDIsStale(t *testing.T) {
	publish := map[string]any{
		"websiteProfileId": "121-420",
		"versionProfile":   "8个解压",
		"websiteProfiles": []any{
			map[string]any{"id": "121-1", "config_id": "1", "name": "8个解压", "organization": "1"},
		},
	}
	profile := selectedPublishProfile(publish)
	if profile == nil || profile["id"] != "121-1" {
		t.Fatalf("selected profile = %#v, want the same named current profile", profile)
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
