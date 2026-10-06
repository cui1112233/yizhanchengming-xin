package shuihuo

import (
	"reflect"
	"testing"
)

func TestParseSmartCandidatesJSONAndFence(t *testing.T) {
	got, err := ParseSmartCandidates("```json\n[{\"text\":\"镜头一\"},{\"text\":\"镜头二\",\"speaker\":\"角色甲\"}]\n```")
	if err != nil {
		t.Fatal(err)
	}
	want := []Candidate{{Text: "镜头一", Speaker: "旁白"}, {Text: "镜头二", Speaker: "角色甲"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%#v", got)
	}
}

func TestParseSmartCandidatesLineFallback(t *testing.T) {
	got, err := ParseSmartCandidates("镜头一\n镜头二")
	if err != nil {
		t.Fatal(err)
	}
	want := []Candidate{{Text: "镜头一", Speaker: "旁白"}, {Text: "镜头二", Speaker: "旁白"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%#v", got)
	}
}
