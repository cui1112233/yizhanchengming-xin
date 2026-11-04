package httpapi

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
)

type Capability struct {
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}

// capabilityGate 是“路由已注册”之外仍需要的运行态开关（生产/合并/发布是否
// 配置好）。路由是否注册由 ServeMux 真实探测决定，不再用 slice 数值硬算。
type capabilityGate int

const (
	capabilityGateNone capabilityGate = iota
	capabilityGateProduction
	capabilityGateMerge
	capabilityGatePublish121
	capabilityGatePublishYadi
)

// capabilitySpec 是能力真值表的单行。ProbePath 用真实可匹配的具体路径，
// 启动后由 routeRegistered 探测 v11 主 mux 是否真的注册了该路由。
type capabilitySpec struct {
	Key       string
	MinSlice  int
	Method    string
	ProbePath string
	Reason    string
	Gate      capabilityGate
}

var capabilitySpecs = []capabilitySpec{
	{Key: "batch.read", MinSlice: 1, Method: http.MethodGet, ProbePath: "/api/batch-factory/v11/batches", Reason: "V11 settings slice not released"},
	{Key: "batch.create", MinSlice: 1, Method: http.MethodPost, ProbePath: "/api/batch-factory/v11/batches", Reason: "V11 settings slice not released"},
	{Key: "settings.edit", MinSlice: 1, Method: http.MethodPut, ProbePath: "/api/batch-factory/v11/batches/probe-batch/settings", Reason: "V11 settings slice not released"},
	// Batch detail carries the effective settings snapshot; use it as the
	// routable probe for the read-only snapshot capability.
	{Key: "snapshot.read", MinSlice: 1, Method: http.MethodGet, ProbePath: "/api/batch-factory/v11/batches/probe-batch", Reason: "V11 settings slice not released"},
	{Key: "override.edit", MinSlice: 1, Method: http.MethodPut, ProbePath: "/api/batch-factory/v11/batches/probe-batch/books/probe-book/override", Reason: "V11 settings slice not released"},
	{Key: "director.run", MinSlice: 2, Method: http.MethodPost, ProbePath: "/api/batch-factory/v11/batches/probe-batch/director", Reason: "Director slice not released"},
	{Key: "hook.review", MinSlice: 2, Method: http.MethodPost, ProbePath: "/api/batch-factory/v11/batches/probe-batch/books/probe-book/hook", Reason: "Director slice not released"},
	{Key: "working-front.viral", MinSlice: 2, Method: http.MethodPost, ProbePath: "/api/batch-factory/v11/batches/probe-batch/books/probe-book/working-front/viral", Reason: "Director slice not released"},
	{Key: "compiler.preview", MinSlice: 3, Method: http.MethodGet, ProbePath: "/api/batch-factory/v11/batches/probe-batch/books/probe-book/videos/probe-video/final-prompt", Reason: "Final prompt compiler slice not released"},
	{Key: "production.submit", MinSlice: 4, Method: http.MethodPost, ProbePath: "/api/batch-factory/v11/batches/probe-batch/production", Reason: "Production slice not released", Gate: capabilityGateProduction},
	{Key: "production.cancel", MinSlice: 4, Method: http.MethodPost, ProbePath: "/api/batch-factory/v11/batches/probe-batch/production/cancel", Reason: "Production slice not released", Gate: capabilityGateProduction},
	{Key: "merge.run", MinSlice: 5, Method: http.MethodPost, ProbePath: "/api/batch-factory/v11/batches/probe-batch/merge", Reason: "Merge slice not released", Gate: capabilityGateMerge},
	{Key: "publish.121", MinSlice: 5, Method: http.MethodPost, ProbePath: "/api/batch-factory/v11/batches/probe-batch/books/probe-book/publish-121", Reason: "121 视频发布未启用：尚未验证 MP4 提交接口、回执与回读", Gate: capabilityGatePublish121},
	{Key: "publish.yadi", MinSlice: 6, Method: http.MethodPost, ProbePath: "/api/batch-factory/v11/publish/yadi/intents", Reason: "Yadi is not enabled", Gate: capabilityGatePublishYadi},
}

type capabilityGates struct {
	ProductionEnabled bool
	MergeEnabled      bool
	Publish121        bool
	PublishYadi       bool
}

