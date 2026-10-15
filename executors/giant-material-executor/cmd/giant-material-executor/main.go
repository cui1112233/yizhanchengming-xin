package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"qiantie/giant-material-executor/internal/agent"
	"qiantie/giant-material-executor/internal/credential"
	"qiantie/giant-material-executor/internal/httpapi"
	"qiantie/giant-material-executor/internal/modelcache"
	"qiantie/giant-material-executor/internal/platform"
	"qiantie/giant-material-executor/internal/ui"
	"qiantie/giant-material-executor/internal/worker"
)

var version = "dev"

// publicAPIClientTimeout 必须大于服务端的 25 秒 claim 长轮询窗口。
// 否则空闲执行器会在服务端返回前自行断开，后续新任务也无法稳定领取。
const publicAPIClientTimeout = 35 * time.Second

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	closeLog := configureLogging()
	defer closeLog()

	nonce, err := loadOrCreateNonce()
	if err != nil {
		log.Fatalf("load executor nonce: %v", err)
	}
	publicURL := loadPublicAPIURL()
	origin := envOr("GIANT_MATERIAL_EXECUTOR_ORIGIN", "http://127.0.0.1:5173")
	deviceName := envOr("GIANT_MATERIAL_EXECUTOR_DEVICE_NAME", localDeviceName())
	exeDir := executableDir()
	workerCommand, err := platform.WorkerCommand(runtime.GOOS, exeDir)
	if err != nil {
		log.Fatal(err)
	}
	var prepareRuntime func(context.Context, string) error
	if runtime.GOOS == "windows" {
		python := defaultPythonCommand()
		script := workerScriptPath()
		workerCommand = []string{python, script}
	}
	supervisor := &worker.Supervisor{Command: workerCommand, Dir: exeDir}
	if runtime.GOOS == "windows" {
		python := workerCommand[0]
		modelPrepare := modelPreparer(supervisor)
		prepareRuntime = func(ctx context.Context, modelVersion string) error {
			if err := ensureRuntime(ctx, python, exeDir); err != nil {
				return err
			}
			if modelPrepare != nil {
				return modelPrepare(ctx, modelVersion)
			}
			return nil
		}
	}
	var stateMu sync.RWMutex
	var runningAgent *agent.Agent
	bindingState := agent.BindingUnpaired
	setBindingState := func(next agent.BindingState) {
		stateMu.Lock()
		bindingState = next
		stateMu.Unlock()
	}
	snapshot := func() agent.Snapshot {
		stateMu.RLock()
		defer stateMu.RUnlock()
		var current agent.Snapshot
		if runningAgent != nil {
			current = runningAgent.Snapshot()
		} else {
			current = agent.NewStateMachine().Snapshot()
		}
		current.BindingState = bindingState
		return current
	}
	publicURLMu := sync.RWMutex{}
	executorOriginMu := sync.RWMutex{}
	executorOrigin := origin
	if parsed, parseErr := url.Parse(publicURL); parseErr == nil && parsed.Scheme != "" && parsed.Host != "" {
		executorOrigin = parsed.Scheme + "://" + parsed.Host
	}
	publicClient := agent.NewHTTPClient(publicURL, &http.Client{Timeout: publicAPIClientTimeout})
	getPublicClient := func() *agent.HTTPClient {
		publicURLMu.RLock()
		defer publicURLMu.RUnlock()
		return publicClient
	}
	setPublicURL := func(raw string) error {
		normalized, normalizeErr := normalizePublicAPIURL(raw)
		if normalizeErr != nil {
			return normalizeErr
		}
		if saveErr := savePublicAPIURL(normalized); saveErr != nil {
			return errors.New("无法保存控制服务地址：" + saveErr.Error())
		}
		publicURLMu.Lock()
		publicURL = normalized
		publicClient = agent.NewHTTPClient(normalized, &http.Client{Timeout: publicAPIClientTimeout})
		publicURLMu.Unlock()
		if parsed, parseErr := url.Parse(normalized); parseErr == nil && parsed.Scheme != "" && parsed.Host != "" {
			executorOriginMu.Lock()
			executorOrigin = parsed.Scheme + "://" + parsed.Host
			executorOriginMu.Unlock()
		}
		return nil
	}
	credentialStore, err := newCredentialStore()
	if err != nil {
		log.Fatalf("load executor credential store: %v", err)
	}
	startAgent := func(token string) {
		stateMu.Lock()
		if runningAgent != nil {
			stateMu.Unlock()
			return
		}
		bindingState = agent.BindingConnecting
		stateMu.Unlock()
		go func() {
			for {
				client := getPublicClient()
				next, newErr := agent.New(agent.Config{
					Client:         client,
					Supervisor:     supervisor,
					PrepareModel:   prepareRuntime,
					OnBindingState: setBindingState,
					Token:          token,
					DeviceName:     deviceName,
					OS:             platform.ExecutorPlatform(runtime.GOOS),
					Version:        version,
				})
				if newErr != nil {
					setBindingState(agent.BindingOffline)
					log.Printf("agent configuration failed: %v", newErr)
					return
				}
				stateMu.Lock()
				runningAgent = next
				stateMu.Unlock()
				runErr := next.Run(ctx)
				stateMu.Lock()
				if runningAgent == next {
					runningAgent = nil
				}
				stateMu.Unlock()
				if ctx.Err() != nil {
					return
				}
				if nextState := handleAgentRunError(credentialStore, runErr); nextState == agent.BindingNeedsPairing {
					setBindingState(nextState)
					log.Printf("executor credential rejected; pairing required")
					return
				}
				setBindingState(agent.BindingOffline)
				if runErr != nil {
					log.Printf("agent stopped; retrying connection: %v", runErr)
				}
				timer := time.NewTimer(5 * time.Second)
				select {
				case <-ctx.Done():
					timer.Stop()
					return
				case <-timer.C:
					setBindingState(agent.BindingConnecting)
				}
			}
		}()
	}
	pair := func(pairCtx context.Context, code string) error {
		result, pairErr := getPublicClient().Pair(pairCtx, agent.PairInput{Code: code, DeviceName: deviceName, OS: platform.ExecutorPlatform(runtime.GOOS), Version: version, Platform: "giant_material"})
		if pairErr != nil {
			return pairErr
		}
		if strings.TrimSpace(result.Token) == "" {
			return &pairError{message: "public pair returned no executor token"}
		}
		if err := savePairResult(credentialStore, result); err != nil {
			return err
		}
		startAgent(result.Token)
		return nil
	}
	updater := startSelfUpdater(ctx, stop, snapshot, &executorOriginMu, &executorOrigin)
	server, err := httpapi.NewServer(httpapi.ServerConfig{Addr: "127.0.0.1:17861", Origin: origin, AllowedOrigin: func() string {
		executorOriginMu.RLock()
		defer executorOriginMu.RUnlock()
		return executorOrigin
	}, Nonce: nonce, Version: version, PublicURL: publicURL, Snapshot: snapshot, Callbacks: httpapi.Callbacks{SetPublicURL: setPublicURL, Pair: pair, CheckUpdate: func(checkCtx context.Context) (string, error) {
		if updater == nil {
			return "", errors.New("self update unavailable")
		}
		return updater.CheckNow(checkCtx)
	}, ApplyUpdate: func(applyCtx context.Context) (string, error) {
		if updater == nil {
			return "", errors.New("self update unavailable")
		}
		return updater.ApplyNow(applyCtx)
	}}})
	if err != nil {
		log.Fatalf("create loopback server: %v", err)
	}
	startupPaired := false
	if code := configuredPairingCode(); code != "" {
		pairCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		if pairErr := pair(pairCtx, code); pairErr != nil {
			log.Printf("startup pairing failed: %v", pairErr)
		} else {
			startupPaired = true
			log.Printf("startup pairing succeeded")
		}
		cancel()
	}
	if !startupPaired {
		if token := strings.TrimSpace(os.Getenv("GIANT_MATERIAL_EXECUTOR_TOKEN")); token != "" {
			startAgent(token)
		} else if record, loadErr := loadExecutorCredential(credentialStore); loadErr == nil {
			startAgent(record.Token)
		} else if !errors.Is(loadErr, credential.ErrNotFound) {
			log.Printf("saved executor credential unavailable; pairing required: %v", loadErr)
		}
	}
	go func() {
		if serveErr := server.ListenAndServe(ctx); serveErr != nil && ctx.Err() == nil {
			log.Printf("loopback server stopped: %v", serveErr)
		}
	}()
	go func() {
		if readyErr := waitForLoopbackSetup(ctx); readyErr != nil {
			if ctx.Err() == nil {
				log.Printf("loopback setup page unavailable: %v", readyErr)
			}
			return
		}
		if uiErr := ui.Run(ctx, ui.Config{PublicURL: publicURL, SetPublicURL: setPublicURL, Pair: pair, Snapshot: snapshot, Shutdown: stop, Version: version}); uiErr != nil && ctx.Err() == nil {
			log.Printf("executor UI stopped: %v", uiErr)
		}
	}()
	log.Printf("giant material executor started version=%s loopback=127.0.0.1:17861 workerResident=true", version)
	<-ctx.Done()
	log.Printf("giant material executor stopped")
}

