package imagecapacity

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"qiantie/backend/internal/mergeworker"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type testRepo struct{ p Policy }

func (r *testRepo) Resolve(_ context.Context, q Request) (Policy, error) {
	p := r.p
	p.AccountKey = AccountKey(q)
	p.ProfileKey = ProfileKey(q)
	return p, nil
}
func (r *testRepo) Save(_ context.Context, p Policy) error { r.p = p; return nil }
func testController(t *testing.T, p Policy) (*Controller, *testRepo) {
	t.Helper()
	rawURL := os.Getenv("QIANTIE_IMAGE_CAPACITY_REDIS_TEST_URL")
	if rawURL == "" {
		t.Skip("set QIANTIE_IMAGE_CAPACITY_REDIS_TEST_URL to run isolated Redis integration tests")
	}
	s, e := mergeworker.NewRedisStore(rawURL, "")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Do(context.Background(), "PING"); e != nil {
		t.Fatal(e)
	}
	r := &testRepo{p}
	c := NewController(r, s.Do, fmt.Sprintf("test:imagecapacity:%d", time.Now().UnixNano()))
	c.cacheTTL = 0 // Mutable policy fixtures deliberately bypass the production cache.
	t.Cleanup(func() {
		cursor := "0"
		for {
			v, e := s.Do(context.Background(), "SCAN", cursor, "MATCH", c.prefix+"*", "COUNT", "100")
			if e != nil {
				t.Errorf("scoped Redis cleanup: %v", e)
				return
			}
			a, ok := v.([]any)
			if !ok || len(a) != 2 {
				t.Error("invalid SCAN response")
				return
			}
			cursor = fmt.Sprint(a[0])
			if keys, ok := a[1].([]any); ok {
				for _, k := range keys {
					if _, e = s.Do(context.Background(), "DEL", fmt.Sprint(k)); e != nil {
						t.Errorf("scoped Redis cleanup: %v", e)
					}
				}
			}
			if cursor == "0" {
				break
			}
		}
	})
	return c, r
}
func TestIdentity(t *testing.T) {
	a := Request{Endpoint: "https://EXAMPLE.com/v1/images/generations", Credential: "secret", Model: "a"}
	b := a
	b.Endpoint = "https://example.com/alias"
	if AccountKey(a) != AccountKey(b) {
		t.Fatal("aliases split account")
	}
	b.Model = "b"
	if ProfileKey(a) == ProfileKey(b) {
		t.Fatal("models share profile")
	}
	b = a
	b.References = true
	if ProfileKey(a) == ProfileKey(b) {
		t.Fatal("references lost")
	}
}
func TestSharedCapsCancellationAndDoubleRelease(t *testing.T) {
	c, r := testController(t, Policy{AccountLimit: 2, ModelLimit: 1, InitialLimit: 1})
	d := NewController(r, c.do, c.prefix)
	q := Request{Endpoint: "https://test", Credential: "secret", Model: "one"}
	a, e := c.Acquire(context.Background(), q)
	if e != nil {
		t.Fatal(e)
	}
	q.Model = "two"
	b, e := d.Acquire(context.Background(), q)
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Millisecond)
	defer cancel()
	if _, e = d.Acquire(ctx, q); e == nil {
		t.Fatal("exceeded account cap")
	}
	st, e := c.Snapshot(context.Background(), q)
	if e != nil || st.Running != 2 || st.Waiting != 0 {
		t.Fatalf("status %+v %v", st, e)
	}
	a.Finish(Outcome{})
	a.Finish(Outcome{})
	b.Finish(Outcome{})
	st, e = c.Snapshot(context.Background(), q)
	if e != nil || st.Running != 0 {
		t.Fatalf("release %+v %v", st, e)
	}
}
func TestUnknownRetainsAndLimitedAdapts(t *testing.T) {
	c, _ := testController(t, Policy{AccountLimit: 4, ModelLimit: 4, InitialLimit: 4})
	q := Request{Endpoint: "https://test", Credential: "secret", Model: "a"}
	a, _ := c.Acquire(context.Background(), q)
	a.Finish(Outcome{Unknown: true})
	st, _ := c.Snapshot(context.Background(), q)
	if st.Running != 1 {
		t.Fatal("unknown freed")
	}
	b, _ := c.Acquire(context.Background(), q)
	b.Finish(Outcome{Limited: true, RetryAfter: time.Second})
	st, e := c.Snapshot(context.Background(), q)
	if e != nil || st.EffectiveLimit != 2 || st.CoolingUntil.IsZero() {
		t.Fatalf("limit %+v %v", st, e)
	}
}

