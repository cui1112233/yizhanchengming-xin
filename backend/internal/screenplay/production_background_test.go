package screenplay

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestProductionStartRedisUnavailableDoesNotMutate(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	_, _ = s.CreateProject(ctx, Project{ID: "p", OwnerUsername: "a"})
	repo := NewMemoryProductionRepository()
	r := ProductionRuntime{Store: s, Repository: repo, RedisDo: func(context.Context, ...string) (any, error) { return nil, fmt.Errorf("offline") }, Executor: func(string) BatchExecutor { return BatchExecutor{Store: s} }}
	if _, e := r.Start(ctx, "a", "p"); e == nil {
		t.Fatal("offline Redis accepted")
	}
	if _, e := repo.GetProductionJob(ctx, "a", "p"); !errors.Is(e, ErrNotFound) {
		t.Fatalf("job created: %v", e)
	}
	if _, e := s.GetBatch(ctx, "a", "p"); !errors.Is(e, ErrNotFound) {
		t.Fatalf("batch created: %v", e)
	}
}
func TestProductionWorkerContinuesAfterStartRequestCloses(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	p, _ := s.CreateProject(ctx, Project{ID: "p", OwnerUsername: "a", SourceText: "真实内容"})
	_, _ = s.SaveProjectSource(ctx, "a", p.ID, "真实内容")
	b := NewBatch("b", "p", "a", "s", PermissionBypass)
	for _, st := range b.Steps {
		st.Status = StepDone
	}
	b.Step("0_ingest").Status = StepPending
	b.Status = BatchPaused
	_ = s.CreateBatch(ctx, b)
	repo := NewMemoryProductionRepository()
	r := ProductionRuntime{Store: s, Repository: repo, Interval: time.Millisecond, RedisDo: newProductionTestRedis(), Executor: func(string) BatchExecutor { return BatchExecutor{Store: s} }}
	request, cancelRequest := context.WithCancel(ctx)
	if _, e := r.Start(request, "a", "p"); e != nil {
		t.Fatal(e)
	}
	cancelRequest()
	worker, stop := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { done <- r.Run(worker) }()
	defer func() { stop(); <-done }()
	deadline := time.After(2 * time.Second)
	for {
		j, _ := repo.GetProductionJob(ctx, "a", "p")
		if j.Status == "completed" {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("worker did not complete: %+v", j)
		case <-time.After(time.Millisecond):
		}
	}
}

func newProductionTestRedis() func(context.Context, ...string) (any, error) {
	var mu sync.Mutex
	leases := map[string]string{}
	due := map[string]float64{}
	return func(_ context.Context, args ...string) (any, error) {
		mu.Lock()
		defer mu.Unlock()
		switch args[0] {
		case "PING":
			return "PONG", nil
		case "SET":
			if leases[args[1]] != "" {
				return nil, nil
			}
			leases[args[1]] = args[2]
			return "OK", nil
		case "ZADD":
			member := args[len(args)-1]
			if _, exists := due[member]; len(args) > 4 && args[2] == "NX" && exists {
				return int64(0), nil
			}
			score, _ := strconv.ParseFloat(args[len(args)-2], 64)
			due[member] = score
			return int64(1), nil
		case "ZREM":
			delete(due, args[2])
			return int64(1), nil
		case "ZRANGEBYSCORE":
			out := []any{}
			max, _ := strconv.ParseFloat(args[3], 64)
			for key, score := range due {
				if score <= max {
					out = append(out, key)
				}
			}
			return out, nil
		case "EVAL":
			if strings.Contains(args[1], "'ZSCORE'") {
				score, _ := strconv.ParseFloat(args[4], 64)
				old, ok := due[args[5]]
				if !ok || old > score {
					due[args[5]] = score
					return int64(1), nil
				}
				return int64(0), nil
			}
			if leases[args[3]] != args[4] {
				return int64(0), nil
			}
			if strings.Contains(args[1], "'DEL'") {
				delete(leases, args[3])
			}
			return int64(1), nil
		}
		return nil, fmt.Errorf("unexpected redis command %s", args[0])
	}
}