func waitForLoopbackSetup(ctx context.Context) error {
	client := &http.Client{Timeout: 500 * time.Millisecond}
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:17861/v1/health", nil)
		if err == nil {
			resp, requestErr := client.Do(req)
			if requestErr == nil {
				_ = resp.Body.Close()
				if resp.StatusCode >= 200 && resp.StatusCode < 500 {
					return nil
				}
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return errors.New("等待本机配对页面服务超时")
		case <-ticker.C:
		}
	}
}

func configureLogging() func() {
	root, err := os.UserConfigDir()
	if err != nil {
		return func() {}
	}
	path := strings.TrimSpace(os.Getenv("GIANT_MATERIAL_EXECUTOR_LOG_PATH"))
	if path == "" {
		path = filepath.Join(root, "YizhanShengming", "GiantMaterialExecutor", "executor.log")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return func() {}
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return func() {}
	}
	log.SetOutput(file)
	return func() { _ = file.Close() }
}

func newCredentialStore() (credential.Store, error) {
	path := strings.TrimSpace(os.Getenv("GIANT_MATERIAL_EXECUTOR_CREDENTIAL_PATH"))
	if path == "" {
		root, err := os.UserConfigDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(root, "YizhanShengming", "GiantMaterialExecutor", "credential.bin")
	}
	if runtime.GOOS == "darwin" {
		// Older macOS executors persist a paired device in this restricted local
		// file. Keep using a valid existing record so an upgrade cannot force a
		// user to pair again; fresh installs still use the device store.
		return preferLegacyCredentialStore(credential.NewDeviceStore(path), credential.NewFileStore(path))
	}
	return credential.NewDeviceStore(path), nil
}