func TestMultipleInstancesNeverExceedAccountOrModel(t *testing.T) {
	c, r := testController(t, Policy{AccountLimit: 3, ModelLimit: 2, InitialLimit: 2})
	d := NewController(r, c.do, c.prefix)
	var wg sync.WaitGroup
	var mu sync.Mutex
	active := 0
	models := map[string]int{}
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			q := Request{Endpoint: "https://test", Credential: "same", Model: fmt.Sprint(i % 2), Owner: fmt.Sprint(i % 3), Book: fmt.Sprint(i % 4)}
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()
			ctrl := c
			if i%2 == 0 {
				ctrl = d
			}
			l, e := ctrl.Acquire(ctx, q)
			if e != nil {
				t.Error(e)
				return
			}
			mu.Lock()
			active++
			models[q.Model]++
			if active > 3 || models[q.Model] > 2 {
				t.Errorf("exceeded limits: %d %+v", active, models)
			}
			mu.Unlock()
			time.Sleep(12 * time.Millisecond)
			mu.Lock()
			active--
			models[q.Model]--
			mu.Unlock()
			l.Finish(Outcome{})
		}(i)
	}
	wg.Wait()
}

func TestFairOwnerBookRotation(t *testing.T) {
	c, _ := testController(t, Policy{AccountLimit: 1, ModelLimit: 1, InitialLimit: 1})
	q := Request{Endpoint: "https://test", Credential: "same", Model: "m", Owner: "first", Book: "one"}
	a, e := c.Acquire(context.Background(), q)
	if e != nil {
		t.Fatal(e)
	}
	p, _ := c.Resolve(context.Background(), q)
	q1 := q
	q1.Book = "two"
	q2 := q
	q2.Owner = "second"
	q3 := q
	q3.Book = "three"
	for i, r := range []Request{q1, q2, q3} {
		v, e := c.transition(context.Background(), "acquire", p, r, fmt.Sprint(i), Outcome{})
		if e != nil || v.Granted != 0 {
			t.Fatalf("enqueue %v %v", v, e)
		}
	}
	a.Finish(Outcome{})
	v, e := c.transition(context.Background(), "acquire", p, q1, "0", Outcome{})
	if e != nil || v.Granted != 0 {
		t.Fatal("first owner jumped queue", e)
	}
	v, e = c.transition(context.Background(), "acquire", p, q2, "1", Outcome{})
	if e != nil || v.Granted != 1 {
		t.Fatal("second owner starved", e)
	}
	c.transition(context.Background(), "finish", p, q2, "1", Outcome{})
	v, e = c.transition(context.Background(), "acquire", p, q1, "0", Outcome{})
	if e != nil || v.Granted != 1 {
		t.Fatal("book rotation failed", e)
	}
}

func TestLoweringCapDoesNotCancelActive(t *testing.T) {
	c, r := testController(t, Policy{AccountLimit: 2, ModelLimit: 2, InitialLimit: 2})
	q := Request{Endpoint: "https://test", Credential: "same", Model: "m"}
	a, _ := c.Acquire(context.Background(), q)
	b, _ := c.Acquire(context.Background(), q)
	r.p.AccountLimit = 1
	r.p.ModelLimit = 1
	ctx, cancel := context.WithTimeout(context.Background(), 160*time.Millisecond)
	defer cancel()
	if _, e := c.Acquire(ctx, q); e == nil {
		t.Fatal("lowering bypass")
	}
	if a.Context().Err() != nil || b.Context().Err() != nil {
		t.Fatal("lowering canceled active")
	}
	a.Finish(Outcome{})
	st, _ := c.Snapshot(context.Background(), q)
	if st.Running != 1 {
		t.Fatal(st)
	}
	b.Finish(Outcome{})
}

