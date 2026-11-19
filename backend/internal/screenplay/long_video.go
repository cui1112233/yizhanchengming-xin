package screenplay

import (
	"fmt"
	"strings"
	"time"
)

// 洛水 2026-10-29 版「超长视频」：按推荐时长把分镜累积分组，每组不超过单次生成上限，
// 组内合并角色/场景/道具与视频提示词，本质仍是「一个分镜 + 一个提示词」交给视频模型。
const (
	LongVideoSectionPrefix         = "超长视频分镜*"
	DefaultLongVideoTargetSeconds  = 300.0
	DefaultLongVideoCapSeconds     = 30.0
	LongVideoStatusIdle            = "idle"
	LongVideoStatusGenerating      = "generating"
	LongVideoStatusDone            = "done"
	LongVideoStatusFailed          = "failed"
	LongVideoFinalStatusIdle       = "idle"
	LongVideoFinalStatusCombining  = "combining"
	LongVideoFinalStatusDone       = "done"
	LongVideoFinalStatusFailed     = "failed"
	longVideoTargetTolerance       = 1e-6
	longVideoDurationNameSeparator = "、"
)

// LongVideoSection 是一组「超长视频分镜*N」，对应一次视频生成。
type LongVideoSection struct {
	ID           string   `json:"id"`
	Index        int      `json:"index"`
	Label        string   `json:"label"`
	ShotIDs      []string `json:"shot_ids"`
	ShotSeqs     []int    `json:"shot_seqs"`
	TotalSeconds float64  `json:"total_seconds"`
	Prompt       string   `json:"prompt"`
	Characters   string   `json:"characters,omitempty"`
	Scene        string   `json:"scene,omitempty"`
	Props        string   `json:"props,omitempty"`
	Status       string   `json:"status"`
	Progress     int      `json:"progress"`
	ProviderID   string   `json:"provider_id,omitempty"`
	ItemID       string   `json:"item_id,omitempty"`
	Error        string   `json:"error,omitempty"`
}