func preferLegacyCredentialStore(deviceStore, legacyStore credential.Store) (credential.Store, error) {
	if deviceStore == nil || legacyStore == nil {
		return nil, errors.New("credential store is required")
	}
	if _, err := legacyStore.Load(); err == nil {
		return legacyStore, nil
	} else if !errors.Is(err, credential.ErrNotFound) {
		return nil, err
	}
	return deviceStore, nil
}

func savePairResult(store credential.Store, result agent.PairResult) error {
	if store == nil {
		return errors.New("executor credential store is required")
	}
	return store.Save(credential.Record{ExecutorID: strings.TrimSpace(result.ExecutorID), Token: strings.TrimSpace(result.Token)})
}

func loadExecutorCredential(store credential.Store) (credential.Record, error) {
	if store == nil {
		return credential.Record{}, errors.New("executor credential store is required")
	}
	return store.Load()
}

func handleAgentRunError(store credential.Store, runErr error) agent.BindingState {
	if errors.Is(runErr, agent.ErrUnauthorized) {
		if store != nil {
			_ = store.Clear()
		}
		return agent.BindingNeedsPairing
	}
	return agent.BindingOffline
}

type pairError struct{ message string }

func (e *pairError) Error() string { return e.message }

func loadOrCreateNonce() (string, error) {
	if configured := strings.TrimSpace(os.Getenv("GIANT_MATERIAL_EXECUTOR_NONCE")); configured != "" {
		return configured, nil
	}
	root, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(root, "YizhanShengming", "GiantMaterialExecutor", "nonce")
	if data, readErr := os.ReadFile(path); readErr == nil && strings.TrimSpace(string(data)) != "" {
		return strings.TrimSpace(string(data)), nil
	}
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	nonce := hex.EncodeToString(bytes)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(nonce+"\n"), 0o600); err != nil {
		return "", err
	}
	return nonce, nil
}

