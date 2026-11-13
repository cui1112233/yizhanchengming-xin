package imagecapacity

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var ErrOutcomeUnknown = errors.New("图片请求已发出但结果未确认，请先核对上游任务，系统暂时保留占用名额")
var admissionEnabled atomic.Bool

func Enabled() bool         { return admissionEnabled.Load() }
func SetEnabled(value bool) { admissionEnabled.Store(value) }

type identityKey struct{}
type localExecutorKey struct{}

// Local executors have their own account/device scheduler and never consume a
// remote API account quota. Only trusted backend adapters attach this marker.
func WithoutSharedAdmission(ctx context.Context) context.Context {
	return context.WithValue(ctx, localExecutorKey{}, true)
}

type imageIdentity struct{ Owner, Book string }

func WithIdentity(ctx context.Context, owner, book string) context.Context {
	if previous, ok := ctx.Value(identityKey{}).(imageIdentity); ok {
		if owner == "" {
			owner = previous.Owner
		}
		if book == "" {
			book = previous.Book
		}
	}
	return context.WithValue(ctx, identityKey{}, imageIdentity{owner, book})
}

type waitHeartbeatKey struct{}

func WithWaitHeartbeat(ctx context.Context, beat func()) context.Context {
	return context.WithValue(ctx, waitHeartbeatKey{}, beat)
}

type operationKey struct{}
type operation struct {
	mu        sync.Mutex
	once      sync.Once
	finishers []func(error)
	finished  bool
	err       error
	confirmed bool
}

func ConfirmResult(ctx context.Context) {
	if op, ok := ctx.Value(operationKey{}).(*operation); ok {
		op.mu.Lock()
		op.confirmed = true
		op.mu.Unlock()
	}
}

// BeginOperation keeps the provider lease through decoding and image download.
// Each actual provider request still has its own lease; known rejections return
// theirs immediately, so an internal format fallback cannot deadlock the pool.
func BeginOperation(ctx context.Context) (context.Context, func(error)) {
	if _, ok := ctx.Value(operationKey{}).(*operation); ok {
		return ctx, func(error) {}
	}
	op := &operation{}
	return context.WithValue(ctx, operationKey{}, op), func(err error) {
		op.once.Do(func() {
			op.mu.Lock()
			op.finished = true
			op.err = err
			list := append([]func(error){}, op.finishers...)
			op.mu.Unlock()
			for _, f := range list {
				f(err)
			}
		})
	}
}
func (op *operation) add(f func(error)) {
	op.mu.Lock()
	if op.finished {
		err := op.err
		op.mu.Unlock()
		f(err)
		return
	}
	op.finishers = append(op.finishers, f)
	op.mu.Unlock()
}

type HTTPPermit interface {
	Context() context.Context
	Finish(Outcome)
}
type HTTPAdmission func(context.Context, Request) (HTTPPermit, error)
type admittedTransport struct {
	base              http.RoundTripper
	admit             HTTPAdmission
	timeout           time.Duration
	redirectDeadlines sync.Map
}

func WrapHTTP(base http.RoundTripper, admit HTTPAdmission) http.RoundTripper {
	return &admittedTransport{base: base, admit: admit}
}
func WrapHTTPWithTimeout(base http.RoundTripper, admit HTTPAdmission, timeout time.Duration) http.RoundTripper {
	return &admittedTransport{base: base, admit: admit, timeout: timeout}
}
func (c *Controller) WrapTransportWithTimeout(base http.RoundTripper, timeout time.Duration) http.RoundTripper {
	return WrapHTTPWithTimeout(base, func(ctx context.Context, r Request) (HTTPPermit, error) { return c.Acquire(ctx, r) }, timeout)
}
func (c *Controller) WrapTransport(base http.RoundTripper) http.RoundTripper {
	return WrapHTTP(base, func(ctx context.Context, r Request) (HTTPPermit, error) { return c.Acquire(ctx, r) })
}

