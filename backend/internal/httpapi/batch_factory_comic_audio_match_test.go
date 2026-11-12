package httpapi

import (
	"context"
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/n8catalog"
)

// 书名匹配必须放在预检：占位书名对不上漫画后台作品时，要在出图前拦住，
// 而不是等整本图片生成、上传完才在发布阶段失败。
func TestComicPublishPreflightRejectsUnmatchedPlaceholderTitle(t *testing.T) {
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{
		Title: "漫画批次",
		Books: []batchfactoryv11.CreateBookInput{{Title: "小说 793433"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	publisher := &BatchFactoryN8ComicPublisher{
		Store: store,
		ProducerDirectory: fakeN8ComicAudioDirectory{
			categories: []n8catalog.AudioCategory{{ID: 9, Name: "现代通用", Status: "ENABLE"}},
			audios:     []n8catalog.Audio{{ID: 91, Name: "通用音频", OrganizationID: 1, CategoryIDs: []int64{9}, Status: "ENABLE"}},
		},
		WorksDirectory: fakeN8ComicWorkDirectory{configured: true, works: []n8catalog.Work{
			{ID: 9001, Name: "真实书名一"},
			{ID: 9002, Name: "真实书名二"},
		}},
		WorkTitleMapPath: filepath.Join(t.TempDir(), "n8-work-title-map.json"),
	}
	settings := batchfactoryv11.SettingsPatch{
		"comicPublishSettings": json.RawMessage(`{"organizationId":"1","audioCategoryId":9,"audioCategoryName":"现代通用","audioCategorySelectionMode":"manual","specifiedAudioId":91}`),
	}
	err = publisher.ValidateBookSettings(context.Background(), "alice", batch.ID, batch.Books[0].ID, settings)
	if err == nil || !strings.Contains(err.Error(), "未找到") {
		t.Fatalf("error=%v, want unmatched work title", err)
	}
	// 显式填了作品 ID 时不再依赖书名匹配。
	explicit := batchfactoryv11.SettingsPatch{
		"comicPublishSettings": json.RawMessage(`{"organizationId":"1","audioCategoryId":9,"audioCategoryName":"现代通用","audioCategorySelectionMode":"manual","specifiedAudioId":91,"workIdsText":"9001"}`),
	}
	if err := publisher.ValidateBookSettings(context.Background(), "alice", batch.ID, batch.Books[0].ID, explicit); err != nil {
		t.Fatalf("explicit work id preflight error=%v", err)
	}
}

// 占位书名「小说 793433」无法按名称匹配，但可以用来源书城 ID 反查
// N8 作品的 cp_work_id，发布预检同样要能通过。
func TestComicPublishMatchesWorkBySourceBookIDForPlaceholderTitle(t *testing.T) {
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{
		Title: "漫画批次",
		Books: []batchfactoryv11.CreateBookInput{{
			Title:          "小说 793433",
			SourceMetadata: map[string]any{"sourceLine": "793433"},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	publisher := &BatchFactoryN8ComicPublisher{
		Store: store,
		ProducerDirectory: fakeN8ComicAudioDirectory{
			categories: []n8catalog.AudioCategory{{ID: 9, Name: "现代通用", Status: "ENABLE"}},
			audios:     []n8catalog.Audio{{ID: 91, Name: "通用音频", OrganizationID: 1, CategoryIDs: []int64{9}, Status: "ENABLE"}},
		},
		WorksDirectory: fakeN8ComicSourceDirectory{works: []n8catalog.Work{
			{ID: 1046266, Name: "她为发小有求必应，我转身离开", CPWorkID: "793433", CPType: "QM"},
		}},
		WorkTitleMapPath: filepath.Join(t.TempDir(), "n8-work-title-map.json"),
	}
	ids, err := publisher.resolveComicWorkIDs(context.Background(), "alice", batch.ID, batch.Books[0].ID)
	if err != nil || !reflect.DeepEqual(ids, []int64{1046266}) {
		t.Fatalf("ids=%v err=%v", ids, err)
	}
	settings := batchfactoryv11.SettingsPatch{
		"comicPublishSettings": json.RawMessage(`{"organizationId":"1","audioCategoryId":9,"audioCategoryName":"现代通用","audioCategorySelectionMode":"manual","specifiedAudioId":91}`),
	}
	if err := publisher.ValidateBookSettings(context.Background(), "alice", batch.ID, batch.Books[0].ID, settings); err != nil {
		t.Fatalf("preflight error=%v", err)
	}
}

type fakeN8ComicSourceDirectory struct {
	works []n8catalog.Work
}

func (f fakeN8ComicSourceDirectory) Configured() bool { return true }
func (f fakeN8ComicSourceDirectory) SearchWorks(context.Context, string) ([]n8catalog.Work, error) {
	return nil, nil
}
func (f fakeN8ComicSourceDirectory) SearchWorksByCPWorkID(_ context.Context, sourceID string) ([]n8catalog.Work, error) {
	matched := make([]n8catalog.Work, 0, 1)
	for _, work := range f.works {
		if work.CPWorkID == sourceID {
			matched = append(matched, work)
		}
	}
	return matched, nil
}

func TestComicAudioCategoryMapsNovelStyleAndGender(t *testing.T) {
	tests := []struct {
		style  string
		gender string
		want   string
	}{
		{style: "现代悬疑", gender: "女频", want: "现代悬疑"},
		{style: "男频都市", gender: "男频", want: "男频都市"},
		{style: "古风通用", gender: "女频", want: "古风通用"},
		{style: "玄幻", gender: "男频", want: "古风通用"},
		{style: "历史", gender: "男频", want: "古风通用"},
		{style: "古风虐文", gender: "女频", want: "古风虐文"},
		{style: "古风甜文", gender: "女频", want: "古风甜文"},
		{style: "现代虐文", gender: "女频", want: "现代虐文"},
		{style: "年代虐文", gender: "女频", want: "现代虐文"},
		{style: "家庭伤感", gender: "女频", want: "现代虐文"},
		{style: "现代甜文", gender: "女频", want: "现代通用"},
		{style: "现代通用", gender: "女频", want: "现代通用"},
		{style: "年代甜文", gender: "女频", want: "现代通用"},
		{style: "年代通用", gender: "女频", want: "现代通用"},
		{style: "现代女主", gender: "女频", want: "现代通用"},
		{style: "家庭奇葩", gender: "女频", want: "现代通用"},
		{style: "职场打脸", gender: "男频", want: "男频都市"},
		{style: "职场打脸", gender: "女频", want: "现代通用"},
		{style: "未知", gender: "男频", want: "男频都市"},
		{style: "未知", gender: "女频", want: "现代通用"},
	}
	for _, tt := range tests {
		t.Run(tt.style+tt.gender, func(t *testing.T) {
			book := batchfactoryv11.Book{SourceMetadata: map[string]any{
				"style": tt.style, "gender": tt.gender, "classifyStatus": "classified",
			}}
			got, err := comicAudioCategoryName(book)
			if err != nil || got != tt.want {
				t.Fatalf("comicAudioCategoryName() = %q, %v; want %q", got, err, tt.want)
			}
		})
	}
}

func createComicAudioTestBook(t *testing.T, metadata map[string]any) (*batchfactoryv11.MemoryStore, batchfactoryv11.Batch, batchfactoryv11.Book) {
	t.Helper()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{
		Title: "音频匹配测试",
		Books: []batchfactoryv11.CreateBookInput{{Title: "第一本", SourceMetadata: metadata}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return store, batch, batch.Books[0]
}

func TestComicAudioSelectionPersistsAndReusesRandomChoice(t *testing.T) {
	store, batch, book := createComicAudioTestBook(t, map[string]any{
		"style": "现代悬疑", "gender": "女频", "classifyStatus": "classified",
	})
	randomCalls := 0
	publisher := &BatchFactoryN8ComicPublisher{
		Store: store,
		Now:   func() time.Time { return time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC) },
		RandomIndex: func(upper int) (int, error) {
			randomCalls++
			if upper != 2 {
				t.Fatalf("random upper=%d, want 2", upper)
			}
			return 1, nil
		},
		ProducerDirectory: fakeN8ComicAudioDirectory{
			categories: []n8catalog.AudioCategory{{ID: 14, Name: "现代悬疑", Status: "ENABLE"}},
			audios: []n8catalog.Audio{
				{ID: 101, Name: "悬疑一", OrganizationID: 1, CategoryIDs: []int64{14}, Status: "ENABLE", UpdatedAt: "2026-11-07 12:00:00"},
				{ID: 102, Name: "悬疑二", OrganizationID: 1, CategoryIDs: []int64{14}, Status: "ENABLE", UpdatedAt: "2026-11-06 12:00:00"},
			},
		},
	}
	config := n8ComicPublishSettings{OrganizationID: "1"}
	first, err := publisher.resolveBookComicAudio(context.Background(), "alice", batch.ID, book.ID, config)
	if err != nil {
		t.Fatal(err)
	}
	if first.CategoryID != 14 || first.AudioID != 102 || first.Mode != "auto" {
		t.Fatalf("first selection=%+v", first)
	}

	loaded, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	var persisted comicAudioSelection
	if err := json.Unmarshal(loaded.Books[0].SettingsState.Patch["comicAudioSelection"], &persisted); err != nil {
		t.Fatalf("persisted selection: %v", err)
	}
	if persisted.AudioID != 102 || persisted.CategoryName != "现代悬疑" {
		t.Fatalf("persisted=%+v", persisted)
	}

	publisher.RandomIndex = func(int) (int, error) {
		randomCalls++
		return 0, nil
	}
	second, err := publisher.resolveBookComicAudio(context.Background(), "alice", batch.ID, book.ID, config)
	if err != nil {
		t.Fatal(err)
	}
	if second.AudioID != 102 || randomCalls != 1 {
		t.Fatalf("second=%+v randomCalls=%d, want persisted audio and one random call", second, randomCalls)
	}
}

// 风格分类没有音频时必须自动兜底到有音频的通用分类，否则整本书会永久卡在
// 发布预检，只能人工换分类。
func TestComicAudioFallsBackWhenStyleCategoryHasNoAudio(t *testing.T) {
	store, batch, book := createComicAudioTestBook(t, map[string]any{
		"style": "现代悬疑", "gender": "女频", "classifyStatus": "classified",
	})
	publisher := &BatchFactoryN8ComicPublisher{
		Store:       store,
		Now:         func() time.Time { return time.Date(2026, 11, 12, 11, 0, 0, 0, time.UTC) },
		RandomIndex: func(int) (int, error) { return 0, nil },
		ProducerDirectory: fakeN8ComicAudioDirectory{
			categories: []n8catalog.AudioCategory{
				{ID: 14, Name: "现代悬疑", Status: "ENABLE"},
				{ID: 9, Name: "现代通用", Status: "ENABLE"},
				{ID: 10, Name: "古风通用", Status: "DISABLE"},
			},
			audios: []n8catalog.Audio{
				{ID: 91, Name: "通用音频", OrganizationID: 1, CategoryIDs: []int64{9}, Status: "ENABLE"},
			},
		},
	}
	selection, err := publisher.resolveBookComicAudio(context.Background(), "alice", batch.ID, book.ID, n8ComicPublishSettings{OrganizationID: "1"})
	if err != nil {
		t.Fatalf("resolve error=%v", err)
	}
	if selection.CategoryID != 9 || selection.CategoryName != "现代通用" || selection.AudioID != 91 {
		t.Fatalf("selection=%+v, want fallback to 现代通用", selection)
	}
}

// 所有分类都没有音频时仍然要给出明确错误，不能静默失败。
func TestComicAudioReportsWhenNoCategoryHasAudio(t *testing.T) {
	store, batch, book := createComicAudioTestBook(t, map[string]any{
		"style": "现代悬疑", "gender": "女频", "classifyStatus": "classified",
	})
	publisher := &BatchFactoryN8ComicPublisher{
		Store: store,
		Now:   func() time.Time { return time.Date(2026, 11, 12, 11, 0, 0, 0, time.UTC) },
		ProducerDirectory: fakeN8ComicAudioDirectory{
			categories: []n8catalog.AudioCategory{
				{ID: 14, Name: "现代悬疑", Status: "ENABLE"},
				{ID: 9, Name: "现代通用", Status: "ENABLE"},
			},
		},
	}
	_, err := publisher.resolveBookComicAudio(context.Background(), "alice", batch.ID, book.ID, n8ComicPublishSettings{OrganizationID: "1"})
	if err == nil || !strings.Contains(err.Error(), "没有可用音频") {
		t.Fatalf("error=%v, want clear no-audio message", err)
	}
}

// 预检也要走兜底：否则风格分类为空时，书会在预检被挡住，永远到不了会兜底的发布阶段。
func TestComicPublishPreflightAllowsAudioFallback(t *testing.T) {
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{
		Title: "漫画批次",
		Books: []batchfactoryv11.CreateBookInput{{
			Title:          "音频兜底",
			SourceMetadata: map[string]any{"style": "现代悬疑", "gender": "女频", "classifyStatus": "classified"},
			Videos:         []batchfactoryv11.CreateVideoInput{{Label: "第一幕"}},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	publisher := &BatchFactoryN8ComicPublisher{
		Store: store,
		ProducerDirectory: fakeN8ComicAudioDirectory{
			categories: []n8catalog.AudioCategory{
				{ID: 14, Name: "现代悬疑", Status: "ENABLE"},
				{ID: 9, Name: "现代通用", Status: "ENABLE"},
			},
			audios: []n8catalog.Audio{
				{ID: 91, Name: "通用音频", OrganizationID: 1, CategoryIDs: []int64{9}, Status: "ENABLE"},
			},
		},
		WorksDirectory:   fakeN8ComicWorkDirectory{configured: true, works: []n8catalog.Work{{ID: 9001, Name: "音频兜底"}}},
		WorkTitleMapPath: filepath.Join(t.TempDir(), "n8-work-title-map.json"),
	}
	settings := batchfactoryv11.SettingsPatch{
		"comicPublishSettings": json.RawMessage(`{"organizationId":"1","workMatchMode":"title_auto"}`),
	}
	if err := publisher.ValidateBookSettings(context.Background(), "alice", batch.ID, batch.Books[0].ID, settings); err != nil {
		t.Fatalf("preflight error=%v, want audio fallback to pass", err)
	}
}

func TestComicAudioSelectionUsesManualAudioWithoutRandomOverride(t *testing.T) {
	store, batch, book := createComicAudioTestBook(t, map[string]any{
		"style": "现代悬疑", "gender": "女频", "classifyStatus": "classified",
	})
	randomCalls := 0
	publisher := &BatchFactoryN8ComicPublisher{
		Store: store,
		Now:   func() time.Time { return time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC) },
		RandomIndex: func(int) (int, error) {
			randomCalls++
			return 0, nil
		},
		ProducerDirectory: fakeN8ComicAudioDirectory{
			categories: []n8catalog.AudioCategory{{ID: 14, Name: "现代悬疑", Status: "ENABLE"}},
			audios:     []n8catalog.Audio{{ID: 199, Name: "手动音频", OrganizationID: 1, CategoryIDs: []int64{14}, Status: "ENABLE"}},
		},
	}
	selection, err := publisher.resolveBookComicAudio(context.Background(), "alice", batch.ID, book.ID, n8ComicPublishSettings{
		OrganizationID: "1", AudioCategoryID: 14, AudioCategoryName: "现代悬疑",
		SpecifiedAudioID: 199, SpecifiedAudioName: "手动音频", AudioCategorySelectionMode: "manual",
	})
	if err != nil {
		t.Fatal(err)
	}
	if selection.AudioID != 199 || selection.Mode != "manual" || randomCalls != 0 {
		t.Fatalf("selection=%+v randomCalls=%d", selection, randomCalls)
	}
}

func TestComicAudioSelectionUsesCategoryIDWhenSavedNameIsStale(t *testing.T) {
	store, batch, book := createComicAudioTestBook(t, map[string]any{
		"style": "古风通用", "gender": "女频", "classifyStatus": "classified",
	})
	publisher := &BatchFactoryN8ComicPublisher{
		Store: store,
		ProducerDirectory: fakeN8ComicAudioDirectory{
			categories: []n8catalog.AudioCategory{{ID: 12, Name: "古风通用", Status: "ENABLE"}},
			audios:     []n8catalog.Audio{{ID: 16, Name: "已保存音频", OrganizationID: 1, CategoryIDs: []int64{12}, Status: "ENABLE"}},
		},
	}
	selection, err := publisher.resolveBookComicAudio(context.Background(), "alice", batch.ID, book.ID, n8ComicPublishSettings{
		OrganizationID: "1", AudioCategoryID: 12, AudioCategoryName: "古风-爽文（打脸）",
		SpecifiedAudioID: 16, SpecifiedAudioName: "已保存音频", AudioCategorySelectionMode: "manual",
	})
	if err != nil {
		t.Fatalf("stale category display name must not invalidate enabled category ID: %v", err)
	}
	if selection.CategoryID != 12 || selection.CategoryName != "古风通用" || selection.AudioID != 16 {
		t.Fatalf("selection=%+v, want enabled category 12 and audio 16", selection)
	}
}

func TestComicAudioSelectionRerandomizesOnlyWhenSavedAutomaticAudioIsUnavailable(t *testing.T) {
	store, batch, book := createComicAudioTestBook(t, map[string]any{
		"style": "现代悬疑", "gender": "女频", "classifyStatus": "classified",
	})
	directory := &fakeN8ComicAudioDirectory{
		categories: []n8catalog.AudioCategory{{ID: 14, Name: "现代悬疑", Status: "ENABLE"}},
		audios:     []n8catalog.Audio{{ID: 101, Name: "旧音频", OrganizationID: 1, CategoryIDs: []int64{14}, Status: "ENABLE"}},
	}
	publisher := &BatchFactoryN8ComicPublisher{
		Store: store, ProducerDirectory: directory,
		RandomIndex: func(int) (int, error) { return 0, nil },
	}
	config := n8ComicPublishSettings{OrganizationID: "1"}
	first, err := publisher.resolveBookComicAudio(context.Background(), "alice", batch.ID, book.ID, config)
	if err != nil || first.AudioID != 101 {
		t.Fatalf("first=%+v err=%v", first, err)
	}
	directory.audios = []n8catalog.Audio{{ID: 102, Name: "替代音频", OrganizationID: 1, CategoryIDs: []int64{14}, Status: "ENABLE"}}
	second, err := publisher.resolveBookComicAudio(context.Background(), "alice", batch.ID, book.ID, config)
	if err != nil {
		t.Fatal(err)
	}
	if second.AudioID != 102 || second.Mode != "auto" {
		t.Fatalf("second=%+v, want replacement automatic audio", second)
	}
}

func TestComicAudioSelectionRejectsUnavailableSavedManualAudio(t *testing.T) {
	store, batch, book := createComicAudioTestBook(t, map[string]any{
		"style": "现代悬疑", "gender": "女频", "classifyStatus": "classified",
	})
	directory := &fakeN8ComicAudioDirectory{
		categories: []n8catalog.AudioCategory{{ID: 14, Name: "现代悬疑", Status: "ENABLE"}},
		audios:     []n8catalog.Audio{{ID: 199, Name: "手动音频", OrganizationID: 1, CategoryIDs: []int64{14}, Status: "ENABLE"}},
	}
	publisher := &BatchFactoryN8ComicPublisher{Store: store, ProducerDirectory: directory}
	config := n8ComicPublishSettings{OrganizationID: "1", AudioCategoryID: 14, AudioCategoryName: "现代悬疑", SpecifiedAudioID: 199, AudioCategorySelectionMode: "manual"}
	if _, err := publisher.resolveBookComicAudio(context.Background(), "alice", batch.ID, book.ID, config); err != nil {
		t.Fatal(err)
	}
	directory.audios = nil
	_, err := publisher.resolveBookComicAudio(context.Background(), "alice", batch.ID, book.ID, config)
	if err == nil || !strings.Contains(err.Error(), "重新选择") {
		t.Fatalf("error=%v, want manual audio unavailable explanation", err)
	}
}

func TestComicAudioPreflightDoesNotChooseOneCategoryForAutomaticBooks(t *testing.T) {
	publisher := &BatchFactoryN8ComicPublisher{ProducerDirectory: fakeN8ComicAudioDirectory{
		categories: []n8catalog.AudioCategory{{ID: 14, Name: "现代悬疑", Status: "ENABLE"}},
	}}
	automatic := batchfactoryv11.SettingsPatch{"comicPublishSettings": json.RawMessage(`{"organizationId":"1"}`)}
	if err := publisher.ValidateSettings(context.Background(), automatic); err != nil {
		t.Fatalf("automatic preflight error=%v, want catalog-only validation", err)
	}
	manual := batchfactoryv11.SettingsPatch{"comicPublishSettings": json.RawMessage(`{"organizationId":"1","audioCategoryId":14,"audioCategoryName":"现代悬疑","audioCategorySelectionMode":"manual","specifiedAudioId":199}`)}
	if err := publisher.ValidateSettings(context.Background(), manual); err == nil || !strings.Contains(err.Error(), "重新选择") {
		t.Fatalf("manual preflight error=%v, want unavailable manual audio", err)
	}
}

func TestComicAudioBookPreflightValidatesMappedCategoryWithoutPersisting(t *testing.T) {
	store, batch, book := createComicAudioTestBook(t, map[string]any{
		"style": "现代悬疑", "gender": "女频", "classifyStatus": "classified",
	})
	randomCalls := 0
	publisher := &BatchFactoryN8ComicPublisher{
		Store: store,
		RandomIndex: func(int) (int, error) {
			randomCalls++
			return 0, nil
		},
		ProducerDirectory: fakeN8ComicAudioDirectory{
			categories: []n8catalog.AudioCategory{{ID: 9, Name: "现代通用", Status: "ENABLE"}},
			audios:     []n8catalog.Audio{{ID: 91, Name: "通用音频", OrganizationID: 1, CategoryIDs: []int64{9}, Status: "ENABLE"}},
		},
	}
	settings := batchfactoryv11.SettingsPatch{
		"comicPublishSettings": json.RawMessage(`{"organizationId":"1"}`),
	}
	before, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	err = publisher.ValidateBookSettings(context.Background(), "alice", batch.ID, book.ID, settings)
	if err == nil || !strings.Contains(err.Error(), "现代悬疑") || !strings.Contains(err.Error(), "不存在、未启用或名称不唯一") {
		t.Fatalf("error=%v, want missing mapped category", err)
	}
	after, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if after.Revision != before.Revision || randomCalls != 0 {
		t.Fatalf("preflight mutated state: revision %d -> %d, randomCalls=%d", before.Revision, after.Revision, randomCalls)
	}
	if _, exists := after.Books[0].SettingsState.Patch["comicAudioSelection"]; exists {
		t.Fatal("preflight persisted comicAudioSelection")
	}
}

func TestComicAudioCategoryRejectsMissingClassification(t *testing.T) {
	book := batchfactoryv11.Book{SourceMetadata: map[string]any{"style": "未知", "classifyStatus": "classified"}}
	if _, err := comicAudioCategoryName(book); err == nil {
		t.Fatal("comicAudioCategoryName() error = nil, want missing classification error")
	}
}
