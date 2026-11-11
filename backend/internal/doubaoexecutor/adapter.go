package doubaoexecutor

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"

	"qiantie/backend/internal/safehttp"
)

const DefaultWorkspaceURL = "https://www.doubao.com/chat/create-image"

type Config struct {
	ChromePath   string
	UserDataDir  string
	WorkspaceURL string
	Headless     bool
	Timeout      time.Duration
}

type Adapter struct {
	cfg      Config
	mu       sync.Mutex
	accounts map[string]*account
}

type account struct {
	cancel    context.CancelFunc
	ctx       context.Context
	tempFiles []string
}

type Job struct {
	ID      string
	Payload map[string]any
}

type Submission struct {
	ID string
}

type Completion struct {
	MediaURL string
}

func New(cfg Config) (*Adapter, error) {
	if strings.TrimSpace(cfg.UserDataDir) == "" {
		return nil, errors.New("doubao user data directory is required")
	}
	if cfg.WorkspaceURL == "" {
		cfg.WorkspaceURL = DefaultWorkspaceURL
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Minute
	}
	if err := os.MkdirAll(cfg.UserDataDir, 0o700); err != nil {
		return nil, err
	}
	return &Adapter{cfg: cfg, accounts: make(map[string]*account)}, nil
}

func (a *Adapter) open(ctx context.Context, accountID string) (context.Context, context.CancelFunc, error) {
	accountID = safeID(accountID)
	a.mu.Lock()
	if existing := a.accounts[accountID]; existing != nil {
		a.mu.Unlock()
		return existing.ctx, existing.cancel, nil
	}
	a.mu.Unlock()
	profile := filepath.Join(a.cfg.UserDataDir, accountID)
	if err := os.MkdirAll(profile, 0o700); err != nil {
		return nil, nil, err
	}
	flags := []chromedp.ExecAllocatorOption{chromedp.Flag("headless", a.cfg.Headless), chromedp.Flag("disable-gpu", a.cfg.Headless), chromedp.UserDataDir(profile), chromedp.WindowSize(1280, 860)}
	if strings.TrimSpace(a.cfg.ChromePath) != "" {
		flags = append(flags, chromedp.ExecPath(a.cfg.ChromePath))
	}
	alloc, allocCancel := chromedp.NewExecAllocator(ctx, flags...)
	page, cancel := chromedp.NewContext(alloc)
	page, timeoutCancel := context.WithTimeout(page, a.cfg.Timeout)
	cancelAll := func() { timeoutCancel(); cancel(); allocCancel() }
	a.mu.Lock()
	a.accounts[accountID] = &account{ctx: page, cancel: cancelAll}
	a.mu.Unlock()
	return page, cancelAll, nil
}

