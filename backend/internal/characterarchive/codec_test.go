package characterarchive

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestNativeFixture(t *testing.T) {
	archive, err := os.ReadFile("testdata/native.lschar")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("testdata/native.json")
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeJSON(archive)
	if err != nil {
		t.Fatal(err)
	}
	var g, w any
	if json.Unmarshal(got, &g) != nil || json.Unmarshal(want, &w) != nil {
		t.Fatal("invalid fixture JSON")
	}
	gb, _ := json.Marshal(g)
	wb, _ := json.Marshal(w)
	if !bytes.Equal(gb, wb) {
		t.Fatal("native fixture semantic mismatch")
	}
}

func TestRoundTripAndRandomness(t *testing.T) {
	for _, source := range []string{`[]`, `{"unknown":{"a":[1,true,null]},"名字":"自造角色"}`, `42`, `null`} {
		a, err := EncodeJSON([]byte(source))
		if err != nil {
			t.Fatal(err)
		}
		b, err := EncodeJSON([]byte(source))
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Equal(a, b) {
			t.Fatal("encryption must use fresh randomness")
		}
		got, err := DecodeJSON(a)
		if err != nil || string(got) != source {
			t.Fatalf("roundtrip: %s %v", got, err)
		}
	}
}
