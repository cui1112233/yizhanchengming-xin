package httpapi

import (
	"context"
	"fmt"
	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/yzcmusage"
)

// NovelCreationTextResolver rechecks active membership and the current manager
// binding before every actual external call. The supplied resolver honours an
// explicit model ID and never substitutes a different configured model.
func NovelCreationTextResolver(accounts auth.Store, profiles *MemberProfileStore, resolve AgentTextModelResolver) AgentTextModelResolver {
	return func(ctx context.Context, owner, modelID string) (batchfactoryv11.DirectorProvider, error) {
		if accounts != nil {
			account, err := accounts.FindAccount(ctx, owner)
			if err != nil || !account.Active {
				return nil, fmt.Errorf("所选文本模型当前不可用：账号无权调用")
			}
		}
		effective := effectiveModelCatalogOwner(BridgeIdentity{Username: owner}, profiles, accounts)
		if accounts != nil && effective != owner {
			account, err := accounts.FindAccount(ctx, effective)
			if err != nil || !account.Active {
				return nil, fmt.Errorf("所选文本模型当前不可用：管理账号不可用")
			}
		}
		p, err := resolve(ctx, effective, modelID)
		if err != nil || p == nil {
			return p, err
		}
		return novelScopedTextProvider{DirectorProvider: p, owner: effective, actor: owner}, nil
	}
}

type novelScopedTextProvider struct {
	batchfactoryv11.DirectorProvider
	owner string
	actor string
}

func (p novelScopedTextProvider) TextModelMetadata() batchfactoryv11.TextModelMetadata {
	m := batchfactoryv11.TextModelMetadata{}
	if actual, ok := p.DirectorProvider.(batchfactoryv11.TextModelMetadataProvider); ok {
		m = actual.TextModelMetadata()
	}
	m.CatalogOwner = p.owner
	return m
}

// The worker reconstructs its context from a durable job. Keep the caller
// separately from the manager whose model settings supplied the credentials.
func (p novelScopedTextProvider) Complete(ctx context.Context, input batchfactoryv11.TextCompletionRequest) (string, error) {
	ctx = yzcmusage.WithIdentity(ctx, p.actor, "novel-creation")
	return p.DirectorProvider.Complete(ctx, input)
}
