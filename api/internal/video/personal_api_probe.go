package video

import (
	"context"
	"strings"
)

func (p *personalAPIProvider) Probe(ctx context.Context) error {
	return probeHTTP(ctx, p.client, strings.TrimRight(p.config.TasksURL, "/"), "Bearer "+p.secret)
}