func TestLeaseLossCancelsAndRetainsUntilSafeHorizon(t *testing.T) {
	c, _ := testController(t, Policy{AccountLimit: 1, ModelLimit: 1, InitialLimit: 1})
	c.heartbeat = 15 * time.Millisecond
	c.horizon = 250 * time.Millisecond
	q := Request{Endpoint: "https://test", Credential: "same", Model: "m"}
	orig := c.do
	var failed atomic.Bool
	c.do = func(ctx context.Context, args ...string) (any, error) {
		if failed.Load() {
			return nil, fmt.Errorf("connection lost")
		}
		return orig(ctx, args...)
	}
	l, e := c.Acquire(context.Background(), q)
	if e != nil {
		t.Fatal(e)
	}
	failed.Store(true)
	select {
	case <-l.Context().Done():
	case <-time.After(time.Second):
		t.Fatal("lease continued on renewal loss")
	}
	failed.Store(false)
	st, _ := c.Snapshot(context.Background(), q)
	if st.Running != 1 {
		t.Fatal("lost lease released early")
	}
	time.Sleep(300 * time.Millisecond)
	st, _ = c.Snapshot(context.Background(), q)
	if st.Running != 0 {
		t.Fatal("safe horizon did not expire")
	}
	l.Finish(Outcome{Unknown: true})
}

func TestOrdinaryFailuresDoNotAdaptAndRecoveryIsGradual(t *testing.T) {
	c, _ := testController(t, Policy{AccountLimit: 4, ModelLimit: 4, InitialLimit: 4})
	q := Request{Endpoint: "https://test", Credential: "same", Model: "m"}
	for i := 0; i < 4; i++ {
		l, _ := c.Acquire(context.Background(), q)
		l.Finish(Outcome{})
	}
	st, _ := c.Snapshot(context.Background(), q)
	if st.EffectiveLimit != 4 {
		t.Fatal(st)
	}
	l, _ := c.Acquire(context.Background(), q)
	l.Finish(Outcome{Limited: true, RetryAfter: time.Millisecond})
	p, _ := c.Resolve(context.Background(), q)
	key := c.prefix + ":" + hash(p.AccountKey)
	v, e := c.do(context.Background(), "GET", key)
	if e != nil {
		t.Fatal(e)
	}
	var state map[string]any
	json.Unmarshal([]byte(v.(string)), &state)
	profile := state["profiles"].(map[string]any)[p.ProfileKey].(map[string]any)
	profile["stable"] = time.Now().Add(-time.Minute).UnixMilli()
	profile["lastBad"] = time.Now().Add(-time.Minute).UnixMilli()
	profile["successes"] = 19
	profile["cooling"] = 0
	b, _ := json.Marshal(state)
	c.do(context.Background(), "SET", key, string(b))
	l, _ = c.Acquire(context.Background(), q)
	l.Finish(Outcome{Success: true})
	st, _ = c.Snapshot(context.Background(), q)
	if st.EffectiveLimit != 3 {
		t.Fatalf("recovery skipped step %+v", st)
	}
}

func TestPolicyRevisionStartsAtConfiguredInitialWithoutResettingAdaptation(t *testing.T) {
	c, r := testController(t, Policy{AccountLimit: 4, ModelLimit: 4, InitialLimit: 4})
	q := Request{Endpoint: "https://test", Credential: "same", Model: "m"}
	c.Snapshot(context.Background(), q)
	r.p = Policy{AccountLimit: 20, ModelLimit: 20, InitialLimit: 12, Verified: true}
	st, e := c.Snapshot(context.Background(), q)
	if e != nil || st.EffectiveLimit != 12 {
		t.Fatalf("saved initial ignored %+v %v", st, e)
	}
	l, _ := c.Acquire(context.Background(), q)
	l.Finish(Outcome{Limited: true, RetryAfter: time.Millisecond})
	st, _ = c.Snapshot(context.Background(), q)
	if st.EffectiveLimit != 6 {
		t.Fatalf("adaptation reset on read %+v", st)
	}
}

func TestOldLeaseFinishDoesNotUndoPolicyRevision(t *testing.T) {
	c, r := testController(t, Policy{AccountLimit: 4, ModelLimit: 4, InitialLimit: 4})
	q := Request{Endpoint: "https://test", Credential: "same", Model: "m"}
	l, _ := c.Acquire(context.Background(), q)
	r.p = Policy{AccountLimit: 20, ModelLimit: 20, InitialLimit: 12, Verified: true}
	c.Snapshot(context.Background(), q)
	l.Finish(Outcome{Limited: true, RetryAfter: time.Millisecond})
	st, _ := c.Snapshot(context.Background(), q)
	if st.EffectiveLimit != 6 {
		t.Fatalf("old lease reverted new config %+v", st)
	}
}

