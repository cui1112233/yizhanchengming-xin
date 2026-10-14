package httpapi

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestPublishOrganizationUsesNestedBatchPublishSettings(t *testing.T) {
	settings := map[string]json.RawMessage{
		"publishSettings": json.RawMessage(`{"organization":1}`),
	}
	if got := publishOrganization(map[string]any{}, settings); got != "1" {
		t.Fatalf("publishOrganization() = %q, want 1", got)
	}
}

func TestPublisherRequiresGoVideoManagementConnector(t *testing.T) {
	publisher := &BatchFactory121Publisher{}
	_, err := publisher.Submit(context.Background(), "owner", "batch", "book", nil)
	if err == nil || !strings.Contains(err.Error(), "视频管理系统上传服务未配置") {
		t.Fatalf("Submit() error = %v, want explicit Go connector configuration error", err)
	}
}
