package novelcreation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

// 洛水 novels.pipeline.play_runner 的等价实现（最小完整闭环）：
// 建世界 → 开场 → 玩家行动 → 推演 → 渲染 → 同步状态。
// 分支运行线（fork_run）、回合重做与版本恢复尚未移植，调用时会明确提示，
// 不伪造世界状态。
const (
	playWorldKind          = "play_world"
	playTurnLimit          = 50
	playSceneExcerptRunes  = 1500
	playDefaultMode        = "open"
	playDefaultActionsJSON = `["观察周围","开口说话","回忆来意"]`
)

var playIDSequence atomic.Uint64

type PlayEntity struct {
	Name     string `json:"name"`
	Type     string `json:"type,omitempty"`
	Status   string `json:"status,omitempty"`
	Location string `json:"location,omitempty"`
}

type PlayRelation struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value string `json:"value,omitempty"`
	Type  string `json:"type,omitempty"`
}

type PlayTurn struct {
	Turn             int       `json:"turn"`
	Input            string    `json:"input"`
	ActionKind       string    `json:"actionKind,omitempty"`
	Outcome          string    `json:"outcome,omitempty"`
	SceneText        string    `json:"sceneText"`
	SuggestedActions []string  `json:"suggestedActions"`
	CreatedAt        time.Time `json:"createdAt"`
}

