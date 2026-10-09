package main

import (
	"context"
	"net"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/app"
)

func TestBuildInfoUsesLinkerInjectedGitSHA(t *testing.T) {
	previous := BuildGitSHA
	BuildGitSHA = "release-sha-123"
	t.Cleanup(func() { BuildGitSHA = previous })

	if got := buildInfo().GitSHA; got != "release-sha-123" {
		t.Fatalf("gitSha=%q want release-sha-123", got)
	}
}

type serverRuntime struct {
	started atomic.Bool
	closed  atomic.Bool
}

func (r *serverRuntime) Start(context.Context) error { r.started.Store(true); return nil }
func (r *serverRuntime) Ready() bool                 { return r.started.Load() && !r.closed.Load() }
func (r *serverRuntime) Status() app.RuntimeStatus {
	return app.RuntimeStatus{State: "available", ReasonCode: "ready"}
}
func (r *serverRuntime) Wait() error  { return nil }
func (r *serverRuntime) Close() error { r.closed.Store(true); return nil }

func TestServeStartsRuntimeBeforeHTTPAndClosesItOnCancellation(t *testing.T) {
	runtime := &serverRuntime{}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if !runtime.started.Load() {
			http.Error(w, "runtime not started", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- serve(ctx, server, listener, runtime) }()

	deadline := time.Now().Add(time.Second)
	for !runtime.started.Load() && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	response, err := http.Get("http://" + listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("status=%d", response.StatusCode)
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("serve did not stop")
	}
	if !runtime.closed.Load() {
		t.Fatal("runtime was not closed")
	}
}