func imageRequest(r *http.Request) (Request, bool, error) {
	if skip, _ := r.Context().Value(localExecutorKey{}).(bool); skip {
		return Request{}, false, nil
	}
	path := strings.ToLower(r.URL.Path)
	_, operationImage := r.Context().Value(operationKey{}).(*operation)
	if r.Method != "POST" || (!operationImage && !strings.HasSuffix(path, "/images/generations") && !strings.HasSuffix(path, "/images/edits")) {
		return Request{}, false, nil
	}
	target := Request{Endpoint: r.URL.String(), Credential: strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")), Count: 1}
	if identity, ok := r.Context().Value(identityKey{}).(imageIdentity); ok {
		target.Owner = identity.Owner
		target.Book = identity.Book
	}
	if r.GetBody == nil {
		return target, true, fmt.Errorf("无法核对图片请求正文，未向上游提交")
	}
	body, err := r.GetBody()
	if err != nil {
		return target, true, fmt.Errorf("无法读取图片请求正文，未向上游提交：%w", err)
	}
	defer body.Close()
	media, params, mediaErr := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if mediaErr != nil && r.Header.Get("Content-Type") != "" {
		return target, true, fmt.Errorf("图片请求正文格式无效，未向上游提交")
	}
	if strings.HasPrefix(media, "multipart/") {
		reader := multipart.NewReader(body, params["boundary"])
		countSeen := false
		for {
			part, e := reader.NextPart()
			if e == io.EOF {
				break
			}
			if e != nil {
				return target, true, fmt.Errorf("图片请求表单无效，未向上游提交：%w", e)
			}
			name := part.FormName()
			if part.FileName() != "" || strings.Contains(name, "image") {
				target.References = true
				part.Close()
				continue
			}
			// Only admission metadata needs a small field limit. Prompts and
			// unrelated provider options can legitimately be much longer.
			switch name {
			case "model", "size", "quality", "resolution", "n":
			default:
				if closeErr := part.Close(); closeErr != nil {
					return target, true, fmt.Errorf("图片请求表单无效，未向上游提交：%w", closeErr)
				}
				continue
			}
			data, readErr := io.ReadAll(io.LimitReader(part, 1025))
			part.Close()
			if readErr != nil || len(data) > 1024 {
				return target, true, fmt.Errorf("图片请求表单字段无效，未向上游提交")
			}
			switch name {
			case "model":
				target.Model = string(data)
			case "size":
				target.Size = string(data)
			case "quality":
				target.Quality = string(data)
			case "resolution":
				target.Resolution = string(data)
			case "n":
				if countSeen {
					return target, true, fmt.Errorf("图片数量字段不能重复，未向上游提交")
				}
				countSeen = true
				n, countErr := strconv.Atoi(string(data))
				if countErr != nil || n < 1 || n > 20 {
					return target, true, fmt.Errorf("图片数量必须为 1–20 的整数，未向上游提交")
				}
				target.Count = n
			}
		}
	} else {
		data, readErr := io.ReadAll(io.LimitReader(body, (64<<20)+1))
		var fields map[string]json.RawMessage
		if readErr != nil || len(data) > 64<<20 || json.Unmarshal(data, &fields) != nil {
			return target, true, fmt.Errorf("图片请求正文无效，未向上游提交")
		}
		for key, destination := range map[string]*string{"model": &target.Model, "size": &target.Size, "quality": &target.Quality, "resolution": &target.Resolution} {
			_ = json.Unmarshal(fields[key], destination)
		}
		if raw, present := fields["n"]; present {
			var count int
			if json.Unmarshal(raw, &count) != nil || count < 1 || count > 20 {
				return target, true, fmt.Errorf("图片数量必须为 1–20 的整数，未向上游提交")
			}
			target.Count = count
		}
		for _, key := range []string{"image", "images", "image_url", "image_urls", "reference_images", "referenceImages", "mask"} {
			if data := bytes.TrimSpace(fields[key]); len(data) > 0 && !bytes.Equal(data, []byte("null")) && !bytes.Equal(data, []byte("[]")) && !bytes.Equal(data, []byte(`""`)) {
				target.References = true
			}
		}
	}
	return target, true, nil
}

func (t *admittedTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	target, image, parseErr := imageRequest(request)
	if parseErr != nil {
		return nil, parseErr
	}
	if image && target.Model == "" {
		return nil, fmt.Errorf("无法核对图片请求的模型和数量，未向上游提交")
	}
	if !image {
		if t.timeout <= 0 {
			return t.base.RoundTrip(request)
		}
		deadline := time.Now().Add(t.timeout)
		if request.Response != nil {
			if previous, ok := t.redirectDeadlines.LoadAndDelete(request.Response); ok {
				deadline = previous.(time.Time)
			}
		}
		ctx, cancel := context.WithDeadline(request.Context(), deadline)
		response, err := t.base.RoundTrip(request.Clone(ctx))
		if err != nil {
			cancel()
			return response, err
		}
		if response.StatusCode >= 300 && response.StatusCode < 400 && response.Header.Get("Location") != "" {
			t.redirectDeadlines.Store(response, deadline)
			time.AfterFunc(time.Until(deadline), func() { t.redirectDeadlines.Delete(response) })
		}
		response.Body = &cancelOnCloseBody{ReadCloser: response.Body, cancel: cancel}
		return response, nil
	}
	waitDone := make(chan struct{})
	if beat, ok := request.Context().Value(waitHeartbeatKey{}).(func()); ok {
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-waitDone:
					return
				case <-request.Context().Done():
					return
				case <-ticker.C:
					beat()
				}
			}
		}()
	}
	permit, err := t.admit(request.Context(), target)
	close(waitDone)
	if err != nil {
		return nil, fmt.Errorf("图片正在等待全站并发名额，或共享队列暂不可用：%w", err)
	}
	started := time.Now()
	providerTimeout := t.timeout
	if providerTimeout <= 0 {
		providerTimeout = 10 * time.Minute
	}
	minimumWindow := providerTimeout
	if minimumWindow > 30*time.Second {
		minimumWindow = 30 * time.Second
	}
	if deadline, ok := request.Context().Deadline(); ok && time.Until(deadline) < minimumWindow {
		permit.Finish(Outcome{})
		return nil, fmt.Errorf("排队后剩余时间不足，未向上游提交，请稍后继续")
	}
	ctx, cancel := context.WithTimeout(request.Context(), providerTimeout)
	stop := context.AfterFunc(permit.Context(), cancel)
	finish := func(outcome Outcome) { stop(); cancel(); permit.Finish(outcome) }
	response, err := t.base.RoundTrip(request.Clone(ctx))
	if err != nil {
		finish(Outcome{Unknown: true, Duration: time.Since(started)})
		return response, fmt.Errorf("%w: %w", ErrOutcomeUnknown, err)
	}
	body := &admittedBody{ReadCloser: response.Body, finish: finish, started: started, status: response.StatusCode, retryAfter: parseRetryAfter(response.Header.Get("Retry-After")), op: nil}
	if op, ok := request.Context().Value(operationKey{}).(*operation); ok {
		body.op = op
		op.add(body.endOperation)
	}
	response.Body = body
	return response, nil
}