func (a *Adapter) Prepare(ctx context.Context, accountID string, job Job) error {
	page, _, err := a.open(ctx, accountID)
	if err != nil {
		return err
	}
	payload := job.Payload
	prompt := strings.TrimSpace(asString(payload["prompt"], payload["videoPrompt"], payload["text"]))
	if prompt == "" {
		return errors.New("video prompt is required")
	}
	if err := a.clearTempFiles(accountID); err != nil {
		return err
	}
	if err := chromedp.Run(page,
		chromedp.Navigate(a.cfg.WorkspaceURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.ActionFunc(func(actionCtx context.Context) error {
			return evaluateSetPrompt(actionCtx, page, prompt)
		}),
	); err != nil {
		return err
	}
	imageURLs := asStrings(payload["imageUrls"], payload["image_urls"], payload["referenceImageURLs"], payload["reference_image_urls"])
	if len(imageURLs) == 0 {
		return nil
	}
	paths := make([]string, 0, len(imageURLs))
	for index, imageURL := range imageURLs {
		path, err := downloadReference(ctx, imageURL, index)
		if err != nil {
			_ = a.clearTempFiles(accountID)
			return fmt.Errorf("reference image %d: %w", index+1, err)
		}
		paths = append(paths, path)
	}
	a.mu.Lock()
	if item := a.accounts[safeID(accountID)]; item != nil {
		item.tempFiles = paths
	}
	a.mu.Unlock()
	if err := chromedp.Run(page, chromedp.SetUploadFiles(`input[type="file"]`, paths, chromedp.ByQuery)); err != nil {
		_ = a.clearTempFiles(accountID)
		return fmt.Errorf("upload reference images: %w", err)
	}
	return nil
}

func (a *Adapter) Submit(ctx context.Context, accountID string) (Submission, error) {
	a.mu.Lock()
	item := a.accounts[safeID(accountID)]
	a.mu.Unlock()
	if item == nil {
		return Submission{}, errors.New("doubao account is not open")
	}
	err := chromedp.Run(item.ctx, chromedp.ActionFunc(func(actionCtx context.Context) error {
		return evaluateClick(actionCtx, item.ctx, []string{"生成视频", "开始制作", "立即生成", "开始生成", "生成", "发送"})
	}))
	if err != nil {
		return Submission{}, err
	}
	return Submission{ID: fmt.Sprintf("doubao-%d", time.Now().UnixNano())}, nil
}

func (a *Adapter) Wait(ctx context.Context, accountID string, _ Submission) (Completion, error) {
	a.mu.Lock()
	item := a.accounts[safeID(accountID)]
	a.mu.Unlock()
	if item == nil {
		return Completion{}, errors.New("doubao account is not open")
	}
	deadline := time.Now().Add(25 * time.Minute)
	for time.Now().Before(deadline) {
		var src string
		if err := chromedp.Run(item.ctx, chromedp.Evaluate(`(() => { const v=[...document.querySelectorAll('video')].find(x=>x.currentSrc||x.src); return v ? (v.currentSrc||v.src) : ''; })()`, &src)); err == nil && strings.HasPrefix(src, "http") {
			return Completion{MediaURL: src}, nil
		}
		select {
		case <-ctx.Done():
			return Completion{}, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return Completion{}, errors.New("doubao video did not become available before timeout")
}

func (a *Adapter) Download(ctx context.Context, completion Completion, output string) error {
	if !strings.HasPrefix(completion.MediaURL, "http") {
		return errors.New("official media URL is missing")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, completion.MediaURL, nil)
	if err != nil {
		return err
	}
	resp, err := safehttp.NewClientWithTimeout(30 * time.Minute).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("media download failed: HTTP %d", resp.StatusCode)
	}
	file, err := os.OpenFile(output, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = copyWithContext(ctx, file, resp.Body)
	return err
}

func (a *Adapter) Close(accountID string) {
	a.mu.Lock()
	item := a.accounts[safeID(accountID)]
	delete(a.accounts, safeID(accountID))
	a.mu.Unlock()
	if item != nil {
		for _, path := range item.tempFiles {
			_ = os.Remove(path)
		}
		item.cancel()
	}
}

func (a *Adapter) clearTempFiles(accountID string) error {
	a.mu.Lock()
	item := a.accounts[safeID(accountID)]
	var paths []string
	if item != nil {
		paths = append(paths, item.tempFiles...)
		item.tempFiles = nil
	}
	a.mu.Unlock()
	if len(paths) == 0 {
		return nil
	}
	var firstErr error
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func evaluateSetPrompt(ctx context.Context, page context.Context, prompt string) error {
	var ok bool
	script := fmt.Sprintf(`(() => { const wanted=%q; const visible=e=>{const s=getComputedStyle(e),r=e.getBoundingClientRect();return s.display!=='none'&&s.visibility!=='hidden'&&r.width>0&&r.height>0}; const xs=[...document.querySelectorAll('textarea,[contenteditable="true"]')].filter(visible); xs.sort((a,b)=>((/提示词|描述|输入/.test(b.placeholder||b.getAttribute('aria-label')||''))?1:0)-((/提示词|描述|输入/.test(a.placeholder||a.getAttribute('aria-label')||''))?1:0)); const e=xs[0]; if(!e)return false; e.focus(); if(e.tagName.toLowerCase()==='textarea'){const d=Object.getOwnPropertyDescriptor(HTMLTextAreaElement.prototype,'value');d?.set?.call(e,wanted);e.dispatchEvent(new InputEvent('input',{bubbles:true,inputType:'insertText',data:wanted}));e.dispatchEvent(new Event('change',{bubbles:true}));}else{e.textContent=wanted;e.dispatchEvent(new InputEvent('input',{bubbles:true,inputType:'insertText',data:wanted}));} return true; })()`, prompt)
	if err := chromedp.Run(page, chromedp.Evaluate(script, &ok)); err != nil {
		return err
	}
	if !ok {
		return errors.New("doubao prompt editor not found")
	}
	return nil
}

func evaluateClick(ctx context.Context, page context.Context, labels []string) error {
	var clicked bool
	script := fmt.Sprintf(`(() => { const wanted=%q; const visible=e=>{const s=getComputedStyle(e),r=e.getBoundingClientRect();return s.display!=='none'&&s.visibility!=='hidden'&&r.width>0&&r.height>0}; const text=e=>(e.innerText||e.textContent||e.getAttribute('aria-label')||'').replace(/\s+/g,' ').trim(); for(const w of wanted){const xs=[...document.querySelectorAll('button,[role="button"],[role="menuitem"]')].filter(e=>visible(e)&&!e.disabled&&text(e)===w); if(xs.length===1){xs[0].click();return true;}} return false; })()`, labels)
	if err := chromedp.Run(page, chromedp.Evaluate(script, &clicked)); err != nil {
		return err
	}
	if !clicked {
		return errors.New("doubao submit control not found")
	}
	return nil
}

func safeID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "default"
	}
	var b strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "default"
	}
	return b.String()
}
func asString(values ...any) string {
	for _, v := range values {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}

func asStrings(values ...any) []string {
	for _, value := range values {
		switch typed := value.(type) {
		case []string:
			return nonEmptyStrings(typed)
		case []any:
			out := make([]string, 0, len(typed))
			for _, item := range typed {
				if text, ok := item.(string); ok && strings.TrimSpace(text) != "" {
					out = append(out, strings.TrimSpace(text))
				}
			}
			if len(out) > 0 {
				return out
			}
		}
	}
	return nil
}

func nonEmptyStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if text := strings.TrimSpace(value); text != "" {
			out = append(out, text)
		}
	}
	return out
}

