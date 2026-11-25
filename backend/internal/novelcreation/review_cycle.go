package novelcreation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// 洛水 novels.pipeline.chapter_review_cycle.run_chapter_review_cycle 的等价实现。
// 证据：docs/reverse-engineering/luoshui-novel-native-20261119/constants/novels--pipeline--chapter_review_cycle.json
// 已闭合：长度范围 0.55–1.6、退出路径文案、ReviewSnapshot 字段、rollback 语义。
// 未闭合：NET_IMPROVEMENT_EPSILON 的具体数值，因此采用保守阈值并在报告中标注来源。
const (
	reviewLengthLowRatio        = 0.55
	reviewLengthHighRatio       = 1.6
	reviewNetImprovementEpsilon = 0.5
	reviewEpsilonSource         = "unverified_original_threshold_uses_conservative_0.5"
)

type reviewSnapshot struct {
	Round         int     `json:"round"`
	Digest        string  `json:"digest"`
	WordCount     int     `json:"wordCount"`
	Score         float64 `json:"score"`
	Passed        bool    `json:"passed"`
	LengthInRange bool    `json:"lengthInRange"`
}

type reviewCycleReport struct {
	Iterations            int              `json:"iterations"`
	Revised               bool             `json:"revised"`
	RolledBack            bool             `json:"rolledBack"`
	ExitReason            string           `json:"exitReason"`
	FinalScore            float64          `json:"finalScore"`
	FinalWordCount        int              `json:"finalWordCount"`
	FinalLengthInRange    bool             `json:"finalLengthInRange"`
	TargetWordCount       int              `json:"targetWordCount"`
	LengthLowRatio        float64          `json:"lengthLowRatio"`
	LengthHighRatio       float64          `json:"lengthHighRatio"`
	NetImprovementEpsilon float64          `json:"netImprovementEpsilon"`
	EpsilonSource         string           `json:"epsilonSource"`
	Snapshots             []reviewSnapshot `json:"snapshots"`
}

func reviewLengthInRange(wordCount, target int) bool {
	if target <= 0 {
		return true
	}
	return float64(wordCount) >= float64(target)*reviewLengthLowRatio &&
		float64(wordCount) <= float64(target)*reviewLengthHighRatio
}

func newReviewSnapshot(round int, content string, audit chapterAuditOutput, threshold float64, target int) reviewSnapshot {
	wordCount := countVisibleRunes(content)
	return reviewSnapshot{
		Round:         round,
		Digest:        fingerprint(content),
		WordCount:     wordCount,
		Score:         scoreOf(audit),
		Passed:        auditPasses(audit, threshold),
		LengthInRange: reviewLengthInRange(wordCount, target),
	}
}

// reviewRank 把"是否通过"和"长度是否在范围内"合成优先顺序。
// 原版在比较 best 时同时参考 length_in_range，因此长度更合适但分数相同的候选会胜出。
func reviewRank(snapshot reviewSnapshot) int {
	rank := 0
	if snapshot.Passed {
		rank += 2
	}
	if snapshot.LengthInRange {
		rank++
	}
	return rank
}

// betterReviewSnapshot 判断候选是否优于当前最佳：
// 先比通过/长度档位，再比分数；净提升不足 0.5 分视为未提升。
func betterReviewSnapshot(candidate, best reviewSnapshot) bool {
	if candidateRank, bestRank := reviewRank(candidate), reviewRank(best); candidateRank != bestRank {
		return candidateRank > bestRank
	}
	improvement := candidate.Score - best.Score
	if improvement < 0 {
		return false
	}
	if improvement < reviewNetImprovementEpsilon && !(candidate.LengthInRange && !best.LengthInRange) {
		return false
	}
	return candidate.Score > best.Score || (candidate.Score == best.Score && candidate.LengthInRange && !best.LengthInRange)
}

func newReviewCycleReport(content string, audit chapterAuditOutput, threshold float64, target int) *reviewCycleReport {
	report := &reviewCycleReport{
		TargetWordCount:       target,
		LengthLowRatio:        reviewLengthLowRatio,
		LengthHighRatio:       reviewLengthHighRatio,
		NetImprovementEpsilon: reviewNetImprovementEpsilon,
		EpsilonSource:         reviewEpsilonSource,
		Snapshots:             []reviewSnapshot{newReviewSnapshot(0, content, audit, threshold, target)},
	}
	return report
}

func (r *reviewCycleReport) best() reviewSnapshot {
	best := r.Snapshots[0]
	for _, snapshot := range r.Snapshots[1:] {
		if betterReviewSnapshot(snapshot, best) {
			best = snapshot
		}
	}
	return best
}

// finalize 记录最终采纳版本、是否回退以及退出理由。
func (r *reviewCycleReport) finalize(content string, audit chapterAuditOutput, threshold float64, target int, initialDigest string) {
	final := newReviewSnapshot(r.Iterations, content, audit, threshold, target)
	r.FinalScore = final.Score
	r.FinalWordCount = final.WordCount
	r.FinalLengthInRange = final.LengthInRange
	r.Revised = final.Digest != initialDigest
	// 回退判定：最后一轮候选不是最终采纳版本时，说明撤销了该轮结果。
	if len(r.Snapshots) > 1 {
		last := r.Snapshots[len(r.Snapshots)-1]
		r.RolledBack = last.Digest != final.Digest
	}
	if r.ExitReason == "" {
		if final.Passed {
			r.ExitReason = "passed"
		} else {
			r.ExitReason = "max_iterations"
		}
	}
}

