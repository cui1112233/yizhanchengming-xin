package screenplay

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// StepDefinition 描述批次 9 步骨架（规格 §13.2）。
type StepDefinition struct {
	Key              string `json:"key"`
	Label            string `json:"label"`
	Episodic         bool   `json:"episodic"`
	Doctor           bool   `json:"doctor"`
	Host             bool   `json:"host"`              // 由宿主（V88 服务）完成，不调模型
	RequiresExternal bool   `json:"requires_external"` // 依赖外部视频能力
	PromptID         string `json:"prompt_id"`
}

// StepDefinitions 必须与规格 §13.2 表逐项一致，顺序即执行顺序。
var StepDefinitions = []StepDefinition{
	{Key: "0_ingest", Label: "整理原文", Host: true, PromptID: "batch_step_0"},
	{Key: "1_screenplay", Label: "生成剧本", PromptID: "batch_step_1"},
	{Key: "2_screenplay_doctor", Label: "剧本审核", Doctor: true, PromptID: "batch_step_2"},
	{Key: "3_assets", Label: "资产提取", PromptID: "batch_step_3"},
	{Key: "4_preset_doctor", Label: "预设质检", Doctor: true, PromptID: "batch_step_4"},
	{Key: "5_storyboard", Label: "按集分镜", Episodic: true, PromptID: "batch_step_5"},
	{Key: "6_storyboard_doctor", Label: "分镜审核", Episodic: true, Doctor: true, PromptID: "batch_step_6"},
	{Key: "7_video", Label: "批量出视频", Episodic: true, RequiresExternal: true, PromptID: "batch_step_7"},
	{Key: "8_finalize", Label: "收尾编号", Host: true, PromptID: "batch_step_8"},
}

// StepByKey 返回步骤定义。
func StepByKey(key string) (StepDefinition, bool) {
	for _, def := range StepDefinitions {
		if def.Key == key {
			return def, true
		}
	}
	return StepDefinition{}, false
}

// StepIndex 返回步骤序号（0-8）；未知步骤返回 -1。
func StepIndex(key string) int {
	for i, def := range StepDefinitions {
		if def.Key == key {
			return i
		}
	}
	return -1
}

// IsDoctorStep 判断是否为医生步（2/4/6）。
func IsDoctorStep(key string) bool {
	def, ok := StepByKey(key)
	return ok && def.Doctor
}

// IsEpisodicStep 判断是否按集执行（5/6/7）。
func IsEpisodicStep(key string) bool {
	def, ok := StepByKey(key)
	return ok && def.Episodic
}

// NewBatch 创建一个全新的批次状态：running + 9 步 pending。
func NewBatch(id, projectID, owner, sessionID string, mode PermissionMode) *Batch {
	steps := make(map[string]*StepState, len(StepDefinitions))
	for _, def := range StepDefinitions {
		steps[def.Key] = &StepState{Key: def.Key, Status: StepPending}
	}
	return &Batch{
		ID:             id,
		ProjectID:      projectID,
		OwnerUsername:  owner,
		SessionID:      sessionID,
		Status:         BatchRunning,
		CurrentStep:    StepDefinitions[0].Key,
		PlanVersion:    1,
		PermissionMode: mode,
		Steps:          steps,
	}
}

// Unit 是 run_next 的一次执行单元：某一步，或某一步的某一集。
type Unit struct {
	StepKey string `json:"step_key"`
	Episode int    `json:"episode"`
	Label   string `json:"label"`
}

// IsEpisodeUnit 表示该单元是按集步骤中的一集。
func (u Unit) IsEpisodeUnit() bool { return u.Episode > 0 }

func (u Unit) String() string {
	if u.Episode > 0 {
		return fmt.Sprintf("%s 第 %d 集", u.Label, u.Episode)
	}
	return u.Label
}

