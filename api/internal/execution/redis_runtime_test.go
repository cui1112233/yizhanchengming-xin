package execution

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeRedisServer struct {
	listener net.Listener
	mu sync.Mutex
	commands []string
	responses []string
}

func newFakeRedisServer(t *testing.T, responses ...string) *fakeRedisServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil { t.Fatal(err) }
	s := &fakeRedisServer{listener: ln, responses: responses}
	go s.serve()
	t.Cleanup(func() { _ = ln.Close() })
	return s
}

func (s *fakeRedisServer) serve() {
	for {
		conn, err := s.listener.Accept()
		if err != nil { return }
		go func(c net.Conn) {
			defer c.Close()
			r := bufio.NewReader(c)
			line, err := r.ReadString('\n')
			if err != nil { return }
			if !strings.HasPrefix(line, "*") { return }
			var count int
			_, _ = fmt.Sscanf(strings.TrimSpace(line), "*%d", &count)
			parts := make([]string, 0, count)
			for i := 0; i < count; i++ {
				if _, err := r.ReadString('\n'); err != nil { return }
				value, err := r.ReadString('\n')
				if err != nil { return }
				parts = append(parts, strings.TrimSpace(value))
			}
			s.mu.Lock()
			s.commands = append(s.commands, strings.Join(parts, " "))
			idx := len(s.commands)-1
			response := "+OK\r\n"
			if idx < len(s.responses) { response = s.responses[idx] }
			s.mu.Unlock()
			_, _ = c.Write([]byte(response))
		}(conn)
	}
}

func (s *fakeRedisServer) addr() string { return s.listener.Addr().String() }
func (s *fakeRedisServer) snapshot() []string {
	s.mu.Lock(); defer s.mu.Unlock()
	return append([]string(nil), s.commands...)
}

func TestRedisRuntimeImplementsQueueAndTokenSafeLock(t *testing.T) {
	server := newFakeRedisServer(t,
		":1\r\n",
		"*2\r\n$18\r\nbatch-factory:runs\r\n$2\r\n71\r\n",
		"+OK\r\n",
		":1\r\n",
	)
	runtime := NewRedisRuntime(RedisConfig{Addr: server.addr(), Queue: "batch-factory:runs", DialTimeout: time.Second})
	ctx := context.Background()
	if err := runtime.Enqueue(ctx, 71); err != nil { t.Fatalf("Enqueue: %v", err) }
	runID, err := runtime.Dequeue(ctx)
	if err != nil || runID != 71 { t.Fatalf("Dequeue runID=%d err=%v", runID, err) }
	release, acquired, err := runtime.Acquire(ctx, "batch-factory:run:71", 10*time.Minute)
	if err != nil || !acquired { t.Fatalf("Acquire acquired=%v err=%v", acquired, err) }
	if err := release(ctx); err != nil { t.Fatalf("release: %v", err) }

	deadline := time.Now().Add(time.Second)
	for len(server.snapshot()) < 4 && time.Now().Before(deadline) { time.Sleep(time.Millisecond) }
	commands := server.snapshot()
	if len(commands) != 4 { t.Fatalf("commands=%v", commands) }
	if commands[0] != "RPUSH batch-factory:runs 71" { t.Fatalf("enqueue command=%q", commands[0]) }
	if !strings.HasPrefix(commands[1], "BLPOP batch-factory:runs ") { t.Fatalf("dequeue command=%q", commands[1]) }
	if !strings.Contains(commands[2], "SET batch-factory:run:71") || !strings.Contains(commands[2], "NX PX") { t.Fatalf("lock command=%q", commands[2]) }
	if !strings.HasPrefix(commands[3], "EVAL ") || !strings.Contains(commands[3], "batch-factory:run:71") { t.Fatalf("release command=%q", commands[3]) }
}
