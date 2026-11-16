package novelfetchworkshop

import (
	"fmt"
	"strings"
)

const (
	TaskStatusScheduledWaiting      = "scheduled_waiting"
	TaskStatusInputReady            = "input_ready"
	TaskStatusQueued                = "queued"
	TaskStatusProcessing            = "processing"
	TaskStatusWaitingRetry          = "waiting_retry"
	TaskStatusWaitingClassification = "waiting_classification"
	TaskStatusWaitingSensitiveAI    = "waiting_sensitive_ai"
	TaskStatusWaitingConfig         = "waiting_configuration"
	TaskStatusWaitingLogin          = "waiting_login"
	TaskStatusPendingSubmit         = "pending_submit"
	TaskStatusUploadPending         = "upload_pending"
	TaskStatusCompleted             = "completed"
	TaskStatusFailed                = "failed"
	TaskStatusStopped               = "stopped"
)

// DeriveTaskStatus is the single task-state rule shared by the Go API and UI.
//
// 任务元数据里同时存在 status/queueState/aiStatus/originalStatus/sensitiveStatus/
// siteSubmitStatus 多个阶段状态。展示给用户的有效状态只在这里推导，接口和页面
// 不得再各自拼接，否则列表、详情和旧版兼容页会出现不一致。
func DeriveTaskStatus(document Document) string {
	meta := document.Meta
	raw := statusValue(meta, "status", "taskStatus", "task_status")
	queue := statusValue(meta, "queueState", "queue_state")
	original := statusValue(meta, "originalStatus", "original_status")
	ai := statusValue(meta, "aiStatus", "ai_status")
	submit := statusValue(meta, "siteSubmitStatus", "site_submit_status")
	sensitive := statusValue(meta, "sensitiveStatus", "sensitive_status")

	// 定时挂起是用户显式选择，优先于任何历史阶段状态。
	if statusBool(meta, "scheduledHold", "scheduled_hold") || hasState(raw, TaskStatusScheduledWaiting) || hasState(queue, TaskStatusScheduledWaiting) {
		return TaskStatusScheduledWaiting
	}
	raw = mergeRawTaskStatus(raw, queue)
	switch {
	case hasState(raw, TaskStatusStopped, "cancelled", "canceled"):
		return TaskStatusStopped
	case hasState(raw, TaskStatusQueued, "pending"):
		return TaskStatusQueued
	case hasState(raw, TaskStatusProcessing, "running"):
		return TaskStatusProcessing
	case hasState(raw, TaskStatusWaitingRetry, "retrying"):
		return TaskStatusWaitingRetry
	case hasState(raw, TaskStatusWaitingClassification):
		return TaskStatusWaitingClassification
	case hasState(raw, TaskStatusWaitingSensitiveAI, "waiting_ai_service"):
		return TaskStatusWaitingSensitiveAI
	case hasState(raw, TaskStatusWaitingConfig):
		return TaskStatusWaitingConfig
	case hasState(raw, TaskStatusWaitingLogin):
		return TaskStatusWaitingLogin
	case hasState(raw, TaskStatusUploadPending):
		return TaskStatusUploadPending
	}
	if hasState(raw, "failed", "error") {
		return TaskStatusFailed
	}
	if isDoneState(raw) {
		return doneOrPending(document, original, ai, submit)
	}

	// 主状态缺失或无法识别时，才由阶段状态兜底。顺序与旧版页面保持一致：
	// 失败 > 等待类 > 执行中 > 提交完成 > 待提交 > 就绪。
	if hasState(original, "failed", "error") || hasState(ai, "failed", "error") ||
		hasState(submit, "failed", "error") || hasState(sensitive, "failed", "error") {
		return TaskStatusFailed
	}
	if hasState(original, TaskStatusWaitingConfig) || hasState(ai, TaskStatusWaitingConfig) || hasState(sensitive, TaskStatusWaitingConfig) ||
		hasState(submit, TaskStatusWaitingConfig) || hasState(ai, "waiting_ai_config") {
		return TaskStatusWaitingConfig
	}
	if hasState(original, TaskStatusWaitingLogin) || hasState(ai, TaskStatusWaitingLogin) || hasState(submit, TaskStatusWaitingLogin) {
		return TaskStatusWaitingLogin
	}
	if hasState(original, TaskStatusWaitingSensitiveAI) || hasState(ai, "waiting_ai_service", TaskStatusWaitingSensitiveAI) ||
		hasState(sensitive, TaskStatusWaitingSensitiveAI, "waiting_ai_service") {
		return TaskStatusWaitingSensitiveAI
	}
	if hasState(original, TaskStatusWaitingRetry, "waiting_fetch_retry") || hasState(ai, TaskStatusWaitingRetry, "waiting_fetch_retry") ||
		hasState(sensitive, TaskStatusWaitingRetry) {
		return TaskStatusWaitingRetry
	}
	if hasState(original, TaskStatusWaitingClassification) || hasState(ai, TaskStatusWaitingClassification) {
		return TaskStatusWaitingClassification
	}
	if hasState(original, "running", "processing") || hasState(ai, "running", "processing") || hasState(sensitive, "running", "processing") {
		return TaskStatusProcessing
	}
	if isSubmitDone(submit) {
		return TaskStatusCompleted
	}
	if originalDone(document, original) && aiContentReady(document) && !isSubmitDone(submit) {
		return TaskStatusPendingSubmit
	}
	return TaskStatusInputReady
}

