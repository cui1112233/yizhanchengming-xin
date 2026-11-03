package httpapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
)

// TestVideoProviderCapabilityContract 守护前后端视频通道定义一致：JSON 是唯一
// 真相文件，Go 镜像和前端下拉都从它派生。任何一方新增/删除通道而未同步时本
// 测试先红。
func TestVideoProviderCapabilityContract(t *testing.T) {
	path := filepath.Join("..", "..", "..", "frontend", "src", "user", "pages", "batch-factory-v11", "videoProviderCapabilities.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read provider contract: %v", err)
	}
	var contract struct {
		Providers []VideoProviderCapability `json:"providers"`
	}
	if err := json.Unmarshal(raw, &contract); err != nil {
		t.Fatalf("parse provider contract: %v", err)
	}
	if len(contract.Providers) == 0 {
		t.Fatal("provider contract is empty")
	}

	byID := func(items []VideoProviderCapability) map[string]VideoProviderCapability {
		result := map[string]VideoProviderCapability{}
		for _, item := range items {
			result[item.ID] = item
		}
		return result
	}
	goSide := byID(videoProviderCapabilityContract)
	jsonSide := byID(contract.Providers)
	if len(goSide) != len(jsonSide) {
		t.Fatalf("provider count mismatch: go=%d json=%d", len(goSide), len(jsonSide))
	}
	for id, want := range jsonSide {
		got, ok := goSide[id]
		if !ok {
			t.Errorf("provider %q exists in JSON but not in the Go mirror", id)
			continue
		}
		if got != want {
			t.Errorf("provider %q mismatch: go=%+v json=%+v", id, got, want)
		}
	}
	for _, item := range contract.Providers {
		if normalized := batchfactoryv11.NormalizeVideoProviderForHTTP(item.ID); normalized != item.ID {
			t.Errorf("provider %q is not supported by the backend (normalized to %q)", item.ID, normalized)
		}
		if batchfactoryv11.DefaultVideoModelFor(item.ID) != item.DefaultModel {
			t.Errorf("provider %q default model mismatch with backend", item.ID)
		}
		if item.CatalogID != batchfactoryv11.CatalogIDForProvider(item.ID) {
			t.Errorf("provider %q catalog id mismatch with backend", item.ID)
		}
	}
}