type gatedPolicyRepo struct {
	mu      sync.Mutex
	p       Policy
	pause   chan struct{}
	resume  chan struct{}
	armed   bool
	saveErr error
}

func (r *gatedPolicyRepo) Resolve(ctx context.Context, q Request) (Policy, error) {
	r.mu.Lock()
	p := r.p
	armed := r.armed
	r.armed = false
	r.mu.Unlock()
	p.AccountKey = AccountKey(q)
	p.ProfileKey = ProfileKey(q)
	if armed {
		close(r.pause)
		select {
		case <-r.resume:
		case <-ctx.Done():
			return Policy{}, ctx.Err()
		}
	}
	return p, nil
}
func (r *gatedPolicyRepo) Save(_ context.Context, p Policy) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.saveErr != nil {
		return r.saveErr
	}
	r.p = p
	return nil
}

func TestRebindFencesDelayedOldPolicyRead(t *testing.T) {
	c, _ := testController(t, Policy{})
	q := Request{Endpoint: "https://test", Credential: "same", Model: "m"}
	r := &gatedPolicyRepo{p: defaultPolicy(q), pause: make(chan struct{}), resume: make(chan struct{}), armed: true}
	c.repo = r
	result := make(chan Policy, 1)
	go func() { p, _ := c.Resolve(context.Background(), q); result <- p }()
	<-r.pause
	p := defaultPolicy(q)
	p.GroupKey = "new-group"
	if e := c.SavePolicy(context.Background(), q, p); e != nil {
		t.Fatal(e)
	}
	close(r.resume)
	stale := <-result
	v, e := c.transition(context.Background(), "acquire", stale, q, "stale-token", Outcome{})
	if e != nil || v.Granted != -1 {
		t.Fatalf("stale grant after rebind %+v %v", v, e)
	}
	l, e := c.Acquire(context.Background(), q)
	if e != nil {
		t.Fatal(e)
	}
	if l.p.GroupKey != "new-group" {
		t.Fatal("new lease used old group")
	}
	l.Finish(Outcome{})
}

func TestRebindRejectsActiveWaitingAndTargetPool(t *testing.T) {
	c, _ := testController(t, Policy{})
	q := Request{Endpoint: "https://test", Credential: "same", Model: "m"}
	r := &gatedPolicyRepo{p: defaultPolicy(q)}
	c.repo = r
	l, _ := c.Acquire(context.Background(), q)
	p := defaultPolicy(q)
	p.GroupKey = "target"
	if e := c.SavePolicy(context.Background(), q, p); e == nil {
		t.Fatal("moved active account")
	}
	p = defaultPolicy(q)
	p.AccountLimit = 1
	p.ModelLimit = 1
	p.InitialLimit = 1
	if e := c.SavePolicy(context.Background(), q, p); e != nil {
		t.Fatal("live cap edit rejected", e)
	}
	current, _ := c.Resolve(context.Background(), q)
	c.transition(context.Background(), "acquire", current, q, "wait-token", Outcome{})
	l.Finish(Outcome{})
	p.GroupKey = "target"
	if e := c.SavePolicy(context.Background(), q, p); e == nil {
		t.Fatal("moved waiting account")
	}
	c.transition(context.Background(), "cancel", current, q, "wait-token", Outcome{})
	target := p
	target.configEpoch = current.configEpoch
	target.configEpoch = "0"
	targetFresh, _ := c.Resolve(context.Background(), q)
	target.configEpoch = targetFresh.configEpoch
	v, e := c.transition(context.Background(), "acquire", target, q, "target-live", Outcome{})
	if e != nil || v.Granted != 1 {
		t.Fatalf("target seed %v %v", v, e)
	}
	if e = c.SavePolicy(context.Background(), q, p); e == nil {
		t.Fatal("joined busy target")
	}
	c.transition(context.Background(), "finish", target, q, "target-live", Outcome{})
	if e = c.SavePolicy(context.Background(), q, p); e != nil {
		t.Fatal("idle rebind failed", e)
	}
}

