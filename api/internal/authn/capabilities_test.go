package authn

import "testing"

func TestEffectiveCapabilitiesMapsAdminRolesWithoutHTTPRoleBypass(t *testing.T) {
	for _, role := range []string{"owner", "dev", "admin"} {
		got := EffectiveCapabilities(role, []string{"batch.view"})
		if len(got) != 12 || !HasCapability(got, "admin.member.manage") || !HasCapability(got, "batch.view") {
			t.Fatalf("role %q capabilities=%v", role, got)
		}
	}
	if got := EffectiveCapabilities("manager", []string{"admin.prompt.view"}); len(got) != 1 || !HasCapability(got, "admin.prompt.view") {
		t.Fatalf("manager capabilities=%v", got)
	}
	for _, role := range []string{"member", "unknown"} {
		if got := EffectiveCapabilities(role, nil); len(got) != 0 {
			t.Fatalf("role %q capabilities=%v", role, got)
		}
	}
}
