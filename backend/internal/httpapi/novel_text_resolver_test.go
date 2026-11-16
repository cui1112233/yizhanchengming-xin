package httpapi

import (
	"context"
	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
	"testing"
)

func TestNovelResolverChecksActiveBindingAndCurrentCatalog(t *testing.T) {
	ctx := context.Background()
	accounts := &applicationTestStore{accounts: map[string]auth.Account{"member": {Username: "member", Active: true}, "manager": {Username: "manager", Active: true}, "other": {Username: "other", Active: true}}}
	profiles := NewMemberProfileStore(t.TempDir())
	bind := func(manager string) {
		t.Helper()
		if err := profiles.writeArray("member-profiles.json", []map[string]any{{"username": "member", "role": "member", "boundTo": manager}}); err != nil {
			t.Fatal(err)
		}
	}
	bind("manager")
	catalog := novelfetchworkshop.NewMemoryStore()
	model := map[string]any{"id": "chosen", "modelId": "actual", "kind": "text", "enabled": true, "credential": "secret", "baseUrl": "https://example.invalid/v1"}
	save := func() {
		t.Helper()
		if err := catalog.PutConfig(ctx, "manager", map[string]any{"modelCatalog": []any{model}}); err != nil {
			t.Fatal(err)
		}
	}
	save()
	resolver := NovelCreationTextResolver(accounts, profiles, func(c context.Context, owner, id string) (batchfactoryv11.DirectorProvider, error) {
		return resolveAgentTextModel(c, catalog, nil, owner, id)
	})
	p, err := resolver(ctx, "member", "chosen")
	if err != nil {
		t.Fatal(err)
	}
	if p.(batchfactoryv11.TextModelMetadataProvider).TextModelMetadata().CatalogOwner != "manager" {
		t.Fatal("actual catalog owner was not recorded")
	}
	// All following checks only resolve providers; no external HTTP can be sent.
	member := accounts.accounts["member"]
	member.Active = false
	accounts.accounts["member"] = member
	if _, err = resolver(ctx, "member", "chosen"); err == nil {
		t.Fatal("inactive account retained model permission")
	}
	member.Active = true
	accounts.accounts["member"] = member
	bind("other")
	if _, err = resolver(ctx, "member", "chosen"); err == nil {
		t.Fatal("binding change silently reused previous manager model")
	}
	bind("manager")
	model["enabled"] = false
	save()
	if _, err = resolver(ctx, "member", "chosen"); err == nil {
		t.Fatal("disabled selected model silently fell back")
	}
	model["enabled"] = true
	save()
	if err = catalog.PutConfig(ctx, "manager", map[string]any{"modelCatalog": []any{}}); err != nil {
		t.Fatal(err)
	}
	if _, err = resolver(ctx, "member", "chosen"); err == nil {
		t.Fatal("deleted selected model silently fell back")
	}
}