// GateCleared 判断医生闸门是否已解除：审核通过，或该步已不处于 blocked。
func GateCleared(step *StepState) bool {
	if step == nil {
		return true
	}
	if step.Status != StepBlocked {
		return true
	}
	return step.Review != nil && step.Review.Pass
}

// BlockedSteps 返回仍处于 blocked 且审核未通过的步骤键（按步序）。
func BlockedSteps(b *Batch) []string {
	if b == nil {
		return nil
	}
	var blocked []string
	for _, def := range StepDefinitions {
		step := b.Step(def.Key)
		if step != nil && step.Status == StepBlocked && !GateCleared(step) {
			blocked = append(blocked, def.Key)
		}
	}
	return blocked
}

// NextUnit 计算下一个执行单元。已完成/已跳过/已 stale 的步骤按下列规则处理：
//   - done / skipped → 跳过；
//   - 按集步骤 → 逐集返回，全部完成后视为整步完成；
//   - blocked 且闸门未清 → 返回错误，要求先 rewrite 或更新 review JSON；
//   - pending / in_progress / failed / stale → 作为可重跑单元返回。
func NextUnit(b *Batch) (Unit, error) {
	if b == nil {
		return Unit{}, fmt.Errorf("%w: 批次为空", ErrInvalidArgument)
	}
	status := NormalizeBatchStatus(b.Status)
	switch status {
	case BatchCompleted:
		return Unit{}, fmt.Errorf("%w: 批次已完成", ErrInvalidState)
	case BatchWaitingUser:
		blocked := BlockedSteps(b)
		if len(blocked) > 0 {
			return Unit{}, fmt.Errorf("%w: 审核仍未通过，请先 rewrite 或更新 review JSON", ErrInvalidState)
		}
	}

	for _, def := range StepDefinitions {
		step := b.Step(def.Key)
		if step == nil {
			step = &StepState{Key: def.Key, Status: StepPending}
		}
		switch step.Status {
		case StepDone, StepSkipped:
			continue
		case StepBlocked:
			if !GateCleared(step) {
				return Unit{}, fmt.Errorf("%w: 审核仍未通过，请先 rewrite 或更新 review JSON", ErrInvalidState)
			}
			// 审核已改为通过：重新执行该步以落 done。
			return Unit{StepKey: def.Key, Episode: 0, Label: def.Label}, nil
		}

		if def.Episodic {
			if episode := nextEpisode(step); episode > 0 {
				return Unit{StepKey: def.Key, Episode: episode, Label: def.Label}, nil
			}
			// 集号未知（TotalEpisodes 尚未规划）时，整步作为单次单元执行。
			if !episodeStepComplete(step) {
				return Unit{StepKey: def.Key, Episode: 0, Label: def.Label}, nil
			}
			continue // 该步所有集已完成
		}
		return Unit{StepKey: def.Key, Episode: 0, Label: def.Label}, nil
	}
	return Unit{}, fmt.Errorf("%w: 没有可执行单元", ErrInvalidState)
}

// nextEpisode 返回该步第一个未完成且可重跑的集号；0 表示全部完成或集数未知。
func nextEpisode(step *StepState) int {
	if step == nil {
		return 0
	}
	if step.TotalEpisodes > 0 {
		for n := 1; n <= step.TotalEpisodes; n++ {
			ep := step.Episodes[n]
			if ep == nil {
				return n
			}
			switch ep.Status {
			case StepDone, StepSkipped:
				continue
			default:
				return n
			}
		}
		return 0
	}
	numbers := make([]int, 0, len(step.Episodes))
	for n := range step.Episodes {
		numbers = append(numbers, n)
	}
	sort.Ints(numbers)
	for _, n := range numbers {
		ep := step.Episodes[n]
		if ep == nil {
			return n
		}
		switch ep.Status {
		case StepDone, StepSkipped:
			continue
		default:
			return n
		}
	}
	return 0
}

