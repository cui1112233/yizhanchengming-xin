package worker

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os/exec"
	"strings"
	"sync"
)

const (
	EventReady    = "ready"
	EventProgress = "progress"
	EventComplete = "complete"
	EventFailed   = "failed"
	EventIdle     = "idle"
	maxEventSize  = 2 << 20
)

var allowedVideoHosts = map[string]struct{}{
	"material.hnqingyuwen.top":      {},
	"mlzr-material.hnqingyuwen.top": {},
	"ml-material.hnqingyuwen.top":   {},
}

type ExtractRequest struct {
	JobID           string  `json:"jobId"`
	VideoURL        string  `json:"videoUrl"`
	DurationSeconds float64 `json:"durationSeconds"`
}

type Event struct {
	Type         string `json:"type"`
	JobID        string `json:"jobId,omitempty"`
	ModelVersion string `json:"modelVersion,omitempty"`
	Completed    int    `json:"completed,omitempty"`
	Total        int    `json:"total,omitempty"`
	Percent      int    `json:"percent,omitempty"`
	Text         string `json:"text,omitempty"`
	Characters   int    `json:"characters,omitempty"`
	Frames       int    `json:"frames,omitempty"`
	Duplicates   int    `json:"duplicates,omitempty"`
	Code         string `json:"code,omitempty"`
	Message      string `json:"message,omitempty"`
}

func EncodeExtractCommand(request ExtractRequest) ([]byte, error) {
	if strings.TrimSpace(request.JobID) == "" || strings.TrimSpace(request.VideoURL) == "" || request.DurationSeconds <= 0 || request.DurationSeconds > 1800 {
		return nil, errors.New("invalid extract request")
	}
	parsed, err := url.Parse(request.VideoURL)
	if err != nil || parsed.Scheme != "https" || parsed.User != nil || parsed.Port() != "" {
		return nil, errors.New("extract video URL is not allowed")
	}
	if _, allowed := allowedVideoHosts[parsed.Hostname()]; !allowed {
		return nil, errors.New("extract video URL is not allowed")
	}
	return json.Marshal(struct {
		Type            string  `json:"type"`
		JobID           string  `json:"jobId"`
		VideoURL        string  `json:"videoUrl"`
		DurationSeconds float64 `json:"durationSeconds"`
	}{"extract", request.JobID, request.VideoURL, request.DurationSeconds})
}

func DecodeEvent(raw []byte) (Event, error) {
	if len(raw) == 0 || len(raw) > maxEventSize {
		return Event{}, errors.New("worker event exceeds size limit")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var event Event
	if err := decoder.Decode(&event); err != nil {
		return Event{}, fmt.Errorf("decode worker event: %w", err)
	}
	if event.Type != EventReady && event.Type != EventProgress && event.Type != EventComplete && event.Type != EventFailed && event.Type != EventIdle {
		return Event{}, fmt.Errorf("unknown worker event type %q", event.Type)
	}
	if event.Type == EventProgress && event.Text != "" {
		return Event{}, errors.New("progress event must not contain frame text")
	}
	return event, nil
}

type Supervisor struct {
	Command      []string
	Dir          string
	MaxLineBytes int

	mu       sync.Mutex
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	events   chan Event
	done     chan error
	readDone chan struct{}
	started  bool
}

func (s *Supervisor) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return nil
	}
	if len(s.Command) == 0 || strings.TrimSpace(s.Command[0]) == "" {
		return errors.New("worker command is required")
	}
	command := exec.CommandContext(ctx, s.Command[0], s.Command[1:]...)
	command.Dir = s.Dir
	hideConsoleWindow(command)
	stdin, err := command.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		stdin.Close()
		return err
	}
	// 把搬运工的错误输出记进执行器日志，崩溃时才有线索可查。
	if stderr, stderrErr := command.StderrPipe(); stderrErr == nil {
		go func() {
			scanner := bufio.NewScanner(stderr)
			scanner.Buffer(make([]byte, 0, 64*1024), 256*1024)
			for scanner.Scan() {
				log.Printf("[worker] %s", scanner.Text())
			}
		}()
	}
	if err := command.Start(); err != nil {
		stdin.Close()
		return err
	}
	s.cmd = command
	s.stdin = stdin
	s.events = make(chan Event, 32)
	s.done = make(chan error, 1)
	s.readDone = make(chan struct{})
	s.started = true
	readDone := s.readDone
	go s.readEvents(stdout, readDone)
	go func() {
		waitErr := command.Wait()
		<-readDone
		s.mu.Lock()
		if s.events != nil {
			close(s.events)
		}
		s.started = false
		s.stdin = nil
		s.mu.Unlock()
		s.done <- waitErr
	}()
	return nil
}

func (s *Supervisor) Events() <-chan Event {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.events
}

func (s *Supervisor) Submit(ctx context.Context, request ExtractRequest) error {
	command, err := EncodeExtractCommand(request)
	if err != nil {
		return err
	}
	return s.write(ctx, command)
}

func (s *Supervisor) EnsureModel(ctx context.Context, modelDir, modelVersion string) error {
	if strings.TrimSpace(modelDir) == "" {
		return errors.New("model directory is required")
	}
	command, err := json.Marshal(struct {
		Type         string `json:"type"`
		ModelDir     string `json:"modelDir"`
		ModelVersion string `json:"modelVersion,omitempty"`
	}{"ensure_model", modelDir, modelVersion})
	if err != nil {
		return err
	}
	return s.write(ctx, command)
}

func (s *Supervisor) Cancel(ctx context.Context, jobID string) error {
	command, err := json.Marshal(struct {
		Type  string `json:"type"`
		JobID string `json:"jobId"`
	}{"cancel", strings.TrimSpace(jobID)})
	if err != nil {
		return err
	}
	return s.write(ctx, command)
}

func (s *Supervisor) Stop(ctx context.Context) error {
	if err := s.write(ctx, []byte(`{"type":"shutdown"}`)); err != nil {
		return err
	}
	s.mu.Lock()
	done := s.done
	s.mu.Unlock()
	if done == nil {
		return nil
	}
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Supervisor) write(ctx context.Context, command []byte) error {
	s.mu.Lock()
	stdin := s.stdin
	started := s.started
	s.mu.Unlock()
	if !started || stdin == nil {
		return errors.New("worker is not started")
	}
	command = append(command, '\n')
	done := make(chan error, 1)
	go func() {
		_, err := stdin.Write(command)
		done <- err
	}()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Supervisor) readEvents(reader io.Reader, readDone chan<- struct{}) {
	defer close(readDone)
	maxLine := s.MaxLineBytes
	if maxLine <= 0 {
		maxLine = maxEventSize
	}
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4<<10), maxLine)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		event, err := DecodeEvent(line)
		if err == nil {
			s.mu.Lock()
			events := s.events
			s.mu.Unlock()
			if events != nil {
				events <- event
			}
		}
	}
}
