package httpapi

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/imagecapacity"
)

func TestComicChannelQueueDoesNotBlockIndependentUpstream(t *testing.T) {
	q := newAutomationComicSlotQueue(1)
	releaseA, err := q.acquireGroup(context.Background(), "upstream-a", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer releaseA()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	releaseB, err := q.acquireGroup(ctx, "upstream-b", nil)
	if err != nil {
		t.Fatalf("idle upstream B blocked behind upstream A: %v", err)
	}
	releaseB()
}

func TestComicChannelQueueCancellationReclaimsGroup(t *testing.T) {
	q := newAutomationComicSlotQueue(1)
	release, err := q.acquireGroup(context.Background(), "shared", nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	waiting, done := make(chan struct{}, 1), make(chan error, 1)
	go func() {
		_, err := q.acquireGroup(ctx, "shared", func(int) { waiting <- struct{}{} })
		done <- err
	}()
	<-waiting
	cancel()
	if err := <-done; err != context.Canceled {
		t.Fatalf("cancel = %v", err)
	}
	release()
	release()
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.groups) != 0 {
		t.Fatal("completed or cancelled channel queues leaked")
	}
}

func TestComicChannelQueueUsesSharedCapacityIdentityAcrossAccounts(t *testing.T) {
	config := NewUserConfigStore(t.TempDir())
	for _, tc := range []struct{ owner, endpoint, key string }{
		{"alice", "https://same.test/v1", "alice-secret"},
		{"bob", "https://same.test/v1", "bob-secret"},
		{"carol", "https://other.test/v1", "carol-secret"},
	} {
		if err := config.write(tc.owner, map[string]any{apiIntegrationsSettingKey: map[string]any{"customChannels": []any{map[string]any{"id": "image", "group": "image", "model": "test-image", "baseUrl": tc.endpoint, "apiKey": tc.key}}}}); err != nil {
			t.Fatal(err)
		}
	}
	controller := imagecapacity.NewController(&imageRouteRepo{}, func(context.Context, ...string) (any, error) { return "0", nil }, "test-book-groups")
	options := RouterOptions{ImageCapacity: controller, UserConfigStore: config, Store: batchfactoryv11.NewMemoryStore()}
	settings := batchfactoryv11.SettingsPatch{"imageModelId": json.RawMessage(`"custom:image"`)}
	groups := make(map[string]string)
	for _, owner := range []string{"alice", "bob", "carol"} {
		group, err := resolveComicBookQueueGroup(context.Background(), options, owner, settings)
		if err != nil {
			t.Fatal(err)
		}
		if group == "" || strings.Contains(group, "secret") || strings.Contains(group, owner) {
			t.Fatalf("invalid private group identity: %s", group)
		}
		groups[owner] = group
	}
	if groups["alice"] != groups["bob"] {
		t.Fatal("same upstream was split by member or credential")
	}
	if groups["alice"] == groups["carol"] {
		t.Fatal("independent upstreams share a whole-book queue")
	}
}

func TestComicChannelQueueKeepsSharedChannelFIFOAcrossMembers(t *testing.T) {
	q := newAutomationComicSlotQueue(1)
	release, err := q.acquireGroup(context.Background(), "shared", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if release != nil {
			release()
		}
	}()
	waiting := make(chan struct{}, 1)
	admitted := make(chan func(), 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		r, err := q.acquireGroup(ctx, "shared", func(int) { waiting <- struct{}{} })
		if err == nil {
			admitted <- r
		}
	}()
	select {
	case <-waiting:
	case <-time.After(time.Second):
		t.Fatal("same upstream must wait")
	}
	select {
	case r := <-admitted:
		r()
		t.Fatal("same upstream exceeded book limit")
	default:
	}
	release()
	release = nil
	select {
	case r := <-admitted:
		r()
	case <-time.After(time.Second):
		t.Fatal("next member did not receive released slot")
	}
}