// StartStep 将单元置为 in_progress，并推进 current_step / current_episode。
func StartStep(b *Batch, unit Unit) error {
	if b == nil {
		return fmt.Errorf("%w: 批次为空", ErrInvalidArgument)
	}
	def, ok := StepByKey(unit.StepKey)
	if !ok {
		return fmt.Errorf("%w: 未知步骤 %q", ErrInvalidArgument, unit.StepKey)
	}
	step := b.Step(def.Key)
	if step == nil {
		step = &StepState{Key: def.Key, Status: StepPending}
		if b.Steps == nil {
			b.Steps = map[string]*StepState{}
		}
		b.Steps[def.Key] = step
	}
	if step.Status == StepBlocked && !GateCleared(step) {
		return fmt.Errorf("%w: 审核仍未通过，请先 rewrite 或更新 review JSON", ErrInvalidState)
	}
	now := nowUTC()
	step.Status = StepInProgress
	step.StartedAt = &now
	step.Error = ""
	b.CurrentStep = def.Key
	if def.Episodic {
		b.CurrentEpisode = unit.Episode
		if step.Episodes == nil {
			step.Episodes = map[int]*EpisodeState{}
		}
		ep := step.Episodes[unit.Episode]
		if ep == nil {
			ep = &EpisodeState{EpisodeNo: unit.Episode}
			step.Episodes[unit.Episode] = ep
		}
		ep.Status = StepInProgress
		ep.UpdatedAt = now
	} else {
		b.CurrentEpisode = 0
	}
	b.Status = BatchRunning
	b.LastRunAt = &now
	b.UpdatedAt = now
	return nil
}

// CompleteStep 标记单元成功，并按需推进整步状态。
func CompleteStep(b *Batch, unit Unit, artifacts []string) {
	if b == nil {
		return
	}
	step := b.Step(unit.StepKey)
	if step == nil {
		return
	}
	now := nowUTC()
	if IsEpisodicStep(unit.StepKey) && unit.Episode > 0 {
		if step.Episodes == nil {
			step.Episodes = map[int]*EpisodeState{}
		}
		ep := step.Episodes[unit.Episode]
		if ep == nil {
			ep = &EpisodeState{EpisodeNo: unit.Episode}
			step.Episodes[unit.Episode] = ep
		}
		ep.Status = StepDone
		ep.UpdatedAt = now
		ep.Artifact = firstArtifact(artifacts)
		if episodeStepDone(step) {
			step.Status = StepDone
			step.FinishedAt = &now
		}
	} else {
		step.Status = StepDone
		step.FinishedAt = &now
	}
	// 重试同一个单元时，服务端可能再次返回已经保存过的产物。
	// 这里去重，避免导出包和生产页出现重复视频地址。
	seenArtifacts := make(map[string]struct{}, len(step.Artifacts)+len(artifacts))
	for _, artifact := range step.Artifacts {
		seenArtifacts[artifact] = struct{}{}
	}
	for _, artifact := range artifacts {
		if artifact == "" {
			continue
		}
		if _, exists := seenArtifacts[artifact]; exists {
			continue
		}
		step.Artifacts = append(step.Artifacts, artifact)
		seenArtifacts[artifact] = struct{}{}
	}
	step.Error = ""
	b.Status = aggregateStatus(b)
	b.CurrentEpisode = 0
	b.UpdatedAt = now
}

// FailStep 标记单元失败。失败不会自动进入 waiting_user（规格 §9：提取失败仅记录）。
func FailStep(b *Batch, unit Unit, message string) {
	if b == nil {
		return
	}
	step := b.Step(unit.StepKey)
	if step == nil {
		return
	}
	now := nowUTC()
	if IsEpisodicStep(unit.StepKey) && unit.Episode > 0 {
		if step.Episodes == nil {
			step.Episodes = map[int]*EpisodeState{}
		}
		ep := step.Episodes[unit.Episode]
		if ep == nil {
			ep = &EpisodeState{EpisodeNo: unit.Episode}
			step.Episodes[unit.Episode] = ep
		}
		ep.Status = StepFailed
		ep.Reason = message
		ep.UpdatedAt = now
	}
	step.Status = StepFailed
	step.Error = message
	b.LastError = message
	b.Status = aggregateStatus(b)
	b.UpdatedAt = now
}