type PlayWorld struct {
	ProjectID      string         `json:"projectId"`
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	Premise        string         `json:"premise"`
	Mode           string         `json:"mode"`
	Turn           int            `json:"turn"`
	PlayerPersona  string         `json:"playerPersona"`
	WorldContract  string         `json:"worldContract"`
	VisualContract string         `json:"visualContract"`
	Location       string         `json:"location"`
	Clock          string         `json:"clock"`
	Entities       []PlayEntity   `json:"entities"`
	Relations      []PlayRelation `json:"relations"`
	Facts          []string       `json:"facts"`
	Goals          []string       `json:"goals"`
	Turns          []PlayTurn     `json:"turns"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
}

type PlayWorldInput struct {
	Title         string `json:"title"`
	Premise       string `json:"premise"`
	Mode          string `json:"mode"`
	PlayerPersona string `json:"playerPersona"`
	WorldContract string `json:"worldContract"`
}

type PlayEditPatch struct {
	Title          string `json:"title"`
	Premise        string `json:"premise"`
	PlayerPersona  string `json:"playerPersona"`
	WorldContract  string `json:"worldContract"`
	VisualContract string `json:"visualContract"`
	Note           string `json:"note"`
}

// PlayTextCompleter 由应用层注入文本补全能力，避免 play 模块持有渠道配置。
type PlayTextCompleter interface {
	CompletePlayText(ctx context.Context, owner, modelID, systemPrompt, userPrompt string, maxTokens int, temperature float64) (string, error)
}

func (s *Service) ConfigurePlayText(completer PlayTextCompleter) {
	s.playText = completer
}

type playOpeningOutput struct {
	SceneText        string   `json:"sceneText"`
	SuggestedActions []string `json:"suggestedActions"`
}

type playSimulationOutput struct {
	ActionKind      string         `json:"actionKind"`
	Intent          string         `json:"intent"`
	Outcome         string         `json:"outcome"`
	Location        string         `json:"location"`
	Clock           string         `json:"clock"`
	Blocked         bool           `json:"blocked"`
	BlockedReason   string         `json:"blockedReason"`
	EntityChanges   []PlayEntity   `json:"entityChanges"`
	RelationChanges []PlayRelation `json:"relationChanges"`
	NewFacts        []string       `json:"newFacts"`
	Goals           []string       `json:"goals"`
}

type playNarrationOutput struct {
	SceneText        string   `json:"sceneText"`
	SuggestedActions []string `json:"suggestedActions"`
}

type playStateOutput struct {
	Location  string         `json:"location"`
	Clock     string         `json:"clock"`
	Entities  []PlayEntity   `json:"entities"`
	Relations []PlayRelation `json:"relations"`
	Facts     []string       `json:"facts"`
	Goals     []string       `json:"goals"`
}

// playID 保留可读的 ASCII 前缀，并追加短后缀，避免同名世界互相覆盖。
func playID(title string) string {
	slug := strings.TrimSpace(title)
	out := []rune{}
	for _, r := range slug {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			out = append(out, r)
		case r == '-' || r == '_':
			out = append(out, r)
		}
		if len(out) >= 32 {
			break
		}
	}
	base := strings.Trim(string(out), "-_")
	if base == "" {
		base = "world"
	}
	unique := fmt.Sprintf("%d-%d", time.Now().UTC().UnixNano(), playIDSequence.Add(1))
	return base + "-" + fingerprint([]string{slug, unique})[:8]
}

// decodePlayJSON 容忍代码围栏与前后杂讯，只取对象本体。
func decodePlayJSON(raw string, target any) error {
	text := strings.TrimSpace(raw)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	if start := strings.Index(text, "{"); start > 0 {
		text = text[start:]
	}
	if end := strings.LastIndex(text, "}"); end >= 0 && end < len(text)-1 {
		text = text[:end+1]
	}
	return json.Unmarshal([]byte(strings.TrimSpace(text)), target)
}

func playStateSummary(world PlayWorld) string {
	parts := []string{fmt.Sprintf("第 %d 幕", world.Turn)}
	if world.Location != "" {
		parts = append(parts, "地点："+world.Location)
	}
	if world.Clock != "" {
		parts = append(parts, "时间："+world.Clock)
	}
	if len(world.Entities) > 0 {
		names := make([]string, 0, len(world.Entities))
		for _, entity := range world.Entities {
			if entity.Name == "" {
				continue
			}
			label := entity.Name
			if entity.Status != "" {
				label += "（" + entity.Status + "）"
			}
			names = append(names, label)
		}
		if len(names) > 0 {
			parts = append(parts, "在场实体："+strings.Join(names, "、"))
		}
	}
	if len(world.Facts) > 0 {
		facts := world.Facts
		if len(facts) > 8 {
			facts = facts[len(facts)-8:]
		}
		parts = append(parts, "已知事实："+strings.Join(facts, "；"))
	}
	if len(world.Goals) > 0 {
		parts = append(parts, "当前目标："+strings.Join(world.Goals, "；"))
	}
	return strings.Join(parts, "\n")
}

func playSceneExcerpt(world PlayWorld) string {
	if len(world.Turns) == 0 {
		return ""
	}
	scene := []rune(strings.TrimSpace(world.Turns[len(world.Turns)-1].SceneText))
	if len(scene) > playSceneExcerptRunes {
		scene = scene[len(scene)-playSceneExcerptRunes:]
	}
	return string(scene)
}

func normalizePlayActions(values []string) []string {
	out := []string{}
	for _, value := range values {
		text := strings.TrimSpace(value)
		if text != "" && len(out) < 4 {
			out = append(out, text)
		}
	}
	if len(out) == 0 {
		out = append(out, "继续观察", "采取行动", "改变策略")
	}
	return out
}

func mergePlayStrings(current, next []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(current)+len(next))
	for _, value := range append(append([]string{}, current...), next...) {
		text := strings.TrimSpace(value)
		if text == "" || seen[text] {
			continue
		}
		seen[text] = true
		out = append(out, text)
	}
	if len(out) > 30 {
		out = out[len(out)-30:]
	}
	return out
}

func mergePlayEntities(current, next []PlayEntity) []PlayEntity {
	out := append([]PlayEntity{}, current...)
	for _, entity := range next {
		name := strings.TrimSpace(entity.Name)
		if name == "" {
			continue
		}
		entity.Name = name
		replaced := false
		for index := range out {
			if out[index].Name == name {
				if entity.Type == "" {
					entity.Type = out[index].Type
				}
				out[index] = entity
				replaced = true
				break
			}
		}
		if !replaced {
			out = append(out, entity)
		}
	}
	if len(out) > 50 {
		out = out[len(out)-50:]
	}
	return out
}

func mergePlayRelations(current, next []PlayRelation) []PlayRelation {
	out := append([]PlayRelation{}, current...)
	for _, relation := range next {
		if relation.From == "" || relation.To == "" {
			continue
		}
		replaced := false
		for index := range out {
			if out[index].From == relation.From && out[index].To == relation.To {
				out[index] = relation
				replaced = true
				break
			}
		}
		if !replaced {
			out = append(out, relation)
		}
	}
	if len(out) > 50 {
		out = out[len(out)-50:]
	}
	return out
}

func truncatePlayText(value string, limit int) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= limit {
		return string(runes)
	}
	return string(runes[:limit])
}
