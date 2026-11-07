package httpapi

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/n8catalog"
)

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

func TestComicAudioCategoryRejectsMissingClassification(t *testing.T) {
	book := batchfactoryv11.Book{SourceMetadata: map[string]any{"style": "未知", "classifyStatus": "classified"}}
	if _, err := comicAudioCategoryName(book); err == nil {
		t.Fatal("comicAudioCategoryName() error = nil, want missing classification error")
	}
}
