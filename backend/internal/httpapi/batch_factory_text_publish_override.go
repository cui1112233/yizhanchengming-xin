package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	txt "qiantie/backend/internal/batchfactorytext"
	"qiantie/backend/internal/batchfactoryv11"
)

type textPublishOverride struct {
	CollectionID string `json:"collectionId"`
	Force        bool   `json:"force"`
	RequestKey   string `json:"requestKey"`
}

var textConfirmationKey = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,127}$`)

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
	bindings := map[string]string{}
	for bookID, o := range overrides {
		if o.Force || o.RequestKey != "" {
			if !textConfirmationKey.MatchString(o.RequestKey) {
				return nil, fmt.Errorf("%w：上传确认编号无效", batchfactoryv11.ErrInvalid)
			}
			scope, _ := json.Marshal([]string{owner, batchID, bookID, o.CollectionID})
			if bound, ok := bindings[o.RequestKey]; ok && bound != string(scope) {
				return nil, batchfactoryv11.ErrConflict
			}
			bindings[o.RequestKey] = string(scope)
		}
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
	if len(bindings) > 0 {
		binder, ok := c.store.(interface {
			BindTextPublishKeys(context.Context, map[string]string) error
		})
		if !ok {
			return nil, batchfactoryv11.ErrUnavailable
		}
		if err := binder.BindTextPublishKeys(ctx, bindings); err != nil {
			return nil, err
		}
	}
	return overrides, nil
}

// An HTTP replay may arrive while the original job is still scheduled or
// running. Only acknowledge a selection carrying the same explicit confirmation.
func confirmedTextReplay(job map[string]any, count int, overrides map[string]textPublishOverride) bool {
	if len(overrides) == 0 || len(overrides) != count {
		return false
	}
	entries := mapOrEmpty(job["books"])
	hasConfirmation := false
	for bookID, o := range overrides {
		e := mapOrEmpty(entries[bookID])
		if e["textCollectionId"] != o.CollectionID {
			return false
		}
		if o.Force {
			if e["textPublishConfirmationKey"] != o.RequestKey {
				return false
			}
			hasConfirmation = true
		} else if e["textPublishConfirmationKey"] != nil {
			return false
		}
	}
	return hasConfirmation
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
	delete(entry, "textPublishConfirmationKey")
	if o.Force {
		entry["textPublishConfirmationKey"] = o.RequestKey
		entry["textPublishIntentId"] = o.RequestKey
	}
	return nil
}
