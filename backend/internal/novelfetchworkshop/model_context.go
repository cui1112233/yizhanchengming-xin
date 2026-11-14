package novelfetchworkshop

import (
	"context"
	"strings"
)

type processModelKey struct{}

func WithProcessTextModel(ctx context.Context, model string) context.Context {
	model = strings.TrimSpace(model)
	if model == "" || model == "未指定" {
		return ctx
	}
	return context.WithValue(ctx, processModelKey{}, model)
}

func ProcessTextModel(ctx context.Context) string {
	model, _ := ctx.Value(processModelKey{}).(string)
	return model
}
