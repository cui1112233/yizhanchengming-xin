package batchfactorytext

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

type workflowRepo struct {
	Repository
	doc         Document
	pages       []Page
	collections []Collection
}

func (r *workflowRepo) GetDocument(context.Context, Scope) (Document, error) {
	if r.doc.ID == "" {
		return Document{}, ErrNotFound
	}
	return r.doc, nil
}
func (r *workflowRepo) ListPages(context.Context, Scope) ([]Page, error) {
	pages := append([]Page(nil), r.pages...)
	for i := range pages {
		if pages[i].ActiveArtifactID != "" {
			pages[i].Status = "ready"
		}
	}
	return pages, nil
}
func (r *workflowRepo) ListArtifacts(context.Context, Scope) ([]PageArtifact, error) {
	arts := []PageArtifact{{ID: r.doc.CoverArtifactID, Kind: "cover", SourceKind: sourceKind(r.doc.CoverLocked)}}
	for _, p := range r.pages {
		if p.ActiveArtifactID != "" {
			arts = append(arts, PageArtifact{ID: p.ActiveArtifactID, PageID: p.ID, Kind: "page", SourceKind: sourceKind(p.Locked), RenderVersion: p.RenderVersion})
		}
	}
	return arts, nil
}
func (r *workflowRepo) ListCollections(context.Context, Scope) ([]Collection, error) {
	return r.collections, nil
}
func (r *workflowRepo) GetCollection(_ context.Context, _ Scope, id string) (Collection, error) {
	for _, c := range r.collections {
		if c.ID == id {
			return c, nil
		}
	}
	return Collection{}, ErrNotFound
}
func (r *workflowRepo) CreateCollection(_ context.Context, _ Scope, in CollectionInput) (Collection, error) {
	c := Collection{ID: fmt.Sprintf("collection-%d", len(r.collections)+1), Status: "ready", VersionID: in.VersionID, PaginationVersion: in.PaginationVersion, CoverArtifactID: r.doc.CoverArtifactID}
	for _, p := range r.pages {
		c.Items = append(c.Items, CollectionItem{Ordinal: p.Ordinal, PageID: p.ID, ArtifactID: p.ActiveArtifactID})
	}
	r.collections = append(r.collections, c)
	return c, nil
}

func TestTextCollectionManifestDoesNotDependOnWorkCover(t *testing.T) {
	ctx := context.Background()
	scope := Scope{"owner", "batch", "book"}
	doc := Document{ID: "doc", ActiveVersionID: "version", PaginationVersion: 1, Revision: 1, CoverArtifactID: "current-cover"}
	page := Page{ID: "p1", VersionID: "version", PaginationVersion: 1, Ordinal: 1, Status: "ready", ActiveArtifactID: "page-1", RenderVersion: 1}
	repo := &workflowRepo{doc: doc, pages: []Page{page}}
	collection := Collection{ID: "manual", Status: "ready", VersionID: "version", PaginationVersion: 1, CoverArtifactID: "older-cover", Items: []CollectionItem{{Ordinal: 1, PageID: "p1", ArtifactID: "page-1"}}}
	manifest := [][2]string{{"p1", "page-1"}}

	if !matchesPages(collection, doc, manifest) {
		t.Fatal("matching page manifest was rejected only because the external cover changed")
	}
	before := workflowCollectionID(scope, collection)
	collection.CoverArtifactID = "another-cover"
	if after := workflowCollectionID(scope, collection); after != before {
		t.Fatalf("collection id changed with external cover: %s != %s", after, before)
	}
	w := Workflow{Repo: repo, VerifyFile: func(context.Context, PageArtifact, string) error { return nil }}
	if err := w.verifyManifest(ctx, scope, collection, false); err != nil {
		t.Fatalf("page-only manifest rejected: %v", err)
	}
	repairs, err := w.inspectFiles(ctx, scope, Document{ID: "doc", ActiveVersionID: "version", PaginationVersion: 1}, []Page{page})
	if err != nil || repairs[""] != nil {
		t.Fatalf("missing work cover entered collection repair set: repairs=%v err=%v", repairs, err)
	}
}

