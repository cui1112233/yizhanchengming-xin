package novelcreation

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
)

// 洛水 novels.services.hook_manager 的服务端等价实现（50 册 §3.20）。
// 只使用已冻结的结算伏笔与真实章节号，不调用模型。
const (
	hookMaxActive       = 15
	hookStaleThreshold  = 8
	hookDroughtThreshold = 12
	hookBurstThreshold  = 3
	hookDecayThreshold  = 0.7
)

var hookHalfLifeMap = map[string]float64{"immediate": 10, "mid-arc": 30, "slow-burn": 80}

type HookAlert struct {
	Level   string   `json:"level"`
	Message string   `json:"message"`
	HookIDs []string `json:"hookIds,omitempty"`
}

type HookHealthReport struct {
	CurrentChapter int                     `json:"currentChapter"`
	ActiveCount    int                     `json:"activeCount"`
	StaleCount     int                     `json:"staleCount"`
	ResolvedCount  int                     `json:"resolvedCount"`
	UnknownStart   int                     `json:"unknownStartCount"`
	ActiveHooks    []settledForeshadowing  `json:"activeHooks"`
	StaleHooks     []settledForeshadowing  `json:"staleHooks"`
	Alerts         []HookAlert             `json:"alerts"`
}

func hookHalfLife(hook settledForeshadowing) float64 {
	if half, ok := hookHalfLifeMap[strings.TrimSpace(hook.PayoffTiming)]; ok {
		return half
	}
	return hookHalfLifeMap["mid-arc"]
}

// hookAge 返回距最近一次推进（或埋下）的章数。
// 历史快照缺章号时返回 -1，表示来源未知，不能参与 stale/decay 判定。
func hookAge(hook settledForeshadowing, current int) int {
	base := hook.LastAdvancedChapter
	if base <= 0 {
		base = hook.StartChapter
	}
	if base <= 0 {
		return -1
	}
	age := current - base
	if age < 0 {
		return 0
	}
	return age
}

// hookDecay 与洛水一致：1 - exp(-0.693 * age / halfLife)。
func hookDecay(hook settledForeshadowing, current int) float64 {
	age := hookAge(hook, current)
	if age <= 0 {
		return 0
	}
	return 1 - math.Exp(-0.693*float64(age)/hookHalfLife(hook))
}

// analyzeHookHealth 评估活跃/沉寂/旱灾/爆发/衰减状况。
func analyzeHookHealth(hooks []settledForeshadowing, currentChapter int) HookHealthReport {
	report := HookHealthReport{CurrentChapter: currentChapter, ActiveHooks: []settledForeshadowing{}, StaleHooks: []settledForeshadowing{}, Alerts: []HookAlert{}}
	sorted := append([]settledForeshadowing(nil), hooks...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].StartChapter == sorted[j].StartChapter {
			return sorted[i].Thread < sorted[j].Thread
		}
		return sorted[i].StartChapter < sorted[j].StartChapter
	})
	lastAdvanced := -1
	recentPlanted := 0
	recentResolved := 0
	for _, hook := range sorted {
		switch strings.TrimSpace(hook.Status) {
		case "resolved":
			report.ResolvedCount++
			if currentChapter > 0 && hook.LastAdvancedChapter > 0 && currentChapter-hook.LastAdvancedChapter <= 2 {
				recentResolved++
			}
			continue
		case "open", "advancing":
			report.ActiveCount++
			report.ActiveHooks = append(report.ActiveHooks, hook)
		default:
			continue
		}
		if hook.StartChapter <= 0 {
			report.UnknownStart++
		}
		if hook.LastAdvancedChapter > lastAdvanced {
			lastAdvanced = hook.LastAdvancedChapter
		}
		age := hookAge(hook, currentChapter)
		if age >= hookStaleThreshold {
			report.StaleHooks = append(report.StaleHooks, hook)
		}
		if currentChapter > 0 && hook.StartChapter > 0 && currentChapter-hook.StartChapter <= 2 {
			recentPlanted++
		}
		if decay := hookDecay(hook, currentChapter); decay >= hookDecayThreshold {
			report.Alerts = append(report.Alerts, HookAlert{Level: "info", Message: fmt.Sprintf("伏笔「%s」衰减度 %.0f%%，即将过期", hook.Thread, decay*100), HookIDs: []string{hook.Thread}})
		}
	}
	report.StaleCount = len(report.StaleHooks)
	if report.ActiveCount > hookMaxActive {
		report.Alerts = append(report.Alerts, HookAlert{Level: "warning", Message: fmt.Sprintf("活跃伏笔数 %d 超过上限 %d，读者可能难以追踪", report.ActiveCount, hookMaxActive)})
	}
	if report.StaleCount > 0 {
		ids := make([]string, 0, report.StaleCount)
		for _, hook := range report.StaleHooks {
			ids = append(ids, hook.Thread)
		}
		report.Alerts = append(report.Alerts, HookAlert{Level: "warning", Message: fmt.Sprintf("%d 个伏笔已沉寂 %d+ 章未推进", report.StaleCount, hookStaleThreshold), HookIDs: ids})
	}
	if currentChapter > 0 && lastAdvanced > 0 && currentChapter-lastAdvanced >= hookDroughtThreshold {
		report.Alerts = append(report.Alerts, HookAlert{Level: "error", Message: fmt.Sprintf("已 %d 章未推进任何伏笔（旱灾阈值 %d）", currentChapter-lastAdvanced, hookDroughtThreshold)})
	}
	if recentPlanted >= hookBurstThreshold && recentResolved == 0 {
		report.Alerts = append(report.Alerts, HookAlert{Level: "warning", Message: fmt.Sprintf("近 2 章新开 %d 个伏笔但未解决任何（爆发警告）", recentPlanted)})
	}
	return report
}

// formatHookHealthForPlanner 输出给章节计划的紧凑摘要。
func formatHookHealthForPlanner(report HookHealthReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## 伏笔健康（活跃 %d，沉寂 %d）", report.ActiveCount, report.StaleCount)
	for _, alert := range report.Alerts {
		mark := "i"
		switch alert.Level {
		case "error":
			mark = "!!!"
		case "warning":
			mark = "!!"
		}
		fmt.Fprintf(&b, "\n[%s] %s", mark, alert.Message)
	}
	if len(report.StaleHooks) > 0 {
		b.WriteString("\n需要处理的沉寂伏笔：")
		for i, hook := range report.StaleHooks {
			if i >= 5 {
				break
			}
			fmt.Fprintf(&b, "\n  - %s（第 %d 章埋下）", hook.Thread, hook.StartChapter)
		}
	}
	return b.String()
}

// hooksFromTruth 从已保存的结算快照中读取伏笔列表。
func hooksFromTruth(raw []byte) []settledForeshadowing {
	if len(raw) == 0 {
		return nil
	}
	var truth truthSettlementOutput
	if err := json.Unmarshal(raw, &truth); err != nil {
		return nil
	}
	return truth.Foreshadowing
}
