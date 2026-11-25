package novelcreation

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestValidateStateConsistencyFlagsStructuralConflicts(t *testing.T) {
	previous := []byte(`{"facts":[{"id":"fact-1","statement":"林澜拿到指纹信"}],"characterChanges":[],"timeline":[],"foreshadowing":[{"thread":"监视盲区","status":"resolved"}]}`)
	settlement := truthSettlementOutput{
		Facts: []settledFact{
			{ID: "fact-1", Statement: "林澜丢掉了指纹信"},
			{ID: "fact-1", Statement: "重复编号"},
		},
		CharacterChanges: []settledCharacterChange{{Character: "陌生人", Change: "出现"}},
		Timeline:         []settledTimelineEvent{{Event: "收到匿名信", Position: "第一夜"}},
		Foreshadowing:    []settledForeshadowing{{Thread: "监视盲区", Status: "open"}},
	}
	got := validateStateConsistency(previous, settlement, foundationPlanOutput{Characters: []foundationCharacter{{Name: "林澜"}}})
	if got.Passed {
		t.Fatalf("duplicate fact id must fail: %+v", got)
	}
	joined := strings.Join(got.Warnings, "\n")
	for _, want := range []string{"重复", "表述与已保存版本不同", "再次以未解决状态出现", "不存在的角色"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in %v", want, got.Warnings)
		}
	}
}

func TestValidateStateConsistencyAcceptsCleanSettlement(t *testing.T) {
	settlement := truthSettlementOutput{
		Facts:            []settledFact{{ID: "fact-1", Statement: "林澜拿到指纹信"}},
		CharacterChanges: []settledCharacterChange{{Character: "林澜", Change: "决定继续调查"}},
		Timeline:         []settledTimelineEvent{{Event: "收到匿名信", Position: "第一夜"}},
		Foreshadowing:    []settledForeshadowing{{Thread: "监视盲区", Status: "open"}},
	}
	got := validateStateConsistency(nil, settlement, foundationPlanOutput{Characters: []foundationCharacter{{Name: "林澜"}}})
	if !got.Passed || len(got.Warnings) != 0 {
		t.Fatalf("clean settlement flagged: %+v", got)
	}
}

func TestValidateStateConsistencyRejectsCorruptPreviousTruth(t *testing.T) {
	got := validateStateConsistency([]byte(`not json`), truthSettlementOutput{Facts: []settledFact{}, CharacterChanges: []settledCharacterChange{}, Timeline: []settledTimelineEvent{}, Foreshadowing: []settledForeshadowing{}}, foundationPlanOutput{})
	if got.Passed {
		t.Fatalf("corrupt previous truth must fail: %+v", got)
	}
}

func TestStateValidationVersionGatesPipelineCheck(t *testing.T) {
	for _, tc := range []struct {
		version int
		want    bool
	}{{0, false}, {1, true}} {
		input := fixturePipelineRunContext(NewMemoryRuntime())
		a := NormalizeAgentSettings(AgentSettings{}, 2000)
		a.EnableLengthNorm = false
		input.Agent = &a
		input.StateValidationVersion = tc.version
		provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON}}
		sink := &recordingEventSink{}
		bundle, e := pipelineWithProvider(provider).Run(context.Background(), input, sink)
		if e != nil {
			t.Fatal(e)
		}
		var snapshot map[string]json.RawMessage
		if e = json.Unmarshal(bundle.RuntimeSnapshot, &snapshot); e != nil {
			t.Fatal(e)
		}
		_, present := snapshot["stateValidation"]
		if present != tc.want {
			t.Fatalf("version=%d stateValidation present=%v want=%v", tc.version, present, tc.want)
		}
		if len(provider.requests) != 6 {
			t.Fatalf("version=%d must not add model calls: %d", tc.version, len(provider.requests))
		}
	}
}