type cancelOnCloseBody struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (b *cancelOnCloseBody) Close() error { err := b.ReadCloser.Close(); b.cancel(); return err }

type admittedBody struct {
	io.ReadCloser
	mu         sync.Mutex
	once       sync.Once
	eof        bool
	readErr    error
	finish     func(Outcome)
	started    time.Time
	status     int
	retryAfter time.Duration
	op         *operation
}

func (b *admittedBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	b.mu.Lock()
	if err == io.EOF {
		b.eof = true
	} else if err != nil {
		b.readErr = err
	}
	b.mu.Unlock()
	if err != nil && err != io.EOF {
		return n, fmt.Errorf("%w: %w", ErrOutcomeUnknown, err)
	}
	return n, err
}
func (b *admittedBody) Close() error {
	err := b.ReadCloser.Close()
	if b.status < 200 || b.status >= 300 || b.op == nil {
		b.endOperation(err)
	}
	return err
}
func (b *admittedBody) endOperation(operationErr error) {
	b.once.Do(func() {
		b.mu.Lock()
		eof, readErr := b.eof, b.readErr
		b.mu.Unlock()
		rejected := b.status >= 400 && b.status < 500
		confirmed := false
		if b.op != nil {
			b.op.mu.Lock()
			confirmed = b.op.confirmed
			b.op.mu.Unlock()
		}
		unknown := (!eof && !rejected) || b.status == http.StatusAccepted || (b.status >= 200 && b.status < 300 && !confirmed)
		outcome := Outcome{Success: b.status >= 200 && b.status < 300 && eof && readErr == nil && operationErr == nil && !unknown, Limited: b.status == 429, Transient: !unknown && (b.status >= 500 || readErr != nil), Unknown: unknown, RetryAfter: b.retryAfter, Duration: time.Since(b.started)}
		b.finish(outcome)
	})
}
func parseRetryAfter(value string) time.Duration {
	if seconds, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	if at, err := http.ParseTime(value); err == nil && time.Until(at) > 0 {
		return time.Until(at)
	}
	return 0
}