func TestCrashedConfigurationWriterRecoversAndFencesLateReads(t *testing.T) {
	c, _ := testController(t, Policy{AccountLimit: 1, ModelLimit: 1, InitialLimit: 1})
	q := Request{Endpoint: "https://test", Credential: "same", Model: "m"}
	stale, _ := c.Resolve(context.Background(), q)
	_, e := c.do(context.Background(), "EVAL", `redis.call('INCR',KEYS[2]);redis.call('SET',KEYS[1],'crashed','PX',100);redis.call('SET',KEYS[3],'crashed');return 1`, "3", c.prefix+":config:lock", c.prefix+":config:epoch", c.prefix+":config:dirty")
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	if _, e = c.Acquire(ctx, q); e == nil {
		t.Fatal("live barrier bypassed")
	}
	time.Sleep(120 * time.Millisecond)
	fresh, e := c.Resolve(context.Background(), q)
	if e != nil || fresh.configEpoch == stale.configEpoch {
		t.Fatal("crash recovery did not advance epoch", e)
	}
	v, e := c.transition(context.Background(), "acquire", stale, q, "stale", Outcome{})
	if e != nil || v.Granted != -1 {
		t.Fatal("crashed writer stale read granted", e)
	}
	l, e := c.Acquire(context.Background(), q)
	if e != nil {
		t.Fatal("crash recovery stuck", e)
	}
	l.Finish(Outcome{})
}

