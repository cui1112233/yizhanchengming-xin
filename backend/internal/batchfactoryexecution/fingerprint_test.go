package batchfactoryexecution

import (
	"encoding/json"
	"math"
	"testing"
)

func TestCanonicalJSONStableMapOrder(t *testing.T) {
	a := map[string]any{"z": 1, "nested": map[string]any{"b": true, "a": "x"}}
	b := map[string]any{"nested": map[string]any{"a": "x", "b": true}, "z": 1}
	gotA, err := CanonicalJSON(a)
	if err != nil { t.Fatal(err) }
	gotB, err := CanonicalJSON(b)
	if err != nil { t.Fatal(err) }
	if string(gotA) != string(gotB) {
		t.Fatalf("canonical json differs: %s != %s", gotA, gotB)
	}
	if string(gotA) != `{"nested":{"a":"x","b":true},"z":1}` {
		t.Fatalf("unexpected canonical json: %s", gotA)
	}
}

func TestCanonicalJSONPreservesArrayAndJSONNumber(t *testing.T) {
	got, err := CanonicalJSON(map[string]any{"items": []any{"b", "a"}, "n": json.Number("1.20")})
	if err != nil { t.Fatal(err) }
	if string(got) != `{"items":["b","a"],"n":1.20}` {
		t.Fatalf("unexpected canonical json: %s", got)
	}
	if _, err := CanonicalJSON(math.NaN()); err == nil {
		t.Fatal("NaN must be rejected")
	}
}

func TestConfigDigestStable(t *testing.T) {
	a, err := ConfigDigest(map[string]any{"model": "m1", "promptVersion": 2})
	if err != nil { t.Fatal(err) }
	b, err := ConfigDigest(map[string]any{"promptVersion": 2, "model": "m1"})
	if err != nil { t.Fatal(err) }
	if a != b || len(a) != 64 { t.Fatalf("digest mismatch: %q %q", a, b) }
}

func TestStageInputFingerprintChangesForExecutionInputs(t *testing.T) {
	base := map[string]any{"promptVersion": 1, "sourceDigest": "s1", "modelId": "m1", "assets": []any{"a"}}
	digest, err := ConfigDigest(map[string]any{"mode": "video"})
	if err != nil { t.Fatal(err) }
	original, err := StageInputFingerprint("image", digest, base)
	if err != nil { t.Fatal(err) }
	variants := []struct{name, stage, config string; input map[string]any}{
		{"stage", "video", digest, base},
		{"config", "image", "different", base},
		{"prompt", "image", digest, map[string]any{"promptVersion": 2, "sourceDigest": "s1", "modelId": "m1", "assets": []any{"a"}}},
		{"source", "image", digest, map[string]any{"promptVersion": 1, "sourceDigest": "s2", "modelId": "m1", "assets": []any{"a"}}},
		{"model", "image", digest, map[string]any{"promptVersion": 1, "sourceDigest": "s1", "modelId": "m2", "assets": []any{"a"}}},
		{"assets", "image", digest, map[string]any{"promptVersion": 1, "sourceDigest": "s1", "modelId": "m1", "assets": []any{"b"}}},
	}
	for _, variant := range variants {
		t.Run(variant.name, func(t *testing.T) {
			got, err := StageInputFingerprint(variant.stage, variant.config, variant.input)
			if err != nil { t.Fatal(err) }
			if got == original { t.Fatalf("fingerprint did not change for %s", variant.name) }
		})
	}
}
