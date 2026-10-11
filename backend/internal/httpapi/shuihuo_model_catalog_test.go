package httpapi

import (
	"context"
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

func TestLoadShuihuoOwnerConfigUsesManagedModelCatalog(t *testing.T) {
	shuihuoStore := batchfactoryv11.NewMemoryStore()
	if err := shuihuoStore.SaveShuihuoConfig(context.Background(), "tester", map[string]any{"textModelId": "legacy"}); err != nil {
		t.Fatal(err)
	}
	catalogStore := novelfetchworkshop.NewMemoryStore()
	if err := catalogStore.PutConfig(context.Background(), "tester", map[string]any{
		"modelCatalog":        []any{map[string]any{"id": "managed-text", "kind": "text", "enabled": true}},
		"modelCatalogVersion": 1,
	}); err != nil {
		t.Fatal(err)
	}
	config, err := loadShuihuoOwnerConfig(context.Background(), shuihuoStore, "tester", catalogStore)
	if err != nil {
		t.Fatal(err)
	}
	catalog, ok := config["modelCatalog"].([]any)
	if !ok || len(catalog) != 1 || catalog[0].(map[string]any)["id"] != "managed-text" {
		t.Fatalf("unexpected managed catalog: %#v", config["modelCatalog"])
	}
}