// PauseStep 把正在执行的最小单元退回可继续状态，并把批次标为 paused。
// 已经由外部服务持久化的产物不删除，继续任务时只重跑当前未完成单元。
func PauseStep(b *Batch, unit Unit) {
	if b == nil {
		return
	}
	step := b.Step(unit.StepKey)
	if step != nil {
		step.Status = StepPending
		step.Error = ""
		step.FinishedAt = nil
		if unit.Episode > 0 && step.Episodes != nil {
			if episode := step.Episodes[unit.Episode]; episode != nil {
				episode.Status = StepPending
				episode.Reason = ""
				episode.UpdatedAt = nowUTC()
			}
		}
	}
	b.Status = BatchPaused
	b.CurrentEpisode = 0
	b.LastError = ""
	b.UpdatedAt = nowUTC()
}

// ApplyDoctorReview 落实医生闸门：只有 pass=false && block_next_step=true 才阻断。
func ApplyDoctorReview(b *Batch, unit Unit, review *Review) {
	if b == nil || review == nil {
		return
	}
	step := b.Step(unit.StepKey)
	if step == nil {
		return
	}
	now := nowUTC()
	step.Review = review
	if IsEpisodicStep(unit.StepKey) && unit.Episode > 0 {
		if step.Episodes == nil {
			step.Episodes = map[int]*EpisodeState{}
		}
		ep := step.Episodes[unit.Episode]
		if ep == nil {
			ep = &EpisodeState{EpisodeNo: unit.Episode}
			step.Episodes[unit.Episode] = ep
		}
		ep.UpdatedAt = now
		ep.Artifact = "" // 审核产物由 Artifacts 记录
		if !review.Pass && review.BlockNextStep {
			ep.Status = StepBlocked
			step.Status = StepBlocked
			b.Status = BatchWaitingUser
			b.PendingActions = append(b.PendingActions, PendingAction{
				ID:              fmt.Sprintf("%s-%d-rewrite", unit.StepKey, unit.Episode),
				Kind:            "skill",
				StepKey:         unit.StepKey,
				Title:           fmt.Sprintf("批次等待用户：「%s 第 %d 集」", unit.Label, unit.Episode),
				Detail:          "审核未通过且要求阻断；请先 rewrite 或把 review JSON 改为 pass=true",
				RequiresConfirm: true,
				CreatedAt:       now,
			})
			b.UpdatedAt = now
			return
		}
		ep.Status = StepDone
		if episodeStepDone(step) {
			step.Status = StepDone
			step.FinishedAt = &now
		}
		b.Status = aggregateStatus(b)
		b.UpdatedAt = now
		return
	}
	if !review.Pass && review.BlockNextStep {
		step.Status = StepBlocked
		b.Status = BatchWaitingUser
		b.PendingActions = append(b.PendingActions, PendingAction{
			ID:              unit.StepKey + "-rewrite",
			Kind:            "skill",
			StepKey:         unit.StepKey,
			Title:           fmt.Sprintf("批次等待用户：「%s」", unit.Label),
			Detail:          "审核未通过且要求阻断；请先 rewrite 或把 review JSON 改为 pass=true",
			RequiresConfirm: true,
			CreatedAt:       now,
		})
	} else {
		step.Status = StepDone
		step.FinishedAt = &now
	}
	b.Status = aggregateStatus(b)
	b.UpdatedAt = now
}