func TestTextWorkflowFingerprintsFollowChangedSettings(t *testing.T) {
	ctx := context.Background()
	repo := &workflowRepo{}
	src := WorkflowProduction{VersionInput: VersionInput{ProductionText: "正文"}, Snapshot: json.RawMessage(`{"layout":{"fontSize":24},"sourceVersion":"v"}`)}
	var reflows, renders, publishes int
	journal := map[string]string{}
	w := Workflow{Repo: repo, VerifyPublished: func(context.Context, EnsureRequest, Collection) error { return nil }, VerifyFile: func(context.Context, PageArtifact, string) error { return nil },
		Track: func(_ context.Context, _ Scope, checkpoint StageCheckpoint, reuse bool, run func() error) error {
			stage, key := checkpoint.Stage, checkpoint.Fingerprint
			if reuse && journal[stage] == key {
				return nil
			}
			if err := run(); err != nil {
				return err
			}
			journal[stage] = key
			return nil
		},
		Process: func(context.Context, EnsureRequest, bool) (WorkflowProduction, error) { return src, nil },
		Prepare: func(_ context.Context, _ EnsureRequest, p WorkflowProduction, doc Document, reflow bool) error {
			if reflow {
				reflows++
			}
			repo.doc = Document{ID: "doc", ActiveVersionID: "v", PaginationVersion: doc.PaginationVersion + 1, ConfigSnapshot: p.Snapshot}
			repo.pages = []Page{{ID: "p", VersionID: "v", PaginationVersion: repo.doc.PaginationVersion, Ordinal: 1}}
			return nil
		},
		Render: func(_ context.Context, _ EnsureRequest, _ Document, p Page) error {
			renders++
			if p.ID == "" {
				repo.doc.CoverArtifactID = fmt.Sprint(renders)
			} else {
				repo.pages[0].ActiveArtifactID = fmt.Sprint(renders)
			}
			return nil
		},
		Validate: func(context.Context, EnsureRequest, Document) error { return nil },
		Publish:  func(context.Context, EnsureRequest, Collection) (bool, error) { publishes++; return true, nil },
	}
	req := EnsureRequest{Scope: Scope{"owner", "batch", "book"}, Target: "published", PublishSettingsKey: "original-settings", FrozenSettings: map[string]json.RawMessage{}}
	ensure := func() {
		t.Helper()
		res, err := w.Ensure(ctx, req)
		if err != nil || !res.Published {
			t.Fatalf("%+v %v", res, err)
		}
	}
	ensure()
	initialPageKey := journal["text_paginate"]
	initialRenderKey := journal["text_render"]
	req.PublishSettingsKey = "organization-2"
	req.FrozenSettings["comicPublishSettings"] = json.RawMessage(`{"organizationId":"2"}`)
	ensure()
	if reflows != 1 || renders != 1 || publishes != 2 || len(repo.collections) != 1 {
		t.Fatalf("mapping change: layout=%d render=%d publish=%d collections=%d", reflows, renders, publishes, len(repo.collections))
	}
	src.Snapshot = json.RawMessage(`{"layout":{"fontSize":24,"textColor":"#101010"},"sourceVersion":"v"}`)
	ensure()
	if reflows != 1 || renders != 2 || initialPageKey != journal["text_paginate"] || initialRenderKey == journal["text_render"] || len(repo.collections) != 2 {
		t.Fatalf("color change: reflows=%d renders=%d journal=%v", reflows, renders, journal)
	}
	src.Snapshot = json.RawMessage(`{"layout":{"fontSize":48},"sourceVersion":"v"}`)
	ensure()
	if reflows != 2 || renders != 3 || initialPageKey == journal["text_paginate"] || len(repo.collections) != 3 {
		t.Fatalf("layout change: reflows=%d renders=%d", reflows, renders)
	}
}

