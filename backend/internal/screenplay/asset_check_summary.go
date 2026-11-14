package screenplay

import "time"

type AssetCheckEpisodeProgress struct {
	ChapterIndex   int    `json:"chapter_index"`
	EpisodeNo      int    `json:"episode_no"`
	Title          string `json:"title"`
	Status         string `json:"status"`
	DisplayMessage string `json:"display_message"`
	Error          string `json:"error,omitempty"`
}
type AssetCheckSummary struct {
	ID              string                      `json:"id"`
	RequestKey      string                      `json:"request_key"`
	Status          string                      `json:"status"`
	Revision        int64                       `json:"revision"`
	CancelRequested bool                        `json:"cancel_requested"`
	Applied         bool                        `json:"applied"`
	Discarded       bool                        `json:"discarded"`
	AddedCount      int                         `json:"added_count"`
	SkippedCount    int                         `json:"skipped_count"`
	SourceMode      string                      `json:"source_mode"`
	Model           string                      `json:"model"`
	CheckedCount    int                         `json:"checked_count"`
	Total           int                         `json:"total"`
	CandidateCounts map[AssetType]int           `json:"candidate_counts"`
	Candidates      []AssetCheckCandidate       `json:"candidates"`
	Episodes        []AssetCheckEpisodeProgress `json:"episodes"`
	Error           string                      `json:"error,omitempty"`
	CreatedAt       time.Time                   `json:"created_at"`
	UpdatedAt       time.Time                   `json:"updated_at"`
}

func (j AssetCheckJob) Summary() AssetCheckSummary {
	out := AssetCheckSummary{ID: j.ID, RequestKey: j.RequestKey, Status: j.Status, Revision: j.Revision, CancelRequested: j.CancelRequested, Applied: j.Applied, Discarded: j.Discarded, AddedCount: j.AddedCount, SkippedCount: j.SkippedCount, SourceMode: j.SourceMode, Model: j.Model, Total: len(j.Episodes), CandidateCounts: map[AssetType]int{AssetCharacter: 0, AssetScene: 0, AssetProp: 0}, Candidates: j.Candidates, Error: j.Error, CreatedAt: j.CreatedAt, UpdatedAt: j.UpdatedAt, Episodes: []AssetCheckEpisodeProgress{}}
	if out.Candidates == nil {
		out.Candidates = []AssetCheckCandidate{}
	}
	for _, c := range j.Candidates {
		out.CandidateCounts[c.Asset.Type]++
	}
	for _, ep := range j.Episodes {
		p := AssetCheckEpisodeProgress{ChapterIndex: ep.ChapterIndex, EpisodeNo: ep.EpisodeNo, Title: ep.Title, Status: "pending"}
		total, finished, skipped := 0, 0, 0
		for _, step := range j.Steps {
			if step.ChapterIndex != ep.ChapterIndex {
				continue
			}
			total++
			switch step.Status {
			case "pending":
			case "calling", "received":
				if p.Status == "pending" {
					p.Status = "running"
				}
			case "unknown":
				p.Status = "unknown"
				finished++
			case "failed":
				if p.Status != "unknown" {
					p.Status = "failed"
				}
				p.Error = step.Error
				finished++
			case "cancelled":
				skipped++
				if p.Status == "pending" {
					p.Status = "cancelled"
				}
				finished++
			case "completed":
				finished++
			}
		}
		if total > 0 && finished == total {
			if skipped == 0 {
				out.CheckedCount++
			}
			if p.Status == "pending" || p.Status == "running" {
				p.Status = "completed"
			}
		}
		if j.Status == "cancelled" && p.Status == "pending" {
			p.Status = "cancelled"
		}
		p.DisplayMessage = map[string]string{"pending": "等待检查", "running": "检查中", "completed": "检查完成", "failed": "检查失败", "cancelled": "已跳过", "unknown": "请求结果待核对"}[p.Status]
		out.Episodes = append(out.Episodes, p)
	}
	return out
}
