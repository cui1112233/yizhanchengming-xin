package httpapi

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net"
	"net/http"
	"runtime"
	"strings"

	"qiantie/giant-material-executor/internal/agent"
)

type JobStatus struct {
	ID           string         `json:"id"`
	State        agent.State    `json:"state"`
	Progress     agent.Progress `json:"progress"`
	ErrorCode    string         `json:"errorCode,omitempty"`
	ErrorMessage string         `json:"errorMessage,omitempty"`
}

type ServerConfig struct {
	Addr          string
	Origin        string
	AllowedOrigin func() string
	Nonce         string
	Version       string
	// PublicURL 是当前生效的控制服务地址，预填在配置页输入框里，
	// 普通用户不需要知道服务器地址，直接填配对码即可。
	PublicURL  string
	Snapshot   func() agent.Snapshot
	ModelReady func() bool
	Callbacks  Callbacks
}

type Callbacks struct {
	SetPublicURL func(string) error
	Pair         func(context.Context, string) error
	StartJob     func(context.Context, string) error
	GetJob       func(context.Context, string) (JobStatus, error)
	Cancel       func(context.Context, string) error
	// CheckUpdate 立即检查一次更新；返回新版本号（已是最新返回空）。
	CheckUpdate func(context.Context) (string, error)
	// ApplyUpdate 仅由本机回环页面在用户确认后调用。
	ApplyUpdate func(context.Context) (string, error)
}

type Server struct {
	config ServerConfig
}

func NewServer(config ServerConfig) (*Server, error) {
	if err := validateLoopbackAddr(config.Addr); err != nil {
		return nil, err
	}
	if strings.TrimSpace(config.Origin) == "" || strings.TrimSpace(config.Nonce) == "" {
		return nil, errors.New("loopback origin and nonce are required")
	}
	if config.Snapshot == nil {
		config.Snapshot = func() agent.Snapshot { return agent.Snapshot{State: agent.StateIdle, WorkerResident: true} }
	}
	return &Server{config: config}, nil
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		return err
	}
	server := &http.Server{Handler: s.Handler()}
	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()
	err = server.Serve(listener)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /setup", s.setupPage)
	mux.HandleFunc("GET /v1/health", s.health)
	mux.HandleFunc("GET /v1/capabilities", s.capabilities)
	mux.HandleFunc("POST /v1/pair", s.pair)
	mux.HandleFunc("POST /v1/jobs", s.startJob)
	mux.HandleFunc("GET /v1/jobs/{id}", s.getJob)
	mux.HandleFunc("POST /v1/jobs/{id}/cancel", s.cancelJob)
	mux.HandleFunc("POST /v1/update", s.checkUpdate)
	mux.HandleFunc("POST /v1/update/apply", s.applyUpdate)
	return s.guard(mux)
}

func (s *Server) guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet && req.URL.Path == "/setup" {
			next.ServeHTTP(w, req)
			return
		}
		requestOrigin := strings.TrimSpace(req.Header.Get("Origin"))
		if req.Method == http.MethodOptions {
			// 浏览器对“公网页面 → 本机回环地址”的请求会先发预检（含 Chrome 私网预检），
			// 这里直接放行并回显来源，否则网页永远拿不到执行器状态。
			w.Header().Set("Access-Control-Allow-Origin", requestOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Giant-Executor-Nonce")
			w.Header().Set("Access-Control-Allow-Private-Network", "true")
			w.Header().Set("Access-Control-Max-Age", "600")
			w.Header().Set("Vary", "Origin")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		originOK := requestOrigin == s.config.Origin || requestOrigin == "http://"+s.config.Addr
		if !originOK && s.config.AllowedOrigin != nil {
			originOK = requestOrigin == strings.TrimSpace(s.config.AllowedOrigin())
		}
		nonceOK := req.Header.Get("X-Giant-Executor-Nonce") == s.config.Nonce
		bootstrapPair := req.Method == http.MethodPost && req.URL.Path == "/v1/pair" && strings.TrimSpace(req.Header.Get("X-Giant-Executor-Nonce")) == ""
		localHealthProbe := req.Method == http.MethodGet && req.URL.Path == "/v1/health" && isLoopbackRemote(req.RemoteAddr)
		localSetupRequest := isLoopbackRemote(req.RemoteAddr) && sameHostPort(req.Host, s.config.Addr) && (req.URL.Path == "/v1/pair" || req.URL.Path == "/v1/health" || req.URL.Path == "/v1/update" || req.URL.Path == "/v1/update/apply")
		if (!originOK && !localHealthProbe && !localSetupRequest) || (!nonceOK && !bootstrapPair && !localHealthProbe && !localSetupRequest) {
			writeError(w, http.StatusForbidden, "loopback origin or nonce rejected")
			return
		}
		allowOrigin := s.config.Origin
		if requestOrigin != "" {
			// 谁来问就回显谁的来源，浏览器才能把结果交给发起请求的网页。
			allowOrigin = requestOrigin
		}
		w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		w.Header().Set("Vary", "Origin")
		next.ServeHTTP(w, req)
	})
}