// deriveCapabilities 先看路由是否真实注册，再看运行态开关。任何声明
// available 的能力都必须能通过 ServeMux 匹配到真实 handler。
func deriveCapabilities(slice int, gates capabilityGates, muxes ...*http.ServeMux) map[string]Capability {
	caps := make(map[string]Capability, len(capabilitySpecs))
	for _, spec := range capabilitySpecs {
		capability := Capability{Reason: spec.Reason}
		if slice >= spec.MinSlice && gateAllows(spec.Gate, gates) && routeRegisteredAny(spec.Method, spec.ProbePath, muxes...) {
			capability = Capability{Available: true}
		}
		caps[spec.Key] = capability
	}
	return caps
}

func gateAllows(gate capabilityGate, gates capabilityGates) bool {
	switch gate {
	case capabilityGateProduction:
		return gates.ProductionEnabled
	case capabilityGateMerge:
		return gates.MergeEnabled
	case capabilityGatePublish121:
		return gates.Publish121
	case capabilityGatePublishYadi:
		return gates.PublishYadi
	default:
		return true
	}
}

func routeRegisteredAny(method, path string, muxes ...*http.ServeMux) bool {
	for _, mux := range muxes {
		if routeRegistered(mux, method, path) {
			return true
		}
	}
	return false
}

func routeRegistered(mux *http.ServeMux, method, path string) bool {
	if mux == nil {
		return false
	}
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		return false
	}
	_, pattern := mux.Handler(req)
	return pattern != ""
}

func capabilityHandler(caps map[string]Capability) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, caps)
	}
}

// capabilityBootSummary 返回已开启数量与按字母序排列的未开启能力，供启动
// 日志使用，避免“前端显示未开启但不知道到底注册了什么”。
func capabilityBootSummary(caps map[string]Capability) (int, []string) {
	enabled := 0
	disabled := make([]string, 0, len(caps))
	for key, capability := range caps {
		if capability.Available {
			enabled++
			continue
		}
		disabled = append(disabled, key)
	}
	sort.Strings(disabled)
	return enabled, disabled
}

// capabilityNotFoundHandler 只在“当前 mux 没有匹配到任何注册路由”且响应码为
// 404 时把裸文本改写成 JSON 错误体；已注册路由自己返回的 404（例如资源不
// 存在）、405 及其它状态全部原样透传。这样不会把真实的“找不到资源”伪装成
// “能力未开启”。
func capabilityNotFoundHandler(mux *http.ServeMux) http.Handler {
	if mux == nil {
		return http.NotFoundHandler()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r == nil || r.URL == nil {
			mux.ServeHTTP(w, r)
			return
		}
		if _, pattern := mux.Handler(r); pattern != "" {
			mux.ServeHTTP(w, r)
			return
		}
		writer := &capabilityNotFoundWriter{ResponseWriter: w, path: r.URL.Path}
		mux.ServeHTTP(writer, r)
	})
}

type capabilityNotFoundWriter struct {
	http.ResponseWriter
	path        string
	wroteHeader bool
	discard     bool
}

func (w *capabilityNotFoundWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	if status != http.StatusNotFound {
		w.ResponseWriter.WriteHeader(status)
		return
	}
	feature, requiredSlice := capabilityFeatureForPath(w.path)
	body, err := json.Marshal(map[string]any{
		"error":         "该能力未开启",
		"code":          ErrorCodeCapabilityNotEnabled,
		"feature":       feature,
		"requiredSlice": requiredSlice,
	})
	if err != nil {
		w.ResponseWriter.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Del("Content-Length")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.ResponseWriter.WriteHeader(http.StatusNotFound)
	_, _ = w.ResponseWriter.Write(body)
	w.discard = true
}

func (w *capabilityNotFoundWriter) Write(payload []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	if w.discard {
		return len(payload), nil
	}
	return w.ResponseWriter.Write(payload)
}

// capabilityFeatureForPath 用路径特征给出最接近的能力名与所需 slice；无法
// 归类时按 batch-factory 基础能力处理，保证错误体字段永远存在。
func capabilityFeatureForPath(path string) (string, int) {
	switch {
	case strings.Contains(path, "/publish-121"):
		return "publish.121", 5
	case strings.Contains(path, "/merge"):
		return "merge.run", 5
	case strings.Contains(path, "/production"):
		return "production.cancel", 4
	case strings.Contains(path, "/publish/"):
		return "publish.yadi", 6
	case strings.Contains(path, "/director"), strings.Contains(path, "/hook"), strings.Contains(path, "/working-front/"):
		return "director.run", 2
	case strings.Contains(path, "/compiler"), strings.Contains(path, "/final-prompt"):
		return "compiler.preview", 3
	default:
		return "batch.read", 1
	}
}