func executableDir() string {
	path, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(path)
}

func defaultPythonCommand() string {
	if value := strings.TrimSpace(os.Getenv("GIANT_MATERIAL_PYTHON")); value != "" {
		return value
	}
	if runtime.GOOS == "windows" {
		return "python"
	}
	return "python3"
}

func workerScriptPath() string {
	if value := strings.TrimSpace(os.Getenv("GIANT_MATERIAL_WORKER_SCRIPT")); value != "" {
		return value
	}
	candidates := []string{
		filepath.Join(executableDir(), "worker", "ocr_worker.py"),
		filepath.Join("worker", "ocr_worker.py"),
		filepath.Join("giant-material-executor", "worker", "ocr_worker.py"),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return candidates[0]
}

func localDeviceName() string {
	if current, err := user.Current(); err == nil && strings.TrimSpace(current.Username) != "" {
		if runtime.GOOS == "darwin" {
			if hostname, hostErr := os.Hostname(); hostErr == nil && strings.TrimSpace(hostname) != "" {
				return strings.TrimSpace(hostname) + "/" + strings.TrimSpace(current.Username)
			}
		}
		return current.Username
	}
	if hostname, err := os.Hostname(); err == nil && strings.TrimSpace(hostname) != "" {
		return hostname
	}
	return "giant-executor"
}

func modelPreparer(supervisor *worker.Supervisor) func(context.Context, string) error {
	manifestPath := strings.TrimSpace(os.Getenv("GIANT_MATERIAL_MODEL_MANIFEST"))
	if manifestPath == "" {
		return nil
	}
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		log.Printf("OCR model manifest unavailable: %v", err)
		return func(context.Context, string) error { return err }
	}
	var manifest modelcache.Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		log.Printf("OCR model manifest invalid: %v", err)
		return func(context.Context, string) error { return err }
	}
	root := strings.TrimSpace(os.Getenv("GIANT_MATERIAL_MODEL_ROOT"))
	if root == "" {
		configRoot, configErr := os.UserConfigDir()
		if configErr != nil {
			return func(context.Context, string) error { return configErr }
		}
		root = filepath.Join(configRoot, "YizhanShengming", "GiantMaterialExecutor", "models")
	}
	return func(ctx context.Context, _ string) error {
		install, ensureErr := (modelcache.Cache{Root: root}).Ensure(ctx, manifest, func(progress modelcache.DownloadProgress) {
			log.Printf("OCR model download %d/%d (%d%%)", progress.Completed, progress.Total, progress.Percent)
		})
		if ensureErr != nil {
			return ensureErr
		}
		return supervisor.EnsureModel(ctx, install.Dir, install.Version)
	}
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func configuredPairingCode() string {
	return strings.TrimSpace(os.Getenv("GIANT_MATERIAL_EXECUTOR_PAIRING_CODE"))
}
