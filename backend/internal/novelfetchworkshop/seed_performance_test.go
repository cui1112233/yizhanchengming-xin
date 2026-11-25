package novelfetchworkshop

import (
	"encoding/json"
	"reflect"
	"testing"
)

func BenchmarkEnsureWorkshopDefaults(b *testing.B) {
	for _, populated := range []bool{false, true} {
		name := "empty"
		settings := map[string]any{}
		if populated {
			name = "saved"
			settings = EnsureWorkshopDefaults(settings)
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = EnsureWorkshopDefaults(settings)
			}
		})
	}
}

func TestWorkshopDefaultsCopiesStayIsolated(t *testing.T) {
	first, second := DefaultWorkshopSettings(), DefaultWorkshopSettings()
	if !reflect.DeepEqual(first, second) {
		t.Fatal("fresh defaults differ")
	}
	first["knowledgeLibraries"].(map[string]any)["opening_phrases"].(map[string]any)["items"] = []any{"changed"}
	first["appConfig"].(map[string]any)["custom"] = "first user"
	if reflect.DeepEqual(first, second) {
		t.Fatal("test mutation did not change defaults")
	}
	if !reflect.DeepEqual(second, DefaultWorkshopSettings()) {
		t.Fatal("one user's edits leaked into another defaults read")
	}
}

func TestEnsureWorkshopDefaultsAllocationBudget(t *testing.T) {
	// A scheduler read must not repeatedly serialize the entire bundled catalog.
	// Keep an allocation bound, rather than a timing bound that varies by host.
	allocs := testing.AllocsPerRun(3, func() {
		_ = EnsureWorkshopDefaults(map[string]any{})
	})
	if allocs > 30000 {
		t.Fatalf("default config read allocates %.0f objects; budget is 30000", allocs)
	}
}

func TestWorkshopFastClonePreservesJSONValues(t *testing.T) {
	input := map[string]any{
		"integer": 12,
		"nested":  map[string]any{"enabled": false, "empty": "", "number": 1.5},
		"items":   []any{map[string]any{"text": "中文规则"}, true, nil},
		"typed":   []map[string]string{{"id": "custom"}},
		"raw":     json.RawMessage(`{"value":123}`),
		"nilMap":  map[string]any(nil), "nilSlice": []any(nil),
	}
	if got, want := cloneWorkshopValue(input), cloneMap(input); !reflect.DeepEqual(got, want) {
		t.Fatalf("fast clone changed JSON values: got %#v, want %#v", got, want)
	}
}

func TestWorkshopMergePreservesSavedOverrides(t *testing.T) {
	base := DefaultWorkshopSettings()
	override := map[string]any{
		"knowledgeLibraries": map[string]any{
			"opening_phrases": map[string]any{"items": []any{}, "source": "user"},
			"custom":          map[string]any{"items": []any{map[string]any{"text": "本团队规则"}}},
		},
		"appConfig":             map[string]any{"enabled": false, "limit": 200},
		"novel_fetch_schedules": []any{map[string]any{"status": "scheduled"}},
	}
	var legacyMerge func(map[string]any, map[string]any) map[string]any
	legacyMerge = func(base, override map[string]any) map[string]any {
		result := cloneMap(base)
		for key, value := range override {
			if isEmptyWorkshopValue(value) {
				continue
			}
			if nested, ok := value.(map[string]any); ok {
				if previous, ok := result[key].(map[string]any); ok {
					result[key] = legacyMerge(previous, nested)
				} else {
					result[key] = cloneMap(nested)
				}
			} else {
				result[key] = value
			}
		}
		return result
	}
	want := legacyMerge(base, override)
	if got := mergeWorkshopMaps(base, override); !reflect.DeepEqual(got, want) {
		t.Fatal("optimized merge changed default or saved values")
	}
	if !reflect.DeepEqual(base, DefaultWorkshopSettings()) {
		t.Fatal("merge mutated shared defaults")
	}
}

func TestWorkshopDefaultsConcurrentIsolation(t *testing.T) {
	for i := 0; i < 12; i++ {
		t.Run("owner", func(t *testing.T) {
			t.Parallel()
			settings := EnsureWorkshopDefaults(map[string]any{"owner": t.Name()})
			settings["knowledgeLibraries"].(map[string]any)["opening_phrases"] = "private edit"
			fresh := EnsureWorkshopDefaults(map[string]any{})
			if _, ok := fresh["knowledgeLibraries"].(map[string]any)["opening_phrases"].(map[string]any); !ok {
				t.Fatal("another owner's private edit leaked into defaults")
			}
		})
	}
}
