package screenplay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"qiantie/backend/internal/mergeworker"
	"strconv"
	"strings"
	"time"
)

var ErrPlanningProgressUnavailable = errors.New("暂时无法读取规划进度")

// PlanningProgress is a measured round snapshot. Chars describes this attempt,
// not the previously saved plan. It never replaces the MySQL chapter result.
type PlanningProgress struct {
	Type         string   `json:"type"`
	RequestKey   string   `json:"request_key"`
	ChapterID    string   `json:"chapter_id"`
	ChapterIndex int      `json:"chapter_index"`
	Sequence     int      `json:"sequence"`
	Round        int      `json:"round"`
	Total        int      `json:"total"`
	Chars        int      `json:"chars"`
	Complete     bool     `json:"complete"`
	Missing      []string `json:"missing"`
	Phase        string   `json:"phase"`
	Status       string   `json:"status"`
	Message      string   `json:"message"`
}
type PlanningProgressRepository interface {
	Put(context.Context, string, string, PlanningProgress) error
	Get(context.Context, string, string, string) (PlanningProgress, error)
}
type RedisPlanningProgressRepository struct{ Do mergeworker.RedisDoFunc }

func PlanningProgressRedisKey(owner, project, request string) string {
	return "qiantie:screenplay:planning-progress:" + contentDigest(strings.Join([]string{owner, project, request}, "\x00"))
}

const planningProgressWrite = `local old=redis.call('GET',KEYS[1]); if old then local ok,v=pcall(cjson.decode,old); if ok and tonumber(v.sequence or 0)>=tonumber(ARGV[2]) then return 0 end end; redis.call('SET',KEYS[1],ARGV[1],'EX',ARGV[3]); return 1`

func (s *RedisPlanningProgressRepository) Put(ctx context.Context, owner, project string, p PlanningProgress) error {
	if s == nil || s.Do == nil {
		return ErrPlanningProgressUnavailable
	}
	raw, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_, err = s.Do(ctx, "EVAL", planningProgressWrite, "1", PlanningProgressRedisKey(owner, project, p.RequestKey), string(raw), strconv.Itoa(p.Sequence), "86400")
	return err
}
func (s *RedisPlanningProgressRepository) Get(ctx context.Context, owner, project, request string) (PlanningProgress, error) {
	var out PlanningProgress
	if s == nil || s.Do == nil {
		return out, ErrPlanningProgressUnavailable
	}
	value, err := s.Do(ctx, "GET", PlanningProgressRedisKey(owner, project, request))
	if err != nil {
		return out, err
	}
	if value == nil {
		return out, ErrNotFound
	}
	var raw []byte
	switch v := value.(type) {
	case string:
		raw = []byte(v)
	case []byte:
		raw = v
	default:
		return out, ErrPlanningProgressUnavailable
	}
	if json.Unmarshal(raw, &out) != nil || out.RequestKey != request || out.Type != "plan" {
		return PlanningProgress{}, ErrPlanningProgressUnavailable
	}
	return out, nil
}

func planningProgressFrame(round int, text, density, phase string) PlanningProgress {
	coverage := EvaluateStoryboardPlan(text, density)
	out := PlanningProgress{Type: "plan", Round: round, Total: 8, Chars: coverage.Chars, Complete: coverage.OK, Missing: coverage.Missing, Phase: phase, Status: "running"}
	if phase == "requesting" {
		out.Message = fmt.Sprintf("正在生成分镜规划…第 %d 轮", round)
		if round > 1 {
			out.Message = fmt.Sprintf("规划未写完，自动续写第 %d 轮…（已 %d 字）", round, out.Chars)
		}
	} else {
		label := "规划"
		if round > 1 {
			label = "规划续写"
		}
		out.Message = fmt.Sprintf("%s第 %d 轮 · %d 字 · %s", label, round, out.Chars, coverage.Reason)
	}
	return out
}

func (s *AuthoringService) planningProgressObserver(owner, project, chapter, key string, index int) (func(PlanningProgress), func(string)) {
	var latest PlanningProgress
	sequence := 0
	write := func(p PlanningProgress) {
		sequence++
		p.Sequence = sequence
		p.ChapterID = chapter
		p.ChapterIndex = index
		p.RequestKey = key
		latest = p
		if s.PlanningProgress != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = s.PlanningProgress.Put(ctx, owner, project, p)
		}
	}
	finish := func(status string) {
		if latest.Type == "" {
			return
		}
		p := latest
		p.Phase = "finished"
		p.Status = status
		switch status {
		case "succeeded":
			p.Message = "规划已完整并保存"
		case "partial":
			p.Message = "规划草稿已保存，请核对完整性"
		case "cancelled":
			p.Message = "规划已取消，之前保存的规划保留"
		case "failed":
			p.Message = "规划生成失败，之前保存的规划保留"
		}
		write(p)
	}
	return write, finish
}