func TestSaveFailureQuarantinesThenRecoversWithNewEpoch(t *testing.T) {
	c, _ := testController(t, Policy{})
	q := Request{Endpoint: "https://test", Credential: "same", Model: "m"}
	r := &gatedPolicyRepo{p: defaultPolicy(q), saveErr: fmt.Errorf("DB failed")}
	c.repo = r
	stale, _ := c.Resolve(context.Background(), q)
	if e := c.SavePolicy(context.Background(), q, defaultPolicy(q)); e == nil {
		t.Fatal("save failure hidden")
	}
	v, e := c.transition(context.Background(), "acquire", stale, q, "stale", Outcome{})
	if e != nil || v.Granted != -1 {
		t.Fatal("save failure did not fence old reads", e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if _, e = c.Acquire(ctx, q); e == nil {
		t.Fatal("uncertain commit quarantine bypassed")
	}
	c.do(context.Background(), "PEXPIRE", c.prefix+":config:lock", "20")
	time.Sleep(30 * time.Millisecond)
	l, e := c.Acquire(context.Background(), q)
	if e != nil {
		t.Fatal("barrier not released", e)
	}
	l.Finish(Outcome{})
}

func TestAccountHostAliasesAndCountProfile(t *testing.T) {
	q := Request{Endpoint: "https://example.com/images", Credential: "same", Model: "m"}
	for _, endpoint := range []string{"https://EXAMPLE.com.:443/alias", "http://example.com:80/alias", "https://example.com./v1"} {
		alias := q
		alias.Endpoint = endpoint
		if AccountKey(alias) != AccountKey(q) {
			t.Errorf("default port/trailing dot alias split %s", endpoint)
		}
	}
	q2 := q
	q2.Endpoint = "https://example.com:8443/v1"
	if AccountKey(q2) == AccountKey(q) {
		t.Fatal("nondefault port collapsed")
	}
	q2 = q
	q2.Count = 1
	if ProfileKey(q2) != ProfileKey(q) {
		t.Fatal("zero count must normalize1")
	}
	q2.Count = 4
	if ProfileKey(q2) == ProfileKey(q) {
		t.Fatal("n4 reused tested n1 profile")
	}
}
func TestImageCountReservesUnitsAndRejectsImpossibleBatch(t *testing.T) {
	c, _ := testController(t, Policy{AccountLimit: 4, ModelLimit: 4, InitialLimit: 4})
	q := Request{Endpoint: "https://test", Credential: "same", Model: "m", Count: 2}
	a, e := c.Acquire(context.Background(), q)
	if e != nil {
		t.Fatal(e)
	}
	b, e := c.Acquire(context.Background(), q)
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 130*time.Millisecond)
	defer cancel()
	if _, e = c.Acquire(ctx, q); e == nil {
		t.Fatal("three n2 requests exceeded4 units")
	}
	st, _ := c.Snapshot(context.Background(), q)
	if st.Running != 4 || st.ModelRunning != 4 {
		t.Fatalf("units not counted %+v", st)
	}
	a.Finish(Outcome{})
	a.Finish(Outcome{})
	st, _ = c.Snapshot(context.Background(), q)
	if st.Running != 2 {
		t.Fatal("duplicate release removed other units", st)
	}
	b.Finish(Outcome{})
	q.Count = 5
	start := time.Now()
	if _, e = c.Acquire(context.Background(), q); e == nil {
		t.Fatal("impossible batch was allowed")
	}
	if time.Since(start) > time.Second {
		t.Fatal("impossible batch waited")
	}
}

func TestBatchAdaptationNeverFallsBelowOneBatchOrExceedsLoweredHardCap(t *testing.T) {
	c, r := testController(t, Policy{AccountLimit: 4, ModelLimit: 4, InitialLimit: 4})
	q := Request{Endpoint: "https://test", Credential: "same", Model: "m", Count: 2}
	a, e := c.Acquire(context.Background(), q)
	if e != nil {
		t.Fatal(e)
	}
	a.Finish(Outcome{Limited: true, RetryAfter: time.Millisecond})
	st, _ := c.Snapshot(context.Background(), q)
	if st.EffectiveLimit != 2 {
		t.Fatal(st)
	}
	time.Sleep(2 * time.Millisecond)
	a, e = c.Acquire(context.Background(), q)
	if e != nil {
		t.Fatal(e)
	}
	a.Finish(Outcome{Limited: true, RetryAfter: time.Millisecond})
	st, _ = c.Snapshot(context.Background(), q)
	if st.EffectiveLimit != 2 {
		t.Fatal("batch deadlocked below its own size", st)
	}
	time.Sleep(2 * time.Millisecond)
	a, e = c.Acquire(context.Background(), q)
	if e != nil {
		t.Fatal(e)
	}
	r.p.ModelLimit = 1
	r.p.InitialLimit = 1
	c.Snapshot(context.Background(), q)
	a.Finish(Outcome{Limited: true, RetryAfter: time.Millisecond})
	st, _ = c.Snapshot(context.Background(), q)
	if st.EffectiveLimit > 1 {
		t.Fatal("old batch exceeded lowered hard cap", st)
	}
}

func TestUnknownOccupancyIsVisibleSeparatelyFromActiveWork(t *testing.T) {
	c, _ := testController(t, Policy{AccountLimit: 4, ModelLimit: 4, InitialLimit: 4})
	q := Request{Endpoint: "https://test", Credential: "same", Model: "m", Count: 2}
	l, e := c.Acquire(context.Background(), q)
	if e != nil {
		t.Fatal(e)
	}
	st, _ := c.Snapshot(context.Background(), q)
	if st.Running != 2 || st.Uncertain != 0 {
		t.Fatalf("active occupancy %+v", st)
	}
	l.Finish(Outcome{Unknown: true})
	st, _ = c.Snapshot(context.Background(), q)
	if st.Running != 2 || st.Uncertain != 2 {
		t.Fatalf("unknown occupancy not distinguished %+v", st)
	}
	v, e := c.transition(context.Background(), "renew", l.p, q, l.token, Outcome{})
	if e != nil || v.Granted != 1 {
		t.Fatal(e)
	}
	st, _ = c.Snapshot(context.Background(), q)
	if st.Uncertain != 2 {
		t.Fatal("renew cleared uncertainty", st)
	}
}

func TestModelTotalSpansProfilesAndCaseAliases(t *testing.T) {
	c, _ := testController(t, Policy{AccountLimit: 20, ModelLimit: 4, InitialLimit: 4, ModelTotalLimit: 4})
	q := Request{Endpoint: "https://test", Credential: "same", Model: "  Untested_Model ", Size: "sizeA", Count: 2}
	a, e := c.Acquire(context.Background(), q)
	if e != nil {
		t.Fatal(e)
	}
	q.Model = "untested_MODEL"
	q.Size = "sizeB"
	b, e := c.Acquire(context.Background(), q)
	if e != nil {
		t.Fatal(e)
	}
	q.Size = "sizeC"
	ctx, cancel := context.WithTimeout(context.Background(), 130*time.Millisecond)
	defer cancel()
	if _, e = c.Acquire(ctx, q); e == nil {
		t.Fatal("model total bypassed through size/case aliases")
	}
	st, _ := c.Snapshot(context.Background(), q)
	if st.ModelTotalRunning != 4 || st.Running != 4 {
		t.Fatalf("model total not visible %+v", st)
	}
	a.Finish(Outcome{})
	b.Finish(Outcome{})
}
func TestKnownModelTotalStillProtectsUnknownProfile(t *testing.T) {
	c, _ := testController(t, Policy{AccountLimit: 20, ModelLimit: 4, InitialLimit: 4, ModelTotalLimit: 20})
	q := Request{Endpoint: "https://test", Credential: "same", Model: "tested", Size: "untestedA", Count: 4}
	a, e := c.Acquire(context.Background(), q)
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 130*time.Millisecond)
	defer cancel()
	if _, e = c.Acquire(ctx, q); e == nil {
		t.Fatal("unknown profile exceeded4 under tested model20")
	}
	q.Size = "untestedB"
	b, e := c.Acquire(context.Background(), q)
	if e != nil {
		t.Fatal(e)
	}
	st, _ := c.Snapshot(context.Background(), q)
	if st.ModelTotalRunning != 8 || st.ModelRunning != 4 {
		t.Fatalf("wrong hierarchy counts %+v", st)
	}
	a.Finish(Outcome{})
	b.Finish(Outcome{})
}

