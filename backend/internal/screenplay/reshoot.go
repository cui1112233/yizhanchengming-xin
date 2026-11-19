package screenplay

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// 洛水 2026-10-29 版「片段重拍」：在 20 帧时间轴上框选当前镜头视频的一段，
// 抽首尾帧作为参考图交给视频模型重新生成这一小段，再把「原前段 + 新片段 + 原后段」
// 拼回一条完整视频。这里只保存重拍任务本身，不复制原视频。
const (
	ReshootStatusGenerating = "generating"
	ReshootStatusDone       = "done"
	ReshootStatusFailed     = "failed"
	ReshootStatusApplied    = "applied"

	// ReshootModeFrames 抽首帧/尾帧作为图片1、图片2，是 V88 视频通道支持的方式。
	ReshootModeFrames = "frames"

	ReshootFramePromptPrefix = "@图片1是视频的第一帧起始画面，@图片2是视频的最后一帧画面"
	ReshootMinClipSeconds    = 0.1
	ReshootMinSeconds        = 1.0
	ReshootMaxSeconds        = 30.0
	ReshootFrameCount        = 20
)

// ReshootTask 是一条片段重拍记录，随媒体状态持久化，关窗或重启后可按任务号继续查询。
type ReshootTask struct {
	ShotID     string     `json:"shot_id"`
	Seq        int        `json:"seq,omitempty"`
	Episode    int        `json:"episode,omitempty"`
	Engine     string     `json:"engine,omitempty"`
	Start      float64    `json:"start"`
	End        float64    `json:"end"`
	Duration   float64    `json:"duration"`
	Prompt     string     `json:"prompt,omitempty"`
	Mode       string     `json:"mode"`
	Status     string     `json:"status"`
	Progress   int        `json:"progress"`
	ProviderID string     `json:"provider_id,omitempty"`
	ItemID     string     `json:"item_id,omitempty"`
	Error      string     `json:"error,omitempty"`
	AppliedAt  *time.Time `json:"applied_at,omitempty"`
	AppliedID  string     `json:"applied_item_id,omitempty"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// ValidateReshootRange 校验框选范围；没有生成好的视频时无法确定片段，调用方需先判断时长。
func ValidateReshootRange(start, end float64) error {
	if math.IsNaN(start) || math.IsNaN(end) || math.IsInf(start, 0) || math.IsInf(end, 0) {
		return fmt.Errorf("%w: 片段范围无效", ErrInvalidArgument)
	}
	if start < 0 {
		return fmt.Errorf("%w: 片段起点不能小于 0", ErrInvalidArgument)
	}
	if end-start < ReshootMinClipSeconds {
		return fmt.Errorf("%w: 请先框选要重拍的片段", ErrInvalidArgument)
	}
	return nil
}

// NormalizeReshootDuration 生成时长按整数秒，1-30 秒；未填时按框选长度四舍五入。
func NormalizeReshootDuration(duration, start, end float64) float64 {
	if duration <= 0 {
		duration = math.Round(end - start)
	}
	if duration < ReshootMinSeconds {
		duration = ReshootMinSeconds
	}
	if duration > ReshootMaxSeconds {
		duration = ReshootMaxSeconds
	}
	return math.Round(duration)
}

// ReshootPrompt 在用户提示词前补上首尾帧说明，与洛水提交给模型的措辞一致。
func ReshootPrompt(prompt string) string {
	body := strings.TrimSpace(prompt)
	if body == "" {
		return ReshootFramePromptPrefix
	}
	return ReshootFramePromptPrefix + "，" + body
}

// SplicePlan 描述拼接时要保留的原视频前后段。
type SplicePlan struct {
	HasHead    bool    `json:"has_head"`
	HasTail    bool    `json:"has_tail"`
	HeadEnd    float64 `json:"head_end"`
	TailStart  float64 `json:"tail_start"`
	TailEnd    float64 `json:"tail_end"`
	NewSeconds float64 `json:"new_seconds"`
}

// PlanSplice 根据原视频总时长和框选范围决定保留哪些部分；留 0.05 秒容差避免空片段。
func PlanSplice(originalDuration, start, end float64) (SplicePlan, error) {
	if originalDuration <= 0 {
		return SplicePlan{}, fmt.Errorf("%w: 原视频时长未知，无法拼接", ErrInvalidState)
	}
	if e := ValidateReshootRange(start, end); e != nil {
		return SplicePlan{}, e
	}
	if start > originalDuration+ReshootMinClipSeconds {
		return SplicePlan{}, fmt.Errorf("%w: 片段起点超出视频时长", ErrInvalidArgument)
	}
	if end > originalDuration {
		end = originalDuration
	}
	const tolerance = 0.05
	plan := SplicePlan{HeadEnd: start, TailStart: end, TailEnd: originalDuration}
	plan.HasHead = start > tolerance
	plan.HasTail = originalDuration-end > tolerance
	return plan, nil
}

// FindReshoot 取某条分镜最近一次重拍记录。
func (s MediaState) FindReshoot(shotID string) (ReshootTask, bool) {
	if s.Reshoots == nil {
		return ReshootTask{}, false
	}
	task, ok := s.Reshoots[shotID]
	return task, ok
}

// PutReshoot 写入某条分镜的重拍记录。
func (s *MediaState) PutReshoot(task ReshootTask) {
	if s.Reshoots == nil {
		s.Reshoots = map[string]ReshootTask{}
	}
	task.UpdatedAt = time.Now().UTC()
	s.Reshoots[task.ShotID] = task
}

// MutateReshoot 就地修改某条分镜的重拍记录并写回。
func (s *MediaState) MutateReshoot(shotID string, mutate func(*ReshootTask)) bool {
	existing, ok := s.FindReshoot(shotID)
	if !ok {
		return false
	}
	mutate(&existing)
	s.PutReshoot(existing)
	return true
}