func downloadReference(ctx context.Context, value string, index int) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("image URL is empty")
	}
	file, err := os.CreateTemp("", fmt.Sprintf("qiantie-reference-%d-*.img", index))
	if err != nil {
		return "", err
	}
	path := file.Name()
	cleanup := func(err error) (string, error) {
		_ = file.Close()
		_ = os.Remove(path)
		return "", err
	}
	if strings.HasPrefix(value, "data:") {
		comma := strings.IndexByte(value, ',')
		if comma < 0 {
			return cleanup(errors.New("invalid data image"))
		}
		encoded := value[comma+1:]
		var data []byte
		if strings.Contains(value[:comma], ";base64") {
			data, err = base64.StdEncoding.DecodeString(encoded)
		} else {
			var decoded string
			decoded, err = url.QueryUnescape(encoded)
			data = []byte(decoded)
		}
		if err != nil {
			return cleanup(fmt.Errorf("decode image: %w", err))
		}
		if _, err := file.Write(data); err != nil {
			return cleanup(err)
		}
	} else if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, value, nil)
		if err != nil {
			return cleanup(err)
		}
		resp, err := safehttp.NewClientWithTimeout(30 * time.Minute).Do(req)
		if err != nil {
			return cleanup(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return cleanup(fmt.Errorf("HTTP %d", resp.StatusCode))
		}
		if _, err := copyWithContext(ctx, file, io.LimitReader(resp.Body, 50<<20)); err != nil {
			return cleanup(err)
		}
	} else {
		input, err := os.Open(value)
		if err != nil {
			return cleanup(err)
		}
		_, copyErr := copyWithContext(ctx, file, io.LimitReader(input, 50<<20))
		_ = input.Close()
		if copyErr != nil {
			return cleanup(copyErr)
		}
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	return path, nil
}
func copyWithContext(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	if ctx == nil {
		return io.Copy(dst, src)
	}
	result := make(chan struct {
		n   int64
		err error
	}, 1)
	go func() {
		n, err := io.Copy(dst, src)
		result <- struct {
			n   int64
			err error
		}{n, err}
	}()
	select {
	case out := <-result:
		return out.n, out.err
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}
