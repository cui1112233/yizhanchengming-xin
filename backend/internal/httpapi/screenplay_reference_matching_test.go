package httpapi

import (
	"bytes"
	"context"
	"reflect"
	"sort"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/screenplay"
)

// A legacy import can contain duplicate names even though CreateAsset rejects them.
type screenplayReferenceDuplicateStore struct {
	screenplay.Store
	duplicate screenplay.Asset
}

func (s screenplayReferenceDuplicateStore) ListAssets(ctx context.Context, owner, project string, kind screenplay.AssetType) ([]screenplay.Asset, error) {
	assets, err := s.Store.ListAssets(ctx, owner, project, kind)
	if err == nil && (kind == "" || kind == s.duplicate.Type) {
		assets = append(assets, s.duplicate)
	}
	return assets, err
}

func TestScreenplayReferencesMatchCompleteNamesAndTypes(t *testing.T) {
	for _, tc := range []struct {
		name          string
		shot          screenplay.StoryboardShot
		want          []string
		duplicate     bool
		shareSelected bool
	}{
		{name: "long name excludes its prefix", shot: screenplay.StoryboardShot{Characters: "林晚晴"}, want: []string{"long"}},
		{name: "short name remains supported", shot: screenplay.StoryboardShot{Characters: "林晚"}, want: []string{"short"}},
		{name: "explicit short and long names", shot: screenplay.StoryboardShot{Characters: " 林晚、 林晚晴 "}, want: []string{"short", "long"}},
		{name: "existing slash separated names", shot: screenplay.StoryboardShot{Characters: "林晚/林晚晴"}, want: []string{"short", "long"}},
		{name: "existing at separated names", shot: screenplay.StoryboardShot{Characters: "@林晚@林晚晴"}, want: []string{"short", "long"}},
		{name: "all existing list separators", shot: screenplay.StoryboardShot{Characters: "林晚,林晚晴，Alice Smith;林晚；林晚晴\n林晚|林晚晴"}, want: []string{"short", "long", "english"}},
		{name: "field types remain isolated", shot: screenplay.StoryboardShot{Characters: "花园", Scene: "林晚", Props: "林晚晴"}, want: []string{"scene-short", "prop-long"}},
		{name: "same name in multiple types is explicit", shot: screenplay.StoryboardShot{Characters: "林晚", Scene: "林晚", Props: "林晚"}, want: []string{"short", "scene-short", "prop-short"}},
		{name: "descriptive text is not an exact name", shot: screenplay.StoryboardShot{Characters: "林晚晴走向林晚", Scene: "林晚的家", Props: "林晚晴的手机"}},
		{name: "missing image does not fall back to prefix", shot: screenplay.StoryboardShot{Characters: "林晚晴未选图"}},
		{name: "empty fields"},
		{name: "none markers", shot: screenplay.StoryboardShot{Characters: "无", Scene: "none", Props: "无道具"}},
		{name: "duplicate legacy same type name is ambiguous", shot: screenplay.StoryboardShot{Characters: "林晚"}, duplicate: true},
		{name: "shot reference and repeated selections deduplicate", shot: screenplay.StoryboardShot{ID: "shot-1", Characters: "林晚、林晚晴、林晚"}, want: []string{"short"}, shareSelected: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			library, first, _, _, _ := assetLibraryServiceFixture(t)
			s := library.Media
			ctx := context.Background()
			p, err := s.Projects.GetProject(ctx, "alice", first.ProjectID)
			if err != nil {
				t.Fatal(err)
			}
			state, err := s.Media.LoadMedia(ctx, "alice", p.ID)
			if err != nil {
				t.Fatal(err)
			}
			firstItem, err := mediaItem(state, state.Selected[first.ID])
			if err != nil {
				t.Fatal(err)
			}
			mediaIDs := map[string]string{"short": firstItem.AssetID}
			for _, a := range []struct {
				key, name  string
				kind       screenplay.AssetType
				unselected bool
			}{
				{"long", "林晚晴", screenplay.AssetCharacter, false},
				{"english", "Alice Smith", screenplay.AssetCharacter, false},
				{"scene-short", "林晚", screenplay.AssetScene, false},
				{"prop-short", "林晚", screenplay.AssetProp, false},
				{"prop-long", "林晚晴", screenplay.AssetProp, false},
				{"unselected", "林晚晴未选图", screenplay.AssetCharacter, true},
			} {
				asset, err := s.Projects.CreateAsset(ctx, screenplay.Asset{OwnerUsername: "alice", ProjectID: p.ID, Type: a.kind, Name: a.name})
				if err != nil {
					t.Fatal(err)
				}
				if a.unselected {
					continue
				}
				item, err := s.saveBytes(ctx, p, "image", a.key+".png", "image/png", bytes.NewReader(testImageToolPNG(t)), asset.ID)
				if err != nil {
					t.Fatal(err)
				}
				mediaIDs[a.key] = item.AssetID
				_, err = s.Media.MutateMedia(ctx, "alice", p.ID, func(st *screenplay.MediaState) error {
					st.Selected[asset.ID] = item.ID
					if tc.shareSelected && a.key == "long" {
						st.Selected[asset.ID] = firstItem.ID
						st.Selected[tc.shot.ID] = firstItem.ID
					}
					return nil
				})
				if err != nil {
					t.Fatal(err)
				}
			}
			if tc.duplicate {
				dup := first
				dup.ID = "legacy-duplicate"
				s.Projects = screenplayReferenceDuplicateStore{Store: s.Projects, duplicate: dup}
			}
			s.ReferencesProvider = agentImageReferenceProvider(func(_ context.Context, asset agentworkspace.Asset, _ time.Duration) (string, error) {
				return "https://fixture.invalid/" + asset.ID, nil
			})
			wantIDs, wantURLs := []string{}, []string{}
			for _, key := range tc.want {
				wantIDs = append(wantIDs, mediaIDs[key])
				wantURLs = append(wantURLs, "https://fixture.invalid/"+mediaIDs[key])
			}
			gotIDs, err := s.ReferenceIDs(ctx, p, tc.shot)
			if err != nil {
				t.Fatal(err)
			}
			gotURLs, err := s.References(ctx, p, tc.shot)
			if err != nil {
				t.Fatal(err)
			}
			sort.Strings(wantIDs)
			sort.Strings(wantURLs)
			sort.Strings(gotIDs)
			sort.Strings(gotURLs)
			if !reflect.DeepEqual(gotIDs, wantIDs) {
				t.Errorf("ReferenceIDs = %v; want %v", gotIDs, wantIDs)
			}
			if !reflect.DeepEqual(gotURLs, wantURLs) {
				t.Errorf("References = %v; want %v", gotURLs, wantURLs)
			}
		})
	}
}
