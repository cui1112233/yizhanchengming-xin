package httpapi

import (
	"context"
	"net/http"

	"qiantie/backend/internal/giantmaterialexecutor"
	"qiantie/backend/internal/localexecutor"
)

// 执行器状态的统一只读口径：在线 / 离线 / 需配对 / 未安装 + 设备列表。
// 这里只做读取层归一，不改变 local_executor_* 与 giant_executor_* 的心跳、
// 配对协议或表结构。
const (
	ExecutorStateOnline       = "online"
	ExecutorStateOffline      = "offline"
	ExecutorStateNeedsPairing = "needs_pairing"
	ExecutorStateNotInstalled = "not_installed"
)

type executorStatusDevice struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Platform   string `json:"platform,omitempty"`
	OS         string `json:"os,omitempty"`
	Version    string `json:"version,omitempty"`
	Online     bool   `json:"online"`
	LastSeenAt any    `json:"lastSeenAt,omitempty"`
}

type executorStatusGroup struct {
	State   string                 `json:"state"`
	Devices []executorStatusDevice `json:"devices"`
	Error   string                 `json:"error,omitempty"`
}

type executorStatusResponse struct {
	Local   executorStatusGroup    `json:"local"`
	Giant   executorStatusGroup    `json:"giant"`
	Devices []executorStatusDevice `json:"devices"`
}

// executorStateFor 把设备列表归一为四种状态之一。前端若刚生成配对码且还没有
// 已配对设备，会在此基础上叠加“需配对”的本地状态（后端没有待配对列表）。
func executorStateFor(devices []executorStatusDevice) string {
	if len(devices) == 0 {
		return ExecutorStateNotInstalled
	}
	for _, device := range devices {
		if device.Online {
			return ExecutorStateOnline
		}
	}
	return ExecutorStateOffline
}

func localExecutorStatusDevices(views []localexecutor.ExecutorView) []executorStatusDevice {
	devices := make([]executorStatusDevice, 0, len(views))
	for _, view := range views {
		devices = append(devices, executorStatusDevice{
			ID: view.ID, Kind: "local", Name: view.Name, Platform: view.Platform,
			OS: view.OS, Version: view.Version, Online: view.Online, LastSeenAt: view.LastSeenAt,
		})
	}
	return devices
}

func giantExecutorStatusDevices(views []giantmaterialexecutor.ExecutorView) []executorStatusDevice {
	devices := make([]executorStatusDevice, 0, len(views))
	for _, view := range views {
		devices = append(devices, executorStatusDevice{
			ID: view.ID, Kind: "giant", Name: view.Name, Platform: view.Platform,
			OS: view.OS, Version: view.Version, Online: view.Online, LastSeenAt: view.LastSeenAt,
		})
	}
	return devices
}

// collectExecutorStatus 聚合两种执行器；任一服务读取失败时只标记该组，不影响
// 另一组返回。
func collectExecutorStatus(ctx context.Context, owner string, local *localexecutor.Service, giant *giantmaterialexecutor.Service) executorStatusResponse {
	localGroup := executorStatusGroup{State: ExecutorStateNotInstalled, Devices: []executorStatusDevice{}}
	if local != nil {
		if views, err := local.List(ctx, owner); err != nil {
			localGroup.State = ExecutorStateOffline
			localGroup.Error = err.Error()
		} else {
			localGroup.Devices = localExecutorStatusDevices(views)
			localGroup.State = executorStateFor(localGroup.Devices)
		}
	}
	giantGroup := executorStatusGroup{State: ExecutorStateNotInstalled, Devices: []executorStatusDevice{}}
	if giant != nil {
		if views, err := giant.ListExecutors(ctx, owner); err != nil {
			giantGroup.State = ExecutorStateOffline
			giantGroup.Error = err.Error()
		} else {
			giantGroup.Devices = giantExecutorStatusDevices(views)
			giantGroup.State = executorStateFor(giantGroup.Devices)
		}
	}
	devices := make([]executorStatusDevice, 0, len(localGroup.Devices)+len(giantGroup.Devices))
	devices = append(devices, localGroup.Devices...)
	devices = append(devices, giantGroup.Devices...)
	return executorStatusResponse{Local: localGroup, Giant: giantGroup, Devices: devices}
}

func registerExecutorStatusRoute(root *http.ServeMux, auth BridgeAuth, local *localexecutor.Service, giant *giantmaterialexecutor.Service) {
	if root == nil {
		return
	}
	root.Handle("GET /api/shuihuo-production/executors/status", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := BridgeIdentityFromContext(r.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, collectExecutorStatus(r.Context(), identity.Username, local, giant))
	})))
}
