package httpapi

import (
	"context"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelcreation"
	"qiantie/backend/internal/novelfetchworkshop"
)

// novelCoverGenerator 把小说封面请求转接到应用统一的配置化绘图入口。
// 它只复用账号模型目录与实际出图逻辑，不复制渠道配置。
type novelCoverGenerator struct {
	store   batchfactoryv11.Store
	catalog novelfetchworkshop.Store
}

func NewNovelCoverGenerator(store batchfactoryv11.Store, catalog novelfetchworkshop.Store) novelcreation.CoverImageGenerator {
	return novelCoverGenerator{store: store, catalog: catalog}
}

func (g novelCoverGenerator) GenerateCoverImage(ctx context.Context, owner, prompt, imageModelID string) (string, error) {
	result, _, err := configuredBatchImage(ctx, g.store, g.catalog, owner, prompt, strings.TrimSpace(imageModelID), "1024x1536")
	if err != nil {
		return "", err
	}
	imageURL, _ := result["imageUrl"].(string)
	return strings.TrimSpace(imageURL), nil
}
