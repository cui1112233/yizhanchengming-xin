package httpapi

import (
	"context"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

func (f *fakePublishingService) ClaimBatchProject(context.Context, authn.User, int64) error { return nil }
