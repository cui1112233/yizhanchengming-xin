package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	txt "qiantie/backend/internal/batchfactorytext"
	"qiantie/backend/internal/batchfactoryv11"
)

type textPublishOverride struct {
	CollectionID string `json:"collectionId"`
	Force        bool   `json:"force"`
}

// Resolve every override before writing a job. The database lookup pins the
// collection to the authenticated owner, batch and explicitly selected book.
func (c *BatchAutomationCoordinator) textPublishOverrides(ctx context.Context, owner, batchID string, rawIDs, raw any) (map[string]textPublishOverride, error) {
	if raw == nil {
		return nil, nil
	}
	var overrides map[string]textPublishOverride
	b, err := json.Marshal(raw)
	if err != nil || json.Unmarshal(b, &overrides) != nil {
		return nil, fmt.Errorf("%w：指定的文本合集格式无效", batchfactoryv11.ErrInvalid)
	}
	if len(overrides) == 0 {
		return nil, nil
	}
	var ids []string
	b, err = json.Marshal(rawIDs)
	if err != nil || json.Unmarshal(b, &ids) != nil {
		return nil, batchfactoryv11.ErrInvalid
	}
	selected := make(map[string]bool, len(ids))
	for _, bookID := range ids {
		selected[bookID] = true
	}
	for bookID, o := range overrides {
		if !selected[bookID] || strings.TrimSpace(bookID) == "" || strings.TrimSpace(o.CollectionID) == "" || len(o.CollectionID) > 64 {
			return nil, fmt.Errorf("%w：请为所选书籍填写准确的文本合集", batchfactoryv11.ErrInvalid)
		}
	}
	if c.textPublishCollection == nil {
		return nil, fmt.Errorf("%w：文本合集查询未配置", batchfactoryv11.ErrUnavailable)
	}
	for bookID, o := range overrides {
		collection, err := c.textPublishCollection(ctx, txt.Scope{Owner: owner, BatchID: batchID, BookID: bookID}, o.CollectionID)
		if errors.Is(err, txt.ErrNotFound) {
			return nil, batchfactoryv11.ErrNotFound
		}
		if err != nil {
			return nil, err
		}
		if collection.ID != o.CollectionID || collection.Status != "ready" {
			return nil, fmt.Errorf("%w：所选文本合集尚未就绪", batchfactoryv11.ErrConflict)
		}
	}
	return overrides, nil
}

func (o textPublishOverride) apply(entry map[string]any) error {
	if o.CollectionID == "" {
		return nil
	}
	if _, ok := entry["frozenTextSettings"]; !ok {
		return fmt.Errorf("%w：只有文本书可以指定上传合集", batchfactoryv11.ErrInvalid)
	}
	entry["textPendingPublishCollectionId"] = o.CollectionID
	entry["textCollectionId"] = o.CollectionID
	delete(entry, "textPublishIntentId")
	if o.Force {
		entry["textPublishIntentId"] = rand.Text()
	}
	return nil
}
