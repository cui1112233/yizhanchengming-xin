package httpapi

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
	cp "qiantie/backend/internal/collectionpublish"
	"qiantie/backend/internal/n8catalog"
)

func TestTextPublishMappingUsesGenderWithoutClassificationMarker(t *testing.T) {
	// The real failing text book has a bookinfo gender, an empty style, and no
	// classifyStatus. Audio routing must not require another AI classification.
	metadata := map[string]any{"gender": "女频", "genderSource": "bookinfo", "style": ""}
	book := batchfactoryv11.Book{Title: "测试作品", SourceMetadata: metadata}
	publisher := textCollectionPublisher{n8: &BatchFactoryN8ComicPublisher{
		ProducerAccounts: []string{"maker"},
		ProducerDirectory: fakeN8ComicAudioDirectory{
			categories: []n8catalog.AudioCategory{{ID: 9, Name: "现代通用", Status: "ENABLE"}},
			audios:     []n8catalog.Audio{{ID: 91, OrganizationID: 1, CategoryIDs: []int64{9}, Status: "ENABLE"}},
		},
	}}
	settings := batchfactoryv11.SettingsPatch{"comicPublishSettings": json.RawMessage(`{"producerAdminName":"maker","organizationId":"1","workIds":[45]}`)}
	request := cp.PublishRequest{Kind: cp.CollectionKindText, Owner: "alice", BatchID: "batch", BookID: "book"}
	if err := publisher.resolveMapping(context.Background(), &request, book, settings); err != nil {
		t.Fatalf("available gender must resolve automatic audio without classification marker: %v", err)
	}
	if request.AudioCategoryID != 9 || request.AudioID != 91 {
		t.Fatalf("audio mapping = %d/%d, want 9/91", request.AudioCategoryID, request.AudioID)
	}
	if !reflect.DeepEqual(metadata, map[string]any{"gender": "女频", "genderSource": "bookinfo", "style": ""}) {
		t.Fatalf("audio matching changed book classification: %#v", metadata)
	}
}

func TestComicAudioCategoryPreservesAvailableMetadataWithoutMarker(t *testing.T) {
	for _, tc := range []struct {
		name     string
		metadata map[string]any
		want     string
	}{
		{"male alias", map[string]any{"gender": "男"}, "男频都市"},
		{"female bookinfo", map[string]any{"gender": "女频", "genderSource": "bookinfo", "style": ""}, "现代通用"},
		{"known style", map[string]any{"style": "古风虐文"}, "古风虐文"},
		{"stale marker", map[string]any{"gender": "男", "classifyStatus": "pending_confirmation"}, "男频都市"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := comicAudioCategoryName(batchfactoryv11.Book{SourceMetadata: tc.metadata})
			if err != nil || got != tc.want {
				t.Fatalf("audio category = %q, %v; want %q", got, err, tc.want)
			}
		})
	}
}
