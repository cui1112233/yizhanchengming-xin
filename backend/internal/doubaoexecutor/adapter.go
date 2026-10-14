package doubaoexecutor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
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
	cancel context.CancelFunc
	ctx    context.Context
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
	return chromedp.Run(page,
		chromedp.Navigate(a.cfg.WorkspaceURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.ActionFunc(func(actionCtx context.Context) error {
			return evaluateSetPrompt(actionCtx, page, prompt)
		}),
	)
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
	resp, err := http.DefaultClient.Do(req)
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
		item.cancel()
	}
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
