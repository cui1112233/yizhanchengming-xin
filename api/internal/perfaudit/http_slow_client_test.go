package perfaudit

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestPerformanceSlowClientConnectionRetentionProbe(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	defer func() {
		_ = server.Close()
		<-done
	}()

	warm, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	_, _ = fmt.Fprintf(warm, "GET / HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")
	_, _ = bufio.NewReader(warm).ReadString('\n')
	_ = warm.Close()

	before := runtime.NumGoroutine()
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_, err = fmt.Fprintf(conn, "GET / HTTP/1.1\r\nHost: localhost\r\nX-Slow: ")
	if err != nil {
		t.Fatal(err)
	}

	hold := 250 * time.Millisecond
	time.Sleep(hold)
	afterHold := runtime.NumGoroutine()
	_ = conn.SetReadDeadline(time.Now().Add(30 * time.Millisecond))
	buf := make([]byte, 1)
	_, readErr := conn.Read(buf)
	retained := false
	if netErr, ok := readErr.(net.Error); ok && netErr.Timeout() {
		retained = true
	}

	_ = conn.Close()
	time.Sleep(50 * time.Millisecond)
	afterClose := runtime.NumGoroutine()
	t.Logf("PERF_SLOW_CLIENT server_config=zero_timeouts hold=%s connection_retained=%t goroutines_before=%d goroutines_during_hold=%d goroutines_after_close=%d risk=missing_read_header_timeout",
		hold, retained, before, afterHold, afterClose)
	if !retained {
		t.Fatalf("slow incomplete header was not retained as expected; read_err=%v", readErr)
	}
	if afterHold <= before {
		t.Fatalf("expected at least one retained connection goroutine; before=%d during=%d", before, afterHold)
	}
	if strings.Contains(fmt.Sprint(readErr), "EOF") {
		t.Fatalf("connection closed before client deadline: %v", readErr)
	}
}
