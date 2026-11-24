package httpapi

import (
	"context"
	"errors"
	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/yzcmusage"
	"strings"
)

func (p *MemberProfileStore) usage(ctx context.Context, owner string) (map[string]any, error) {
	if p.UsageReader == nil {
		return emptyMemberUsage(), nil
	}
	return p.UsageReader(ctx, owner)
}
func usageMonthTokens(usage map[string]any) int {
	month, _ := usage["month"].(map[string]any)
	switch n := month["totalTokens"].(type) {
	case int64:
		return int(n)
	case int:
		return n
	case float64:
		return int(n)
	}
	return 0
}

// Resolve only credentials from the effective account's existing configuration.
// No raw credential is persisted in the usage ledger or sent to a browser.
func YZCMUsageCredentialResolver(config *UserConfigStore, profiles *MemberProfileStore, accounts auth.Store, primary ...func(context.Context, string) (map[string]any, error)) func(context.Context, string, string) (string, error) {
	return func(ctx context.Context, owner, hash string) (string, error) {
		effective := effectiveModelCatalogOwner(BridgeIdentity{Username: owner}, profiles, accounts)
		if effective == "" {
			return "", errors.New("usage account unavailable")
		}
		config.mu.Lock()
		settings := config.read(effective)
		config.mu.Unlock()
		var find func(any) string
		find = func(value any) string {
			switch v := value.(type) {
			case string:
				if yzcmusage.Fingerprint(v) == hash {
					return v
				}
			case map[string]any:
				for _, child := range v {
					if key := find(child); key != "" {
						return key
					}
				}
			case []any:
				for _, child := range v {
					if key := find(child); key != "" {
						return key
					}
				}
			}
			return ""
		}
		if key := find(settings); key != "" {
			return key, nil
		}
		if len(primary) > 0 && primary[0] != nil {
			if settings, err := primary[0](ctx, owner); err == nil {
				if key := find(settings); key != "" {
					return key, nil
				}
			}
		}
		return "", errors.New("usage credential no longer configured")
	}
}

func usageFeature(path string) string {
	// The module is attached to the authenticated incoming request, never a
	// client-supplied owner header or a model-name heuristic.
	for _, p := range []struct{ match, name string }{{"novel-fetch", "novel-fetch"}, {"novel-creation", "novel-creation"}, {"shuihuo", "shuihuo-production"}, {"batch-factory", "shuihuo-production"}, {"screenplay", "comic"}, {"agent", "canvas"}, {"aicut", "aicut"}, {"script", "script"}} {
		if strings.Contains(path, p.match) {
			return p.name
		}
	}
	return "model"
}

func (p *MemberProfileStore) usageGovernance(ctx context.Context, manager map[string]any) (map[string]any, error) {
	result := memberTeamGovernance(manager)
	if p.TeamTokenReader == nil {
		return result, nil
	}
	owner := toString(manager["username"])
	owners := []string{owner}
	seen := map[string]bool{owner: true}
	for _, member := range p.readArray("member-profiles.json") {
		name := toString(member["username"])
		if name != "" && !seen[name] && toString(member["role"]) == "member" && toString(member["boundTo"]) == owner {
			owners = append(owners, name)
			seen[name] = true
		}
	}
	used, err := p.TeamTokenReader(ctx, owners)
	if err != nil {
		return nil, err
	}
	result["quota"] = memberQuota(manager["monthlyTokenLimit"], used)
	result["usage"] = map[string]any{"totalTokens": used}
	return result, nil
}
