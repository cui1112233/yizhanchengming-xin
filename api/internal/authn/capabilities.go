package authn

import (
	"sort"
	"strings"
)

const (
	CapabilityAdminDashboardView = "admin.dashboard.view"
	CapabilityAdminPromptView    = "admin.prompt.view"
	CapabilityAdminPromptEdit    = "admin.prompt.edit"
	CapabilityAdminPromptPublish = "admin.prompt.publish"
	CapabilityAdminModelView     = "admin.model.view"
	CapabilityAdminModelManage   = "admin.model.manage"
	CapabilityAdminSkillView     = "admin.skill.view"
	CapabilityAdminSkillManage   = "admin.skill.manage"
	CapabilityAdminAuditView     = "admin.audit.view"
	CapabilityAdminMemberView    = "admin.member.view"
	CapabilityAdminMemberManage  = "admin.member.manage"
)

var adminCapabilities = []string{CapabilityAdminDashboardView, CapabilityAdminPromptView, CapabilityAdminPromptEdit, CapabilityAdminPromptPublish, CapabilityAdminModelView, CapabilityAdminModelManage, CapabilityAdminSkillView, CapabilityAdminSkillManage, CapabilityAdminAuditView, CapabilityAdminMemberView, CapabilityAdminMemberManage}

func EffectiveCapabilities(role string, granted []string) []string {
	set := map[string]struct{}{}
	for _, item := range granted {
		if item = strings.TrimSpace(item); item != "" {
			set[item] = struct{}{}
		}
	}
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "owner", "dev", "admin":
		for _, item := range adminCapabilities {
			set[item] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for item := range set {
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}

func HasCapability(capabilities []string, required string) bool {
	for _, item := range capabilities {
		if item == required {
			return true
		}
	}
	return false
}
func HasAnyAdminCapability(capabilities []string) bool {
	for _, item := range capabilities {
		if strings.HasPrefix(item, "admin.") {
			return true
		}
	}
	return false
}