// mergeRawTaskStatus 让真正在执行的队列状态覆盖过期的终态，避免重跑时仍显示“完成”。
func mergeRawTaskStatus(raw, queue string) string {
	if queue == "" {
		return raw
	}
	if hasState(queue, TaskStatusStopped, "cancelled", "canceled") {
		return queue
	}
	if hasState(queue, TaskStatusQueued, "pending", TaskStatusProcessing, "running") {
		return queue
	}
	if raw == "" || isDoneState(raw) {
		return queue
	}
	return raw
}

func doneOrPending(document Document, original, ai, submit string) string {
	if originalDone(document, original) && aiContentReady(document) && !isSubmitDone(submit) {
		return TaskStatusPendingSubmit
	}
	return TaskStatusCompleted
}

func isDoneState(value string) bool {
	return hasState(value, "done", "success", "succeeded", "complete", "completed", "submitted", "uploaded", "restored")
}

func isSubmitDone(value string) bool {
	return hasState(value, "submitted", "done", "completed", "success", "succeeded", "uploaded")
}

func aiContentReady(document Document) bool {
	if hasState(statusValue(document.Meta, "aiStatus", "ai_status"), "ai_done", "done", "completed", "success", "succeeded", "ready") {
		return true
	}
	for version, value := range document.Versions {
		if !strings.HasPrefix(strings.ToLower(version), "ai") {
			continue
		}
		if text := strings.TrimSpace(fmt.Sprint(value)); text != "" && text != "<nil>" {
			return true
		}
	}
	for version, ref := range document.BodyRefs {
		if strings.HasPrefix(strings.ToLower(version), "ai") && strings.EqualFold(strings.TrimSpace(ref.State), "ready") {
			return true
		}
	}
	return false
}

func statusBool(meta map[string]any, keys ...string) bool {
	for _, key := range keys {
		switch value := meta[key].(type) {
		case bool:
			if value {
				return true
			}
		case string:
			if strings.EqualFold(strings.TrimSpace(value), "true") {
				return true
			}
		}
	}
	return false
}

func statusValue(meta map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := strings.ToLower(strings.TrimSpace(fmt.Sprint(meta[key]))); value != "" && value != "<nil>" {
			return value
		}
	}
	return ""
}

func hasState(value string, states ...string) bool {
	for _, state := range states {
		if value == state {
			return true
		}
	}
	return false
}

func originalDone(document Document, original string) bool {
	return document.OriginalRawChars > 0 || hasState(original, "done", "completed", "success", "succeeded", "restored")
}