type countingPolicyRepo struct {
	calls atomic.Int64
	mu    sync.Mutex
	p     Policy
}

func (r *countingPolicyRepo) Resolve(_ context.Context, q Request) (Policy, error) {
	r.calls.Add(1)
	r.mu.Lock()
	defer r.mu.Unlock()
	p := r.p
	p.AccountKey = AccountKey(q)
	p.ProfileKey = ProfileKey(q)
	return p, nil
}
func (r *countingPolicyRepo) Save(_ context.Context, p Policy) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.p = p
	return nil
}
func TestPolicyCacheReusesReadsAndEpochSaveInvalidatesImmediately(t *testing.T) {
	c, _ := testController(t, Policy{})
	q := Request{Endpoint: "https://test", Credential: "same", Model: "m"}
	repo := &countingPolicyRepo{p: defaultPolicy(q)}
	c.repo = repo
	c.cacheTTL = time.Second
	c.Resolve(context.Background(), q)
	c.Resolve(context.Background(), q)
	if repo.calls.Load() != 1 {
		t.Fatal("repeat policy reads missed cache", repo.calls.Load())
	}
	p := defaultPolicy(q)
	p.ModelLimit = 8
	p.InitialLimit = 6
	if e := c.SavePolicy(context.Background(), q, p); e != nil {
		t.Fatal(e)
	}
	resolved, e := c.Resolve(context.Background(), q)
	if e != nil || resolved.ModelLimit != 8 || repo.calls.Load() != 3 {
		t.Fatalf("saved policy stayed cached %+v %d %v", resolved, repo.calls.Load(), e)
	}
	c.do = func(context.Context, ...string) (any, error) { return nil, fmt.Errorf("Redis unavailable") }
	if _, e = c.Resolve(context.Background(), q); e == nil {
		t.Fatal("cached policy bypassed Redis outage")
	}
}
func TestPolicyCacheIsBoundedAndExpires(t *testing.T) {
	repo := &countingPolicyRepo{p: Policy{AccountLimit: 4, ModelLimit: 4, InitialLimit: 4}}
	c := NewController(repo, func(context.Context, ...string) (any, error) { return "0", nil }, "cache-unit-test")
	for i := 0; i < 1050; i++ {
		if _, e := c.Resolve(context.Background(), Request{Endpoint: "https://test", Credential: "same", Model: fmt.Sprint(i)}); e != nil {
			t.Fatal(e)
		}
	}
	if len(c.cache) > 1024 {
		t.Fatal("unbounded policy cache", len(c.cache))
	}
	c.cacheTTL = time.Millisecond
	q := Request{Endpoint: "https://test", Credential: "same", Model: "expiring"}
	c.Resolve(context.Background(), q)
	before := repo.calls.Load()
	time.Sleep(3 * time.Millisecond)
	c.Resolve(context.Background(), q)
	if repo.calls.Load() != before+1 {
		t.Fatal("expired policy cache was reused")
	}
}