func isLoopbackRemote(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(strings.TrimSpace(remoteAddr))
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func sameHostPort(left, right string) bool {
	leftHost, leftPort, leftErr := net.SplitHostPort(strings.TrimSpace(left))
	rightHost, rightPort, rightErr := net.SplitHostPort(strings.TrimSpace(right))
	return leftErr == nil && rightErr == nil && leftPort == rightPort && (leftHost == rightHost || (isLoopbackHost(leftHost) && isLoopbackHost(rightHost)))
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(strings.TrimSpace(host), "localhost") {
		return true
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	return ip != nil && ip.IsLoopback()
}

func (s *Server) setupPage(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	page := strings.ReplaceAll(setupPageHTML, "__EXECUTOR_VERSION__", html.EscapeString(strings.TrimSpace(s.config.Version)))
	page = strings.Replace(page, "__PUBLIC_URL__", html.EscapeString(strings.TrimSpace(s.config.PublicURL)), 1)
	page = strings.Replace(page, "</body>", `<script>document.getElementById('applyUpdate').addEventListener('click',async e=>{const b=e.target;b.disabled=true;const s=document.getElementById('status');s.textContent='正在下载、校验并重启执行器…';try{const r=await fetch('/v1/update/apply',{method:'POST'});const d=await r.json();if(!r.ok)throw new Error(d.error||'更新失败');s.textContent='正在重启并恢复连接到 '+d.version+'，窗口会短暂消失。';s.className='ok'}catch(err){s.textContent='更新失败：'+err.message;s.className='error';b.disabled=false}});</script></body>`, 1)
	_, _ = w.Write([]byte(page))
}

const setupPageHTML = `<!doctype html>
<html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>巨量素材执行器</title><style>
:root{color-scheme:dark;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif}body{margin:0;background:#08151e;color:#e8f0f6;min-height:100vh;display:grid;place-items:center}.card{width:min(560px,calc(100vw - 40px));padding:32px;border:1px solid #20566d;border-radius:18px;background:#0d2230;box-sizing:border-box}h1{margin:0 0 8px;font-size:28px}.muted{color:#9bb1c0;line-height:1.6}label{display:block;margin-top:22px;font-weight:600}input{display:block;margin-top:8px;width:100%;height:42px;box-sizing:border-box;border:1px solid #326a80;border-radius:9px;background:#071720;color:#f6fbff;padding:0 12px;font-size:15px}button{margin-top:24px;width:100%;height:44px;border:0;border-radius:9px;background:#64dfc0;color:#062119;font-size:16px;font-weight:700;cursor:pointer}button:disabled{opacity:.6;cursor:wait}#status{margin-top:18px;padding:12px;border-radius:8px;background:#102c3b;color:#b7cbd7;white-space:pre-wrap}.error{background:#3b1820!important;color:#ffc1c7!important}.ok{background:#10382d!important;color:#abf4d9!important}</style></head>
<body><main class="card"><h1>巨量素材执行器</h1><p class="muted">当前版本 <strong id="executorVersion">__EXECUTOR_VERSION__</strong>。填写平台控制服务地址与网页生成的配对码。绑定完成后，执行器会继续在后台运行。</p>
<form id="pairForm"><label for="publicURL">控制服务地址（已预填，无需修改）</label><input id="publicURL" name="publicURL" autocomplete="url" placeholder="https://你的平台域名" value="__PUBLIC_URL__" required><label for="code">配对码</label><input id="code" name="code" autocomplete="one-time-code" placeholder="例如 ABCDE-12345" required><button id="submit" type="submit">绑定执行器</button></form><button type="button" id="checkUpdate" style="margin-top:12px;width:100%;height:40px;border:0;border-radius:9px;background:#2b5f78;color:#dff3ff;font-size:14px;font-weight:600;cursor:pointer">检查更新（当前版本 __EXECUTOR_VERSION__）</button><button type="button" id="applyUpdate" style="margin-top:12px;width:100%;height:40px;border:0;border-radius:9px;background:#64dfc0;color:#062119;font-size:14px;font-weight:700;cursor:pointer">立即更新</button><div id="status">正在检查本机执行器状态…</div></main>
<script>const status=document.getElementById('status'),submit=document.getElementById('submit');function setStatus(message,kind=''){status.textContent=message;status.className=kind}function taskText(d){if(d.state==='running'||d.state==='cleaning'||d.state==='uploading'){const p=d.progress&&d.progress.total>0?('，已完成 '+d.progress.completed+'/'+d.progress.total+'（'+d.progress.percent+'%）'):'';return '正在执行 OCR 任务'+p+'…，可以关掉此页面，任务会继续。'}if(d.state==='downloading_model')return '正在准备 OCR 运行环境（首次使用需下载依赖，请耐心等待）…';if(d.state==='failed')return '上次任务失败：'+(d.errorMessage||d.errorCode||'未知原因');return ''}async function health(){try{const r=await fetch('/v1/health');if(!r.ok)throw new Error();const d=await r.json();const task=taskText(d);if(task){setStatus(task);return}if(d.bindingState==='online')setStatus('已绑定并在线，执行器正在后台等待任务。','ok');else if(!submit.disabled)setStatus('执行器已启动，请填写地址和配对码。')}catch{if(!submit.disabled)setStatus('执行器已启动；等待绑定。')}}document.getElementById('pairForm').addEventListener('submit',async e=>{e.preventDefault();submit.disabled=true;setStatus('正在绑定，请保持此页面打开…');try{const r=await fetch('/v1/pair',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({publicURL:document.getElementById('publicURL').value.trim(),code:document.getElementById('code').value.trim()})});const d=await r.json();if(!r.ok)throw new Error(d.error||'绑定失败');setStatus('绑定成功，执行器正在连接后台。','ok');document.getElementById('code').value=''}catch(err){setStatus('绑定失败：'+err.message,'error')}finally{submit.disabled=false}});document.getElementById('checkUpdate').addEventListener('click',async e=>{const b=e.target;b.disabled=true;setStatus('正在检查更新…');try{const r=await fetch('/v1/update',{method:'POST'});const d=await r.json();if(!r.ok)throw new Error(d.error||'检查更新失败');if(d.updated)setStatus('发现新版本 '+d.version+'，请点击“立即更新”。','ok');else setStatus('已是最新版本 '+d.current+'。','ok')}catch(err){setStatus('检查更新失败：'+err.message,'error')}finally{b.disabled=false}});health();setInterval(health,1000);</script></body></html>`

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	snapshot := s.config.Snapshot()
	modelReady := snapshot.State != agent.StateDownloadingModel
	if s.config.ModelReady != nil {
		modelReady = s.config.ModelReady()
	}
	writeJSON(w, http.StatusOK, map[string]any{"online": true, "version": s.config.Version, "state": snapshot.State, "bindingState": snapshot.BindingState, "modelReady": modelReady, "workerResident": snapshot.WorkerResident, "jobId": snapshot.JobID, "progress": snapshot.Progress, "errorCode": snapshot.ErrorCode, "errorMessage": snapshot.ErrorMessage})
}

func (s *Server) capabilities(w http.ResponseWriter, _ *http.Request) {
	snapshot := s.config.Snapshot()
	modelReady := snapshot.State != agent.StateDownloadingModel
	if s.config.ModelReady != nil {
		modelReady = s.config.ModelReady()
	}
	writeJSON(w, http.StatusOK, map[string]any{"platform": runtime.GOOS, "ocr": true, "modelVersion": snapshot.ModelVersion, "modelReady": modelReady})
}

func (s *Server) pair(w http.ResponseWriter, req *http.Request) {
	var input struct {
		PublicURL string `json:"publicURL"`
		Code      string `json:"code"`
	}
	if err := decodeJSON(w, req, &input); err != nil || strings.TrimSpace(input.Code) == "" {
		writeError(w, http.StatusBadRequest, "pairing code is required")
		return
	}
	if s.config.Callbacks.Pair == nil {
		writeError(w, http.StatusNotImplemented, "pairing is not configured")
		return
	}
	if publicURL := strings.TrimSpace(input.PublicURL); publicURL != "" && s.config.Callbacks.SetPublicURL != nil {
		if err := s.config.Callbacks.SetPublicURL(publicURL); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if err := s.config.Callbacks.Pair(req.Context(), strings.TrimSpace(input.Code)); err != nil {
		writeError(w, http.StatusConflict, "pairing failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"paired": true, "nonce": s.config.Nonce})
}

func (s *Server) startJob(w http.ResponseWriter, req *http.Request) {
	var input struct {
		JobID string `json:"jobId"`
	}
	if err := decodeJSON(w, req, &input); err != nil || strings.TrimSpace(input.JobID) == "" {
		writeError(w, http.StatusBadRequest, "jobId is required")
		return
	}
	if s.config.Callbacks.StartJob == nil {
		writeError(w, http.StatusNotImplemented, "job start is not configured")
		return
	}
	if err := s.config.Callbacks.StartJob(req.Context(), strings.TrimSpace(input.JobID)); err != nil {
		writeError(w, http.StatusConflict, "job start failed")
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"jobId": strings.TrimSpace(input.JobID), "accepted": true})
}

func (s *Server) getJob(w http.ResponseWriter, req *http.Request) {
	if s.config.Callbacks.GetJob == nil {
		writeError(w, http.StatusNotImplemented, "job status is not configured")
		return
	}
	job, err := s.config.Callbacks.GetJob(req.Context(), req.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (s *Server) cancelJob(w http.ResponseWriter, req *http.Request) {
	if s.config.Callbacks.Cancel == nil {
		writeError(w, http.StatusNotImplemented, "job cancellation is not configured")
		return
	}
	if err := s.config.Callbacks.Cancel(req.Context(), req.PathValue("id")); err != nil {
		writeError(w, http.StatusConflict, "job cancellation failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"cancelled": true})
}

func (s *Server) checkUpdate(w http.ResponseWriter, req *http.Request) {
	if s.config.Callbacks.CheckUpdate == nil {
		writeError(w, http.StatusNotImplemented, "update check is not configured")
		return
	}
	nextVersion, err := s.config.Callbacks.CheckUpdate(req.Context())
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]any{"updated": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"updated": strings.TrimSpace(nextVersion) != "", "version": strings.TrimSpace(nextVersion), "current": s.config.Version})
}

func (s *Server) applyUpdate(w http.ResponseWriter, req *http.Request) {
	if s.config.Callbacks.ApplyUpdate == nil {
		writeError(w, http.StatusNotImplemented, "update apply is not configured")
		return
	}
	nextVersion, err := s.config.Callbacks.ApplyUpdate(req.Context())
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]any{"applied": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"applied": true, "version": strings.TrimSpace(nextVersion), "current": s.config.Version})
}

func validateLoopbackAddr(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil || host != "127.0.0.1" {
		return fmt.Errorf("loopback server must bind to 127.0.0.1: %s", addr)
	}
	return nil
}

func decodeJSON(w http.ResponseWriter, req *http.Request, target any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, req.Body, 2<<20))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
