package imagecapacity

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func testCapacityStateKey(c *Controller, p Policy) string {
	group := p.AccountKey
	if p.GroupKey != "" {
		group = "group:" + p.GroupKey
	}
	return c.prefix + ":" + hash(group)
}

func TestLowConcurrencyRecoversAfterTwoConfirmedSuccesses(t *testing.T) {
	c, _ := testController(t, Policy{AccountLimit: 8, ModelLimit: 8, InitialLimit: 1})
	q := Request{Endpoint: "https://test", Credential: "test-only", Model: "slow-image"}
	c.Snapshot(context.Background(), q)
	p, _ := c.Resolve(context.Background(), q)
	key := testCapacityStateKey(c, p)
	v, err := c.do(context.Background(), "GET", key)
	if err != nil {
		t.Fatal(err)
	}
	var state map[string]any
	if err := json.Unmarshal([]byte(v.(string)), &state); err != nil {
		t.Fatal(err)
	}
	profile := state["profiles"].(map[string]any)[p.ProfileKey].(map[string]any)
	profile["stable"], profile["lastBad"] = time.Now().Add(-time.Minute).UnixMilli(), time.Now().Add(-time.Minute).UnixMilli()
	raw, _ := json.Marshal(state)
	if _, err := c.do(context.Background(), "SET", key, string(raw)); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		lease, err := c.Acquire(context.Background(), q)
		if err != nil {
			t.Fatal(err)
		}
		lease.Finish(Outcome{Success: true})
	}
	status, err := c.Snapshot(context.Background(), q)
	if err != nil || status.EffectiveLimit != 2 {
		t.Fatalf("stable one-lane channel did not recover one step: %+v err=%v", status, err)
	}
}

func TestLowConcurrencyRecoveryStillRequiresQuietWindow(t *testing.T) {
	c, _ := testController(t, Policy{AccountLimit: 8, ModelLimit: 8, InitialLimit: 1})
	q := Request{Endpoint: "https://test", Credential: "test-only", Model: "slow-image"}
	for i := 0; i < 3; i++ {
		lease, err := c.Acquire(context.Background(), q)
		if err != nil {
			t.Fatal(err)
		}
		lease.Finish(Outcome{Success: true})
	}
	status, err := c.Snapshot(context.Background(), q)
	if err != nil || status.EffectiveLimit != 1 {
		t.Fatalf("recovered without quiet window: %+v %v", status, err)
	}
}

func TestAdmissionBarrierLetsActiveRequestsDrain(t *testing.T) {
	c, _ := testController(t, Policy{AccountLimit: 2, ModelLimit: 2, InitialLimit: 2})
	q := Request{Endpoint: "https://test", Credential: "test-only", Model: "image"}
	lease, err := c.Acquire(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	lock := c.prefix + ":config:lock"
	if _, err := c.do(context.Background(), "SET", lock, "maintenance", "PX", "10000", "NX"); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if _, err := c.Acquire(ctx, q); err == nil {
		t.Fatal("new request entered while draining")
	}
	lease.Finish(Outcome{Success: true})
	if _, err := c.do(context.Background(), "DEL", lock); err != nil {
		t.Fatal(err)
	}
	status, err := c.Snapshot(context.Background(), q)
	if err != nil || status.Running != 0 {
		t.Fatalf("active request could not finish through barrier: %+v %v", status, err)
	}
}
