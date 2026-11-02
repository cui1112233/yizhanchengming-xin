package httpapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// TestAutomationStagesContract 守护前后端阶段定义一致：JSON 是唯一真相文件，
// Go 常量只是它的镜像。改阶段不改两侧时本测试先红。
func TestAutomationStagesContract(t *testing.T) {
	path := filepath.Join("..", "..", "..", "frontend", "src", "user", "pages", "shuihuo", "automationStages.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read stage contract: %v", err)
	}
	var contract struct {
		Stages []automationStageDefinition `json:"stages"`
	}
	if err := json.Unmarshal(raw, &contract); err != nil {
		t.Fatalf("parse stage contract: %v", err)
	}
	if len(contract.Stages) == 0 {
		t.Fatal("stage contract is empty")
	}

	sortStages := func(stages []automationStageDefinition) {
		sort.Slice(stages, func(i, j int) bool {
			if stages[i].ID != stages[j].ID {
				return stages[i].ID < stages[j].ID
			}
			return stages[i].Mode < stages[j].Mode
		})
	}
	got := append([]automationStageDefinition(nil), automationStageContract...)
	want := append([]automationStageDefinition(nil), contract.Stages...)
	sortStages(got)
	sortStages(want)

	if len(got) != len(want) {
		t.Fatalf("stage count mismatch: go=%d json=%d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("stage mismatch at %d: go=%+v json=%+v", i, got[i], want[i])
		}
	}
}
