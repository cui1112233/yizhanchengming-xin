package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/screenplay"
	"strings"
	"sync"
	"testing"
	"time"
)

func assetLibraryServiceFixture(t *testing.T) (*ScreenplayAssetLibrary, screenplay.Asset, *memoryImageToolObjects, *http.ServeMux, time.Time) {
	t.Helper()
	ctx := context.Background()
	projects := screenplay.NewMemoryStore()
	project, err := projects.CreateProject(ctx, screenplay.Project{ID: "library-source", OwnerUsername: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	source, err := projects.CreateAsset(ctx, screenplay.Asset{OwnerUsername: "alice", ProjectID: project.ID, Type: screenplay.AssetCharacter, Name: "林晚", Identity: "主角", Description: "角色原设定", Prompt: "不得被选图库覆盖", Voice: "源声音", Extra: map[string]string{"custom": "kept"}})
	if err != nil {
		t.Fatal(err)
	}
	objects := &memoryImageToolObjects{files: map[string][]byte{}}
	media := &ScreenplayMediaWorkbench{Projects: projects, Media: screenplay.NewMediaMemoryStore(), Cut: aicut.NewMemoryStore(), Objects: objects, Assets: agentworkspace.NewMemoryAssetStore()}
	item, err := media.saveBytes(ctx, project, "image", "source.png", "image/png", bytes.NewReader(testImageToolPNG(t)), source.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = media.Media.MutateMedia(ctx, "alice", project.ID, func(state *screenplay.MediaState) error { state.Selected[source.ID] = item.ID; return nil })
	if err != nil {
		t.Fatal(err)
	}
	service := &ScreenplayAssetLibrary{Repository: screenplay.NewMemoryAssetLibraryRepository(), Media: media}
	now := time.Unix(1700000000, 0)
	mux := http.NewServeMux()
	RegisterScreenplayAssetLibraryRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, service)
	return service, source, objects, mux, now
}

func sourceLibraryInput(service *ScreenplayAssetLibrary, asset screenplay.Asset) screenplayLibrarySaveInput {
	state, _ := service.Media.Media.LoadMedia(context.Background(), asset.OwnerUsername, asset.ProjectID)
	return screenplayLibrarySaveInput{ExpectedSelectedItemID: state.Selected[asset.ID], RequestKey: "library-request-001", AssetID: asset.ID, ExpectedAssetRevision: &asset.Revision, Category: "短剧角色"}
}

func TestScreenplayAssetLibrarySaveCopiesRealSelectedImageAndSurvivesSourceDeletion(t *testing.T) {
	service, asset, objects, mux, now := assetLibraryServiceFixture(t)
	in := sourceLibraryInput(service, asset)
	path := "/api/comic-screenplays/projects/" + asset.ProjectID + "/asset-library"
	response := signedJSONRequest(t, mux, now, "alice", "POST", path, in)
	if response.Code != 201 {
		t.Fatal(response.Code, response.Body.String())
	}
	var data struct {
		Entry struct {
			ID    string               `json:"id"`
			Image screenplay.MediaItem `json:"image"`
		}
	}
	if err := json.Unmarshal(response.Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}
	if data.Entry.Image.URL == "" || data.Entry.Image.AssetID == "" {
		t.Fatal(data)
	}
	for _, hidden := range []string{"object_key", "image_object", "snapshot", "不得被选图库覆盖"} {
		if strings.Contains(response.Body.String(), hidden) {
			t.Fatal("private storage leak", hidden)
		}
	}
	repeat := signedJSONRequest(t, mux, now, "alice", "POST", path, in)
	if repeat.Code != 201 {
		t.Fatal(repeat.Code, repeat.Body.String())
	}
	entries, _ := service.Repository.ListAssetLibrary(context.Background(), "alice", screenplay.AssetCharacter, "", "")
	if len(entries) != 1 || len(objects.files) != 2 || entries[0].Snapshot.Extra["custom"] != "kept" {
		t.Fatal(entries, len(objects.files))
	}
	stored, _ := service.Media.Cut.Asset(context.Background(), "alice", asset.ProjectID, entries[0].SourceMediaID)
	if entries[0].ImageObject.ObjectKey == stored.ObjectKey {
		t.Fatal("library depends on source object")
	}
	if err := objects.Delete(context.Background(), aicutObjectAsset(stored)); err != nil {
		t.Fatal(err)
	}
	if err := service.Media.Projects.DeleteProject(context.Background(), "alice", asset.ProjectID); err != nil {
		t.Fatal(err)
	}
	content := signedJSONRequest(t, mux, now, "alice", "GET", data.Entry.Image.URL, nil)
	if content.Code != 200 || !bytes.Equal(content.Body.Bytes(), testImageToolPNG(t)) {
		t.Fatal(content.Code, content.Body.Len())
	}
	forbidden := signedJSONRequest(t, mux, now, "bob", "GET", data.Entry.Image.URL, nil)
	if forbidden.Code != 404 {
		t.Fatal(forbidden.Code)
	}
	listing := signedJSONRequest(t, mux, now, "bob", "GET", "/api/comic-screenplays/asset-library", nil)
	if listing.Code != 200 || !strings.Contains(listing.Body.String(), `"entries":[]`) {
		t.Fatal(listing.Code, listing.Body.String())
	}
}

func TestScreenplayAssetLibrarySavePreservesSourceAndRejectsStaleOrUnselected(t *testing.T) {
	service, asset, _, _, _ := assetLibraryServiceFixture(t)
	ctx := context.Background()
	in := sourceLibraryInput(service, asset)
	stale := asset.Revision + 1
	in.ExpectedAssetRevision = &stale
	if _, err := service.SaveSource(ctx, "alice", asset.ProjectID, in); !errors.Is(err, screenplay.ErrRevisionConflict) {
		t.Fatal(err)
	}
	in = sourceLibraryInput(service, asset)
	state, _ := service.Media.Media.LoadMedia(ctx, "alice", asset.ProjectID)
	_, _ = service.Media.Media.MutateMedia(ctx, "alice", asset.ProjectID, func(s *screenplay.MediaState) error { s.Selected = map[string]string{}; return nil })
	if _, err := service.SaveSource(ctx, "alice", asset.ProjectID, in); !errors.Is(err, screenplay.ErrRevisionConflict) {
		t.Fatal(err)
	}
	_, _ = service.Media.Media.MutateMedia(ctx, "alice", asset.ProjectID, func(s *screenplay.MediaState) error { s.Selected = state.Selected; return nil })
	if _, err := service.SaveSource(ctx, "alice", asset.ProjectID, in); err != nil {
		t.Fatal(err)
	}
	assets, _ := service.Media.Projects.ListAssets(ctx, "alice", asset.ProjectID, "")
	if len(assets) != 1 || assets[0].Revision != asset.Revision || assets[0].Prompt != asset.Prompt || assets[0].Voice != asset.Voice {
		t.Fatal(assets)
	}
	in.Category = "changed"
	if _, err := service.SaveSource(ctx, "alice", asset.ProjectID, in); !errors.Is(err, screenplay.ErrRevisionConflict) {
		t.Fatal(err)
	}
}

type commitLostAssetLibraryRepository struct {
	screenplay.AssetLibraryRepository
	lookupUnavailable bool
	saved             bool
}

func (r *commitLostAssetLibraryRepository) SaveAssetLibraryEntry(ctx context.Context, entry screenplay.AssetLibraryEntry, revision int64) (screenplay.AssetLibraryEntry, error) {
	result, err := r.AssetLibraryRepository.SaveAssetLibraryEntry(ctx, entry, revision)
	if err != nil {
		return result, err
	}
	r.saved = true
	return screenplay.AssetLibraryEntry{}, fmt.Errorf("fixture commit acknowledgement lost")
}
func (r *commitLostAssetLibraryRepository) GetAssetLibraryEntry(ctx context.Context, owner, id string) (screenplay.AssetLibraryEntry, error) {
	if r.saved && r.lookupUnavailable {
		return screenplay.AssetLibraryEntry{}, fmt.Errorf("fixture reconciliation unavailable")
	}
	return r.AssetLibraryRepository.GetAssetLibraryEntry(ctx, owner, id)
}
func TestScreenplayAssetLibraryLostCommitAcknowledgementNeverDeletesCommittedImage(t *testing.T) {
	for _, unknown := range []bool{false, true} {
		t.Run(fmt.Sprint(unknown), func(t *testing.T) {
			service, asset, objects, _, _ := assetLibraryServiceFixture(t)
			repository := &commitLostAssetLibraryRepository{AssetLibraryRepository: service.Repository, lookupUnavailable: unknown}
			service.Repository = repository
			entry, err := service.SaveSource(context.Background(), "alice", asset.ProjectID, sourceLibraryInput(service, asset))
			if unknown && err == nil || !unknown && err != nil {
				t.Fatal(entry, err)
			}
			if len(objects.files) != 2 {
				t.Fatal("committed object deleted", len(objects.files))
			}
			repository.lookupUnavailable = false
			recovered, err := service.SaveSource(context.Background(), "alice", asset.ProjectID, sourceLibraryInput(service, asset))
			if err != nil || recovered.ID == "" || len(objects.files) != 2 {
				t.Fatal(recovered, err, len(objects.files))
			}
			body, err := objects.Open(context.Background(), agentworkspace.Asset{ObjectKey: recovered.ImageObject.ObjectKey})
			if err != nil {
				t.Fatal(err)
			}
			defer body.Close()
			raw := new(bytes.Buffer)
			_, _ = raw.ReadFrom(body)
			if !bytes.Equal(raw.Bytes(), testImageToolPNG(t)) {
				t.Fatal("library object not readable")
			}
		})
	}
}
func TestScreenplayAssetLibraryRejectsChangedImageEvenWhenAssetRevisionUnchanged(t *testing.T) {
	service, asset, _, _, _ := assetLibraryServiceFixture(t)
	in := sourceLibraryInput(service, asset)
	ctx := context.Background()
	project, _ := service.Media.Projects.GetProject(ctx, "alice", asset.ProjectID)
	second, err := service.Media.saveBytes(ctx, project, "image", "second.png", "image/png", bytes.NewReader(testImageToolPNG(t)), asset.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = service.Media.Media.MutateMedia(ctx, "alice", asset.ProjectID, func(s *screenplay.MediaState) error { s.Selected[asset.ID] = second.ID; return nil })
	assets, _ := service.Media.Projects.ListAssets(ctx, "alice", asset.ProjectID, "")
	if assets[0].Revision != asset.Revision {
		t.Fatal("fixture changed asset revision")
	}
	if _, err = service.SaveSource(ctx, "alice", asset.ProjectID, in); !errors.Is(err, screenplay.ErrRevisionConflict) {
		t.Fatal("saved a different image from the one user selected", err)
	}
}

type failingAssetLibraryRepository struct {
	screenplay.AssetLibraryRepository
}

func (f failingAssetLibraryRepository) SaveAssetLibraryEntry(context.Context, screenplay.AssetLibraryEntry, int64) (screenplay.AssetLibraryEntry, error) {
	return screenplay.AssetLibraryEntry{}, fmt.Errorf("fixture database write failed")
}

func TestScreenplayAssetLibraryFailedSaveCleansOnlyNewObject(t *testing.T) {
	service, asset, objects, _, _ := assetLibraryServiceFixture(t)
	service.Repository = failingAssetLibraryRepository{service.Repository}
	if _, err := service.SaveSource(context.Background(), "alice", asset.ProjectID, sourceLibraryInput(service, asset)); err == nil {
		t.Fatal("expected failure")
	}
	if len(objects.files) != 1 {
		t.Fatal("orphan library object or removed source", len(objects.files))
	}
}

func TestScreenplayAssetLibraryConcurrentDuplicateSaveKeepsOneCopy(t *testing.T) {
	service, asset, objects, _, _ := assetLibraryServiceFixture(t)
	var workers sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			_, err := service.SaveSource(context.Background(), "alice", asset.ProjectID, sourceLibraryInput(service, asset))
			errs <- err
		}()
	}
	workers.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	entries, _ := service.Repository.ListAssetLibrary(context.Background(), "alice", screenplay.AssetCharacter, "", "")
	if len(entries) != 1 || len(objects.files) != 2 {
		t.Fatal(len(entries), len(objects.files))
	}
}

func TestScreenplayAssetLibrarySharedReadDoesNotExportOwnersLibrary(t *testing.T) {
	service, asset, _, _, _ := assetLibraryServiceFixture(t)
	ctx := context.Background()
	authors := &screenplay.AuthoringService{Store: service.Media.Projects, Repository: screenplay.NewMemoryAuthoringRepository()}
	state, _ := authors.Load(ctx, "alice", asset.ProjectID)
	state.Members = []screenplay.MemberGrant{{Username: "reader", Read: true}}
	if _, err := authors.Save(ctx, "alice", asset.ProjectID, state, state.Revision); err != nil {
		t.Fatal(err)
	}
	service.Media.Authoring = authors
	if _, err := service.SaveSource(ctx, "reader", asset.ProjectID, sourceLibraryInput(service, asset)); !errors.Is(err, screenplay.ErrAccessDenied) {
		t.Fatal(err)
	}
}