// reviewLoopRequest 把内部循环所需的依赖显式列出，避免修改旧循环路径。
type reviewLoopRequest struct {
	Input      RunContext
	Sink       EventSink
	Foundation foundationPlanOutput
	Plan       chapterPlanOutput
	Initial    string
	InitialRaw json.RawMessage
	Threshold  float64
	MaxRounds  int
	Assess     func(ctx context.Context, content string, round int) (chapterAuditOutput, json.RawMessage, error)
	Revise     func(ctx context.Context, content string, audit chapterAuditOutput, round int) (chapterRevisionOutput, json.RawMessage, error)
}

type reviewLoopResult struct {
	Content    string
	ContentRaw json.RawMessage
	Audit      chapterAuditOutput
	AuditRaw   json.RawMessage
	Audits     []Audit
	Revisions  []Revision
	Report     *reviewCycleReport
}

// runStructuredReviewLoop 按原版顺序执行审稿循环：
// 通过 → 退出；无新正文 → 退出；未净提升 → 退出；轮数用尽 → 退出。
func (p *Pipeline) runStructuredReviewLoop(ctx context.Context, req reviewLoopRequest) (reviewLoopResult, error) {
	result := reviewLoopResult{
		Content:    req.Initial,
		ContentRaw: req.InitialRaw,
		Audit:      chapterAuditOutput{Issues: []auditIssue{}},
		AuditRaw:   req.InitialRaw,
		Audits:     []Audit{},
		Revisions:  []Revision{},
	}
	if req.Input.ReviewLoopVersion < 1 {
		return result, nil
	}
	initialAudit, initialRaw, err := req.Assess(ctx, req.Initial, 0)
	if err != nil {
		return result, err
	}
	result.Audit, result.AuditRaw = initialAudit, initialRaw
	result.Audits = append(result.Audits, newAudit(req.Input, 1, initialAudit, initialRaw))
	report := newReviewCycleReport(req.Initial, initialAudit, req.Threshold, req.Input.Project.TargetChapterWords)
	result.Report = report

	if strings.TrimSpace(req.Initial) == "" {
		report.ExitReason = "empty_content"
		report.finalize(req.Initial, initialAudit, req.Threshold, req.Input.Project.TargetChapterWords, fingerprint(req.Initial))
		return result, nil
	}
	if auditPasses(initialAudit, req.Threshold) {
		report.ExitReason = "passed"
		report.finalize(req.Initial, initialAudit, req.Threshold, req.Input.Project.TargetChapterWords, fingerprint(req.Initial))
		return result, nil
	}

	bestContent, bestContentRaw, bestAudit, bestAuditRaw := req.Initial, req.InitialRaw, initialAudit, initialRaw
	emitted := false
	for round := 1; round <= req.MaxRounds; round++ {
		if err := p.checkCancelled(ctx, req.Input); err != nil {
			report.ExitReason = "cancelled"
			break
		}
		if len(bestAudit.Issues) == 0 {
			report.ExitReason = "no_actionable_issues"
			break
		}
		if !emitted {
			if err := p.emitStage(ctx, req.Input, req.Sink, RunRevising, "正在根据审核修订章节", 65); err != nil {
				return result, err
			}
			emitted = true
		}
		report.Iterations = round
		revision, revisionRaw, callErr := req.Revise(ctx, bestContent, bestAudit, round)
		if callErr != nil {
			return result, callErr
		}
		result.Revisions = append(result.Revisions, newRevision(req.Input, round, revision, revisionRaw))
		if strings.TrimSpace(revision.Content) == "" || fingerprint(revision.Content) == fingerprint(bestContent) {
			report.ExitReason = "empty_revision"
			break
		}
		candidateAudit, candidateRaw, auditErr := req.Assess(ctx, revision.Content, round)
		if auditErr != nil {
			return result, auditErr
		}
		result.Audits = append(result.Audits, newAudit(req.Input, len(result.Audits)+1, candidateAudit, candidateRaw))
		snapshot := newReviewSnapshot(round, revision.Content, candidateAudit, req.Threshold, req.Input.Project.TargetChapterWords)
		report.Snapshots = append(report.Snapshots, snapshot)
		if !betterReviewSnapshot(snapshot, newReviewSnapshot(0, bestContent, bestAudit, req.Threshold, req.Input.Project.TargetChapterWords)) {
			report.ExitReason = "no_net_improvement"
			break
		}
		bestContent, bestContentRaw, bestAudit, bestAuditRaw = strings.TrimSpace(revision.Content), revisionRaw, candidateAudit, candidateRaw
		if snapshot.Passed {
			report.ExitReason = "passed"
			break
		}
	}
	if report.ExitReason == "" {
		report.ExitReason = "max_iterations"
	}
	report.finalize(bestContent, bestAudit, req.Threshold, req.Input.Project.TargetChapterWords, fingerprint(req.Initial))
	result.Content, result.ContentRaw, result.Audit, result.AuditRaw = bestContent, bestContentRaw, bestAudit, bestAuditRaw
	if report.RolledBack {
		if err := p.emit(ctx, req.Input, req.Sink, RunEvent{Type: "run.review_rollback", Step: string(RunRevising), Status: RunRevising,
			Message: fmt.Sprintf("回退到最高分版本（%.0f 分，当前 %.0f 分）", report.FinalScore, scoreOf(bestAudit)), Progress: 70}); err != nil {
			return result, err
		}
	}
	return result, nil
}