func TestTextWorkflowResumesFailedPageAndReusesCollection(t *testing.T) {
	ctx := context.Background()
	repo := &workflowRepo{}
	var calls []string
	failed := false
	validationFailed := false
	brokenID := ""
	source := WorkflowProduction{VersionInput: VersionInput{ProductionText: "甲乙丙丁", SensitiveMode: "off"}, Snapshot: json.RawMessage(`{"layout":{},"sourceVersion":"v"}`)}
	processed := false
	stages := map[string]string{}
	workflow := func() Workflow {
		return Workflow{Repo: repo, VerifyFile: func(_ context.Context, art PageArtifact, _ string) error {
			if art.ID == brokenID {
				return ErrNotReady
			}
			return nil
		},
			Track: func(_ context.Context, _ Scope, checkpoint StageCheckpoint, reuse bool, run func() error) error {
				stage, key := checkpoint.Stage, checkpoint.Fingerprint
				if reuse && stages[stage] == key {
					return nil
				}
				if err := run(); err != nil {
					return err
				}
				stages[stage] = key
				return nil
			},
			Process: func(_ context.Context, _ EnsureRequest, verify bool) (WorkflowProduction, error) {
				if verify && !processed {
					return WorkflowProduction{}, ErrNotReady
				}
				if !verify {
					calls = append(calls, "process")
					processed = true
				}
				return source, nil
			},
			Prepare: func(_ context.Context, _ EnsureRequest, p WorkflowProduction, _ Document, _ bool) error {
				calls = append(calls, "paginate")
				repo.doc = Document{ID: "doc", ActiveVersionID: "version", PaginationVersion: 1, Revision: 1, ConfigSnapshot: p.Snapshot, CoverArtifactID: "uploaded-cover", CoverLocked: true}
				repo.pages = []Page{{ID: "1", Ordinal: 1}, {ID: "2", Ordinal: 2}, {ID: "3", Ordinal: 3, Locked: true, ActiveArtifactID: "uploaded"}, {ID: "4", Ordinal: 4}}
				for i := range repo.pages {
					repo.pages[i].VersionID = "version"
					repo.pages[i].PaginationVersion = 1
				}
				return nil
			},
			Render: func(_ context.Context, _ EnsureRequest, _ Document, p Page) error {
				calls = append(calls, p.ID)
				if p.ID == "2" && !failed {
					failed = true
					return errors.New("测试渲染失败")
				}
				for i := range repo.pages {
					if repo.pages[i].ID == p.ID {
						if p.ActiveArtifactID == brokenID {
							brokenID = ""
						}
						repo.pages[i].ActiveArtifactID = "img" + p.ID
					}
				}
				return nil
			},
			Validate: func(context.Context, EnsureRequest, Document) error {
				calls = append(calls, "validate")
				if !validationFailed {
					validationFailed = true
					brokenID = "img4"
					return ErrNotReady
				}
				return nil
			},
		}
	}
	req := EnsureRequest{Scope: Scope{"owner", "batch", "book"}, Target: "collection"}
	result, err := workflow().Ensure(ctx, req)
	if err == nil || result.Stage != "text_render" {
		t.Fatalf("unexpected first result %+v %v", result, err)
	}
	result, err = workflow().Ensure(ctx, req)
	if !errors.Is(err, ErrNotReady) || result.Stage != "text_validate" || len(repo.collections) != 0 {
		t.Fatalf("first validation: %+v %v", result, err)
	}
	for i := 0; i < 2; i++ {
		result, err = workflow().Ensure(ctx, req)
		if err != nil || result.Collection.ID != "collection-1" {
			t.Fatalf("retry %d: %+v %v", i, result, err)
		}
	}
	if want := []string{"process", "paginate", "1", "2", "4", "2", "validate", "4", "validate"}; !reflect.DeepEqual(calls, want) || len(repo.collections) != 1 {
		t.Fatalf("calls %v, collections %d", calls, len(repo.collections))
	}
	// The artifact rows are still valid, but this stored collection claims a different SHA.
	repo.collections[0].Items[0].SHA256 = "wrong"
	result, err = workflow().Ensure(ctx, req)
	if !errors.Is(err, ErrNotReady) || result.Stage != "text_collection" || result.Collection.ID != "" {
		t.Fatalf("invalid frozen manifest was reused: %+v %v", result, err)
	}
	repo.collections[0].Items[0].SHA256 = ""
	req.Target = "published"
	req.PublishSettingsKey = "frozen-mapping"
	w := workflow()
	w.Publish = func(context.Context, EnsureRequest, Collection) (bool, error) { return false, nil }
	result, err = w.Ensure(ctx, req)
	if err == nil || result.Published || result.Stage != "text_publish" {
		t.Fatalf("unconfirmed publish %+v, %v", result, err)
	}
	// A deterministic collection also pins its pagination and all frozen references.
	req.CollectionID = workflowCollectionID(req.Scope, repo.collections[0])
	repo.collections[0].ID = req.CollectionID
	repo.collections[0].PaginationVersion++
	if _, err := w.Ensure(ctx, req); !errors.Is(err, ErrNotReady) {
		t.Fatalf("wrong pinned pagination: %v", err)
	}
}