// MarkStale 按 rewrite manifest 把命中的步骤置 stale，并清理对应待办动作。
func MarkStale(b *Batch, stepKeys []string) int {
	if b == nil {
		return 0
	}
	hit := 0
	for _, key := range stepKeys {
		step := b.Step(key)
		if step == nil {
			continue
		}
		step.Status = StepStale
		step.FinishedAt = nil
		step.Review = nil
		hit++
	}
	if hit > 0 {
		b.Status = aggregateStatus(b)
		filtered := b.PendingActions[:0]
		for _, action := range b.PendingActions {
			if containsString(stepKeys, action.StepKey) {
				continue
			}
			filtered = append(filtered, action)
		}
		b.PendingActions = filtered
		b.UpdatedAt = nowUTC()
	}
	return hit
}

// ResumeBatch 在闸门解除后把 waiting_user 恢复为 running。
func ResumeBatch(b *Batch) error {
	if b == nil {
		return fmt.Errorf("%w: 批次为空", ErrInvalidArgument)
	}
	if NormalizeBatchStatus(b.Status) != BatchWaitingUser {
		return nil
	}
	if blocked := BlockedSteps(b); len(blocked) > 0 {
		return fmt.Errorf("%w: 审核仍未通过，请先 rewrite 或更新 review JSON", ErrInvalidState)
	}
	b.Status = BatchRunning
	b.UpdatedAt = nowUTC()
	return nil
}

// ValidateTransition 校验批次状态迁移是否合法，供存储层与测试共用。
func ValidateTransition(from, to BatchStatus) error {
	from = NormalizeBatchStatus(from)
	to = NormalizeBatchStatus(to)
	if from == to {
		return nil
	}
	allowed := map[BatchStatus][]BatchStatus{
		BatchRunning:     {BatchWaitingUser, BatchPaused, BatchCompleted, BatchFailed},
		BatchWaitingUser: {BatchRunning, BatchPaused, BatchCompleted, BatchFailed},
		BatchPaused:      {BatchRunning, BatchFailed, BatchCompleted},
		BatchFailed:      {BatchRunning},
		BatchCompleted:   {},
	}
	for _, candidate := range allowed[from] {
		if candidate == to {
			return nil
		}
	}
	return fmt.Errorf("%w: 批次状态 %s → %s 不合法", ErrInvalidState, from, to)
}

// SetBatchStatus 校验并设置批次状态。
func SetBatchStatus(b *Batch, to BatchStatus) error {
	if b == nil {
		return fmt.Errorf("%w: 批次为空", ErrInvalidArgument)
	}
	if err := ValidateTransition(b.Status, to); err != nil {
		return err
	}
	b.Status = NormalizeBatchStatus(to)
	b.UpdatedAt = nowUTC()
	return nil
}

func aggregateStatus(b *Batch) BatchStatus {
	if b == nil {
		return BatchFailed
	}
	allFinished := true
	for _, def := range StepDefinitions {
		step := b.Step(def.Key)
		if step == nil {
			allFinished = false
			continue
		}
		switch step.Status {
		case StepDone, StepSkipped:
		case StepBlocked:
			return BatchWaitingUser
		case StepFailed:
			return BatchFailed
		default:
			allFinished = false
		}
	}
	if allFinished {
		return BatchCompleted
	}
	return BatchRunning
}

func episodeStepComplete(step *StepState) bool {
	if step == nil || len(step.Episodes) == 0 {
		return false
	}
	if step.TotalEpisodes > 0 && len(step.Episodes) < step.TotalEpisodes {
		return false
	}
	for _, ep := range step.Episodes {
		if ep == nil {
			return false
		}
		if ep.Status != StepDone && ep.Status != StepSkipped {
			return false
		}
	}
	return true
}

func episodeStepDone(step *StepState) bool {
	return episodeStepComplete(step)
}

func firstArtifact(artifacts []string) string {
	if len(artifacts) == 0 {
		return ""
	}
	return artifacts[0]
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(target)) {
			return true
		}
	}
	return false
}

// nowUTC 便于测试替换。
var nowUTC = func() time.Time { return time.Now().UTC() }