// LongVideoBatch 是某一集的超长视频编排与结果，随媒体状态一起持久化，
// 因此关闭窗口或重启服务后仍能按已保存的任务编号继续。
type LongVideoBatch struct {
	ID            string             `json:"id"`
	Episode       int                `json:"episode"`
	TargetSeconds float64            `json:"target_seconds"`
	CapSeconds    float64            `json:"cap_seconds"`
	Sections      []LongVideoSection `json:"sections"`
	FinalStatus   string             `json:"final_status"`
	FinalItemID   string             `json:"final_item_id,omitempty"`
	FinalError    string             `json:"final_error,omitempty"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

// NormalizeLongVideoTarget 只接受 100/200/250/300，其他值回落到默认 300 秒。
func NormalizeLongVideoTarget(value float64) float64 {
	for _, allowed := range []float64{100, 200, 250, 300} {
		if value == allowed {
			return value
		}
	}
	return DefaultLongVideoTargetSeconds
}

// NormalizeLongVideoCap 单次生成上限最高 30 秒，与供应商能力一致。
func NormalizeLongVideoCap(value float64) float64 {
	if value <= 0 {
		return DefaultLongVideoCapSeconds
	}
	if value > DefaultLongVideoCapSeconds {
		return DefaultLongVideoCapSeconds
	}
	return value
}

// IsLongVideoSectionShot 判断分镜是否已属于某个超长视频分段。
func IsLongVideoSectionShot(shot StoryboardShot) bool {
	return strings.HasPrefix(strings.TrimSpace(shot.VideoDesc), LongVideoSectionPrefix)
}

// PlanLongVideoSections 生成超长视频分组。目标时长是累计上限，单组不超过 cap。
func PlanLongVideoSections(shots []StoryboardShot, targetSeconds, capSeconds float64) []LongVideoSection {
	target := NormalizeLongVideoTarget(targetSeconds)
	capacity := NormalizeLongVideoCap(capSeconds)
	sections := make([]LongVideoSection, 0, 8)
	var current *LongVideoSection
	var committed float64

	flush := func() {
		if current == nil || len(current.ShotIDs) == 0 {
			current = nil
			return
		}
		current.Index = len(sections) + 1
		current.Label = fmt.Sprintf("%s%d", LongVideoSectionPrefix, current.Index)
		current.ID = fmt.Sprintf("lv_%d_%d", current.Index, len(current.ShotIDs))
		if current.Status == "" {
			current.Status = LongVideoStatusIdle
		}
		sections = append(sections, *current)
		committed += current.TotalSeconds
		current = nil
	}

	for _, shot := range shots {
		desc := strings.TrimSpace(shot.VideoDesc)
		if desc == "" || IsLongVideoSectionShot(shot) {
			continue
		}
		seconds := shot.Duration
		if seconds < 0 {
			seconds = 0
		}
		if seconds > capacity {
			seconds = capacity
		}
		if committed+seconds > target+longVideoTargetTolerance {
			break
		}
		if current != nil && current.TotalSeconds+seconds > capacity+longVideoTargetTolerance {
			flush()
			if committed+seconds > target+longVideoTargetTolerance {
				break
			}
		}
		if current == nil {
			current = &LongVideoSection{Status: LongVideoStatusIdle}
		}
		current.ShotIDs = append(current.ShotIDs, shot.ID)
		current.ShotSeqs = append(current.ShotSeqs, shot.Seq)
		current.TotalSeconds += seconds
		if current.Prompt != "" {
			current.Prompt += "\n"
		}
		current.Prompt += desc
		current.Characters = mergeLongVideoNames(current.Characters, shot.Characters)
		current.Scene = mergeLongVideoNames(current.Scene, shot.Scene)
		current.Props = mergeLongVideoNames(current.Props, shot.Props)
	}
	flush()
	return sections
}

// mergeLongVideoNames 把多镜头的人物/场景/道具合成一份去重名单。
func mergeLongVideoNames(existing, incoming string) string {
	names := make([]string, 0, 8)
	seen := map[string]bool{}
	add := func(raw string) {
		for _, piece := range strings.FieldsFunc(raw, func(r rune) bool {
			switch r {
			case '、', '，', ',', '；', ';', '|', '/':
				return true
			}
			return false
		}) {
			name := strings.TrimSpace(piece)
			if name == "" || seen[name] {
				continue
			}
			seen[name] = true
			names = append(names, name)
		}
	}
	add(existing)
	add(incoming)
	return strings.Join(names, longVideoDurationNameSeparator)
}

// LongVideoPlannedSeconds 返回分组累计时长。
func LongVideoPlannedSeconds(sections []LongVideoSection) float64 {
	total := 0.0
	for _, section := range sections {
		total += section.TotalSeconds
	}
	return total
}

// FindLongVideoBatch 取某一集已保存的编排。
func (s MediaState) FindLongVideoBatch(episode int) (LongVideoBatch, bool) {
	if s.LongVideos == nil {
		return LongVideoBatch{}, false
	}
	batch, ok := s.LongVideos[fmt.Sprint(episode)]
	if !ok {
		return LongVideoBatch{}, false
	}
	return batch, true
}

// PutLongVideoBatch 写入某一集的编排。
func (s *MediaState) PutLongVideoBatch(batch LongVideoBatch) {
	if s.LongVideos == nil {
		s.LongVideos = map[string]LongVideoBatch{}
	}
	batch.UpdatedAt = time.Now().UTC()
	s.LongVideos[fmt.Sprint(batch.Episode)] = batch
}

// MutateLongVideoSection 就地修改某一组并写回。
func (s *MediaState) MutateLongVideoSection(episode, index int, mutate func(*LongVideoSection)) bool {
	batch, ok := s.FindLongVideoBatch(episode)
	if !ok {
		return false
	}
	for i := range batch.Sections {
		if batch.Sections[i].Index == index {
			mutate(&batch.Sections[i])
			s.PutLongVideoBatch(batch)
			return true
		}
	}
	return false
}
