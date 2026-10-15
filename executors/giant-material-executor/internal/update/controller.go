package update

import (
	"context"
	"crypto/ed25519"
	"sync"
	"time"
)

type State string

const (
	StateUnknown   State = "unknown"
	StateChecking  State = "checking"
	StateAvailable State = "available"
	StateCurrent   State = "current"
	StateBusy      State = "busy"
	StateStaging   State = "staging"
	StateError     State = "error"
)

// Status is safe to expose through the executor's loopback UI. It never
// includes artifact URLs, signatures, or executor credentials.
type Status struct {
	State            State     `json:"state"`
	CurrentVersion   string    `json:"currentVersion"`
	AvailableVersion string    `json:"availableVersion,omitempty"`
	Message          string    `json:"message,omitempty"`
	CheckedAt        time.Time `json:"checkedAt,omitempty"`
	CanApply         bool      `json:"canApply"`
}

type ControllerConfig struct {
	CurrentVersion string
	Target         ReleaseTarget
	PublicKey      ed25519.PublicKey
	Idle           func() bool
	Fetch          func(context.Context) (ReleaseManifest, error)
	Stage          func(context.Context, ReleaseManifest) error
}

// Controller separates background availability checks from user-authorized
// installation. Check never downloads, stages, restarts, or exits.
type Controller struct {
	config ControllerConfig
	mu     sync.Mutex
	status Status
}

func NewController(config ControllerConfig) *Controller {
	return &Controller{config: config, status: Status{State: StateUnknown, CurrentVersion: config.CurrentVersion}}
}

func (c *Controller) Status() Status {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status
}

func (c *Controller) Check(ctx context.Context) (Status, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.checkLocked(ctx)
}

func (c *Controller) Apply(ctx context.Context) (Status, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.config.Idle != nil && !c.config.Idle() {
		return c.setLocked(Status{State: StateBusy, CurrentVersion: c.config.CurrentVersion, Message: "执行器正在运行任务，任务结束后可更新"}), &CodedError{Code: "UPDATE_EXECUTOR_BUSY"}
	}
	status, err := c.checkLocked(ctx)
	if err != nil {
		return status, err
	}
	if !status.CanApply {
		return status, &CodedError{Code: "UPDATE_NOT_AVAILABLE"}
	}
	if c.config.Idle != nil && !c.config.Idle() {
		return c.setLocked(Status{State: StateBusy, CurrentVersion: c.config.CurrentVersion, AvailableVersion: status.AvailableVersion, Message: "执行器正在运行任务，任务结束后可更新", CheckedAt: status.CheckedAt}), &CodedError{Code: "UPDATE_EXECUTOR_BUSY"}
	}
	if c.config.Stage == nil {
		return c.setLocked(Status{State: StateError, CurrentVersion: c.config.CurrentVersion, AvailableVersion: status.AvailableVersion, Message: "更新器未配置"}), &CodedError{Code: "UPDATE_STAGE_FAILED"}
	}
	c.setLocked(Status{State: StateStaging, CurrentVersion: c.config.CurrentVersion, AvailableVersion: status.AvailableVersion, Message: "正在下载并校验更新", CheckedAt: status.CheckedAt})
	manifest, fetchErr := c.config.Fetch(ctx)
	if fetchErr != nil {
		return c.setLocked(Status{State: StateError, CurrentVersion: c.config.CurrentVersion, AvailableVersion: status.AvailableVersion, Message: fetchErr.Error(), CheckedAt: status.CheckedAt}), fetchErr
	}
	verified, verifyErr := VerifyReleaseManifest(manifest, c.config.PublicKey, c.config.Target, c.config.CurrentVersion)
	if verifyErr != nil {
		return c.setLocked(Status{State: StateError, CurrentVersion: c.config.CurrentVersion, AvailableVersion: status.AvailableVersion, Message: verifyErr.Error(), CheckedAt: status.CheckedAt}), verifyErr
	}
	if err := c.config.Stage(ctx, verified); err != nil {
		return c.setLocked(Status{State: StateError, CurrentVersion: c.config.CurrentVersion, AvailableVersion: verified.Version, Message: err.Error(), CheckedAt: status.CheckedAt}), err
	}
	return c.setLocked(Status{State: StateStaging, CurrentVersion: c.config.CurrentVersion, AvailableVersion: verified.Version, Message: "更新已就绪，正在重启", CheckedAt: status.CheckedAt}), nil
}

func (c *Controller) checkLocked(ctx context.Context) (Status, error) {
	if c.config.Fetch == nil {
		return c.setLocked(Status{State: StateError, CurrentVersion: c.config.CurrentVersion, Message: "更新清单未配置"}), &CodedError{Code: "UPDATE_MANIFEST_INVALID"}
	}
	c.setLocked(Status{State: StateChecking, CurrentVersion: c.config.CurrentVersion, Message: "正在检查更新"})
	manifest, err := c.config.Fetch(ctx)
	if err != nil {
		return c.setLocked(Status{State: StateError, CurrentVersion: c.config.CurrentVersion, Message: err.Error(), CheckedAt: time.Now().UTC()}), err
	}
	verified, err := VerifyReleaseManifest(manifest, c.config.PublicKey, c.config.Target, c.config.CurrentVersion)
	if CodeOf(err) == "UPDATE_NOT_AVAILABLE" {
		return c.setLocked(Status{State: StateCurrent, CurrentVersion: c.config.CurrentVersion, Message: "已是最新版本", CheckedAt: time.Now().UTC()}), nil
	}
	if err != nil {
		return c.setLocked(Status{State: StateError, CurrentVersion: c.config.CurrentVersion, Message: err.Error(), CheckedAt: time.Now().UTC()}), err
	}
	return c.setLocked(Status{State: StateAvailable, CurrentVersion: c.config.CurrentVersion, AvailableVersion: verified.Version, Message: "发现可用更新", CheckedAt: time.Now().UTC(), CanApply: c.config.Idle == nil || c.config.Idle()}), nil
}

func (c *Controller) setLocked(status Status) Status {
	c.status = status
	return status
}
