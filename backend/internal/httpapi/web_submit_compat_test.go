package httpapi

import "testing"

func TestNormalizeOrganizationOptions(t *testing.T) {
	got := normalizeOrganizationOptions(map[string]any{"success": true, "data": []any{
		map[string]any{"id": 1, "name": "博量", "level": 1},
		map[string]any{"organization_id": 3, "organization_name": "璇奕组织", "level_name": "层2"},
		map[string]any{"id": "", "name": "无效"},
		map[string]any{"id": 1, "name": "重复"},
	}})
	if len(got) != 2 || got[0]["id"] != "1" || got[1]["name"] != "璇奕组织" {
		t.Fatalf("unexpected organization options: %#v", got)
	}
}
