package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/screenplay"
)

func stageFixture(t *testing.T) (*ScreenplayAssetLibrary, screenplay.AssetLibraryEntry, screenplay.Asset, *http.ServeMux, time.Time) {
	t.Helper()
	s, source, _, _, now := assetLibraryServiceFixture(t)
	entry, err := s.SaveSource(context.Background(), "alice", source.ProjectID, sourceLibraryInput(s, source))
	if err != nil {
		t.Fatal(err)
	}
	project, err := s.Media.Projects.CreateProject(context.Background(), screenplay.Project{ID: "stage-target", OwnerUsername: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	target, err := s.Media.Projects.CreateAsset(context.Background(), screenplay.Asset{OwnerUsername: "alice", ProjectID: project.ID, Type: screenplay.AssetCharacter, Name: "目标甲", Description: "目标描述", Prompt: "目标提示词不改", Voice: "目标声音不改", Extra: map[string]string{"unchanged": "yes"}})
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterScreenplayAssetLibraryRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, s)
	return s, entry, target, mux, now
}
func stageInput(entry screenplay.AssetLibraryEntry, target screenplay.Asset) screenplayLibraryStageInput {
	return screenplayLibraryStageInput{RequestKey: "stage-request-001", TargetID: target.ID, ExpectedEntryRevision: &entry.Revision}
}
func TestScreenplayAssetLibraryStageCopiesWithoutApplyingAndCanBeReferenced(t *testing.T) {
	s, entry, target, mux, now := stageFixture(t)
	if err := s.Media.Projects.DeleteProject(context.Background(), "alice", entry.SourceProjectID); err != nil {
		t.Fatal(err)
	}
	path := "/api/comic-screenplays/projects/" + target.ProjectID + "/asset-library/" + entry.ID + "/stage"
	res := signedJSONRequest(t, mux, now, "alice", "POST", path, stageInput(entry, target))
	if res.Code != 200 {
		t.Fatal(res.Code, res.Body.String())
	}
	var data struct {
		Item screenplay.MediaItem `json:"item"`
	}
	json.Unmarshal(res.Body.Bytes(), &data)
	if data.Item.ID == "" || data.Item.AssetID != data.Item.ID || data.Item.Status != "succeeded" || data.Item.TargetID != target.ID || strings.Contains(res.Body.String(), "object_key") {
		t.Fatal(data, res.Body.String())
	}
	state, _ := s.Media.Media.LoadMedia(context.Background(), "alice", target.ProjectID)
	if state.Selected[target.ID] != "" || len(state.Items) != 1 {
		t.Fatal("staging selected the image", state)
	}
	assets, _ := s.Media.Projects.ListAssets(context.Background(), "alice", target.ProjectID, "")
	if len(assets) != 1 || assets[0].Prompt != target.Prompt || assets[0].Voice != target.Voice || assets[0].Name != target.Name || assets[0].Revision != target.Revision {
		t.Fatal(assets)
	}
	cut, err := s.Media.Cut.Asset(context.Background(), "alice", target.ProjectID, data.Item.AssetID)
	if err != nil || cut.ObjectKey == entry.ImageObject.ObjectKey {
		t.Fatal(cut, err)
	}
	if _, err = s.Media.Assets.Asset(context.Background(), "alice", target.ProjectID, data.Item.AssetID); err != nil {
		t.Fatal(err)
	}
	repeat := signedJSONRequest(t, mux, now, "alice", "POST", path, stageInput(entry, target))
	if repeat.Code != 200 {
		t.Fatal(repeat.Code, repeat.Body.String())
	}
	var again struct {
		Item screenplay.MediaItem `json:"item"`
	}
	json.Unmarshal(repeat.Body.Bytes(), &again)
	if again.Item.ID != data.Item.ID {
		t.Fatal("stage id changed")
	}
	state, _ = s.Media.Media.LoadMedia(context.Background(), "alice", target.ProjectID)
	if len(state.Items) != 1 {
		t.Fatal(state)
	}
	// Only the later editor save/select makes this staged image a reference.
	project, _ := s.Media.Projects.GetProject(context.Background(), "alice", target.ProjectID)
	if _, err = s.Media.action(context.Background(), project, screenplayMediaAction{Action: "select", TargetID: target.ID, ItemID: data.Item.ID}); err != nil {
		t.Fatal(err)
	}
	refs, err := s.Media.ReferenceIDs(context.Background(), project, screenplay.StoryboardShot{Characters: target.Name})
	if err != nil || len(refs) != 1 || refs[0] != data.Item.AssetID {
		t.Fatal(refs, err)
	}
}
func TestScreenplayAssetLibraryStageRejectsWrongOwnerTargetRevisionAndCancellation(t *testing.T) {
	s, entry, target, mux, now := stageFixture(t)
	ctx := context.Background()
	input := stageInput(entry, target)
	stale := entry.Revision + 1
	input.ExpectedEntryRevision = &stale
	if _, err := s.StageImage(ctx, "alice", target.ProjectID, entry.ID, input); !errors.Is(err, screenplay.ErrRevisionConflict) {
		t.Fatal(err)
	}
	input = stageInput(entry, target)
	input.TargetID = "missing"
	if _, err := s.StageImage(ctx, "alice", target.ProjectID, entry.ID, input); !errors.Is(err, screenplay.ErrNotFound) {
		t.Fatal(err)
	}
	s.Media.Projects.CreateProject(ctx, screenplay.Project{ID: target.ProjectID, OwnerUsername: "bob"})
	bobAsset, _ := s.Media.Projects.CreateAsset(ctx, screenplay.Asset{OwnerUsername: "bob", ProjectID: target.ProjectID, Type: screenplay.AssetCharacter, Name: "bob"})
	input = stageInput(entry, bobAsset)
	if _, err := s.StageImage(ctx, "bob", target.ProjectID, entry.ID, input); !errors.Is(err, screenplay.ErrNotFound) {
		t.Fatal("cross account entry leaked", err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := s.StageImage(cancelled, "alice", target.ProjectID, entry.ID, stageInput(entry, target)); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	result := signedJSONRequest(t, mux, now, "alice", "POST", "/api/comic-screenplays/projects/"+target.ProjectID+"/asset-library/"+entry.ID+"/stage", map[string]any{"requestKey": "safe-request", "targetId": target.ID})
	if result.Code != 400 {
		t.Fatal(result.Code)
	}
	state, _ := s.Media.Media.LoadMedia(ctx, "alice", target.ProjectID)
	if len(state.Items) != 0 || len(state.Selected) != 0 {
		t.Fatal(state)
	}
}

type interruptedStageAssetStore struct {
	agentworkspace.AssetStore
	failOnce       bool
	commitThenFail bool
}

func (s *interruptedStageAssetStore) CreateAsset(ctx context.Context, owner string, a agentworkspace.Asset) error {
	if s.failOnce {
		s.failOnce = false
		if s.commitThenFail {
			if err := s.AssetStore.CreateAsset(ctx, owner, a); err != nil {
				return err
			}
		}
		return fmt.Errorf("simulated agent commit acknowledgement failure")
	}
	return s.AssetStore.CreateAsset(ctx, owner, a)
}

type interruptedStageMediaStore struct {
	screenplay.MediaStore
	failOnce       bool
	commitThenFail bool
}

func (s *interruptedStageMediaStore) MutateMedia(ctx context.Context, o, p string, fn func(*screenplay.MediaState) error) (screenplay.MediaState, error) {
	if s.failOnce {
		s.failOnce = false
		if s.commitThenFail {
			state, err := s.MediaStore.MutateMedia(ctx, o, p, fn)
			if err != nil {
				return state, err
			}
			return state, fmt.Errorf("simulated media commit acknowledgement lost")
		}
		return screenplay.MediaState{}, fmt.Errorf("media write unavailable")
	}
	return s.MediaStore.MutateMedia(ctx, o, p, fn)
}
func TestScreenplayAssetLibraryStageRepairsPartialWritesAndLostAcknowledgement(t *testing.T) {
	for _, mode := range []string{"agent-failed", "agent-ack-lost", "media-failed", "media-ack-lost"} {
		t.Run(mode, func(t *testing.T) {
			s, entry, target, _, _ := stageFixture(t)
			if strings.HasPrefix(mode, "agent") {
				s.Media.Assets = &interruptedStageAssetStore{AssetStore: s.Media.Assets, failOnce: true, commitThenFail: strings.Contains(mode, "ack-lost")}
			} else {
				s.Media.Media = &interruptedStageMediaStore{MediaStore: s.Media.Media, failOnce: true, commitThenFail: strings.Contains(mode, "ack-lost")}
			}
			first, err := s.StageImage(context.Background(), "alice", target.ProjectID, entry.ID, stageInput(entry, target))
			if strings.Contains(mode, "ack-lost") {
				if err != nil || first.AssetID == "" {
					t.Fatal(first, err)
				}
			} else if err == nil {
				t.Fatal("failure swallowed")
			}
			repaired, err := s.StageImage(context.Background(), "alice", target.ProjectID, entry.ID, stageInput(entry, target))
			if err != nil {
				t.Fatal(err)
			}
			cut, err := s.Media.Cut.Asset(context.Background(), "alice", target.ProjectID, repaired.AssetID)
			if err != nil {
				t.Fatal(err)
			}
			body, err := s.Media.Objects.Open(context.Background(), aicutObjectAsset(cut))
			if err != nil {
				t.Fatal("committed object was cleaned", err)
			}
			body.Close()
			if _, err = s.Media.Assets.Asset(context.Background(), "alice", target.ProjectID, repaired.AssetID); err != nil {
				t.Fatal(err)
			}
			state, _ := s.Media.Media.LoadMedia(context.Background(), "alice", target.ProjectID)
			if len(state.Items) != 1 || len(state.Selected) != 0 {
				t.Fatal(state)
			}
		})
	}
}

type interruptedStageCutStore struct {
	aicut.Store
	failOnce       bool
	commitThenFail bool
}

func (s *interruptedStageCutStore) CreateAsset(ctx context.Context, owner string, a aicut.Asset) error {
	if s.failOnce {
		s.failOnce = false
		if s.commitThenFail {
			if err := s.Store.CreateAsset(ctx, owner, a); err != nil {
				return err
			}
		}
		return fmt.Errorf("cut write acknowledgement failed")
	}
	return s.Store.CreateAsset(ctx, owner, a)
}
func TestScreenplayAssetLibraryStageRecoversCutWriteFailure(t *testing.T) {
	for _, committed := range []bool{false, true} {
		t.Run(fmt.Sprint(committed), func(t *testing.T) {
			s, entry, target, _, _ := stageFixture(t)
			s.Media.Cut = &interruptedStageCutStore{Store: s.Media.Cut, failOnce: true, commitThenFail: committed}
			_, err := s.StageImage(context.Background(), "alice", target.ProjectID, entry.ID, stageInput(entry, target))
			if committed && err != nil {
				t.Fatal("committed cut could not be verified", err)
			}
			if !committed && err == nil {
				t.Fatal("cut failure swallowed")
			}
			if _, err = s.StageImage(context.Background(), "alice", target.ProjectID, entry.ID, stageInput(entry, target)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestScreenplayAssetLibraryStageRechecksOwnerAfterLock(t *testing.T) {
	s, entry, target, _, _ := stageFixture(t)
	memory := s.Media.Projects.(*screenplay.MemoryStore)
	authors := &screenplay.AuthoringService{Store: memory, Repository: screenplay.NewMemoryAuthoringRepository()}
	s.Media.Authoring = authors
	race := &assetCheckGrantRaceStore{MemoryStore: memory}
	authors.Store = race
	race.beforeLock = func() { memory.DeleteProject(context.Background(), "alice", target.ProjectID) }
	if _, err := s.StageImage(context.Background(), "alice", target.ProjectID, entry.ID, stageInput(entry, target)); err == nil {
		t.Fatal("deleted project passed lock recheck")
	}
}

func TestScreenplayAssetLibraryStageConcurrentReplayKeepsOneItem(t *testing.T) {
	s, entry, target, _, _ := stageFixture(t)
	project, _ := s.Media.Projects.GetProject(context.Background(), "alice", target.ProjectID)
	if err := s.Media.ensureCut(context.Background(), project); err != nil {
		t.Fatal(err)
	}
	var workers sync.WaitGroup
	results := make(chan screenplay.MediaItem, 8)
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			item, err := s.StageImage(context.Background(), "alice", target.ProjectID, entry.ID, stageInput(entry, target))
			results <- item
			errs <- err
		}()
	}
	workers.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	id := ""
	for item := range results {
		if id == "" {
			id = item.ID
		}
		if item.ID != id {
			t.Fatal("concurrent stage created multiple identities", item.ID, id)
		}
	}
	state, _ := s.Media.Media.LoadMedia(context.Background(), "alice", target.ProjectID)
	if len(state.Items) != 1 || len(state.Selected) != 0 {
		t.Fatal(state)
	}
}

type cancellingStageObjectStore struct {
	agentworkspace.AssetObjectStore
	cancel context.CancelFunc
}

func (s *cancellingStageObjectStore) Put(ctx context.Context, a agentworkspace.Asset, body io.Reader) error {
	err := s.AssetObjectStore.Put(ctx, a, body)
	if s.cancel != nil {
		cancel := s.cancel
		s.cancel = nil
		cancel()
	}
	return err
}
func TestScreenplayAssetLibraryStageCancellationAfterCopyCanReplaySafely(t *testing.T) {
	s, entry, target, _, _ := stageFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	s.Media.Objects = &cancellingStageObjectStore{AssetObjectStore: s.Media.Objects, cancel: cancel}
	if _, err := s.StageImage(ctx, "alice", target.ProjectID, entry.ID, stageInput(entry, target)); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	state, _ := s.Media.Media.LoadMedia(context.Background(), "alice", target.ProjectID)
	if len(state.Items) != 0 || len(state.Selected) != 0 {
		t.Fatal("cancel applied image", state)
	}
	item, err := s.StageImage(context.Background(), "alice", target.ProjectID, entry.ID, stageInput(entry, target))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Media.Cut.Asset(context.Background(), "alice", target.ProjectID, item.AssetID); err != nil {
		t.Fatal(err)
	}
	state, _ = s.Media.Media.LoadMedia(context.Background(), "alice", target.ProjectID)
	if len(state.Items) != 1 || len(state.Selected) != 0 {
		t.Fatal(state)
	}
}

type privateFailureStageObjects struct {
	agentworkspace.AssetObjectStore
}

func (s privateFailureStageObjects) Open(_ context.Context, a agentworkspace.Asset) (io.ReadCloser, error) {
	return nil, fmt.Errorf("provider private storage key: %s", a.ObjectKey)
}
func TestScreenplayAssetLibraryStageReadFailureDoesNotExposeLibraryObjectKey(t *testing.T) {
	s, entry, target, mux, now := stageFixture(t)
	s.Media.Objects = privateFailureStageObjects{s.Media.Objects}
	result := signedJSONRequest(t, mux, now, "alice", "POST", "/api/comic-screenplays/projects/"+target.ProjectID+"/asset-library/"+entry.ID+"/stage", stageInput(entry, target))
	if result.Code < 400 || strings.Contains(result.Body.String(), entry.ImageObject.ObjectKey) {
		t.Fatal(result.Code, result.Body.String())
	}
}
